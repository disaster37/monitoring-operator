package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"reflect"

	"github.com/Masterminds/sprig/v3"
	"github.com/davecgh/go-spew/spew"
	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
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

type TemplateController struct {
	client.Client
	Scheme *runtime.Scheme
	log    *logrus.Entry
}

// Return true if object type is Template
func isTemplate(o client.Object) bool {
	if reflect.TypeOf(o).Elem().Name() == "Template" {
		return true
	}
	return false
}

// watchTemplate permit to search resource created from Template to reconcil parents of them
func watchTemplate(c client.Client) handler.MapFunc {
	return func(a client.Object) []reconcile.Request {

		reconcileRequests := make([]reconcile.Request, 0, 0)
		template := a.(*v1alpha1.Template)
		selectors, err := labels.Parse(fmt.Sprintf("%s/template-name=%s,%s/template-namespace=%s", monitoringAnnotationKey, a.GetName(), monitoringAnnotationKey, a.GetNamespace()))
		if err != nil {
			panic(err)
		}

		// Get object type
		switch template.Spec.Type {
		case "CentreonService":
			// Get all resources created from this template
			listCentreonService := &v1alpha1.CentreonServiceList{}
			if err := c.List(context.Background(), listCentreonService, &client.ListOptions{LabelSelector: selectors}); err != nil {
				panic(err)
			}

			for _, cs := range listCentreonService.Items {
				// Search parent to reconcile parent
				for _, parent := range cs.OwnerReferences {
					reconcileRequests = append(reconcileRequests, reconcile.Request{NamespacedName: types.NamespacedName{Name: parent.Name, Namespace: cs.Namespace}})
				}
			}
			break
		}

		return reconcileRequests
	}
}

// SetLogger permit to set logger on Centreon controller
func (r *TemplateController) SetLogger(log *logrus.Entry) {
	r.log = log
}

// readTemplating get all templates from annotations (name and namespace)
// Then, retrive all current resources already created from this templates
// Finnaly, compute all expected resources
func (r *TemplateController) readTemplating(ctx context.Context, resource client.Object, data map[string]any, meta any, placeholders map[string]any) (res ctrl.Result, err error) {

	var (
		currentResource  client.Object
		expectedResource client.Object
		objectType       string
	)

	// Get Templates from annotations
	listNamespacedName, err := readListTemplateFromAnnotation(resource)
	if err != nil {
		return res, err
	}
	// No template
	if len(listNamespacedName) == 0 {
		return res, nil
	}

	// compute namespace, because of when resource is namespace, the namespace field not exist
	namespace, err := getComputedNamespaceName(resource)
	if err != nil {
		return res, err
	}

	// Get currents resources and compute expected resources
	comparedResources := map[string][]CompareResource{}
	for _, namespacedName := range listNamespacedName {
		r.log.Debugf("Process template %s/%s", namespacedName.Namespace, namespacedName.Name)

		// Get template
		templateO := &v1alpha1.Template{}
		if err = r.Get(ctx, namespacedName, templateO); err != nil {
			return res, errors.Wrapf(err, "Error when get template %s/%s", namespacedName.Namespace, namespacedName.Name)
		}

		// Process template
		rawTemplate, err := processTemplate(templateO, placeholders)
		if err != nil {
			return res, err
		}

		// Generate the right object depend of template type
		switch templateO.Spec.Type {
		case "CentreonService":
			objectType = "centreonServiceCompareResources"
			if comparedResources[objectType] == nil {
				comparedResources[objectType] = make([]CompareResource, 0)
			}
			currentResource = &v1alpha1.CentreonService{}
			centreonServiceSpec := &v1alpha1.CentreonServiceSpec{}
			// Compute expected resource spec
			if err = yaml.Unmarshal(rawTemplate, centreonServiceSpec); err != nil {
				return res, errors.Wrap(err, "Error when unmarshall expected spec")
			}

			centreonService := &v1alpha1.CentreonService{
				ObjectMeta: metav1.ObjectMeta{
					Name:        namespacedName.Name,
					Namespace:   namespace,
					Labels:      helpers.CopyMapString(resource.GetLabels()),
					Annotations: helpers.CopyMapString(resource.GetAnnotations()),
				},
				Spec: *centreonServiceSpec,
			}

			// Check CentreonService is valid
			if !centreonService.IsValid() {
				return res, fmt.Errorf("Generated CentreonService is not valid: %+v", centreonService.Spec)
			}
			expectedResource = centreonService
			break
		default:
			return res, errors.Errorf("Template of type %s is not supported", templateO.Spec.Type)
		}

		// Get current resource
		err = r.Get(ctx, types.NamespacedName{Name: namespacedName.Name, Namespace: namespace}, currentResource)
		if err != nil && k8serrors.IsNotFound(err) {
			currentResource = reflect.New(reflect.TypeOf(currentResource)).Elem().Interface().(client.Object)
		} else if err != nil {
			return res, errors.Wrapf(err, "Error when get %s %s/%s", templateO.Spec.Type, namespace, namespacedName.Name)
		}

		// Add labels for labelSelectors
		setLabelsOnExpectedResource(expectedResource, namespacedName)

		// Set resource as the owner
		ctrl.SetControllerReference(resource, expectedResource, r.Scheme)

		compareResource := CompareResource{
			Current:  currentResource,
			Expected: expectedResource,
			Diff: &controller.Diff{
				NeedCreate: false,
				NeedUpdate: false,
			},
		}

		comparedResources[objectType] = append(comparedResources[objectType], compareResource)
	}

	r.log.Tracef("List of compare resources: %s", spew.Sdump(comparedResources))

	for key, value := range comparedResources {
		data[key] = value
	}

	return res, nil
}

