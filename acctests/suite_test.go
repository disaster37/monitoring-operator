package acctests

import (
	"fmt"
	"os"
	"testing"

	"github.com/disaster37/go-centreon-rest/v21"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type AccTestSuite struct {
	suite.Suite
	k8sclient    dynamic.Interface
	k8sclientStd kubernetes.Interface
	centreon     centreonhandler.CentreonHandler
	config       *rest.Config
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(AccTestSuite))
}

func (t *AccTestSuite) SetupSuite() {
	logrus.SetLevel(logrus.DebugLevel)
	log := logrus.New()

	// Init k8s client
	homePath, err := os.UserHomeDir()
	if err != nil {
		log.Warnf("Can't get home directory: %s", err.Error())
		homePath = "/root"
	}
	config, err := clientcmd.BuildConfigFromFlags("", fmt.Sprintf("%s/.kube/config", homePath))
	if err != nil {
		panic(err)
	}
	t.config = config
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	t.k8sclient = client
	clientStd, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	t.k8sclientStd = clientStd

	// Init Centreon client
	cfg, err := helpers.GetCentreonConfig()
	if err != nil {
		panic(err)
	}
	centreonClient, err := centreon.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	t.centreon = centreonhandler.NewCentreonHandler(centreonClient, logrus.NewEntry(log))

}

func (t *AccTestSuite) BeforeTest(suiteName, testName string) {
}

func (t *AccTestSuite) AfterTest(suiteName, testName string) {
}
