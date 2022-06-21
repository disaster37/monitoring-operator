package helpers

import (
	"os"

	"github.com/pkg/errors"
)

const (
	operatorNamespaceEnvVar = "OPERATOR_NAMESPACE"
)

func GetOperatorNamespace() (ns string, err error) {
	ns, found := os.LookupEnv(operatorNamespaceEnvVar)
	if !found {
		return "", errors.Errorf("%s must be set", operatorNamespaceEnvVar)
	}

	return ns, nil
}
