package helpers

import (
	"os"
	"time"

	"github.com/disaster37/go-centreon-rest/v21/models"
	"github.com/pkg/errors"
)

const (
	urlEnvVar               = "MONITORING_URL"
	usernameEnvVar          = "MONITORING_USERNAME"
	passwordEnvVar          = "MONITORING_PASSWORD"
	disableSSLCheckEnvVar   = "MONITORING_DISABLE_SSL_CHECK"
	monitoringTimeoutEnvVar = "MONITORING_CLIENT_TIMEOUT"
	operatorNamespaceEnvVar = "OPERATOR_NAMESPACE"
)

func GetCentreonConfig() (cfg *models.Config, err error) {

	url, found := os.LookupEnv(urlEnvVar)
	if !found {
		return nil, errors.Errorf("%s must be set", urlEnvVar)
	}
	username, found := os.LookupEnv(usernameEnvVar)
	if !found {
		return nil, errors.Errorf("%s must be set", usernameEnvVar)
	}
	password, found := os.LookupEnv(passwordEnvVar)
	if !found {
		return nil, errors.Errorf("%s must be set", passwordEnvVar)
	}

	cfg = &models.Config{
		Address:  url,
		Username: username,
		Password: password,
	}

	disableSSLCheck := os.Getenv(disableSSLCheckEnvVar)
	if disableSSLCheck == "true" {
		cfg.DisableVerifySSL = true
	} else {
		cfg.DisableVerifySSL = false
	}

	timeout, found := os.LookupEnv(monitoringTimeoutEnvVar)
	if found {
		d, err := time.ParseDuration(timeout)
		if err != nil {
			return nil, err
		}
		cfg.Timeout = d
	}

	return cfg, nil

}

func GetOperatorNamespace() (ns string, err error) {
	ns, found := os.LookupEnv(operatorNamespaceEnvVar)
	if !found {
		return "", errors.Errorf("%s must be set", operatorNamespaceEnvVar)
	}

	return ns, nil
}
