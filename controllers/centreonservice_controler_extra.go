package controllers

/*

import (
	"fmt"
	"strings"
	"time"

	"github.com/disaster37/go-centreon-rest/v21"
	"github.com/disaster37/go-centreon-rest/v21/models"
	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
)

type CentreonService interface {
	CreateOrUpdate(spec *v1alpha1.CentreonServiceSpec) (isUpdated bool, err error)
	Delete(spec *v1alpha1.CentreonServiceSpec) (err error)
	SetLogger(log *logrus.Entry)
}

type CentreonServiceImpl struct {
	client *centreon.Client
	log    *logrus.Entry
}

func NewCentreonService(client *centreon.Client, log *logrus.Entry) CentreonService {
	return &CentreonServiceImpl{
		client: client,
		log:    log,
	}
}

func (cs *CentreonServiceImpl) CreateOrUpdate(spec *v1alpha1.CentreonServiceSpec) (isChange bool, err error) {
	// Auth on Centreon
	if err = cs.client.API.Auth(); err != nil {
		cs.log.Error("Error when authenticate with Centreon")
		return isChange, err
	}

	// Check if service already exist on Centreon
	service, err := cs.getService(spec)
	if err != nil {
		return isChange, err
	}
	// Create process
	if service == nil {
		cs.log.Info("Service not yet exist on Centreon, create it")
		err = cs.create(spec)
		if err != nil {
			return isChange, err
		}
		return true, nil
	} else {
		cs.log.Info("Service already exist on Centreon, update it")
		return cs.update(spec, service)
	}

}

func (cs *CentreonServiceImpl) Delete(spec *v1alpha1.CentreonServiceSpec) (err error) {

	service, err := cs.getService(spec)
	if err != nil {
		return err
	}
	if service == nil {
		cs.log.Info("Service already deleted on Centreon, skip it")
		return nil
	}

	if err = cs.client.API.Service().Delete(spec.Host, spec.Name); err != nil {
		return err
	}

	cs.log.Info("Successfully delete service on Centreon")

	return nil

}

func (cs *CentreonServiceImpl) SetLogger(log *logrus.Entry) {
	cs.log = log
}

func (cs *CentreonServiceImpl) getService(spec *v1alpha1.CentreonServiceSpec) (service *models.ServiceGet, err error) {
	return cs.client.API.Service().Get(spec.Host, spec.Name)
}

func (cs *CentreonServiceImpl) create(spec *v1alpha1.CentreonServiceSpec) (err error) {

}

func (cs *CentreonServiceImpl) update(spec *v1alpha1.CentreonServiceSpec, service *models.ServiceGet) (isUpdated bool, err error) {

}

*/
