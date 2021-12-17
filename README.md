[![build](https://github.com/disaster37/monitoring-operator/actions/workflows/workflow.yaml/badge.svg)](https://github.com/disaster37/monitoring-operator/actions/workflows/workflow.yaml)
[![GoDoc](https://godoc.org/github.com/disaster37/monitoring-operator?status.svg)](http://godoc.org/github.com/disaster37/monitoring-operator)
[![codecov](https://codecov.io/gh/disaster37/monitoring-operator/branch/main/graph/badge.svg)](https://codecov.io/gh/disaster37/monitoring-operator/branch/main)

# monitoring-operator
Kubernetes operator to manage monitoring resources

## Notes
List des annotations:
  - monitor.k8s.webcenter.fr/discover : true to watch resource
  - centreon.monitor.k8s.webcenter.fr/name: the service name on centreon
  - centreon.monitor.k8s.webcenter.fr/template: the template name to affect on service on Centreon
  - centreon.monitor.k8s.webcenter.fr/host: the host to link with the service on Centreon
  - centreon.monitor.k8s.webcenter.fr/macros: the map of macros (as json form). It can placeolder value from tags
  - centreon.monitor.k8s.webcenter.fr/arguments: the command arguments to set on service. Comma separated
  - centreon.monitor.k8s.webcenter.fr/activated: activate or disable the service on Centreon (0 or 1)
  - centreon.monitor.k8s.webcenter.fr/groups: The list of service groups linked with service on Centreon. Comma separated
  - centreon.monitor.k8s.webcenter.fr/categories: The list of categories linked with service on Centreon. Comma separated
  - centreon.monitor.k8s.webcenter.fr/normal-check-interval: 
  - centreon.monitor.k8s.webcenter.fr/retry-check-interval:
  - centreon.monitor.k8s.webcenter.fr/max-check-attempts:
  - centreon.monitor.k8s.webcenter.fr/active-check-enabled: (0, 1 or 2)
  - centreon.monitor.k8s.webcenter.fr/passive-check-enabled (0, 1 or 2)
  
 placeholders available:
   - <rule.0.host> (the url)
   - <rule.0.scheme> (http or https)
   - <rule.0.path.0> (the path)
   - <name> ingress name
   - <namespace> ingress namespace
   - <label.key>
   - <annotation.key>

## Initialise project for memory

1. Create new project
```bash
operator-sdk init --domain=k8s.webcenter.fr --repo=github.com/disaster37/monitoring-operator
```

2. Create APIs

To manage Centreon configs
```bash
operator-sdk create api --group monitor --version v1alpha1 --kind Centreon --resource
```

To manage service on Centreon
```bash
operator-sdk create api --group monitor --version v1alpha1 --kind CentreonService --resource --controller
```

To watch ingress
```bash
operator-sdk create api --group networking.k8s.io --version v1 --kind Ingress --controller
```

3. Change some default value

Edit the folliwng files:
- `config/default/kustomization.yaml`
- `config/samples/monitor_v1alpha1_centreon.yaml`
- `config/samples/monitor_v1alpha1_centreonservice.yaml`
- `config/samples/networking.k8s.io_v1_ingress.yaml`

4. Generate some Go codes like controllers

> Need each time you change `*_types.go`

```bash
make generate
```

5. Generate CRDs

> Need each time you change `*_types.go` or add some comment annotations in controllers

```bash
make manifests
```
6. Unit test

```bash
go get github.com/onsi/ginkgo/ginkgo

```

7. Test

```bash
docker run --name centreon -d -t --privileged -p 80:80 disaster/centreon:21.10-installed 

export KUBECONFIG=/home/theia/.kube/config
make install run
make instal-sample
```

https://github.com/Azure/azure-databricks-operator/blob/0f722a710fea06b86ecdccd9455336ca712bf775/controllers/secretscope_controller.go

https://redhat-scholars.github.io/operators-sdk-tutorial/template-tutorial/index.html

Quand on crée des sous item dans kube, il faut ajouter dans les metas data le OwnerReference

SI on veut surveiller des sous items, il faut créer un watcher sur les sous ressources (les types)