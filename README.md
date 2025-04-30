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
- Auto create resources from `Secret (TLS certificate only)` with template concept

## Deploy operator with OLM

The right way to deploy operator base on operatot-sdk is to use OLM.
You can use the catalog image `quay.io/webcenter/monitoring-operator-catalog:v1.0.1`

For test purpose, you can use operator-sdk to run bundle
```bash
operator-sdk run bundle quay.io/webcenter/monitoring-operator-bundle:v1.0.1
```

Or upgrade already deployed catalogue
```bash
operator-sdk run bundle-upgrade quay.io/webcenter/monitoring-operator-bundle:v1.0.1
```

## Use it

### Platform

The first way consist to declare a platform. A platform is a monitoring API endpoint. actually, we only support Centreon platform.
So, you need to provide a resource of type platform on same operator namespace.

***platform.yaml***
```yaml
apiVersion: monitor.k8s.webcenter.fr/v1
kind: Platform
metadata:
  name: default
spec:
  isDefault: true
  type: centreon
  debug: false
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
apiVersion: monitor.k8s.webcenter.fr/v1
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


> You can use short name `kubectl get mcs` when you should to get CentreonService resources.

### CentreonServiceGroup

This custom resource permit to handle service group on Centreon.

You can use this properties to set service group:
```yaml
apiVersion: monitor.k8s.webcenter.fr/v1
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

  > You can use short name `kubectl get mcsg` when you should to get CentreonServiceGroup resources.


### Policy concept

The policy permit to handle how controller will reconcile resource.
You can choose to disable the create, update or delete operation on target platform. Or ignore fields when compute diff to know if resource can be updated

```yaml
  policy:
    noCreate: false # Set true to disable create operation on target platform
    noUpdate: false # Set true to disable update operation on target platform
    noDelete: false # Set true to disable delete operation on target platform
    excludeFields:  # Set some fields to ignore them on diff operation
      - activate
```

### Template concept

Template is a conceptual resource that permit to create real resource like CentreonService or CentreonServiceGroup from standard kubernetes resources. You need to create the template and them reference it with annotation on standard kubernetes resource.

> For `Namespace` and `Node` resource, the target resource is created on same operator namespace

You can, for exemple, should create CentreonService when you create new Ingress ressource.

To do that, first you create Template called `check-ingress`

```yaml
apiVersion: monitor.k8s.webcenter.fr/v1
kind: Template
metadata:
  name: check-ingress
  namespace: default
spec:
  type: "CentreonService"
  name: "{{ .templateName }}-{{ .name }}"
  template: |
    {{ $rule := index .rules 0}}
    {{ $path := index $rule.paths 0}
    host: KUBERNETES
    name: "check-url-{{ $path }}"
    template: check-url
    activate: true
    macros:
      SCHEME: "{{ $rule.scheme }}"
      HOST: "{{ $rule.host }}"
      PATH: "{{ $path }}"
```

> of course, you can use go templating to generate real resource from template with some placeholders

Then, when you create your ingress, you need juste to add annotation `monitor.k8s.webcenter.fr/templates`

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    "monitor.k8s.webcenter.fr/templates": `[{"namespace":"default", "name": "check-ingress"}]`,
  name: ingress-sample
  namespace: default
spec:
  rules:
  - host: sample.domain.com
    http:
      paths:
      - backend:
          service:
            name: sample
            port:
              number: 8000
        path: /
        pathType: Prefix
  tls:
  - hosts:
    - sample.domain.com
    secretName: sample-tls
