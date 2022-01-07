package v1alpha1

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	//+kubebuilder:scaffold:imports
)

var testEnv *envtest.Environment

type V1alpha1TestSuite struct {
	suite.Suite
	k8sClient client.Client
}

func TestV1alpha1Suite(t *testing.T) {
	suite.Run(t, new(V1alpha1TestSuite))
}

func (t *V1alpha1TestSuite) SetupSuite() {
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	// Setup testenv
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:        []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ControlPlaneStartTimeout: 120 * time.Second,
		ControlPlaneStopTimeout:  120 * time.Second,
	}

	err := SchemeBuilder.AddToScheme(scheme.Scheme)
	if err != nil {
		panic(err)
	}
	cfg, err := testEnv.Start()
	if err != nil {
		panic(err)
	}

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		panic(err)
	}
	t.k8sClient = k8sClient

}

func (t *V1alpha1TestSuite) TearDownSuite() {
	err := testEnv.Stop()
	if err != nil {
		panic(err)
	}
}
