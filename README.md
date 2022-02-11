[![build](https://github.com/disaster37/monitoring-operator/actions/workflows/workflow.yaml/badge.svg)](https://github.com/disaster37/monitoring-operator/actions/workflows/workflow.yaml)
[![GoDoc](https://godoc.org/github.com/disaster37/monitoring-operator?status.svg)](http://godoc.org/github.com/disaster37/monitoring-operator)
[![Go Report Card](https://goreportcard.com/badge/github.com/disaster37/monitoring-operator)](https://goreportcard.com/report/github.com/disaster37/monitoring-operator)
[![codecov](https://codecov.io/gh/disaster37/monitoring-operator/branch/main/graph/badge.svg)](https://codecov.io/gh/disaster37/monitoring-operator/branch/main)

# monitoring-operator
Kubernetes operator to manage monitoring resources

It actually only support Centreon as monitoring plateform.

## Supported fonctionnalities:
  - Manage service from custom resource `CentreonService`
  - Manage service from ingress annotations and custom resource `Centreon` (kind of global setting)
  - Manage service from routes annotations and custom resource `Centreon` (kind of global setting)

## Deploy operator with helm

The project provide 2 helm chart:
  - monitoring-operator-crds: it only contain the CRD needed by operator
  - monitoring-operator: it deploy operator as deployement

You can deploy the 2 separately, if you not have full right on cluster (CRD is installed by admin cluster). Or only deploy `monitoring-operator` and ask to install CRD on values.yaml.

1. Create custome `values.yaml`

Sample of values.yaml:
```yaml
---
replicaCount: 1
installCRDs: true
config:
  loglevel: info
monitoring:
  secret:
    name: 'monitoring-operator-credentials'
  plateform: 'centreon'
  url: 'https://centreon.domain.com/centreon/api/index.php'
  disableSSLCheck: true

centreon:
  endpoint:
    template: 'TS_App-Protocol-HTTP-MultiCheck'
    nameTemplate: 'App_<namespace>_URL'
    defaultHost: 'HOST_KUBERNETES'
    activeService: true
    macros:
      PROTOCOL: '<rule.0.scheme>'
      CRITICALCONTENT: '%{code} != 200 or ${code} != 401'
      URLPATH: '<rule.0.path.0>'
    arguments:
      - '<rule.0.host>'
```

2. Create secret with credantial to access on Centreon API

> The section of `centreon.endpoint` permit to set some global setting when you shoud monitor ingress automatically. It avoid to set all annotations on each ingress.

And secret specified here with name `monitoring-operator-credentials`, need to have credential to access on Centreon:

```yaml
apiVersion: v1
metadata:
  name: monitoring-operator-credentials
type: Opaque
data:
  MONITORING_PASSWORD: bzQwcFdINy4zNnhs
  MONITORING_USERNAME: Y3MuY2xhcGk=
kind: Secret
```

3. Deploy with helm

Add repository
```bash
helm repo add webcenter https://charts.webcenter.fr
```

Install chart
```bash
helm install monitoring-operator webcenter/monitoring-operator --version 0.0.1
```

## Custom ressources

### Centreon

This custom resource permit to set global setting used by operator.


#### Global setting for endpoints

You can use this global setting when you should to auto discover / monitor your ingress / Route:

```yaml
apiVersion: monitor.k8s.webcenter.fr/v1alpha1
kind: Centreon
metadata:
  name: monitoring-operator
spec:
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

> It avoid to set each time all annotations in Ingress / Route
> Annotation have always the priority on this settings
> You can use some placeholders to get properties from Ingress / Routes described in the next section

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


## List of annotations for Ingress / Route
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
   - <rule.host> : the url
   - <rule.scheme>: the scheme - http or https
   - <rule.path>: the path
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