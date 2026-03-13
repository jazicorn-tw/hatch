{{/*
Expand the name of the chart.
*/}}
{{- define "hatch.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
Truncate to 63 chars because Kubernetes name fields are limited to this.
*/}}
{{- define "hatch.fullname" -}}
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
Create chart label — used to track which chart version installed this resource.
*/}}
{{- define "hatch.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels applied to every resource.
*/}}
{{- define "hatch.labels" -}}
helm.sh/chart: {{ include "hatch.chart" . }}
{{ include "hatch.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels — used by Services and Deployments to match pods.
*/}}
{{- define "hatch.selectorLabels" -}}
app.kubernetes.io/name: {{ include "hatch.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
ServiceAccount name — uses the override if provided, otherwise the fullname.
*/}}
{{- define "hatch.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "hatch.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Resolved image tag — falls back to .Chart.AppVersion when values.image.tag is empty.
*/}}
{{- define "hatch.imageTag" -}}
{{- .Values.image.tag | default .Chart.AppVersion }}
{{- end }}
