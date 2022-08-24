[![build](https://github.com/disaster37/monitoring-operator/actions/workflows/workflow.yaml/badge.svg)](https://github.com/disaster37/monitoring-operator/actions/workflows/workflow.yaml)
[![GoDoc](https://godoc.org/github.com/disaster37/monitoring-operator?status.svg)](http://godoc.org/github.com/disaster37/monitoring-operator)
[![Go Report Card](https://goreportcard.com/badge/github.com/disaster37/monitoring-operator)](https://goreportcard.com/report/github.com/disaster37/monitoring-operator)
[![codecov](https://codecov.io/gh/disaster37/monitoring-operator/branch/main/graph/badge.svg)](https://codecov.io/gh/disaster37/monitoring-operator/branch/main)

# monitoring-operator
Kubernetes operator to manage monitoring resources

It actually only support Centreon as monitoring plateform.

## Supported fonctionnalities:

- Manage service on Centreon from custom resource `CentreonService`
- Manage service group on Centreon from custom resource `CentreonServiceGroup`
- Auto create resources from `Ingress` with template concept
- Auto create resources from `Route` (Openshift) with template concept
- Auto create resources from `Namespace` with template concept
- Auto create resources from `Node` with template concept
- Auto create resources from `Certificate` with template concept

## Deploy operator with OLM

The right way to deploy operator base on operatot-sdk is to use OLM.
You can use the catalog image `webcenter/monitoring-operator-catalog:v0.0.1`

For test purpose, you can use operator-sdk to run bundle
```bash
operator-sdk run bundle docker.io/webcenter/monitoring-operator-bundle:v0.0.1
```

## Use it

### Platform

The first way consist to declare a platform. A platform is a monitoring API endpoint. actually, we only support Centreon platform.
So, you need to provide a resource of type platfrom on same operator namespace.

***platform.yaml***
```yaml
apiVersion: monitor.k8s.webcenter.fr/v1alpha1
kind: Platform
metadata:
  name: default
spec:
  isDefault: true
  type: centreon
  centreonSettings:
    url: "http://localhost:9090/centreon/api/index.php"
    selfSignedCertificat: true
    secret: centreon
```

Like you can see, you need to set credential to access on external monitoring API. The right way to do that on K8s is to use secret.
So, you need to create a new secret on same operator namespace with the name which are privided on platform.

secret.yaml
```yaml
apiVersion: v1
metadata:
  name: centreon
type: Opaque
data:
  username: bzQwcFdINy4zNnhs
  password: Y3MuY2xhcGk=
kind: Secret
``` 

### CentreonService

This custom resource permit to handle service on Centreon.

You can use this properties to set service:
```yaml
apiVersion: monitor.k8s.webcenter.fr/v1alpha1
kind: CentreonService
metadata:
  name: monitor-workloads
spec:
  # Optional
  # Target platform to create monitoring resource
  platformRef: default

  # Optional
  # It enable service
  activate: true
  
  # The host to link service on it
  host: HOST_KUBERNETES_HM-HPD

  # The service name
  name: App_Rancher_hm-hpd_logmanagement-rec_workloads

  # The service's template
  # Optional
  template: TS_App_Rancher

  # The service's macro
  # Optional
  macros:
    APIHOST: k8s.domain.com
    APITOKEN: my_token_secret
    APIUSER: my_token_access
    CHECKTYPE: workload
    EXTRAOPTIONS: -sS
    NAMESPACE: my-app

  # The service's arguments
  # Optional
  arguments:
  - arg1

  # The service's group
  # Optional
  groups:
  - SG_K8S_WORKLOAD

  # The service's categories
  # Optional
  categories:
  - Workload

  # The service's check command
  # Optional
  checkCommand: ""

  # The service's normal check interval
  # Optional
  normalCheckInterval: ""

  # The service's retry check interval
  # Optional
  retryCheckInterval: ""

  # The service's check attempts
  # Optional
  maxCheckAttempts: ""

  # It enable active check
  # Optional
  activeChecksEnabled: null

  # It enable passive check
  # Optional
  passiveChecksEnabled: null

  # Optional
  # The reconcil policy to use
  # Read the policy concept on documentation
  policy: null
```

> If you not provide spec key `platformRef`, it use the default platform.

When resource is created, you can get the following status:
  - **host**: the host where service is attached on Centreon
  - **serviceName**: the service name on Centreon
  - **conditions**: You can look the condition called `UpdateCentreonService` to know if Centreon service is update to date

### CentreonServiceGroup

This custom resource permit to handle service group on Centreon.

You can use this properties to set service group:
```yaml
apiVersion: monitor.k8s.webcenter.fr/v1alpha1
kind: CentreonServiceGroup
metadata:
  name: web-servers
spec:
  # Optional
  # Target platform to create monitoring resource
  platformRef: default

  # Optional
  # It enable service
  activate: true

  # The service group name
  name: SG_WEB_SERVER

  # Optional
  # The description
  description: "my web server"

  # Optional
  # The reconcil policy to use
  # Read the policy concept on documentation
  policy: null
```

> If you not provide spec key `platformRef`, it use the default platform.

When resource is created, you can get the following status:
  - **serviceGroupName**: the service group name on Centreon
  - **conditions**: You can look the condition called `UpdateCentreonServiceGroup` to know if Centreon service group is update to date


### Policy concept

@TODO

### Template concept

@TODO

### Namespace

If you need create Centreon service when you create namespace (like monitoring check base on prometheus to check namespace quota, resource usage, etc.), you can add annotation `monitor.k8s.webcenter.fr/templates` with content of json array
```json
[
  {
    "namespace": "monitoring-operator",
    "name": "check-namespace-quota"
  },
  {
    "namespace": "monitoring-operator",
    "name": "check-workloads-state"
  }
]
```

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: test
  annotations:
    monitor.k8s.webcenter.fr/templates: '[{"namespace":"monitoring-operator","name":"check-namespace-quota"},{"namespace":"monitoring-operator","name":"check-workloads-state"}]'
    monitoring-host: k8s-virtual-host


```

Then you need create associated template. The template is just the CentreonService spec with golang templating syntaxe (look placeholders available on templating).

```yaml
apiVersion: monitor.k8s.webcenter.fr/v1alpha1
kind: CentreonService
metadata:
  name: check-namespace-quota
  namespace: monitoring-operator
spec:
  template: |
    activate: true
    host: {{ .annotations.monitoring-host }}
    name: "check-namespace-quota_{{ .namespace }}"
    template: TS_prometheus_quota
    macros:
      CHECKTYPE: quota
      EXTRAOPTIONS: -sS
      NAMESPACE: {{ .namespace }}
```

```yaml
apiVersion: monitor.k8s.webcenter.fr/v1alpha1
kind: CentreonService
metadata:
  name: check-workloads-state
  namespace: monitoring-operator
spec:
  template: |
    activate: true
    host: {{ .annotations.monitoring-host }}
    name: "check-workloads-state_{{ .namespace }}"
    template: TS_prometheus_workloads
    macros:
      CHECKTYPE: workloads
      EXTRAOPTIONS: -sS
      NAMESPACE: {{ .namespace }}
```

Have fun, the two services was created on Centreon when namespace is created.

### List of annotations for Ingress / Route
**Global annotations:**
  - monitor.k8s.webcenter.fr/discover : true to watch resource

**Centreon annotations:**
  - centreon.monitor.k8s.webcenter.fr/name: the service name on centreon
  - centreon.monitor.k8s.webcenter.fr/template: the template name to affect on service on Centreon
  - centreon.monitor.k8s.webcenter.fr/host: the host to link with the service on Centreon
  - centreon.monitor.k8s.webcenter.fr/macros: the map of macros (as json form). It can placeolder value from tags
  -  - centreon.monitor.k8s.webcenter.fr/check-command: the check command on Centreon
  - centreon.monitor.k8s.webcenter.fr/arguments: the command arguments to set on service. Comma separated
  - centreon.monitor.k8s.webcenter.fr/activated: activate or disable the service on Centreon (0 or 1)
  - centreon.monitor.k8s.webcenter.fr/groups: The list of service groups linked with service on Centreon. Comma separated
  - centreon.monitor.k8s.webcenter.fr/categories: The list of categories linked with service on Centreon. Comma separated
  - centreon.monitor.k8s.webcenter.fr/normal-check-interval: 
  - centreon.monitor.k8s.webcenter.fr/retry-check-interval:
  - centreon.monitor.k8s.webcenter.fr/max-check-attempts:
  - centreon.monitor.k8s.webcenter.fr/active-check-enabled: (0, 1 or 2)
  - centreon.monitor.k8s.webcenter.fr/passive-check-enabled (0, 1 or 2)
  
 placeholders available for macros, arguments and nameTemplate from Ingress:
   - <rule.0.host> (the url)
   - <rule.0.scheme> (http or https)
   - <rule.0.path.0> (the path)
   - <name> ingress name
   - <namespace>: ingress namespace
   - <label.key>: labels
   - <annotation.key>: annotations

placeholders available for macros, arguments and nameTemplate from Route:
   - <rule.0.host> : the url
   - <rule.0.scheme>: the scheme - http or https
   - <rule.0.path>: the path
   - <name>: route name
   - <namespace>: route namespace
   - <label.key>: labels
   - <annotation.key>: annotations

## Deploy Centreon for test purpose

If you haven't Centreon ready, and you should to test operator, you can deploy it (only for quick test):

```bash
kubectl apply -n default -f sample/centreon/
```

Get the public port
```bash
kubectl get svc/centreon-test -n default -o go-template='{{range .spec.ports}}{{if .nodePort}}{{.nodePort}}{{"\n"}}{{end}}{{end}}'
```

You can now access on centreon with this port.
> Login are admin / admin

When you have finhished your test:
```bash
kubectl delete -n default -f sample/centreon/
```