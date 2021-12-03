# monitoring-operator
Kubernetes operator to manage monitoring resources

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

3. Generate some Go codes like controllers

> Need each time you change `*_types.go`

```bash
make generate
```

4. Generate CRDs

> Need each time you change `*_types.go` or add some comment annotations in controllers

```bash
make manifests
```

5. Test

```bash
export KUBECONFIG=/home/theia/.kube/config
make install run
```