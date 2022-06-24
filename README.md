[![build](https://github.com/disaster37/monitoring-operator/actions/workflows/workflow.yaml/badge.svg)](https://github.com/disaster37/monitoring-operator/actions/workflows/workflow.yaml)
[![GoDoc](https://godoc.org/github.com/disaster37/monitoring-operator?status.svg)](http://godoc.org/github.com/disaster37/monitoring-operator)
[![Go Report Card](https://goreportcard.com/badge/github.com/disaster37/monitoring-operator)](https://goreportcard.com/report/github.com/disaster37/monitoring-operator)
[![codecov](https://codecov.io/gh/disaster37/monitoring-operator/branch/main/graph/badge.svg)](https://codecov.io/gh/disaster37/monitoring-operator/branch/main)

# monitoring-operator
Kubernetes operator to manage monitoring resources

It actually only support Centreon as monitoring plateform.

## Supported fonctionnalities:
  - Manage monitoring service from custom resource `CentreonService`
  - Auto create monitoring service from Ingress
  - Auto create monitoring service from Route

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
So, you need to provide a resource of type platfrom on same operator namespace

platform.yaml
```yaml
apiVersion: monitor.k8s.webcenter.fr/v1alpha1
kind: Platform
metadata:
  name: default
spec:
  name: default
  isDefault: true
  type: centreon
  centreonSettings:
    url: "http://localhost:9090/centreon/api/index.php"
    selfSignedCertificat: true
    secret: centreon
    endpoint:
      template: "check-http"
      nameTemplate: "App_<namespace>_URL"
      defaultHost: "kubernetes"
      macros:
        PROTOCOL: "<rule.0.scheme>"
        URLPATH: "<rule.0.path.0>"
      arguments:
        - "<rule.0.host>"
      activeService: true
      serviceGroups:
        - "SG1"
      categories:
        - "cat1"
```

Like you can see, you need to set credential to access on external monitoring API. The right way to do that on K8s is to use secret.
So, you need to create a new secret on same operator namespace with the name who are privided on platform.

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

> On Platform spec, there are a subsection call endpoint. It's a generic setting when you should to auto create monitoring service from Ingress or route spec. It avoid to provide each time, the wall setting by annotation.


You can use this global setting when you should to auto discover / monitor your ingress / Route:

```yaml
apiVersion: monitor.k8s.webcenter.fr/v1alpha1
kind: Platform
metadata:
  name: default
spec:
  name: default
  isDefault: true
  type: centreon
  centreonSettings:
    url: "http://localhost:9090/centreon/api/index.php"
    selfSignedCertificat: true
    secret: centreon
    endpoint:
      # It enable service when it create it
      # Optional
      activeService: true

      # The name template to use when it generate service name
      # You can use placeholders
      # Optional
      nameTemplate: App_<namespace>_URL

      # The service template to affect on service
      # Optional
      template: TS_App-Protocol-HTTP-MultiCheck

      # The default host to link service on it
      # Optional
      defaultHost: HOST_KUBERNETES_HM-HPD

      # The list service's arguments
      # You can use placeholders
      # Optional
      arguments:
      - <rule.0.host>

      # The list of service's macros
      # You can use placeholders
      # Optional
      macros:
        CRITICALCONTENT: '%{code} != 200 or ${code} != 401'
        PROTOCOL: <rule.0.scheme>
        URLPATH: <rule.0.path.0>
      
      # The list of service's groups
      # Optional
      serviceGroups:
      - SG_K8S_INGRESS

      # The list of service's categories
      # Optional
      categories:
      - Endpoint
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
  checkCommand:

  # The service's normal check interval
  # Optional
  normalCheckInterval:

  # The service's retry check interval
  # Optional
  retryCheckInterval:

  # The service's check attempts
  # Optional
  maxCheckAttempts:

  # It enable active check
  # Optional
  activeChecksEnabled:

  # It enable passive check
  # Optional
  passiveChecksEnabled:
```

> If you not provide spec key `platformRef`, it use the default platform.


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