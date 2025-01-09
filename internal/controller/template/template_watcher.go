package template

import (
	"context"
	"fmt"

	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// WatchTemplate permit to search resource created from Template to reconcil parents of them
func WatchTemplate(c client.Client, parent client.ObjectList) handler.MapFunc {
	return func(ctx context.Context, a client.Object) []reconcile.Request {
		reconcileRequests := make([]reconcile.Request, 0)

		// templates
		parentList := helpers.CloneObject(parent)
		fs := fields.ParseSelectorOrDie(fmt.Sprintf("%s.templates=%s/%s", centreoncrd.MonitoringAnnotationKey, a.GetNamespace(), a.GetName()))
		if err := c.List(context.Background(), parentList, &client.ListOptions{FieldSelector: fs}); err != nil {
			panic(err)
		}
		for _, k := range helpers.GetItems(parentList) {
			reconcileRequests = append(reconcileRequests, reconcile.Request{NamespacedName: types.NamespacedName{Name: k.GetName(), Namespace: k.GetNamespace()}})
		}

		return reconcileRequests
	}
}
