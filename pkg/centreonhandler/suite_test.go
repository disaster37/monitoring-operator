package centreonhandler

import (
	"testing"

	"github.com/disaster37/go-centreon-rest/v21"
	"github.com/disaster37/go-centreon-rest/v21/mocks"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CentreonHandlerTestSuite struct {
	suite.Suite
	mockClient       *mocks.MockAPI
	mockService      *mocks.MockServiceAPI
	mockServiceGroup *mocks.MockServiceGroupAPI
	mockCtrl         *gomock.Controller
	client           CentreonHandler
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(CentreonHandlerTestSuite))
}

func (t *CentreonHandlerTestSuite) SetupSuite() {

	// Init Centreon mock
	t.mockCtrl = gomock.NewController(t.T())
	t.mockClient = mocks.NewMockAPI(t.mockCtrl)
	t.mockService = mocks.NewMockServiceAPI(t.mockCtrl)
	t.mockServiceGroup = mocks.NewMockServiceGroupAPI(t.mockCtrl)
	t.client = &CentreonHandlerImpl{
		client: &centreon.Client{
			API: t.mockClient,
		},
		log: logrus.NewEntry(logrus.New()),
	}

	logrus.SetLevel(logrus.DebugLevel)
}

func (t *CentreonHandlerTestSuite) BeforeTest(suiteName, testName string) {
	t.mockClient.EXPECT().Service().AnyTimes().Return(t.mockService)
	t.mockClient.EXPECT().ServiceGroup().AnyTimes().Return(t.mockServiceGroup)
	t.mockClient.EXPECT().Auth().AnyTimes().Return(nil)
}

func (t *CentreonHandlerTestSuite) AfterTest(suiteName, testName string) {
	defer t.mockCtrl.Finish()
}

func (t *CentreonHandlerTestSuite) TestSetLogger() {
	log := logrus.NewEntry(logrus.New())
	t.client.SetLogger(log)
}

func (t *CentreonHandlerTestSuite) TestNewCentreonHandler() {
	log := logrus.NewEntry(logrus.New())
	client := &centreon.Client{}
	ch := NewCentreonHandler(client, log)

	assert.Equal(t.T(), client, ch.(*CentreonHandlerImpl).client)
	assert.Equal(t.T(), log, ch.(*CentreonHandlerImpl).log)

}