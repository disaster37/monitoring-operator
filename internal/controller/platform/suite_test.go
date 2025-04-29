package platform

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	//+kubebuilder:scaffold:imports
)

var testEnv *envtest.Environment

type PlatformControllerTestSuite struct {
	suite.Suite
	k8sClient client.Client
	cfg       *rest.Config
	platforms map[string]*ComputedPlatform
}

func TestPlatformControllerSuite(t *testing.T) {
	suite.Run(t, new(PlatformControllerTestSuite))
}

func (t *PlatformControllerTestSuite) SetupSuite() {
	logf.SetLogger(zap.New(zap.UseDevMode(true)))
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableQuote: true,
	})

	// Init controllers
	_ = os.Setenv("POD_NAMESPACE", "default")

	// Setup testenv
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("../../..", "config", "crd", "bases"),
			filepath.Join("../../..", "config", "crd", "externals"),
		},
		ErrorIfCRDPathMissing:    true,
		ControlPlaneStopTimeout:  120 * time.Second,
		ControlPlaneStartTimeout: 120 * time.Second,
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths: []string{filepath.Join("..", "..", "..", "config", "webhook")},
		},
	}
	cfg, err := testEnv.Start()
	if err != nil {
		panic(err)
	}
	t.cfg = cfg

	// Add CRD sheme
	err = scheme.AddToScheme(scheme.Scheme)
	if err != nil {
		panic(err)
	}
	err = centreoncrd.AddToScheme(scheme.Scheme)
	if err != nil {
		panic(err)
	}
	err = routev1.AddToScheme(scheme.Scheme)
	if err != nil {
		panic(err)
	}

	// Init k8smanager and k8sclient
	webhookInstallOptions := &testEnv.WebhookInstallOptions
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
		WebhookServer: webhook.NewServer(webhook.Options{
			Host:    webhookInstallOptions.LocalServingHost,
			Port:    webhookInstallOptions.LocalServingPort,
			CertDir: webhookInstallOptions.LocalServingCertDir,
		}),
		LeaderElection: false,
		Metrics:        metricsserver.Options{BindAddress: "0"},
	})
	if err != nil {
		panic(err)
	}
	k8sClient := k8sManager.GetClient()
	t.k8sClient = k8sClient

	// Setup indexer
	if err := controller.SetupIndexerWithManager(
		k8sManager,
		centreoncrd.SetupPlatformIndexer,
		centreoncrd.SetupCentreonServiceIndexer,
		centreoncrd.SetupCentreonServiceGroupIndexer,
		centreoncrd.SetupCertificateIndexer,
		centreoncrd.SetupIngressIndexer,
		centreoncrd.SetupNamespaceIndexer,
		centreoncrd.SetupNodeIndexer,
		centreoncrd.SetupRouteIndexer,
	); err != nil {
		panic(err)
	}

	// Setup webhook
	if err := controller.SetupWebhookWithManager(
		k8sManager,
		k8sClient,
		centreoncrd.SetupCentreonServiceWebhookWithManager,
		centreoncrd.SetupCentreonServiceGroupWebhookWithManager,
		centreoncrd.SetupPlatformWebhookWithManager,
	); err != nil {
		panic(err)
	}

	platforms := map[string]*ComputedPlatform{}
	t.platforms = platforms

	platformReconsiler := NewPlatformReconciler(
		k8sClient,
		logrus.NewEntry(logrus.StandardLogger()),
		k8sManager.GetEventRecorderFor("plateform-controller"),
		t.platforms,
	)
	if err = platformReconsiler.SetupWithManager(k8sManager); err != nil {
		panic(err)
	}

	go func() {
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		if err != nil {
			panic(err)
		}
	}()
}

func (t *PlatformControllerTestSuite) TearDownSuite() {
	err := testEnv.Stop()
	if err != nil {
		panic(err)
	}
}

func (t *PlatformControllerTestSuite) BeforeTest(suiteName, testName string) {
}

func (t *PlatformControllerTestSuite) AfterTest(suiteName, testName string) {
}
