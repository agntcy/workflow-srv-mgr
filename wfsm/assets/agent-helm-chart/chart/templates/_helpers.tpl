{{/* Generate common labels */}}
{{- define "common.labels" -}}
{{- range $key, $value := .labels }}
{{ $key }}: {{ $value }}
{{- end }}
app.kubernetes.io/managed-by: me
{{- end -}}

{{/* Generate service */}}
{{- define "generate.service" -}}
{{- range .Values.agents }}
{{- if .externalPort }}
---
apiVersion: v1
kind: Service
metadata:
  name: {{ .name }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
spec:
  selector:
    app: {{ .name }}
  ports:
    - port: {{ .externalPort }}
      targetPort: {{ .internalPort }}
{{- end }}
{{- end }}
{{- end -}}

{{/* Generate configmap */}}
{{- define "generate.configmap" -}}
{{- range .Values.agents }}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .name }}-config
data:
  {{- range .env }}
  {{ .name }}: {{ .value | quote }}
  {{- end }}
{{- end }}
{{- end -}}

{{/* Generate secret */}}
{{- define "generate.secret" -}}
{{- range .Values.agents }}
---
apiVersion: v1
kind: Secret
metadata:
  name: {{ .name }}-secret
type: Opaque
data:
  {{- range .secretEnvs }}
  {{ .name }}: {{ .value | b64enc | quote }}
  {{- end }}
{{- end }}
{{- end -}}


{{/* Generate statefulset */}}
{{- define "generate.statefulset" -}}
{{- range .Values.agents }}
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: {{ .name }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
spec:
  serviceName: {{ .name }}
  replicas: 1
  selector:
    matchLabels:
      app: {{ .name }}
  template:
    metadata:
      labels:
        app: {{ .name }}
    spec:
      containers:
      - name: {{ .name }}
        image: "{{ .image.repository }}:{{ .image.tag }}"
        envFrom:
        - configMapRef:
            name: {{ .name }}-config
        - secretRef:
            name: {{ .name }}-secret
        volumeMounts:
        - name: storage
          mountPath: {{ .volumePath }}
        ports:
        - containerPort: {{ .internalPort }}
  volumeClaimTemplates:
  - metadata:
      name: storage
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 1Gi
{{- end }}
{{- end -}}
