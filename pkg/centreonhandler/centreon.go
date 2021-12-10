package centreonhandler

import (
	"github.com/disaster37/go-centreon-rest/v21"
	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/sirupsen/logrus"
)

type CentreonHandler interface {
	CreateService(spec *v1alpha1.CentreonServiceSpec) (err error)
	UpdateService(spec *v1alpha1.CentreonServiceSpec) (err error)
	DeleteService(spec *v1alpha1.CentreonServiceSpec) (err error)
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
