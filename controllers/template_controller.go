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

type TemplateController struct {
	client.Client
	Scheme *runtime.Scheme
	log    *logrus.Entry
}

// Return true if object type is Template
func isTemplate(o client.Object) bool {
	return reflect.TypeOf(o).Elem().Name() == "Template"
}

// watchTemplate permit to search resource created from Template to reconcil parents of them
func watchTemplate(c client.Client) handler.MapFunc {
	return func(a client.Object) []reconcile.Request {

		var listRessources client.ObjectList

		reconcileRequests := make([]reconcile.Request, 0)
		template := a.(*v1alpha1.Template)
		selectors, err := labels.Parse(fmt.Sprintf("%s/template-name=%s,%s/template-namespace=%s", monitoringAnnotationKey, a.GetName(), monitoringAnnotationKey, a.GetNamespace()))
		if err != nil {
			panic(err)
		}

		// Get object type
		switch template.Spec.Type {
		case "CentreonService":
			listRessources = &v1alpha1.CentreonServiceList{}
		case "CentreonServiceGroup":
			listRessources = &v1alpha1.CentreonServiceGroupList{}
		default:
			return reconcileRequests
		}

		// Get all resources created from this template
		if err := c.List(context.Background(), listRessources, &client.ListOptions{LabelSelector: selectors}); err != nil {
			panic(err)
		}

		items, err := GetItems(listRessources)
		if err != nil {
			panic(err)
		}

		for _, item := range items {
			// Search parent to reconcile parent
			for _, parent := range item.GetOwnerReferences() {
				reconcileRequests = append(reconcileRequests, reconcile.Request{NamespacedName: types.NamespacedName{Name: parent.Name, Namespace: item.GetNamespace()}})
			}
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
	listcomparedResource := make([]CompareResource, 0, len(listNamespacedName))
	for _, namespacedName := range listNamespacedName {
		r.log.Debugf("Process template %s/%s", namespacedName.Namespace, namespacedName.Name)

		// Get template
		templateO := &v1alpha1.Template{}
		if err = r.Get(ctx, namespacedName, templateO); err != nil {
			return res, errors.Wrapf(err, "Error when get template %s/%s", namespacedName.Namespace, namespacedName.Name)
		}

		placeholders["templateName"] = templateO.Name

		// Process resource name
		targetResourceName, err := processName(templateO, placeholders)
		if err != nil {
			return res, err
		}

		// Process template
		rawTemplate, err := processTemplate(templateO, placeholders)
		if err != nil {
			return res, err
		}

		// Generate the right object depend of template type
		switch templateO.Spec.Type {
		case "CentreonService":
			currentResource = &v1alpha1.CentreonService{}
			centreonServiceSpec := &v1alpha1.CentreonServiceSpec{}
			// Compute expected resource spec
			if err = yaml.Unmarshal(rawTemplate, centreonServiceSpec); err != nil {
				return res, errors.Wrap(err, "Error when unmarshall expected spec")
			}

			centreonService := &v1alpha1.CentreonService{
				ObjectMeta: metav1.ObjectMeta{
					Name:        targetResourceName,
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
		case "CentreonServiceGroup":
			currentResource = &v1alpha1.CentreonServiceGroup{}
			centreonServiceGroupSpec := &v1alpha1.CentreonServiceGroupSpec{}
			// Compute expected resource spec
			if err = yaml.Unmarshal(rawTemplate, centreonServiceGroupSpec); err != nil {
				return res, errors.Wrap(err, "Error when unmarshall expected spec")
			}

			centreonServiceGroup := &v1alpha1.CentreonServiceGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:        targetResourceName,
					Namespace:   namespace,
					Labels:      helpers.CopyMapString(resource.GetLabels()),
					Annotations: helpers.CopyMapString(resource.GetAnnotations()),
				},
				Spec: *centreonServiceGroupSpec,
			}

			// Check CentreonServiceGroup is valid
			if !centreonServiceGroup.IsValid() {
				return res, fmt.Errorf("Generated CentreonServiceGroup is not valid: %+v", centreonServiceGroup.Spec)
			}
			expectedResource = centreonServiceGroup
		default:
			return res, errors.Errorf("Template of type %s is not supported", templateO.Spec.Type)
		}

		// Get current resource
		err = r.Get(ctx, types.NamespacedName{Name: targetResourceName, Namespace: namespace}, currentResource)
		if err != nil && k8serrors.IsNotFound(err) {
			currentResource = reflect.New(reflect.TypeOf(currentResource)).Elem().Interface().(client.Object)
		} else if err != nil {
			return res, errors.Wrapf(err, "Error when get %s %s/%s", templateO.Spec.Type, namespace, namespacedName.Name)
		}

		// Add labels for labelSelectors
		setLabelsOnExpectedResource(expectedResource, namespacedName)

		// Set resource as the owner
		err = ctrl.SetControllerReference(resource, expectedResource, r.Scheme)
		if err != nil {
			return res, errors.Wrapf(err, "Error when set as owner reference")
		}

		compareResource := CompareResource{
			Current:  currentResource,
			Expected: expectedResource,
			Diff: &controller.Diff{
				NeedCreate: false,
				NeedUpdate: false,
			},
		}

		listcomparedResource = append(listcomparedResource, compareResource)
	}

	r.log.Tracef("List of compare resources: %s", spew.Sdump(listcomparedResource))

	data["compareResources"] = listcomparedResource

	return res, nil
}

// diffRessourcesFromTemplate permit to diff ressources generated by templating
func (r *TemplateController) diffRessourcesFromTemplate(resource client.Object, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	var d any

	diff = controller.Diff{
		NeedCreate: false,
		NeedUpdate: false,
	}

	d, err = helper.Get(data, "compareResources")
	if err != nil {
		return diff, err
	}
	listCompareResource := d.([]CompareResource)
	if listCompareResource == nil {
		return diff, nil
	}

	var sb strings.Builder

	for _, compareResource := range listCompareResource {

		// New ressource
		if reflect.ValueOf(compareResource.Current).IsNil() {
			compareResource.Diff.NeedCreate = true
			compareResource.Diff.Diff = fmt.Sprintf("Ressource %s not exist", compareResource.Expected.GetName())
			diff.NeedCreate = true
			sb.WriteString(compareResource.Diff.Diff)
		} else {
			// Existing ressource
			currentSpec, err := GetSpec(compareResource.Current)
			if err != nil {
				return diff, err
			}
			expectedSpec, err := GetSpec(compareResource.Expected)
			if err != nil {
				return diff, err
			}
			diffSpec := cmp.Diff(currentSpec, expectedSpec)
			diffLabels := cmp.Diff(compareResource.Current.GetLabels(), compareResource.Expected.GetLabels())
			diffAnnotations := cmp.Diff(compareResource.Current.GetAnnotations(), compareResource.Expected.GetAnnotations())
			if diffSpec != "" || diffLabels != "" || diffAnnotations != "" {
				compareResource.Diff.NeedUpdate = true
				compareResource.Diff.Diff = fmt.Sprintf("%s\n%s\n%s", diffLabels, diffAnnotations, diffSpec)
				compareResource.Current.SetLabels(compareResource.Expected.GetLabels())
				compareResource.Current.SetAnnotations(compareResource.Expected.GetAnnotations())
				err = SetSpec(compareResource.Current, expectedSpec)
				if err != nil {
					return diff, err
				}
				diff.NeedUpdate = true
				sb.WriteString(compareResource.Diff.Diff)
			}
		}
	}

	diff.Diff = sb.String()

	return diff, nil

}

// createOrUpdateRessourcesFromTemplate permit to create or update ressources computing from templates
func (r *TemplateController) createOrUpdateRessourcesFromTemplate(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	var d any

	d, err = helper.Get(data, "compareResources")
	if err != nil {
		return res, err
	}
	listCompareResource := d.([]CompareResource)
	if listCompareResource == nil {
		return res, nil
	}

	for _, compareResource := range listCompareResource {
		if compareResource.Diff.NeedCreate {
			r.log.Debugf("Create ressource %s/%s", compareResource.Expected.GetNamespace(), compareResource.Expected.GetName())
			if err = r.Client.Create(ctx, compareResource.Expected); err != nil {
				return res, errors.Wrapf(err, "Error when create ressource %s", compareResource.Expected.GetName())
			}
		} else if compareResource.Diff.NeedUpdate {
			r.log.Debugf("Update ressource %s/%s", compareResource.Current.GetNamespace(), compareResource.Current.GetName())
			if err = r.Client.Update(ctx, compareResource.Current); err != nil {
				return res, errors.Wrapf(err, "Error when update ressource %s", compareResource.Current.GetName())
			}
		}
	}

	return res, nil
}

// readListTemplateFromAnnotation return the list of template resource
func readListTemplateFromAnnotation(resource client.Object) (listNamespacedName []types.NamespacedName, err error) {
	targetTemplates := resource.GetAnnotations()[fmt.Sprintf("%s/templates", monitoringAnnotationKey)]
	listNamespacedName = make([]types.NamespacedName, 0)
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
	case "Node":
		ns, err := helpers.GetOperatorNamespace()
		if err != nil {
			return namespace, err
		}
		namespace = ns
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
	if resource.GetLabels() == nil {
		resource.SetLabels(map[string]string{})
	}
	resource.GetLabels()[fmt.Sprintf("%s/template-name", monitoringAnnotationKey)] = namespacedName.Name
	resource.GetLabels()[fmt.Sprintf("%s/template-namespace", monitoringAnnotationKey)] = namespacedName.Namespace
}

// processName permit to get the resource name generated from template
// It return the template name if name is not provided
func processName(templateO *v1alpha1.Template, placeholders map[string]any) (name string, err error) {
	if templateO.Spec.Name == "" {
		return templateO.Name, nil
	}

	t, err := template.New("template").Funcs(sprig.FuncMap()).Parse(templateO.Spec.Name)
	if err != nil {
		return "", errors.Wrapf(err, "Error when parse template name %s/%s", templateO.Namespace, templateO.Name)
	}
	buf := bytes.NewBufferString("")
	if err = t.Execute(buf, placeholders); err != nil {
		return "", errors.Wrap(err, "Error when execute template")
	}

	return buf.String(), nil
}
