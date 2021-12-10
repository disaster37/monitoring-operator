package main

import (
	"os"
	osruntime "runtime"

	"github.com/disaster37/go-centreon-rest/v21/models"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap/zapcore"
)

func getZapLogLevel() zapcore.Level {
	switch logLevel, _ := os.LookupEnv("LOG_LEVEL"); logLevel {
	case zapcore.DebugLevel.String():
		return zapcore.DebugLevel
	case zapcore.InfoLevel.String():
		return zapcore.InfoLevel
	case zapcore.WarnLevel.String():
		return zapcore.WarnLevel
	case zapcore.ErrorLevel.String():
		return zapcore.ErrorLevel
	case zapcore.PanicLevel.String():
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}

func getLogrusLogLevel() logrus.Level {
	switch logLevel, _ := os.LookupEnv("LOG_LEVEL"); logLevel {
	case logrus.DebugLevel.String():
		return logrus.DebugLevel
	case logrus.InfoLevel.String():
		return logrus.InfoLevel
	case logrus.WarnLevel.String():
		return logrus.WarnLevel
	case logrus.ErrorLevel.String():
		return logrus.ErrorLevel
	case logrus.PanicLevel.String():
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel
	}
}

func printVersion(logger logr.Logger, metricsAddr, probeAddr string) {
	logger.Info("Binary info ", "Go version", osruntime.Version())
	logger.Info("Binary info ", "OS", osruntime.GOOS, "Arch", osruntime.GOARCH)
	logger.Info("Address ", "Metrics", metricsAddr)
	logger.Info("Address ", "Probe", probeAddr)
}

func getWatchNamespace() (ns string, err error) {

	watchNamespaceEnvVar := "WATCH_NAMESPACE"
	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		return "", errors.Errorf("%s must be set", watchNamespaceEnvVar)
	}

	return ns, nil
}

func getConfigNamespace() (ns string, err error) {
	configNamespaceEnvVar := "CONFIG_NAMESPACE"
	ns, found := os.LookupEnv(configNamespaceEnvVar)
	if !found {
		return "", errors.Errorf("%s must be set", configNamespaceEnvVar)
	}

	return ns, nil
}

func getCentreonConfig() (cfg *models.Config, err error) {

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
