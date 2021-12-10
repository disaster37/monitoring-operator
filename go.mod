module github.com/disaster37/monitoring-operator

go 1.16

require (
	github.com/disaster37/go-centreon-rest/v21 v21.0.2
	github.com/go-logr/logr v0.4.0
	github.com/golang/mock v1.6.0
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.17.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/thoas/go-funk v0.9.1
	go.uber.org/zap v1.19.0
	golang.org/x/net v0.0.0-20211029224645-99673261e6eb
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	sigs.k8s.io/controller-runtime v0.10.3
)
