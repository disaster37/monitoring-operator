package helpers

import (
	"os"

	"github.com/disaster37/go-centreon-rest/v21/models"
	"github.com/pkg/errors"
)

func GetCentreonConfig() (cfg *models.Config, err error) {

	urlEnvVar := "MONITORING_URL"
	usernameEnvVar := "MONITORING_USERNAME"
	passwordEnvVar := "MONITORING_PASSWORD"
	disableSSLCheckEnvVar := "MONITORING_DISABLE_SSL_CHECK"

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
	disableSSLCheck := os.Getenv(disableSSLCheckEnvVar)

	cfg = &models.Config{
		Address:  url,
		Username: username,
		Password: password,
	}

	if disableSSLCheck == "true" {
		cfg.DisableVerifySSL = true
	} else {
		cfg.DisableVerifySSL = false
	}

	return cfg, nil

}