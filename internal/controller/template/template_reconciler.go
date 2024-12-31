package template

import (
	"context"
	"encoding/json"
	"fmt"

	"emperror.dev/errors"
	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/disaster37/operator-sdk-extra/pkg/helper"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TemplateReconciler struct {
	controller.SentinelReconcilerAction
}

func NewTemplateReconciler(client client.Client, recorder record.EventRecorder) (sentinelReconcilerAction controller.SentinelReconcilerAction) {
	return &TemplateReconciler{
		SentinelReconcilerAction: controller.NewBasicSentinelAction(
			client,
			recorder,
		),
	}
}

// Read templates
func (r *TemplateReconciler) Read(ctx context.Context, resource client.Object, data map[string]any, logger *logrus.Entry) (read controller.SentinelRead, res ctrl.Result, err error) {
	var template *centreoncrd.Template
	listNamespacedName := make([]types.NamespacedName, 0)
	read = controller.NewBasicSentinelRead()
	var v any
	var placeholders map[string]any
	var expectedObject client.Object
	var currentObject client.Object
	expectedObjects := map[string][]client.Object{}
	currentObjects := map[string][]client.Object{}

	v, err = helper.Get(data, "placeholders")
	if err != nil {
		placeholders = v.(map[string]any)
	}
	templateBuilder := newBuilder(resource, r.Client().Scheme()).
		AddPlaceholders(placeholders).
		For(&centreoncrd.CentreonServiceGroup{}, &centreoncrd.CentreonServiceGroupList{}).
		For(&centreoncrd.CentreonService{}, &centreoncrd.CentreonServiceList{})

	// Check if template annotations
	targetTemplates := resource.GetAnnotations()[fmt.Sprintf("%s/templates", centreoncrd.MonitoringAnnotationKey)]
	if targetTemplates != "" {
		if err = json.Unmarshal([]byte(targetTemplates), &listNamespacedName); err != nil {
			return nil, res, errors.Wrap(err, "Error when unmarshall the list of template")
		}
	}

	// No template to process
	// We stop here
	if len(targetTemplates) == 0 {
		logger.Debug("No template found")
		return read, res, nil
	}

	// Get expected objects
	// Get templates and process thems
	for _, namespacedName := range listNamespacedName {
		template = &centreoncrd.Template{}
		logger.Debugf("Process template %s/%s", namespacedName.Namespace, namespacedName.Name)

		if err = r.Client().Get(ctx, namespacedName, template); err != nil {
			if !k8serrors.IsNotFound(err) {
				logger.Warnf("Template %s/%s not found. We skip it", namespacedName.Namespace, namespacedName.Name)
				continue
			}
			return nil, res, errors.Wrapf(err, "Error when get template %s/%s", namespacedName.Namespace, namespacedName.Name)
		}

		expectedObject, err = templateBuilder.Process(template)
		if err != nil {
			return read, res, errors.Wrapf(err, "Error when process template %s/%s; %s", namespacedName.Namespace, namespacedName.Name, err.Error())
		}
		expectedObject.SetLabels(getLabels(
			resource,
			map[string]string{
				fmt.Sprintf("%s/template", centreoncrd.MonitoringAnnotationKey): fmt.Sprintf("%s/%s", template.GetNamespace(), template.GetName()),
			},
		))
		expectedObject.SetNamespace(resource.GetNamespace())
		if expectedObject.GetName() == "" {
			expectedObject.SetName(template.GetName())
		}
		if expectedObjects[helpers.GetObjectType(expectedObject)] == nil {
			expectedObjects[helpers.GetObjectType(expectedObject)] = []client.Object{
				expectedObject,
			}
		} else {
			expectedObjects[helpers.GetObjectType(expectedObject)] = append(expectedObjects[helpers.GetObjectType(expectedObject)], expectedObject)
		}

		// Get current object
		// It's to temporary support object already created. There are not yet labels
		currentObject = helpers.CloneObject(expectedObject)
		if err = r.Client().Get(ctx, types.NamespacedName{Namespace: expectedObject.GetNamespace(), Name: expectedObject.GetName()}, currentObject); err != nil {
			if !k8serrors.IsNotFound(err) {
				return read, res, errors.Wrapf(err, "Error when read object %s/%s", expectedObject.GetNamespace(), expectedObject.GetName())
			}
		} else {
			if currentObjects[helpers.GetObjectType(currentObject)] == nil {
				currentObjects[helpers.GetObjectType(currentObject)] = []client.Object{
					currentObject,
				}
			} else {
				currentObjects[helpers.GetObjectType(currentObject)] = append(currentObjects[helpers.GetObjectType(currentObject)], currentObject)
			}
		}
	}

	// Get all object supported to check if there are orphan
	// We need to gel all children object from labels
	labelSelectors, err := labels.Parse(fmt.Sprintf("%s/parent=%s/%s", centreoncrd.MonitoringAnnotationKey, resource.GetNamespace(), resource.GetName()))
	if err != nil {
		return read, res, errors.Wrap(err, "Error when generate label selector")
	}
	for _, currentObjectList := range templateBuilder.Lists() {
		if err = r.Client().List(ctx, currentObjectList, &client.ListOptions{Namespace: resource.GetNamespace(), LabelSelector: labelSelectors}); err != nil {
			return read, res, errors.Wrapf(err, "Error when read objects")
		}

		if len(currentObjectList.GetItems()) > 0 {
			if currentObjects[helpers.GetObjectType(currentObjectList.GetItems()[0])] == nil {
				currentObjects[helpers.GetObjectType(currentObjectList.GetItems()[0])] = currentObjectList.GetItems()
			} else {
				currentObjects[helpers.GetObjectType(currentObjectList.GetItems()[0])] = append(currentObjects[helpers.GetObjectType(currentObjectList.GetItems()[0])], currentObjectList.GetItems()...)
			}
		}
	}

	// Delete duplicate
	for key, objects := range currentObjects {
		read.SetCurrentObjects(key, funk.UniqBy(objects, func(o client.Object) string {
			return fmt.Sprintf("%s/%s/%s/%s", o.GetObjectKind().GroupVersionKind().Group, o.GetObjectKind().GroupVersionKind().Kind, o.GetNamespace(), o.GetName())
		}).([]client.Object))
	}
	for key, objects := range expectedObjects {
		read.SetExpectedObjects(key, funk.UniqBy(objects, func(o client.Object) string {
			return fmt.Sprintf("%s/%s/%s/%s", o.GetObjectKind().GroupVersionKind().Group, o.GetObjectKind().GroupVersionKind().Kind, o.GetNamespace(), o.GetName())
		}).([]client.Object))
	}

	return read, res, nil
}
