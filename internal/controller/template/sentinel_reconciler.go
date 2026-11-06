package template

import (
	"context"
	"fmt"

	"emperror.dev/errors"
	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/disaster37/monitoring-operator/pkg/object"
	"github.com/disaster37/operator-sdk-extra/pkg/apis/shared"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	NamespaceFinalizer   shared.FinalizerName = "namespace.monitor.k8s.webcenter.fr/finalizer"
	IngressFinalizer     shared.FinalizerName = "ingress.monitor.k8s.webcenter.fr/finalizer"
	RouteFinalizer       shared.FinalizerName = "route.monitor.k8s.webcenter.fr/finalizer"
	CertificateFinalizer shared.FinalizerName = "certificate.monitor.k8s.webcenter.fr/finalizer"
	NodeFinalizer        shared.FinalizerName = "node.monitor.k8s.webcenter.fr/finalizer"
	TemplateFinalizer    shared.FinalizerName = "template.monitor.k8s.webcenter.fr/finalizer"
)

var LegacyFinalizers = []shared.FinalizerName{
	NamespaceFinalizer,
	IngressFinalizer,
	RouteFinalizer,
	CertificateFinalizer,
	NodeFinalizer,
	TemplateFinalizer,
}

// TemplateSentinelReconciler wraps the sentinel reconciler to manage the finalizer
// and the cleanup of template generated children on resource deletion.
type TemplateSentinelReconciler struct {
	controller.BaseReconciler
	sentinelReconciler controller.SentinelReconciler
	finalizer          shared.FinalizerName
	logger             *logrus.Entry
}

// NewSentinelReconciler creates a sentinel reconciler that manages the given finalizer.
func NewSentinelReconciler(
	client client.Client,
	name string,
	finalizer shared.FinalizerName,
	logger *logrus.Entry,
	recorder record.EventRecorder,
) controller.SentinelReconciler {
	return &TemplateSentinelReconciler{
		BaseReconciler: controller.NewBaseReconciler(client, recorder),
		sentinelReconciler: controller.NewBasicSentinelReconciler(
			client,
			name,
			logger,
			recorder,
		),
		finalizer: finalizer,
		logger: logger.WithFields(logrus.Fields{
			"reconciler": name,
		}),
	}
}

// Reconcile manages the finalizer lifecycle before delegating to the sentinel reconciler.
func (h *TemplateSentinelReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
	o client.Object,
	data map[string]any,
	reconcilerAction controller.SentinelReconcilerAction,
) (res ctrl.Result, err error) {
	logger := h.logger.WithFields(logrus.Fields{
		"name":      req.Name,
		"namespace": req.Namespace,
	})

	// Get current resource
	if err = h.Client().Get(ctx, req.NamespacedName, o); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		logger.Errorf("Error when get object: %s", err.Error())
		return res, errors.Wrap(err, controller.ErrWhenGetObjectFromReconciler.Error())
	}

	// Ignore if needed by annotation. A deleting object that still carries a
	// monitoring finalizer must be processed so the finalizer cannot block deletion.
	if o.GetDeletionTimestamp().IsZero() && o.GetAnnotations()[fmt.Sprintf("%s/ignoreReconcile", controller.BaseAnnotation)] == "true" {
		logger.Info("Found annotation on ressource to ignore reconcile")
		return res, nil
	}

	// Handle deletion
	if !o.GetDeletionTimestamp().IsZero() {
		if !hasMonitoringFinalizer(o, h.finalizer) {
			return ctrl.Result{}, nil
		}

		return h.cleanupAndRemoveFinalizers(ctx, o, data, reconcilerAction, logger, "from reconciler")
	}

	// Handle absence of template annotation on non-deleting object
	if !isMonitoringTemplateAnnotation(o) {
		return h.cleanupAndRemoveFinalizers(ctx, o, data, reconcilerAction, logger, "after annotation removal")
	}

	// Ensure finalizer is present when template annotation is active
	if h.finalizer != "" && !controllerutil.ContainsFinalizer(o, h.finalizer.String()) {
		controllerutil.AddFinalizer(o, h.finalizer.String())
		if err = h.Client().Update(ctx, o); err != nil {
			logger.Errorf("Error when add finalizer: %s", err.Error())
			return reconcilerAction.OnError(ctx, o, data, errors.Wrap(err, controller.ErrWhenAddFinalizer.Error()), logger)
		}
		logger.Debug("Add finalizer successfully, force requeue object")
		return ctrl.Result{Requeue: true}, nil
	}

	// Delegate normal reconciliation to underlying SentinelReconciler
	return h.sentinelReconciler.Reconcile(ctx, req, o, data, reconcilerAction)
}

