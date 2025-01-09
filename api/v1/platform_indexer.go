package v1

import (
	"context"

	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// SetupPlatformIndexer setup indexer for platform
func SetupPlatformIndexer(k8sManager manager.Manager) (err error) {
	if err := k8sManager.GetFieldIndexer().IndexField(context.Background(), &Platform{}, "spec.centreonSettings.secret", func(o client.Object) []string {
		p := o.(*Platform)
		return []string{p.Spec.CentreonSettings.Secret}
	}); err != nil {
		return err
	}

	// Addindexer to chekc unicity of default platform
	if err := k8sManager.GetFieldIndexer().IndexField(context.Background(), &Platform{}, "spec.isDefault", func(o client.Object) []string {
		p := o.(*Platform)
		return []string{helpers.BoolToString(&p.Spec.IsDefault)}
	}); err != nil {
		return err
	}

	return nil
}
