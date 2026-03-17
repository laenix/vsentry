{{/*
Expand the name of the chart.
*/}}
{{- define "vsentry.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "vsentry.fullname" -}}
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
{{- define "vsentry.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "vsentry.labels" -}}
helm.sh/chart: {{ include "vsentry.chart" . }}
{{ include "vsentry.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "vsentry.selectorLabels" -}}
app.kubernetes.io/name: {{ include "vsentry.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "vsentry.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "vsentry.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
VictoriaLogs URL
*/}}
{{- define "vsentry.victorialogsUrl" -}}
{{- if .Values.victorialogs.enabled }}
{{- printf "http://%s-victorialogs:9428" (include "vsentry.fullname" .) }}
{{- else }}
{{- .Values.externalVictorialogs.url | default .Values.config.externalUrl }}
{{- end }}
{{- end }}
