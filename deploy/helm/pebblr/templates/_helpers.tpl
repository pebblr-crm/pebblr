{{/*
Expand the name of the chart.
*/}}
{{- define "pebblr.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "pebblr.fullname" -}}
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
Create chart label.
*/}}
{{- define "pebblr.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels.
*/}}
{{- define "pebblr.labels" -}}
helm.sh/chart: {{ include "pebblr.chart" . }}
{{ include "pebblr.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels.
*/}}
{{- define "pebblr.selectorLabels" -}}
app.kubernetes.io/name: {{ include "pebblr.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Service account name.
*/}}
{{- define "pebblr.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "pebblr.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Secret name for the Kubernetes Secret created by ExternalSecret.
*/}}
{{- define "pebblr.secretName" -}}
{{- printf "%s-secrets" (include "pebblr.fullname" .) }}
{{- end }}

{{/*
Pod-level security context — shared between Deployment and migration Job.
*/}}
{{- define "pebblr.podSecurityContext" -}}
runAsNonRoot: true
runAsUser: 1000
fsGroup: 1000
{{- end }}

{{/*
Container-level security context — shared between Deployment and migration Job.
*/}}
{{- define "pebblr.containerSecurityContext" -}}
allowPrivilegeEscalation: false
readOnlyRootFilesystem: true
capabilities:
  drop:
    - ALL
{{- end }}

{{/*
Secrets volume definition — shared between Deployment and migration Job.
*/}}
{{- define "pebblr.secretsVolume" -}}
- name: secrets
  secret:
    secretName: {{ include "pebblr.secretName" . }}
    defaultMode: 0400
{{- end }}

{{/*
Secrets volumeMount definition — shared between Deployment and migration Job.
*/}}
{{- define "pebblr.secretsVolumeMount" -}}
- name: secrets
  mountPath: {{ .Values.secrets.mountPath }}
  readOnly: true
{{- end }}

{{/*
Container image spec — shared between Deployment and migration Job.
*/}}
{{- define "pebblr.image" -}}
image: "{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
imagePullPolicy: {{ .Values.image.pullPolicy }}
{{- end }}
