package helpers

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOperatorNamespace(t *testing.T) {
	_ = os.Setenv(operatorNamespaceEnvVar, "test")
	ns, err := GetOperatorNamespace()
	assert.NoError(t, err)
	assert.Equal(t, "test", ns)

	_ = os.Unsetenv(operatorNamespaceEnvVar)
	_, err = GetOperatorNamespace()
	assert.Error(t, err)
}
