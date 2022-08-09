package centreonhandler

import (
	"github.com/disaster37/go-centreon-rest/v21"
	"github.com/sirupsen/logrus"
)

type CentreonHandler interface {
	CreateService(service *CentreonService) (err error)
	UpdateService(service *CentreonServiceDiff) (err error)
	DeleteService(host, service string) (err error)
	GetService(host, name string) (service *CentreonService, err error)
	DiffService(actual, expected *CentreonService) (diff *CentreonServiceDiff, err error)
	CreateServiceGroup(sg *CentreonServiceGroup) (err error)
	UpdateServiceGroup(sg *CentreonServiceGroupDiff) (err error)
	DeleteServiceGroup(name string) (err error)
	GetServiceGroup(name string) (sg *CentreonServiceGroup, err error)
	DiffServiceGroup(actual, expected *CentreonServiceGroup) (diff *CentreonServiceGroupDiff, err error)

	SetLogger(log *logrus.Entry)
}

type CentreonHandlerImpl struct {
	client *centreon.Client
	log    *logrus.Entry
}

func NewCentreonHandler(client *centreon.Client, log *logrus.Entry) CentreonHandler {
	return &CentreonHandlerImpl{
		client: client,
		log:    log,
	}
}

func (h *CentreonHandlerImpl) SetLogger(log *logrus.Entry) {
	h.log = log
}
