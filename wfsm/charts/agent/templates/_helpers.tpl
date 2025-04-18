{{/* Generate common labels */}}
{{- define "common.labels" -}}
{{- range $key, $value := .labels }}
{{ $key }}: {{ $value }}
{{- end }}
app.kubernetes.io/managed-by: me
{{- end -}}
