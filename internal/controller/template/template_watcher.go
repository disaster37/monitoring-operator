package template

import (
	"context"
	"encoding/json"
	"fmt"

	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// WatchTemplate permit to search resource created from Template to reconcil parents of them
func WatchTemplate(c client.Client, parent client.ObjectList) handler.MapFunc {
	return func(ctx context.Context, a client.Object) []reconcile.Request {

		var (
			listRessources client.ObjectList
			ls             labels.Selector
			err            error
		)

		reconcileRequests := make([]reconcile.Request, 0)
		template := a.(*centreoncrd.Template)

		// Old style
		if template.Spec.Type != "" {
			ls, err = labels.Parse(fmt.Sprintf("%s/template-name=%s,%s/template-namespace=%s", centreoncrd.MonitoringAnnotationKey, a.GetName(), centreoncrd.MonitoringAnnotationKey, a.GetNamespace()))
			if err != nil {
				panic(err)
			}

			// Get object type
			switch template.Spec.Type {
			case "CentreonService":
				listRessources = &centreoncrd.CentreonServiceList{}
			case "CentreonServiceGroup":
				listRessources = &centreoncrd.CentreonServiceGroupList{}
			default:
				return reconcileRequests
			}

			// Get all resources created from this template
			if err := c.List(context.Background(), listRessources, &client.ListOptions{LabelSelector: ls}); err != nil {
				panic(err)
			}

			items := helpers.GetItems(listRessources)
			for _, item := range items {
				// Search parent to reconcile parent
				for _, parent := range item.GetOwnerReferences() {
					reconcileRequests = append(reconcileRequests, reconcile.Request{NamespacedName: types.NamespacedName{Name: parent.Name, Namespace: item.GetNamespace()}})
				}
			}

			return reconcileRequests
		}

		parent = helpers.CloneObject[client.ObjectList](parent)

		// New style
		listNamespacedName := make([]types.NamespacedName, 0)
		ls, err = labels.Parse(fmt.Sprintf("%s=true", centreoncrd.MonitoringAnnotationKey))
		if err != nil {
			panic(err)
		}

		// Get all resources that can be call this template
		if err := c.List(context.Background(), parent, &client.ListOptions{LabelSelector: ls}); err != nil {
			panic(err)
		}

		// Now we need to open all annotations templates to found if this current template is called on
		objects := helpers.GetItems(parent)
		for _, object := range objects {
			targetTemplates := object.GetAnnotations()[fmt.Sprintf("%s/templates", centreoncrd.MonitoringAnnotationKey)]
			if targetTemplates != "" {
				if err = json.Unmarshal([]byte(targetTemplates), &listNamespacedName); err != nil {
					return nil
				}

				for _, nn := range listNamespacedName {
					if nn.Namespace == template.Namespace && nn.Name == template.Name {
						reconcileRequests = append(reconcileRequests, reconcile.Request{NamespacedName: types.NamespacedName{Name: object.GetName(), Namespace: object.GetNamespace()}})
					}
				}
			}
		}

		return reconcileRequests
	}
}
