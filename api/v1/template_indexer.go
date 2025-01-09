package v1

import (
	"context"
	"encoding/json"
	"fmt"

	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func templateIndexer(o client.Object) []string {
	targetTemplates := o.GetAnnotations()[fmt.Sprintf("%s/templates", MonitoringAnnotationKey)]
	if targetTemplates == "" {
		return nil
	}

	listNamespacedName := make([]types.NamespacedName, 0)
	if err := json.Unmarshal([]byte(targetTemplates), &listNamespacedName); err != nil {
		return nil
	}

	res := make([]string, 0, len(listNamespacedName))
	for _, namespacedName := range listNamespacedName {
		res = append(res, fmt.Sprintf("%s/%s", namespacedName.Namespace, namespacedName.Name))
	}
	return res
}

// SetupCertificateIndexer setup indexer for secret (certificate)
func SetupCertificateIndexer(k8sManager manager.Manager) (err error) {
	if err := k8sManager.GetFieldIndexer().IndexField(context.Background(), &corev1.Secret{}, fmt.Sprintf("%s.templates", MonitoringAnnotationKey), templateIndexer); err != nil {
		return err
	}
	return nil
}

// SetupCertificateIndexer setup indexer for secret (certificate)
func SetupIngressIndexer(k8sManager manager.Manager) (err error) {
	if err := k8sManager.GetFieldIndexer().IndexField(context.Background(), &networkv1.Ingress{}, fmt.Sprintf("%s.templates", MonitoringAnnotationKey), templateIndexer); err != nil {
		return err
	}
	return nil
}

// SetupCertificateIndexer setup indexer for secret (certificate)
func SetupNamespaceIndexer(k8sManager manager.Manager) (err error) {
	if err := k8sManager.GetFieldIndexer().IndexField(context.Background(), &corev1.Namespace{}, fmt.Sprintf("%s.templates", MonitoringAnnotationKey), templateIndexer); err != nil {
		return err
	}
	return nil
}

// SetupCertificateIndexer setup indexer for secret (certificate)
func SetupNodeIndexer(k8sManager manager.Manager) (err error) {
	if err := k8sManager.GetFieldIndexer().IndexField(context.Background(), &corev1.Node{}, fmt.Sprintf("%s.templates", MonitoringAnnotationKey), templateIndexer); err != nil {
		return err
	}
	return nil
}

// SetupCertificateIndexer setup indexer for secret (certificate)
func SetupRouteIndexer(k8sManager manager.Manager) (err error) {
	if err := k8sManager.GetFieldIndexer().IndexField(context.Background(), &routev1.Route{}, fmt.Sprintf("%s.templates", MonitoringAnnotationKey), templateIndexer); err != nil {
		return err
	}
	return nil
}
