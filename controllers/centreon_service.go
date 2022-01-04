package controllers

import (
	"context"

	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	namespaceOperator    = "monitoring-operator"
	centreonResourceName = "monitoring-operator"
)

func getCentreonSpec(ctx context.Context, client client.Client) (spec *v1alpha1.CentreonSpec, err error) {

	ns, err := helpers.GetOperatorNamespace()
	if err != nil {
		ns = namespaceOperator
	}

	namespaced := types.NamespacedName{
		Name:      centreonResourceName,
		Namespace: ns,
	}

	instance := &v1alpha1.Centreon{}
	if err := client.Get(ctx, namespaced, instance); err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return &instance.Spec, nil
}
