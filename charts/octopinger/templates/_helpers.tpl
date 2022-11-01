{{/* vim: set filetype=mustache: */}}
{{/*
Create controller name and version as used by the chart label.
Truncated at 52 chars because StatefulSet label 'controller-revision-hash' is limited
to 63 chars and it includes 10 chars of hash and a separating '-'.
*/}}
{{- define "octopinger.controller.fullname" -}}
{{- printf "%s-%s" (include "octopinger.fullname" .) .Values.controller.name | trunc 52 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create the name of the controller service account to use
*/}}
{{- define "octopinger.controllerServiceAccountName" -}}
{{- if .Values.controller.serviceAccount.create -}}
    {{ default (include "octopinger.controller.fullname" .) .Values.controller.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.controller.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Return the default Octopinger controller app version
*/}}
{{- define "octopinger.defaultTag" -}}
  {{- default .Chart.AppVersion .Values.images.tag }}
{{- end -}}