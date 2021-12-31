{{/*
Expand the name of the chart.
*/}}
{{- define "monitoring-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "monitoring-operator.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "monitoring-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "monitoring-operator.labels" -}}
helm.sh/chart: {{ include "monitoring-operator.chart" . }}
{{ include "monitoring-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "monitoring-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "monitoring-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "monitoring-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "monitoring-operator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the secret
*/}}
{{- define "monitoring-operator.secretName" -}}
{{- if .Values.monitoring.secret.create }}
{{- default (include "monitoring-operator.fullname" .) .Values.monitoring.secret.name }}
{{- else }}
{{- default "default" .Values.monitoring.secret.name }}
{{- end }}
{{- end }}

{{/*
RBAC permissions
*/}}
{{- define "monitoring-operator.rbacRules" -}}
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
- apiGroups:
  - monitor.k8s.webcenter.fr
  resources:
  - centreons
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - monitor.k8s.webcenter.fr
  resources:
  - centreons/finalizers
  verbs:
  - update
- apiGroups:
  - monitor.k8s.webcenter.fr
  resources:
  - centreons/status
  verbs:
  - get
  - patch
  - update
- apiGroups:
  - monitor.k8s.webcenter.fr
  resources:
  - centreonservices
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - monitor.k8s.webcenter.fr
  resources:
  - centreonservices/finalizers
  verbs:
  - update
- apiGroups:
  - monitor.k8s.webcenter.fr
  resources:
  - centreonservices/status
  verbs:
  - get
  - patch
  - update
- apiGroups:
  - networking.k8s.io
  resources:
  - ingresses
  verbs:
  - get
  - list
  - watch
{{- end -}}