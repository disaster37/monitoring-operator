package controllers

import (
	"context"
	"fmt"

	"github.com/disaster37/go-centreon-rest/v21"
	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/sirupsen/logrus"
)

var ErrorCentreonNotFound = fmt.Errorf("Centreon not found")
var ErrorCentreonMultipleFound = fmt.Errorf("More than one Centreon found")

func (r *CentreonServiceReconciler) newCentreonClient(ctx context.Context, log *logrus.Entry) (client *centreon.Client, err error) {

	// Get Centreon object
	centreons := &v1alpha1.CentreonList{}
	if err := r.List(ctx, centreons); err != nil {
		return nil, err
	}
	if len(centreons.Items) == 0 {
		return nil, ErrorCentreonNotFound
	}
	if (len(centreons.Items)) > 1 {
		log.Warn("Founf multiple centreon object:")
		for _, centreon := range centreons.Items {
			log.Warnf("%s/%s", centreon.Namespace, centreon.Name)
		}
		return nil, ErrorCentreonMultipleFound
	}
	c := centreons.Items[0]
	log.Infof("Found Centreon %s/%s", c.Namespace, c.Name)

	// Init client
	cfg := centreon.Config{
		Address:          c.Spec.Url,
		Username:         c.Spec.Username,
		Password:         c.Spec.Password,
		DisableVerifySSL: c.Spec.SelfSignedCertificate,
	}
	return centreon.NewClient(cfg)
}
