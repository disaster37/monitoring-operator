package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"reflect"
	"strings"

	"github.com/Masterminds/sprig/v3"
	"github.com/davecgh/go-spew/spew"
	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/disaster37/operator-sdk-extra/pkg/helper"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"
)

// CentreonController permit to handle all Centreon resource templating
type CentreonController struct {
	client.Client
	Scheme *runtime.Scheme
	log    *logrus.Entry
}

type CompareResource struct {
	Current  client.Object
	Expected client.Object
	Diff     *controller.Diff
}

// SetLogger permit to set logger on Centreon controller
func (r *CentreonController) SetLogger(log *logrus.Entry) {
	r.log = log
}

// Return true if object type is TemplateCentreService
func isTemplateCentreonService(o client.Object) bool {
	if reflect.TypeOf(o).Elem().Name() == "TemplateCentreonService" {
		return true
	}
	return false
}

// watchCentreonTemplate permit to search CentreonService created from TemplateCentreonService to reconcil parents of them
func watchCentreonTemplate(c client.Client) handler.MapFunc {
	return func(a client.Object) []reconcile.Request {

		reconcileRequests := make([]reconcile.Request, 0, 0)
		listCentreonService := &v1alpha1.CentreonServiceList{}

		selectors, err := labels.Parse(fmt.Sprintf("%s/template-name=%s,%s/template-namespace=%s", monitoringAnnotationKey, a.GetName(), monitoringAnnotationKey, a.GetNamespace()))
		if err != nil {
			panic(err)
		}

		// Get all CentreonService created from this template
		if err := c.List(context.Background(), listCentreonService, &client.ListOptions{LabelSelector: selectors}); err != nil {
			panic(err)
		}

		for _, cs := range listCentreonService.Items {
			logrus.Debugf("Found CentreonService %s/%s", cs.Namespace, cs.Name)
			// Search parent to reconcile parent
			for _, parent := range cs.OwnerReferences {
				logrus.Debugf("Reconcile %s/%s", cs.Namespace, parent.Name)
				reconcileRequests = append(reconcileRequests, reconcile.Request{NamespacedName: types.NamespacedName{Name: parent.Name, Namespace: cs.Namespace}})
			}
		}

		return reconcileRequests
	}
}

// readTemplatingCentreonService get all templates from annotations (name and namespace)
// Then, retrive all current service already created from this templates
// Finnaly, compute all expected service
func (r *CentreonController) readTemplatingCentreonService(ctx context.Context, resource client.Object, data map[string]any, meta any, placeholders map[string]any) (res ctrl.Result, err error) {

	// Get Templates for CentreonService
	targetTemplates := resource.GetAnnotations()[fmt.Sprintf("%s/templates", monitoringAnnotationKey)]
	r.log.Debugf("Raw target template: %s", targetTemplates)
	listNamespacedName := make([]*types.NamespacedName, 0, 0)
	if targetTemplates != "" {
		if err = json.Unmarshal([]byte(targetTemplates), &listNamespacedName); err != nil {
			return res, errors.Wrap(err, "Error when unmarshall the list of CentreonService template")
		}
	}
	// No template
	if len(listNamespacedName) == 0 {
		return res, nil
	}

	// compute namespace, because of when resource is namespace, the namespace field not exist
	var namespace string
	switch resource.GetObjectKind().GroupVersionKind().Kind {
	case "Namespace":
		namespace = resource.GetName()
		break
	case "Node":
		ns, err := helpers.GetOperatorNamespace()
		if err != nil {
			return res, err
		}
		namespace = ns
		break
	default:
		namespace = resource.GetNamespace()
	}

	// Get currents resources and compute expected resources
	listCompareResource := make([]*CompareResource, 0, len(listNamespacedName))
	for _, namespacedName := range listNamespacedName {
		r.log.Debugf("Process CentreonServiceTemplate %s/%s", namespacedName.Namespace, namespacedName.Name)
		compareResource := &CompareResource{}

		// Get current resource
		currentCS := &v1alpha1.CentreonService{}
		err = r.Get(ctx, types.NamespacedName{Name: namespacedName.Name, Namespace: namespace}, currentCS)
		if err != nil && k8serrors.IsNotFound(err) {
			compareResource.Current = nil
		} else if err != nil {
			return res, errors.Wrapf(err, "Error when get CentreonService %s/%s", namespace, namespacedName.Name)
		} else {
			compareResource.Current = currentCS
		}

		// Compute expected resource
		templateCS := &v1alpha1.TemplateCentreonService{}
		if err = r.Get(ctx, *namespacedName, templateCS); err != nil {
			return res, errors.Wrapf(err, "Error when get CentreonServiceTemplate %s/%s", namespacedName.Namespace, namespacedName.Name)
		}
		t, err := template.New("templateCentreonService").Funcs(sprig.FuncMap()).Parse(templateCS.Spec.Template)
		if err != nil {
			return res, errors.Wrapf(err, "Error when parse templateCentreonService %s/%s", namespacedName.Namespace, namespacedName.Name)
		}
		buf := bytes.NewBufferString("")
		if err = t.Execute(buf, placeholders); err != nil {
			return res, errors.Wrap(err, "Error when execute template")
		}
		r.log.Debugf("Raw CentreonService spec after template it:\n %s", buf.String())
		expectedCSSpec := &v1alpha1.CentreonServiceSpec{}

		if err = yaml.Unmarshal(buf.Bytes(), expectedCSSpec); err != nil {
			return res, errors.Wrap(err, "Error when unmarshall expected Centreon service spec")
		}

		expectedCS := &v1alpha1.CentreonService{
			ObjectMeta: metav1.ObjectMeta{
				Name:        namespacedName.Name,
				Namespace:   namespace,
				Labels:      resource.GetLabels(),
				Annotations: resource.GetAnnotations(),
			},
			Spec: *expectedCSSpec,
		}
		expectedCS.Labels[fmt.Sprintf("%s/template-name", monitoringAnnotationKey)] = namespacedName.Name
		expectedCS.Labels[fmt.Sprintf("%s/template-namespace", monitoringAnnotationKey)] = namespacedName.Namespace
		// Check CentreonService is valide
		if !expectedCS.IsValid() {
			return res, fmt.Errorf("Generated CentreonService is not valid: %+v", expectedCS.Spec)
		}
		r.log.Debugf("Exepcted CentreonService from template %s/%s:\n%s", namespacedName.Namespace, namespacedName.Name, spew.Sdump(expectedCS))
		// Set namespace instance as the owner
		ctrl.SetControllerReference(resource, expectedCS, r.Scheme)
		compareResource.Expected = expectedCS

		listCompareResource = append(listCompareResource, compareResource)

	}

	data["compareResources"] = listCompareResource

	return res, nil
}

