package shared

import (
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// The logrus instance to print some logs from webhook
var Logger *logrus.Entry = logrus.NewEntry(logrus.New())

// The kube client used from webhook to check some other resource on kube
var Client client.Client