// readListTemplateFromAnnotation return the list of template resource
func readListTemplateFromAnnotation(resource client.Object) (listNamespacedName []types.NamespacedName, err error) {
	targetTemplates := resource.GetAnnotations()[fmt.Sprintf("%s/templates", monitoringAnnotationKey)]
	listNamespacedName = make([]types.NamespacedName, 0, 0)
	if targetTemplates != "" {
		if err = json.Unmarshal([]byte(targetTemplates), &listNamespacedName); err != nil {
			return nil, errors.Wrap(err, "Error when unmarshall the list of template")
		}
	}

	return listNamespacedName, nil
}

// getComputedNamespaceName permit to get the right namespace when read template
func getComputedNamespaceName(resource client.Object) (namespace string, err error) {

	switch resource.GetObjectKind().GroupVersionKind().Kind {
	case "Namespace":
		namespace = resource.GetName()
		break
	case "Node":
		ns, err := helpers.GetOperatorNamespace()
		if err != nil {
			return namespace, err
		}
		namespace = ns
		break
	default:
		namespace = resource.GetNamespace()
	}

	return namespace, nil
}

// processTemplate generate the template
func processTemplate(templateO *v1alpha1.Template, placeholders map[string]any) (res []byte, err error) {

	t, err := template.New("template").Funcs(sprig.FuncMap()).Parse(templateO.Spec.Template)
	if err != nil {
		return nil, errors.Wrapf(err, "Error when parse template %s/%s", templateO.Namespace, templateO.Name)
	}
	buf := bytes.NewBufferString("")
	if err = t.Execute(buf, placeholders); err != nil {
		return nil, errors.Wrap(err, "Error when execute template")
	}

	return buf.Bytes(), nil
}

// setLabelsOnExpectedResource set the rigth labels to reconcil it when template change
func setLabelsOnExpectedResource(resource client.Object, namespacedName types.NamespacedName) {
	resource.GetLabels()[fmt.Sprintf("%s/template-name", monitoringAnnotationKey)] = namespacedName.Name
	resource.GetLabels()[fmt.Sprintf("%s/template-namespace", monitoringAnnotationKey)] = namespacedName.Namespace
}