// createOrUpdateCentreonServiceFromTemplate permit to create or update CentreonService computing from templates
func (r *CentreonController) createOrUpdateCentreonServiceFromTemplate(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	var d any

	d, err = helper.Get(data, "compareResources")
	if err != nil {
		return res, err
	}
	listCompareResource := d.([]*CompareResource)

	for _, compareResource := range listCompareResource {
		if compareResource.Diff.NeedCreate {
			r.log.Debugf("Create CentreonService %s/%s", compareResource.Expected.GetNamespace(), compareResource.Expected.GetName())
			if err = r.Client.Create(ctx, compareResource.Expected); err != nil {
				return res, errors.Wrapf(err, "Error when create CentreonService %s", compareResource.Expected.GetName())
			}
		} else if compareResource.Diff.NeedUpdate {
			r.log.Debugf("Update CentreonService %s/%s", compareResource.Current.GetNamespace(), compareResource.Current.GetName())
			if err = r.Client.Update(ctx, compareResource.Current); err != nil {
				return res, errors.Wrapf(err, "Error when update CentreonService %s", compareResource.Current.GetName())
			}
		}
	}

	return res, nil
}

// diffCentreonService permit to diff CentreonService generated by templating
func (r *CentreonController) diffCentreonService(resource client.Object, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	var d any

	d, err = helper.Get(data, "compareResources")
	if err != nil {
		return diff, err
	}
	listCompareResource := d.([]*CompareResource)
	diff = controller.Diff{
		NeedCreate: false,
		NeedUpdate: false,
	}
	var sb strings.Builder

	for _, compareResource := range listCompareResource {
		localDiff := &controller.Diff{
			NeedCreate: false,
			NeedUpdate: false,
		}
		compareResource.Diff = localDiff

		// New CentreonService
		if compareResource.Current == nil {
			localDiff.NeedCreate = true
			localDiff.Diff = fmt.Sprintf("CentreonService %s not exist", compareResource.Expected.GetName())
			diff.NeedCreate = true
			sb.WriteString(localDiff.Diff)
			continue
		}

		// EDxisting CentreonService
		diffSpec := cmp.Diff(compareResource.Current.(*v1alpha1.CentreonService).Spec, compareResource.Expected.(*v1alpha1.CentreonService).Spec)
		diffLabels := cmp.Diff(compareResource.Current.GetLabels(), compareResource.Expected.GetLabels())
		diffAnnotations := cmp.Diff(compareResource.Current.GetAnnotations(), compareResource.Expected.GetAnnotations())
		if diffSpec != "" || diffLabels != "" || diffAnnotations != "" {
			localDiff.NeedUpdate = true
			localDiff.Diff = fmt.Sprintf("%s\n%s\n%s", diffLabels, diffAnnotations, diffSpec)
			compareResource.Current.SetLabels(compareResource.Expected.GetLabels())
			compareResource.Current.SetAnnotations(compareResource.Expected.GetAnnotations())
			compareResource.Current.(*v1alpha1.CentreonService).Spec = compareResource.Expected.(*v1alpha1.CentreonService).Spec
			diff.NeedUpdate = true
			sb.WriteString(localDiff.Diff)
		}
	}

	diff.Diff = sb.String()

	return diff, nil

}
