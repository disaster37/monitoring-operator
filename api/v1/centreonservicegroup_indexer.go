package v1

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// SetupCentreonServiceIndexer setup indexer for CentreonService
func SetupCentreonServiceGroupIndexer(k8sManager manager.Manager) (err error) {
	// Index external name needed by webhook to controle unicity
	if err = k8sManager.GetFieldIndexer().IndexField(context.Background(), &CentreonServiceGroup{}, "spec.externalName", func(o client.Object) []string {
		p := o.(*CentreonServiceGroup)
		return []string{p.GetExternalName()}
	}); err != nil {
		return err
	}

	// Index target platform needed by webhook to controle unicity
	if err = k8sManager.GetFieldIndexer().IndexField(context.Background(), &CentreonServiceGroup{}, "spec.targetPlatform", func(o client.Object) []string {
		p := o.(*CentreonServiceGroup)
		return []string{p.GetPlatform()}
	}); err != nil {
		return err
	}

	return nil
}
