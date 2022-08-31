package controllers

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/disaster37/operator-sdk-extra/pkg/mock"
	"github.com/golang/mock/gomock"
	"github.com/onsi/gomega/gexec"
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

	"github.com/disaster37/monitoring-operator/pkg/mocks"

	monitorapi "github.com/disaster37/monitoring-operator/api/v1"
	//+kubebuilder:scaffold:imports
)

var testEnv *envtest.Environment

type ControllerTestSuite struct {
	suite.Suite
	k8sClient           client.Client
	mockCentreonHandler *mocks.MockCentreonHandler
	mockCtrl            *gomock.Controller
	cfg                 *rest.Config
	platforms           map[string]*ComputedPlatform
}

func TestControllerSuite(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

func (t *ControllerTestSuite) SetupSuite() {
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
	err = monitorapi.AddToScheme(scheme.Scheme)
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
	if err := k8sManager.GetFieldIndexer().IndexField(context.Background(), &monitorapi.Platform{}, "spec.centreonSettings.secret", func(o client.Object) []string {
		p := o.(*monitorapi.Platform)
		return []string{p.Spec.CentreonSettings.Secret}
	}); err != nil {
		panic(err)
	}

	// Init controllers
	os.Setenv("OPERATOR_NAMESPACE", "default")

	platforms := map[string]*ComputedPlatform{
		"default": {
			platform: &monitorapi.Platform{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default",
					Namespace: "default",
				},
				Spec: monitorapi.PlatformSpec{
					IsDefault:        true,
					PlatformType:     "centreon",
					CentreonSettings: &monitorapi.PlatformSpecCentreonSettings{},
				},
			},
			client: t.mockCentreonHandler,
		},
	}
	t.platforms = platforms
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

	centreonServiceReconsiler := NewCentreonServiceReconciler(k8sClient, scheme.Scheme)
	centreonServiceReconsiler.SetLogger(logrus.WithFields(logrus.Fields{
		"type": "centreonServiceController",
	}))
	centreonServiceReconsiler.SetRecorder(k8sManager.GetEventRecorderFor("centreonservice-controller"))
	centreonServiceReconsiler.SetReconsiler(mock.NewMockReconciler(centreonServiceReconsiler, t.mockCentreonHandler))
	centreonServiceReconsiler.SetPlatforms(platforms)
	if err = centreonServiceReconsiler.SetupWithManager(k8sManager); err != nil {
		panic(err)
	}

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
	}()

	select {
	case <-control:
		return false, nil
	case <-timeoutTimer.C:
		control <- true
		return true, err
	}
}
