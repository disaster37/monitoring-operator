package centreon

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/disaster37/operator-sdk-extra/pkg/mock"
	"github.com/disaster37/operator-sdk-extra/pkg/object"
	"github.com/golang/mock/gomock"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/disaster37/monitoring-operator/pkg/mocks"

	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	//+kubebuilder:scaffold:imports
)

var testEnv *envtest.Environment

type CentreonControllerTestSuite struct {
	suite.Suite
	k8sClient           client.Client
	mockCentreonHandler *mocks.MockCentreonHandler
	mockCtrl            *gomock.Controller
	cfg                 *rest.Config
	platforms           map[string]*ComputedPlatform
}

func TestCentreonControllerSuite(t *testing.T) {
	suite.Run(t, new(CentreonControllerTestSuite))
}

func (t *CentreonControllerTestSuite) SetupSuite() {
	// Init Centreon mock
	t.mockCtrl = gomock.NewController(t.T())
	t.mockCentreonHandler = mocks.NewMockCentreonHandler(t.mockCtrl)

	logf.SetLogger(zap.New(zap.UseDevMode(true)))
	logrus.SetLevel(logrus.TraceLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableQuote: true,
	})

	// Setup testenv
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("../../..", "config", "crd", "bases"),
			filepath.Join("../../..", "config", "crd", "externals"),
		},
		ErrorIfCRDPathMissing:    true,
		ControlPlaneStopTimeout:  120 * time.Second,
		ControlPlaneStartTimeout: 120 * time.Second,
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
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	if err != nil {
		panic(err)
	}
	k8sClient := k8sManager.GetClient()
	t.k8sClient = k8sClient

	// Add indexers on Platform to track secret change
	if err := k8sManager.GetFieldIndexer().IndexField(context.Background(), &centreoncrd.Platform{}, "spec.centreonSettings.secret", func(o client.Object) []string {
		p := o.(*centreoncrd.Platform)
		return []string{p.Spec.CentreonSettings.Secret}
	}); err != nil {
		panic(err)
	}

	// Init controllers
	os.Setenv("POD_NAMESPACE", "default")

	platforms := map[string]*ComputedPlatform{
		"default": {
			platform: &centreoncrd.Platform{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default",
					Namespace: "default",
				},
				Spec: centreoncrd.PlatformSpec{
					IsDefault:        true,
					PlatformType:     "centreon",
					CentreonSettings: &centreoncrd.PlatformSpecCentreonSettings{},
				},
			},
			client: t.mockCentreonHandler,
		},
	}
	t.platforms = platforms

	/*
		platformReconsiler := NewPlatformReconciler(k8sClient, scheme.Scheme)
		platformReconsiler.SetLogger(logrus.WithFields(logrus.Fields{
			"type": "platformController",
		}))
		platformReconsiler.SetRecorder(k8sManager.GetEventRecorderFor("platform-controller"))
		platformReconsiler.SetReconsiler(mock.NewMockReconciler(platformReconsiler, t.mockCentreonHandler))
		platformReconsiler.SetPlatforms(platforms)
		if err = platformReconsiler.SetupWithManager(k8sManager); err != nil {
			panic(err)
		}
	*/

	centreonServiceReconsiler := NewCentreonServiceReconciler(
		k8sClient,
		logrus.NewEntry(logrus.StandardLogger()),
		k8sManager.GetEventRecorderFor("centreonservice-controller"),
	)
	centreonServiceReconsiler.(*CentreonServiceReconciler).RemoteReconcilerAction = mock.NewMockRemoteReconcilerAction[*centreoncrd.CentreonService, *CentreonService, centreonhandler.CentreonHandler](
		centreonServiceReconsiler.(*CentreonServiceReconciler).RemoteReconcilerAction,
		func(ctx context.Context, req reconcile.Request, o object.RemoteObject) (handler controller.RemoteExternalReconciler[*centreoncrd.CentreonService, *CentreonService, centreonhandler.CentreonHandler], res reconcile.Result, err error) {
			return newCentreonServiceApiClient(t.mockCentreonHandler, logrus.NewEntry(logrus.StandardLogger())), res, nil
		},
	)
	if err = centreonServiceReconsiler.SetupWithManager(k8sManager); err != nil {
		panic(err)
	}

	/*
		centreonServiceGroupReconsiler := NewCentreonServiceGroupReconciler(k8sClient, scheme.Scheme)
		centreonServiceGroupReconsiler.SetLogger(logrus.WithFields(logrus.Fields{
			"type": "centreonServiceGroupController",
		}))
		centreonServiceGroupReconsiler.SetRecorder(k8sManager.GetEventRecorderFor("centreonservicegroup-controller"))
		centreonServiceGroupReconsiler.SetReconsiler(mock.NewMockReconciler(centreonServiceGroupReconsiler, t.mockCentreonHandler))
		centreonServiceGroupReconsiler.SetPlatforms(platforms)
		if err = centreonServiceGroupReconsiler.SetupWithManager(k8sManager); err != nil {
			panic(err)
		}

		templateController := TemplateController{
			Client: k8sClient,
			Scheme: scheme.Scheme,
		}
		templateController.SetLogger(logrus.WithFields(logrus.Fields{
			"type": "templateController",
		}))

		ingressReconsiler := NewIngressReconciler(k8sClient, scheme.Scheme, templateController)
		ingressReconsiler.Reconciler.SetLogger(logrus.WithFields(logrus.Fields{
			"type": "ingressController",
		}))
		ingressReconsiler.SetRecorder(k8sManager.GetEventRecorderFor("ingress-controller"))
		ingressReconsiler.SetReconsiler(mock.NewMockReconciler(ingressReconsiler, t.mockCentreonHandler))
		ingressReconsiler.SetPlatforms(platforms)
		if err = ingressReconsiler.SetupWithManager(k8sManager); err != nil {
			panic(err)
		}

		routeReconsiler := NewRouteReconciler(k8sClient, scheme.Scheme, templateController)
		routeReconsiler.Reconciler.SetLogger(logrus.WithFields(logrus.Fields{
			"type": "routeController",
		}))
		routeReconsiler.SetRecorder(k8sManager.GetEventRecorderFor("route-controller"))
		routeReconsiler.SetReconsiler(mock.NewMockReconciler(routeReconsiler, t.mockCentreonHandler))
		routeReconsiler.SetPlatforms(platforms)
		if err = routeReconsiler.SetupWithManager(k8sManager); err != nil {
			panic(err)
		}

		namespaceReconsiler := NewNamespaceReconciler(k8sClient, scheme.Scheme, templateController)
		namespaceReconsiler.Reconciler.SetLogger(logrus.WithFields(logrus.Fields{
			"type": "namespaceController",
		}))
		namespaceReconsiler.SetRecorder(k8sManager.GetEventRecorderFor("namespace-controller"))
		namespaceReconsiler.SetReconsiler(mock.NewMockReconciler(namespaceReconsiler, t.mockCentreonHandler))
		namespaceReconsiler.SetPlatforms(platforms)
		if err = namespaceReconsiler.SetupWithManager(k8sManager); err != nil {
			panic(err)
		}

		nodeReconsiler := NewNodeReconciler(k8sClient, scheme.Scheme, templateController)
		nodeReconsiler.Reconciler.SetLogger(logrus.WithFields(logrus.Fields{
			"type": "nodeController",
		}))
		nodeReconsiler.SetRecorder(k8sManager.GetEventRecorderFor("node-controller"))
		nodeReconsiler.SetReconsiler(mock.NewMockReconciler(nodeReconsiler, t.mockCentreonHandler))
		nodeReconsiler.SetPlatforms(platforms)
		if err = nodeReconsiler.SetupWithManager(k8sManager); err != nil {
			panic(err)
		}

		certificateReconsiler := NewCertificateReconciler(k8sClient, scheme.Scheme, templateController)
		certificateReconsiler.Reconciler.SetLogger(logrus.WithFields(logrus.Fields{
			"type": "certificateController",
		}))
		certificateReconsiler.SetRecorder(k8sManager.GetEventRecorderFor("certificate-controller"))
		certificateReconsiler.SetReconsiler(mock.NewMockReconciler(certificateReconsiler, t.mockCentreonHandler))
		certificateReconsiler.SetPlatforms(platforms)
		if err = certificateReconsiler.SetupWithManager(k8sManager); err != nil {
			panic(err)
		}
	*/

	go func() {
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		if err != nil {
			panic(err)
		}
	}()
}

func (t *CentreonControllerTestSuite) TearDownSuite() {
	err := testEnv.Stop()
	if err != nil {
		panic(err)
	}
}

func (t *CentreonControllerTestSuite) BeforeTest(suiteName, testName string) {

}

func (t *CentreonControllerTestSuite) AfterTest(suiteName, testName string) {
	defer t.mockCtrl.Finish()
}
