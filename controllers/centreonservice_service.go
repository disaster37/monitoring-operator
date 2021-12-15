package controllers

import (
	"strings"

	"github.com/disaster37/go-centreon-rest/v21/models"
	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/sirupsen/logrus"
)

type CentreonService interface {
	Reconcile(centreonService *v1alpha1.CentreonService) (isCreate, isUpdate bool, err error)
	Delete(centreonService *v1alpha1.CentreonService) (err error)
	SetLogger(log *logrus.Entry)
}

type CentreonServiceImpl struct {
	client centreonhandler.CentreonHandler
	log    *logrus.Entry
}

func NewCentreonService(client centreonhandler.CentreonHandler) CentreonService {
	return &CentreonServiceImpl{
		client: client,
		log:    logrus.NewEntry(logrus.New()),
	}
}

func (cs *CentreonServiceImpl) SetLogger(log *logrus.Entry) {
	cs.log = log
}

func (cs *CentreonServiceImpl) Delete(centreonService *v1alpha1.CentreonService) (err error) {
	actualCS, err := cs.client.GetService(centreonService.Spec.Host, centreonService.Spec.Name)
	if err != nil {
		return err
	}

	if actualCS == nil {
		cs.log.Info("Service already deleted on Centreon by external process, skip it")
		return nil
	}

	return cs.client.DeleteService(centreonService.Spec.Host, centreonService.Spec.Name)
}

func (cs *CentreonServiceImpl) Reconcile(centreonService *v1alpha1.CentreonService) (isCreate, isUpdate bool, err error) {
	actualCS, err := cs.client.GetService(centreonService.Spec.Host, centreonService.Spec.Name)
	if err != nil {
		return false, false, err
	}

	expectedCS := centreonhandler.CentreonService{
		Host:                centreonService.Spec.Host,
		Name:                centreonService.Spec.Name,
		CheckCommand:        centreonService.Spec.CheckCommand,
		CheckCommandArgs:    helpers.CheckArgumentsToString(centreonService.Spec.Arguments),
		NormalCheckInterval: centreonService.Spec.NormalCheckInterval,
		RetryCheckInterval:  centreonService.Spec.RetryCheckInterval,
		MaxCheckAttempts:    centreonService.Spec.MaxCheckAttempts,
		ActiveCheckEnabled:  helpers.BoolToString(centreonService.Spec.ActiveCheckEnabled),
		PassiveCheckEnabled: helpers.BoolToString(centreonService.Spec.PassiveCheckEnabled),
		Activated:           helpers.BoolToString(&centreonService.Spec.Activated),
		Template:            centreonService.Spec.Template,
		Comment:             "Managed by monitoring-operator",
		Groups:              centreonService.Spec.Groups,
		Categories:          centreonService.Spec.Categories,
		Macros:              make([]*models.Macro, 0, len(centreonService.Spec.Macros)),
	}
	for name, value := range centreonService.Spec.Macros {
		macro := &models.Macro{
			Name:       strings.ToUpper(name),
			Value:      value,
			IsPassword: "0",
		}
		expectedCS.Macros = append(expectedCS.Macros, macro)
	}

	// Create
	if actualCS == nil {
		err := cs.client.CreateService(&expectedCS)
		return true, false, err
	}

	diff, err := cs.client.DiffService(actualCS, &expectedCS)
	if err != nil {
		return false, false, err
	}

	// Update
	if diff.IsDiff {
		err := cs.client.UpdateService(diff)
		if err != nil {
			return false, false, err
		}
		return false, true, nil
	}

	// Already all right
	return false, false, nil
}
