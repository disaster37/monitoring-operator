package controllers

import (
	"context"
	"fmt"

	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	monitorv1alpha1 "github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/disaster37/operator-sdk-extra/pkg/helper"
	"github.com/google/go-cmp/cmp"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *RouteReconciler) readForCentreonPlatform(ctx context.Context, route *routev1.Route, platform *v1alpha1.Platform, data map[string]any, meta any) (res ctrl.Result, err error) {
	// Get endpoint spec from platform
	endpointSpec := platform.Spec.CentreonSettings.Endpoint
	if endpointSpec == nil {
		r.log.Warning("It's recommanded to set some default endpoint values on target platform. It avoid to set on each route all Centreon service properties as annotations")
	}

	// Get if current CentreonService object already exist
	currentCS := &v1alpha1.CentreonService{}
	err = r.Get(ctx, types.NamespacedName{Name: route.Name, Namespace: route.Namespace}, currentCS)
	if err != nil && k8serrors.IsNotFound(err) {
		data["currentCentreonService"] = nil
	} else if err != nil {
		return res, errors.Wrap(err, "Error when get current CentreonService object")
	} else {
		data["currentCentreonService"] = currentCS
	}

	// Compute expected Centreon service
	expectedCS := &v1alpha1.CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:        route.Name,
			Namespace:   route.Namespace,
			Labels:      route.GetLabels(),
			Annotations: route.GetAnnotations(),
		},
		Spec: monitorv1alpha1.CentreonServiceSpec{},
	}
	placeholders := generatePlaceholdersRoute(route)
	initCentreonServiceDefaultValue(endpointSpec, expectedCS, placeholders)
	if err = initCentreonServiceFromAnnotations(route.GetAnnotations(), expectedCS); err != nil {
		return res, errors.Wrap(err, "Error when init CentreonService from route annotations")
	}

	// Check CentreonService is valide
	if !expectedCS.IsValid() {
		return res, fmt.Errorf("Generated CentreonService is not valid: %+v", expectedCS.Spec)
	}

	// Set route instance as the owner
	ctrl.SetControllerReference(route, expectedCS, r.Scheme)

	data["expectedCentreonService"] = expectedCS

	return res, nil

}

func (r *RouteReconciler) createForCentreonPlatform(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	var d any

	d, err = helper.Get(data, "expectedCentreonService")
	if err != nil {
		return res, err
	}
	expectedCS := d.(*v1alpha1.CentreonService)

	if err = r.Client.Create(ctx, expectedCS); err != nil {
		return res, errors.Wrap(err, "Error when create CentreonService object")
	}

	return res, nil
}

func (r *RouteReconciler) updateForCentreonPlatform(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	var d any

	d, err = helper.Get(data, "expectedCentreonService")
	if err != nil {
		return res, err
	}
	expectedCS := d.(*v1alpha1.CentreonService)

	if err = r.Client.Update(ctx, expectedCS); err != nil {
		return res, errors.Wrap(err, "Error when update CentreonService object")
	}

	return res, nil
}

func (r *RouteReconciler) diffForCentreonPlatform(resource client.Object, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	var (
		d          any
		currentCS  *v1alpha1.CentreonService
		expectedCS *v1alpha1.CentreonService
	)

	d, err = helper.Get(data, "currentCentreonService")
	if err != nil {
		return diff, err
	}
	if d == nil {
		currentCS = nil
	} else {
		currentCS = d.(*v1alpha1.CentreonService)
	}

	d, err = helper.Get(data, "expectedCentreonService")
	if err != nil {
		return diff, err
	}
	expectedCS = d.(*v1alpha1.CentreonService)

	diff = controller.Diff{
		NeedCreate: false,
		NeedUpdate: false,
	}
	if currentCS == nil {
		diff.NeedCreate = true
		diff.Diff = "CentreonService object not exist"
		return diff, nil
	}

	diffSpec := cmp.Diff(currentCS.Spec, expectedCS.Spec)
	diffLabels := cmp.Diff(currentCS.GetLabels(), expectedCS.GetLabels())
	diffAnnotations := cmp.Diff(currentCS.GetAnnotations(), expectedCS.GetAnnotations())
	if diffSpec != "" || diffLabels != "" || diffAnnotations != "" {
		diff.NeedUpdate = true
		diff.Diff = fmt.Sprintf("%s\n%s\n%s", diffLabels, diffAnnotations, diffSpec)

		currentCS.SetLabels(expectedCS.GetLabels())
		currentCS.SetAnnotations(expectedCS.GetAnnotations())
		currentCS.Spec = expectedCS.Spec
		data["expectedCentreonService"] = currentCS
	}

	return
}
