package centreon

import (
	"context"
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

	"github.com/disaster37/monitoring-operator/internal/controller/platform"
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
	platforms           map[string]*platform.ComputedPlatform
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

	platforms := map[string]*platform.ComputedPlatform{
		"default": {
			Platform: &centreoncrd.Platform{
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
			Client: t.mockCentreonHandler,
		},
	}
	t.platforms = platforms

	centreonServiceReconsiler := NewCentreonServiceReconciler(
		k8sClient,
		logrus.NewEntry(logrus.StandardLogger()),
		k8sManager.GetEventRecorderFor("centreonservice-controller"),
		t.platforms,
	)
	centreonServiceReconsiler.(*CentreonServiceReconciler).RemoteReconcilerAction = mock.NewMockRemoteReconcilerAction[*centreoncrd.CentreonService, *CentreonService, centreonhandler.CentreonHandler](
		centreonServiceReconsiler.(*CentreonServiceReconciler).RemoteReconcilerAction,
		func(ctx context.Context, req reconcile.Request, o object.RemoteObject, logger *logrus.Entry) (handler controller.RemoteExternalReconciler[*centreoncrd.CentreonService, *CentreonService, centreonhandler.CentreonHandler], res reconcile.Result, err error) {
			return newCentreonServiceApiClient(t.mockCentreonHandler, logger), res, nil
		},
	)
	if err = centreonServiceReconsiler.SetupWithManager(k8sManager); err != nil {
		panic(err)
	}

	centreonServiceGroupReconsiler := NewCentreonServiceGroupReconciler(
		k8sClient,
		logrus.NewEntry(logrus.StandardLogger()),
		k8sManager.GetEventRecorderFor("centreonservicegroup-controller"),
		t.platforms,
	)
	centreonServiceGroupReconsiler.(*CentreonServiceGroupReconciler).RemoteReconcilerAction = mock.NewMockRemoteReconcilerAction[*centreoncrd.CentreonServiceGroup, *CentreonServiceGroup, centreonhandler.CentreonHandler](
		centreonServiceGroupReconsiler.(*CentreonServiceGroupReconciler).RemoteReconcilerAction,
		func(ctx context.Context, req reconcile.Request, o object.RemoteObject, logger *logrus.Entry) (handler controller.RemoteExternalReconciler[*centreoncrd.CentreonServiceGroup, *CentreonServiceGroup, centreonhandler.CentreonHandler], res reconcile.Result, err error) {
			return newCentreonServiceGroupApiClient(t.mockCentreonHandler, logger), res, nil
		},
	)
	if err = centreonServiceGroupReconsiler.SetupWithManager(k8sManager); err != nil {
		panic(err)
	}

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
