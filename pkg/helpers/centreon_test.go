package helpers

import (
	"os"
	"testing"
	"time"

	"github.com/disaster37/go-centreon-rest/v21/models"
	"github.com/stretchr/testify/assert"
)

func TestGetCentreonConfig(t *testing.T) {
	var (
		cfg         *models.Config
		expectedCFG *models.Config
		err         error
	)

	// When all environment variable are successfully set
	expectedCFG = &models.Config{
		Address:          "http://localhost",
		Username:         "user",
		Password:         "pass",
		DisableVerifySSL: true,
		Timeout:          10 * time.Second,
	}

	os.Setenv(urlEnvVar, "http://localhost")
	os.Setenv(usernameEnvVar, "user")
	os.Setenv(passwordEnvVar, "pass")
	os.Setenv(disableSSLCheckEnvVar, "true")
	os.Setenv(monitoringTimeoutEnvVar, "10s")

	cfg, err = GetCentreonConfig()
	assert.NoError(t, err)
	assert.Equal(t, expectedCFG, cfg)

	// When optionnal not setted
	expectedCFG = &models.Config{
		Address:  "http://localhost",
		Username: "user",
		Password: "pass",
	}
	os.Unsetenv(disableSSLCheckEnvVar)
	os.Unsetenv(monitoringTimeoutEnvVar)
	cfg, err = GetCentreonConfig()
	assert.NoError(t, err)
	assert.Equal(t, expectedCFG, cfg)

	// When required env not set
	os.Unsetenv(urlEnvVar)
	cfg, err = GetCentreonConfig()
	assert.Error(t, err)
	os.Setenv(urlEnvVar, "http://localhost")

	os.Unsetenv(usernameEnvVar)
	cfg, err = GetCentreonConfig()
	assert.Error(t, err)
	os.Setenv(usernameEnvVar, "user")

	os.Unsetenv(passwordEnvVar)
	cfg, err = GetCentreonConfig()
	assert.Error(t, err)
	os.Setenv(passwordEnvVar, "pass")
}

func TestGetOperatorNamespace(t *testing.T) {
	os.Setenv(operatorNamespaceEnvVar, "test")
	ns, err := GetOperatorNamespace()
	assert.NoError(t, err)
	assert.Equal(t, "test", ns)

	os.Unsetenv(operatorNamespaceEnvVar)
	ns, err = GetOperatorNamespace()
	assert.Error(t, err)
}