func (h *TemplateSentinelReconciler) cleanupAndRemoveFinalizers(
	ctx context.Context,
	o client.Object,
	data map[string]any,
	reconcilerAction controller.SentinelReconcilerAction,
	logger *logrus.Entry,
	cleanupContext string,
) (ctrl.Result, error) {
	if err := h.cleanupChildren(ctx, o, data, reconcilerAction, logger); err != nil {
		logger.Errorf("Error when cleanup children %s: %s", cleanupContext, err.Error())
		return reconcilerAction.OnError(ctx, o, data, errors.Wrap(err, controller.ErrWhenCallDeleteFromReconciler.Error()), logger)
	}

	if removeMonitoringFinalizers(o, h.finalizer) {
		if err := h.Client().Update(ctx, o); err != nil {
			if k8serrors.IsNotFound(err) {
				return ctrl.Result{}, nil
			}
			logger.Errorf("Failed to remove finalizer: %s", err.Error())
			return reconcilerAction.OnError(ctx, o, data, errors.Wrap(err, controller.ErrWhenDeleteFinalizer.Error()), logger)
		}
		logger.Debug("Remove finalizer successfully")
	}

	return ctrl.Result{}, nil
}

// resolveChildNamespace determines the namespace where template-generated children live for a given parent object.
func resolveChildNamespace(resource client.Object) (string, error) {
	kind := resource.GetObjectKind().GroupVersionKind().Kind
	if kind == "" {
		switch resource.(type) {
		case *corev1.Namespace:
			kind = "Namespace"
		case *corev1.Node:
			kind = "Node"
		}
	}

	switch kind {
	case "Namespace":
		return resource.GetName(), nil
	case "Node":
		ns, err := helpers.GetOperatorNamespace()
		if err != nil {
			return "", errors.Wrap(err, "Error when get operator namespace")
		}
		return ns, nil
	default:
		return resource.GetNamespace(), nil
	}
}

func (h *TemplateSentinelReconciler) cleanupChildren(
	ctx context.Context,
	resource client.Object,
	data map[string]any,
	reconcilerAction controller.SentinelReconcilerAction,
	logger *logrus.Entry,
) error {
	namespace, err := resolveChildNamespace(resource)
	if err != nil {
		return err
	}

	childrenToDelete := make([]client.Object, 0)

	// Label-based discovery (primary mechanism)
	labelSelector, err := labels.Parse(fmt.Sprintf("%s/parent=%s.%s", centreoncrd.MonitoringAnnotationKey, namespace, resource.GetName()))
	if err != nil {
		return errors.Wrap(err, "Error when generate label selector")
	}

	childLists := []object.ObjectList{
		&centreoncrd.CentreonServiceGroupList{},
		&centreoncrd.CentreonServiceList{},
	}

	for _, list := range childLists {
		if err := h.Client().List(ctx, list, &client.ListOptions{Namespace: namespace, LabelSelector: labelSelector}); err != nil {
			return errors.Wrap(err, "Error when list child objects")
		}
		childrenToDelete = append(childrenToDelete, list.GetItems()...)
	}

	// Best-effort template evaluation fallback (for legacy children without parent label)
	if isMonitoringTemplateAnnotation(resource) && reconcilerAction != nil {
		read, _, readErr := reconcilerAction.Read(ctx, resource, data, logger)
		if readErr != nil {
			logger.Warnf("Best-effort template read failed during cleanup (continuing with label-discovered children): %s", readErr.Error())
		} else if read != nil {
			for _, objs := range read.GetAllCurrentObjects() {
				childrenToDelete = append(childrenToDelete, objs...)
			}
			for _, objs := range read.GetAllExpectedObjects() {
				childrenToDelete = append(childrenToDelete, objs...)
			}
		}
	}

	// Deduplicate by type/namespace/name
	uniqueChildren := funk.UniqBy(childrenToDelete, func(o client.Object) string {
		gvk := o.GetObjectKind().GroupVersionKind()
		return fmt.Sprintf("%T/%s/%s/%s/%s", o, gvk.Group, gvk.Kind, o.GetNamespace(), o.GetName())
	}).([]client.Object)

	for _, child := range uniqueChildren {
		if err := h.Client().Delete(ctx, child); err != nil {
			if k8serrors.IsNotFound(err) {
				continue
			}
			return errors.Wrapf(err, "Error when delete object '%s/%s'", child.GetNamespace(), child.GetName())
		}
		logger.Debugf("Delete object '%s/%s' successfully", child.GetNamespace(), child.GetName())
		h.Recorder().Eventf(resource, corev1.EventTypeNormal, "DeleteCompleted", "Object '%s' successfully deleted", child.GetName())
	}

	return nil
}

func hasMonitoringFinalizer(o client.Object, configuredFinalizer shared.FinalizerName) bool {
	if configuredFinalizer != "" && controllerutil.ContainsFinalizer(o, configuredFinalizer.String()) {
		return true
	}
	for _, f := range LegacyFinalizers {
		if controllerutil.ContainsFinalizer(o, f.String()) {
			return true
		}
	}
	return false
}

func removeMonitoringFinalizers(o client.Object, configuredFinalizer shared.FinalizerName) bool {
	removed := false
	if configuredFinalizer != "" && controllerutil.ContainsFinalizer(o, configuredFinalizer.String()) {
		controllerutil.RemoveFinalizer(o, configuredFinalizer.String())
		removed = true
	}
	for _, f := range LegacyFinalizers {
		if controllerutil.ContainsFinalizer(o, f.String()) {
			controllerutil.RemoveFinalizer(o, f.String())
			removed = true
		}
	}
	return removed
}
