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

//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=templates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=templates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=templates/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=patch;get;create

// NewTemplateReconciler create template reconciler
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
	placeholders := map[string]any{}
	var expectedObject client.Object
	var currentObject client.Object
	expectedObjects := map[string][]client.Object{}
	currentObjects := map[string][]client.Object{}
	var namespace string

	v, err = helper.Get(data, "placeholders")
	if err == nil {
		placeholders = v.(map[string]any)
	}

	// Add a special workground for Namespace object that are cluster wide, so not namespace concept
	switch resource.GetObjectKind().GroupVersionKind().Kind {
	case "Namespace":
		namespace = resource.GetName()
		placeholders["namespace"] = namespace
	case "Node":
		namespace, err = helpers.GetOperatorNamespace()
		if err != nil {
			return nil, res, errors.Wrap(err, "Error when get operator namespace")
		}
	default:
		namespace = resource.GetNamespace()
	}

	templateBuilder := newBuilder(resource, r.Client().Scheme()).
		AddPlaceholders(placeholders).
		For(&centreoncrd.CentreonServiceGroup{}, &centreoncrd.CentreonServiceGroupList{}).
		For(&centreoncrd.CentreonService{}, &centreoncrd.CentreonServiceList{})

	// Get all existing objects  created from parent
	// We need to gel all children object from labels
	labelSelectors, err := labels.Parse(fmt.Sprintf("%s/parent=%s.%s", centreoncrd.MonitoringAnnotationKey, namespace, resource.GetName()))
	if err != nil {
		return read, res, errors.Wrap(err, "Error when generate label selector")
	}
	for _, currentObjectList := range templateBuilder.Lists() {
		if err = r.Client().List(ctx, currentObjectList, &client.ListOptions{Namespace: namespace, LabelSelector: labelSelectors}); err != nil {
			return read, res, errors.Wrapf(err, "Error when read objects")
		}

		if len(currentObjectList.GetItems()) > 0 {
			if currentObjects[helpers.GetObjectType(currentObjectList.GetItems()[0].GetObjectKind())] == nil {
				currentObjects[helpers.GetObjectType(currentObjectList.GetItems()[0].GetObjectKind())] = currentObjectList.GetItems()
			} else {
				currentObjects[helpers.GetObjectType(currentObjectList.GetItems()[0].GetObjectKind())] = append(currentObjects[helpers.GetObjectType(currentObjectList.GetItems()[0].GetObjectKind())], currentObjectList.GetItems()...)
			}
		}
	}

	// Compute expectings children from template
	// Get templates and process thems
	targetTemplates := resource.GetAnnotations()[fmt.Sprintf("%s/templates", centreoncrd.MonitoringAnnotationKey)]
	if targetTemplates != "" {
		if err = json.Unmarshal([]byte(targetTemplates), &listNamespacedName); err != nil {
			return nil, res, errors.Wrap(err, "Error when unmarshall the list of template")
		}
	}

	for _, namespacedName := range listNamespacedName {
		template = &centreoncrd.Template{}
		logger.Debugf("Process template %s/%s", namespacedName.Namespace, namespacedName.Name)

		if err = r.Client().Get(ctx, namespacedName, template); err != nil {
			if k8serrors.IsNotFound(err) {
				logger.Warnf("Template %s/%s not found. We skip it", namespacedName.Namespace, namespacedName.Name)
				continue
			}
			return nil, res, errors.Wrapf(err, "Error when get template %s/%s", namespacedName.Namespace, namespacedName.Name)
		}

		expectedObject, err = templateBuilder.Process(template)
		if err != nil {
			return read, res, errors.Wrapf(err, "Error when process template %s/%s; %s", namespacedName.Namespace, namespacedName.Name, err.Error())
		}

		if expectedObject != nil {
			expectedObject.SetLabels(getLabels(
				resource,
				map[string]string{
					fmt.Sprintf("%s/template", centreoncrd.MonitoringAnnotationKey): fmt.Sprintf("%s.%s", template.GetNamespace(), template.GetName()),
					fmt.Sprintf("%s/parent", centreoncrd.MonitoringAnnotationKey):   fmt.Sprintf("%s.%s", namespace, resource.GetName()),
				},
			))
			expectedObject.SetNamespace(namespace)
			if expectedObject.GetName() == "" {
				expectedObject.SetName(template.GetName())
			}
			if expectedObjects[helpers.GetObjectType(expectedObject.GetObjectKind())] == nil {
				expectedObjects[helpers.GetObjectType(expectedObject.GetObjectKind())] = []client.Object{
					expectedObject,
				}
			} else {
				expectedObjects[helpers.GetObjectType(expectedObject.GetObjectKind())] = append(expectedObjects[helpers.GetObjectType(expectedObject.GetObjectKind())], expectedObject)
			}

			// Get current object
			// It's to temporary support object already created. There are not yet labels
			currentObject = helpers.CloneObject(expectedObject)
			if err = r.Client().Get(ctx, types.NamespacedName{Namespace: expectedObject.GetNamespace(), Name: expectedObject.GetName()}, currentObject); err != nil {
				if !k8serrors.IsNotFound(err) {
					return read, res, errors.Wrapf(err, "Error when read object %s/%s", expectedObject.GetNamespace(), expectedObject.GetName())
				}
			} else {
				if currentObjects[helpers.GetObjectType(currentObject.GetObjectKind())] == nil {
					currentObjects[helpers.GetObjectType(currentObject.GetObjectKind())] = []client.Object{
						currentObject,
					}
				} else {
					currentObjects[helpers.GetObjectType(currentObject.GetObjectKind())] = append(currentObjects[helpers.GetObjectType(currentObject.GetObjectKind())], currentObject)
				}
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
