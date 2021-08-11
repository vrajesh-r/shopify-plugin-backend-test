{{- define "service.config" -}}
{{- $serviceConfig := (.Files.Get "local/config.yaml" | fromYaml) }}
{{  mergeOverwrite $serviceConfig .Values.config | toYaml }}
{{- end -}}

{{- define "datadogLabels" -}}
tags.datadoghq.com/env: {{ .Values.datadogEnv | required "datadogEnv is required" }}
tags.datadoghq.com/service: shopify-plugin-backend
tags.datadoghq.com/version: {{ .Values.image.tag | required "image.tag is required" }}
{{- end -}}
