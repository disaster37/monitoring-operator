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