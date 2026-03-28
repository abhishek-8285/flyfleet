{{/* vim: set filetype=mustache: */}}
{{/*
Helper functions for FlyFleet Helm chart
*/}}

{{- define "flyfleet.labels" -}}
labels:
  app.kubernetes.io/name: "{{ .Chart.Name }}"
  app.kubernetes.io/instance: "{{ .Release.Name }}"
  app.kubernetes.io/version: "{{ .Chart.AppVersion }}"
  app.kubernetes.io/managed-by: "Helm"
  service: "{{ .ServiceName }}"
{{- end -}}

{{- define "flyfleet.selectorLabels" -}}
labels:
  app.kubernetes.io/name: "{{ .Chart.Name }}"
  app.kubernetes.io/instance: "{{ .Release.Name }}"
  service: "{{ .ServiceName }}"
{{- end -}}

{{- define "flyfleet.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{ .Values.serviceAccount.name | default (include "flyfleet.fullname" .) }}
{{- else -}}
{{ .Values.serviceAccount.name | default "default" }}
{{- end -}}
{{- end -}}

{{- define "flyfleet.fullname" -}}
{{- printf "%s-%s" .Release.Name .ServiceName | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "flyfleet.image" -}}
{{- $registry := $.Values.global.imageRegistry | default "" -}}
{{- $repo := .image.repository -}}
{{- if $registry -}}
{{-   $repo = printf "%s/%s" $registry $repo -}}
{{- end -}}
image:
  repository: "{{ $repo }}"
  tag: "{{ .image.tag | default $.Values.global.imageTag | default .Chart.AppVersion }}"
  pullPolicy: "{{ .image.pullPolicy | default "Always" }}"
{{- end -}}

{{- define "flyfleet.ports" -}}
ports:
  - name: http
    port: {{ .service.port }}
    targetPort: {{ .service.port }}
    protocol: TCP
{{- end -}}

{{- define "flyfleet.resources" -}}
resources:
  requests:
    memory: "{{ .resources.requests.memory | default "256Mi" }}"
    cpu: "{{ .resources.requests.cpu | default "250m" }}"
  limits:
    memory: "{{ .resources.limits.memory | default "512Mi" }}"
    cpu: "{{ .resources.limits.cpu | default "500m" }}"
{{- end -}}
