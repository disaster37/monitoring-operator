package helpers

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOperatorNamespace(t *testing.T) {
	os.Setenv(operatorNamespaceEnvVar, "test")
	ns, err := GetOperatorNamespace()
	assert.NoError(t, err)
	assert.Equal(t, "test", ns)

	os.Unsetenv(operatorNamespaceEnvVar)
	_, err = GetOperatorNamespace()
	assert.Error(t, err)
}