```

The operator will create new resource `check-ingress-ingress-sample` of type `CentreonService`

> You can use short name `kubectl get mtmpl` when you should to get Template resources.

#### Placeholders for resource name

Per default, if you not set `spec.name` on template, it will use the template name as resource name.

You can use the following placeholders:
- `{{ .templateName }}`: it's the template name used to generate the resource
- all other placeholders available for each standard kubernetes resource

#### Placeholders for Ingress

You can use the followings placeholders:
- **name**: the resource name (string)
- **namespace**: the resource namespace (string)
- **labels**: the resource labels (map of string)
- **annotations**: the resource annotations (map of string)
- **rules**: the ingress rules (array of rule)
  - **rule**: the ingress rule (map)
    - **host**: the host
    - **sheme**: the scheme (http or https)
    - **paths**: the list of path (array of string)

You get a map like this:

```go
placholders := map[string]any{
  "name":      "test",
  "namespace": "default",
  "labels": map[string]string{
    "app": "appTest",
    "env": "dev",
  },
  "annotations": map[string]string{
    "anno1": "value1",
    "anno2": "value2",
  },
  "rules": []map[string]any{
    {
      "host":   "front.local.local",
      "scheme": "http",
      "paths": []string{
        "/",
        "/api",
      },
    },
    {
      "host":   "back.local.local",
      "scheme": "https",
      "paths": []string{
        "/",
      },
    },
  },
}
```

#### Placeholders for Route

You can use the followings placeholders:
- **name**: the resource name (string)
- **namespace**: the resource namespace (string)
- **labels**: the resource labels (map of string)
- **annotations**: the resource annotations (map of string)
- **rules**: the ingress rules (array of rule)
  - **rule**: the ingress rule (map)
    - **host**: the host
    - **sheme**: the scheme (http or https)
    - **paths**: the list of path (array of string)

You get a map like this:

```go
placeholders := map[string]any{
  "name":      "test",
  "namespace": "default",
  "labels": map[string]string{
    "app": "appTest",
    "env": "dev",
  },
  "annotations": map[string]string{
    "anno1": "value1",
    "anno2": "value2",
  },
  "rules": []map[string]any{
    {
      "host":   "front.local.local",
      "scheme": "https",
      "paths": []string{
        "/",
      },
    },
  },
}
```

#### Placeholders for Namespace

You can use the followings placeholders:
- **name**: the resource name (string)
- **namespace**: the resource name (string)
- **labels**: the resource labels (map of string)
- **annotations**: the resource annotations (map of string)

You get a map like this:

```go
placeholders := map[string]any{
  "name":      "test",
  "namespace": "test",
  "labels": map[string]string{
    "app": "appTest",
    "env": "dev",
  },
  "annotations": map[string]string{
    "anno1": "value1",
    "anno2": "value2",
  },
}
```

#### Placeholders for Node

You can use the followings placeholders:
- **name**: the resource name (string)
- **labels**: the resource labels (map of string)
- **annotations**: the resource annotations (map of string)
- **unschedulable**: it's true if node is currently unschedulable (boolean)
- **nodeInfo**: the node infos (Struct of type [NodeInfo](https://pkg.go.dev/k8s.io/api@v0.24.2/core/v1#NodeSystemInfo))
- **addresses**: The node address (Array of [NodeAddress](https://pkg.go.dev/k8s.io/api@v0.24.2/core/v1#NodeAddress))

You get a map like this:

```go
placeholders = map[string]any{
  "name": "test",
  "labels": map[string]string{
    "app": "appTest",
    "env": "dev",
  },
  "annotations": map[string]string{
    "anno1": "value1",
    "anno2": "value2",
  },
  "nodeInfo": core.NodeSystemInfo{
    MachineID:    "id",
    SystemUUID:   "uuid",
    Architecture: "x86",
  },
  "addresses": []core.NodeAddress{
    {
      Type:    core.NodeExternalIP,
      Address: "10.0.0.1",
    },
  },
  "unschedulable": true,
}
```

#### Placeholders for Certificate (Secret of type TLS)

You can use the followings placeholders:
- **name**: the resource name (string)
- **namespace**: the resource namespace (string)
- **labels**: the resource labels (map of string)
- **annotations**: the resource annotations (map of string)
- **certificates**: the list of certificate info (array of [Certificate](https://pkg.go.dev/crypto/x509#Certificate))


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