{{- /* Expand to the namespace lingress installs into. */ -}}
{{- define "lingress.namespace" -}}
    {{- default .Release.Namespace .Values.namespace -}}
{{- end -}}

{{- /* Create chart name and version as used by the chart label. */ -}}
{{- define "lingress.chart" -}}
    {{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- /* Expand the name of the chart. */ -}}
{{- define "lingress.name" -}}
    {{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- /*
    Create a default fully qualified app name.
    We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
    If release name contains chart name it will be used as a full name.
*/ -}}
{{- define "lingress.fullname" -}}
    {{- if .Values.fullnameOverride -}}
        {{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
    {{- else -}}
        {{- $name := default .Chart.Name .Values.nameOverride -}}
        {{- if contains $name .Release.Name -}}
            {{- .Release.Name | trunc 63 | trimSuffix "-" -}}
        {{- else -}}
            {{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
        {{- end -}}
    {{- end -}}
{{- end -}}

{{- /* Defines default labels used in each configuration. */ -}}
{{- define "lingress.labels" -}}
    {{- printf "helm.sh/chart: " -}}{{- include "lingress.chart" . -}}{{- printf "\n" -}}
    {{- include "lingress.selectorLabels" . -}}
    {{- if .Chart.AppVersion -}}
        {{- printf "app.kubernetes.io/version: %v\n" (.Chart.AppVersion | replace "+" "_" | trunc 63 | trimSuffix "-" | quote) -}}
    {{- end -}}
    {{- printf "app.kubernetes.io/managed-by: %v\n" .Release.Service  -}}
{{- end -}}

{{- /* Defines default selector labels used in each configuration. */ -}}
{{- define "lingress.selectorLabels" -}}
    {{- printf "app.kubernetes.io/name: " -}}{{- include "lingress.name" . -}}{{- printf "\n" -}}
    {{- printf "app.kubernetes.io/instance: %v\n" .Release.Name -}}
{{- end }}

{{- /* Defines default labels used in each configuration. */ -}}
{{- define "lingress.annotations" -}}
    {{- if .Chart.AppVersion -}}
        {{- printf "helm.sh/chart-name: %v\n" (.Chart.Name | quote) -}}
        {{- printf "helm.sh/chart-version: %v\n" (.Chart.Version | quote) -}}
        {{- printf "helm.sh/chart-app-version: %v\n" (.Chart.AppVersion | quote) -}}
    {{- end -}}
{{- end }}

{{- /* Create the name of the service account to use*/ -}}
{{- define "lingress.serviceAccountName" -}}
    {{- if .Values.serviceAccount.enabled -}}
        {{- default (include "lingress.fullname" .) .Values.serviceAccount.name -}}
    {{- else -}}
        {{- default "default" .Values.serviceAccount.name -}}
    {{- end -}}
{{- end -}}
