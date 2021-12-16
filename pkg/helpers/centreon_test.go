package helpers

import (
	"os"
	"testing"

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
	}

	os.Setenv(urlEnvVar, "http://localhost")
	os.Setenv(usernameEnvVar, "user")
	os.Setenv(passwordEnvVar, "pass")
	os.Setenv(disableSSLCheckEnvVar, "true")

	cfg, err = GetCentreonConfig()
	assert.NoError(t, err)
	assert.Equal(t, expectedCFG, cfg)

	// When optionan en not setted
	expectedCFG = &models.Config{
		Address:  "http://localhost",
		Username: "user",
		Password: "pass",
	}
	os.Unsetenv(disableSSLCheckEnvVar)
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
