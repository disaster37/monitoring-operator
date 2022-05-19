package controllers

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/disaster37/monitoring-operator/pkg/mocks"
	"github.com/disaster37/operator-sdk-extra/pkg/mock"
	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega/gexec"
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

	monitorv1alpha1 "github.com/disaster37/monitoring-operator/api/v1alpha1"
	//+kubebuilder:scaffold:imports
)

var testEnv *envtest.Environment

type ControllerTestSuite struct {
	suite.Suite
	k8sClient           client.Client
	mockCentreonHandler *mocks.MockCentreonHandler
	mockCtrl            *gomock.Controller
	cfg                 *rest.Config
}

func TestControllerSuite(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

func (t *ControllerTestSuite) SetupSuite() {
	// Init Centreon mock
	t.mockCtrl = gomock.NewController(t.T())
	t.mockCentreonHandler = mocks.NewMockCentreonHandler(t.mockCtrl)

	logf.SetLogger(zap.New(zap.UseDevMode(true)))
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableQuote: true,
	})

	// Setup testenv
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "config", "crd", "bases"),
			filepath.Join("..", "config", "crd", "openshift"),
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
	err = monitorv1alpha1.AddToScheme(scheme.Scheme)
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

	// Init controlles
	centreonServiceReconsiler := &CentreonServiceReconciler{
		Client: k8sClient,
		Scheme: scheme.Scheme,
	}
	centreonServiceReconsiler.SetLogger(logrus.WithFields(logrus.Fields{
		"type": "centreonServiceController",
	}))
	centreonServiceReconsiler.SetRecorder(k8sManager.GetEventRecorderFor("centreonservice-controller"))
	centreonServiceReconsiler.SetReconsiler(mock.NewMockReconciler(centreonServiceReconsiler, t.mockCentreonHandler))
	if err = centreonServiceReconsiler.SetupWithManager(k8sManager); err != nil {
		panic(err)
	}

	ingressReconsiler := &IngressCentreonReconciler{
		Client: k8sClient,
		Scheme: scheme.Scheme,
	}
	ingressReconsiler.SetLogger(logrus.WithFields(logrus.Fields{
		"type": "ingressCentreonController",
	}))
	ingressReconsiler.SetRecorder(k8sManager.GetEventRecorderFor("ingresscentreon-controller"))
	ingressReconsiler.SetReconsiler(mock.NewMockReconciler(ingressReconsiler, t.mockCentreonHandler))
	if err = ingressReconsiler.SetupWithManager(k8sManager); err != nil {
		panic(err)
	}

	routeReconsiler := &RouteCentreonReconciler{
		Client: k8sClient,
		Scheme: scheme.Scheme,
	}
	routeReconsiler.SetLogger(logrus.WithFields(logrus.Fields{
		"type": "routeCentreonController",
	}))
	routeReconsiler.SetRecorder(k8sManager.GetEventRecorderFor("routecentreon-controller"))
	routeReconsiler.SetReconsiler(mock.NewMockReconciler(routeReconsiler, t.mockCentreonHandler))
	if err = routeReconsiler.SetupWithManager(k8sManager); err != nil {
		panic(err)
	}

	go func() {
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		if err != nil {
			panic(err)
		}
	}()
}

func (t *ControllerTestSuite) TearDownSuite() {
	gexec.KillAndWait(5 * time.Second)

	// Teardown the test environment once controller is fnished.
	// Otherwise from Kubernetes 1.21+, teardon timeouts waiting on
	// kube-apiserver to return
	err := testEnv.Stop()
	if err != nil {
		panic(err)
	}
}

func (t *ControllerTestSuite) BeforeTest(suiteName, testName string) {
}

func (t *ControllerTestSuite) AfterTest(suiteName, testName string) {
	defer t.mockCtrl.Finish()
}

func RunWithTimeout(f func() error, timeout time.Duration, interval time.Duration) (isTimeout bool, err error) {
	control := make(chan bool)
	timeoutTimer := time.NewTimer(timeout)
	go func() {
		loop := true
		intervalTimer := time.NewTimer(interval)
		for loop {
			select {
			case <-control:
				return
			case <-intervalTimer.C:
				err = f()
				if err != nil {
					intervalTimer.Reset(interval)
				} else {
					loop = false
				}
			}
		}
		control <- true
		return
	}()

	select {
	case <-control:
		return false, nil
	case <-timeoutTimer.C:
		control <- true
		return true, err
	}
}
