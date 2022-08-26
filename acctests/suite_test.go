package acctests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/disaster37/go-centreon-rest/v21"
	"github.com/disaster37/go-centreon-rest/v21/models"
	monitorapi "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	centreonURL      = "http://localhost:9090/centreon/api/index.php"
	centreonUsername = "admin"
	centreonPassword = "admin"
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

	// Create new monitoring platform
	secret := &core.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name: "centreon",
		},
		Data: map[string][]byte{
			"username": []byte(centreonUsername),
			"password": []byte(centreonPassword),
		},
	}
	if _, err := t.k8sclientStd.CoreV1().Secrets("default").Create(context.Background(), secret, v1.CreateOptions{}); err != nil {
		panic(err)
	}

	platformGVR := monitorapi.GroupVersion.WithResource("platforms")
	p := &monitorapi.Platform{
		TypeMeta: v1.TypeMeta{
			Kind:       "Platform",
			APIVersion: "monitor.k8s.webcenter.fr/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "default",
		},
		Spec: monitorapi.PlatformSpec{
			IsDefault:    true,
			PlatformType: "centreon",
			CentreonSettings: &monitorapi.PlatformSpecCentreonSettings{
				URL:                   centreonURL,
				SelfSignedCertificate: true,
				Secret:                "centreon",
			},
		},
	}

	ucs, err := structuredToUntructured(p)
	if err != nil {
		panic(err)
	}

	_, err = t.k8sclient.Resource(platformGVR).Namespace("default").Create(context.Background(), ucs, v1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	time.Sleep(20 * time.Second)

	// Init Centreon client
	cfg := &models.Config{
		Address:          centreonURL,
		Username:         centreonUsername,
		Password:         centreonPassword,
		DisableVerifySSL: true,
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
