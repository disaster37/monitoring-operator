@TODO

We need to create controller of Template object. The goal of this controller is to reconcile all resource generated from it.
It can use the labels that contains is name to do that.
Moreover, it need to trace his current name to ba abble to rename labels on all child 
To finish, it need to clean all resource when it will deleted

We need to add validating mutating to validate the template (yaml render)


We need to continue to work with actual template and support more generic template
if type is empty, so is the full object on template. Otherwise, is not the case
The name, kind group, of the template must be immuable to avoid to rename, and so stay many orphan resource
Instead to 
Sample of actual template:
```yaml
apiVersion: monitor.k8s.webcenter.fr/v1
kind: Template
metadata:
  annotations:
    argocd.argoproj.io/tracking-id: >-
      etloutils-monitoring-operator-rancher-hpd:monitor.k8s.webcenter.fr/Template:etloutils-monitoring-operator/check-ingress
    kubectl.kubernetes.io/last-applied-configuration: >
      {"apiVersion":"monitor.k8s.webcenter.fr/v1","kind":"Template","metadata":{"annotations":{"argocd.argoproj.io/tracking-id":"etloutils-monitoring-operator-rancher-hpd:monitor.k8s.webcenter.fr/Template:etloutils-monitoring-operator/check-ingress"},"labels":{"app.kubernetes.io/instance":"etloutils-monitoring-operator-rancher-hpd"},"name":"check-ingress","namespace":"etloutils-monitoring-operator"},"spec":{"name":"{{
      .templateName }}-{{ .name }}","template":"{{- $rule := index .rules 0
      -}}\n{{- $path := index $rule.paths 0 -}}\nactivate: true\n\nhost:
      HOST_KUBERNETES_HM-HPD\n\nname: App_{{ (or .labels.app .name) | title
      }}_URL\ntemplate: TS_App-Protocol-HTTP-MultiCheck\nmacros:\n  PROTOCOL:
      \"{{ $rule.scheme }}\"\n  URLPATH: \"{{ (or (index .annotations
      \"ingress.monitor.k8s.webcenter.fr/path\") $path) }}\"\n  CRITICALCONTENT:
      \"%{code} != 200 and %{code} != 401 and %{code} != 403\"\narguments:\n  -
      \"{{ $rule.host }}\"\npolicy:\n  excludeFields:\n    -
      activate\n","type":"CentreonService"}}
  creationTimestamp: '2022-09-05T13:47:29Z'
  generation: 9
  labels:
    app.kubernetes.io/instance: etloutils-monitoring-operator-rancher-hpd
  managedFields:
    - apiVersion: monitor.k8s.webcenter.fr/v1
      fieldsType: FieldsV1
      fieldsV1:
        f:metadata:
          f:annotations: {}
          f:labels: {}
        f:spec:
          .: {}
          f:name: {}
          f:type: {}
      manager: argocd-application-controller
      operation: Update
      time: '2022-09-05T13:47:29Z'
    - apiVersion: monitor.k8s.webcenter.fr/v1
      fieldsType: FieldsV1
      fieldsV1:
        f:metadata:
          f:annotations:
            f:argocd.argoproj.io/tracking-id: {}
            f:kubectl.kubernetes.io/last-applied-configuration: {}
          f:labels:
            f:app.kubernetes.io/instance: {}
        f:spec:
          f:template: {}
      manager: argocd-controller
      operation: Update
      time: '2024-11-25T15:58:02Z'
  name: check-ingress
  namespace: etloutils-monitoring-operator
  resourceVersion: '1221806770'
  uid: 649a924e-25c6-4ac4-b690-0b4822c7ae84
spec:
  name: '{{ .templateName }}-{{ .name }}'
  template: |
    {{- $rule := index .rules 0 -}}
    {{- $path := index $rule.paths 0 -}}
    activate: true

    host: HOST_KUBERNETES_HM-HPD

    name: App_{{ (or .labels.app .name) | title }}_URL
    template: TS_App-Protocol-HTTP-MultiCheck
    macros:
      PROTOCOL: "{{ $rule.scheme }}"
      URLPATH: "{{ (or (index .annotations "ingress.monitor.k8s.webcenter.fr/path") $path) }}"
      CRITICALCONTENT: "%{code} != 200 and %{code} != 401 and %{code} != 403"
    arguments:
      - "{{ $rule.host }}"
    policy:
      excludeFields:
        - activate
  type: CentreonService
```