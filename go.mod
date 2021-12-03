module github.com/disaster37/monitoring-operator

go 1.16

require (
	github.com/disaster37/go-centreon-rest/v21 v21.0.1
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/sirupsen/logrus v1.8.1
	github.com/thoas/go-funk v0.9.1 // indirect
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	sigs.k8s.io/controller-runtime v0.10.0
)
