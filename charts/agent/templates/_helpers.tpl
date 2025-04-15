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
    {{- range $key, $value := .service.labels }}
    {{ $key }}: {{ $value }}
    {{- end }}
  annotations:
    {{- range $key, $value := .service.annotations }}
    {{ $key }}: {{ $value }}
    {{- end }}
spec:
  type: {{ .service.type | default "ClusterIP" }}
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
{{- if not .existingSecretName }}
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
    {{- range $key, $value := .statefulset.labels }}
    {{ $key }}: {{ $value }}
    {{- end }}
  annotations:
    {{- range $key, $value := .statefulset.annotations }}
    {{ $key }}: {{ $value }}
    {{- end }}
spec:
  serviceName: {{ .name }}
  replicas: {{ .statefulset.replicas | default 1 }}
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
        {{- if .existingSecretName }}
        - secretRef:
            name: {{ .existingSecretName }}
        {{- else }}
        - secretRef:
            name: {{ .name }}-secret
        {{- end }}
        volumeMounts:
        - name: storage
          mountPath: {{ .volumePath }}
        ports:
        - containerPort: {{ .internalPort }}
        resources:
          {{- toYaml .statefulset.resources | nindent 10 }}
      nodeSelector:
        {{- toYaml .statefulset.nodeSelector | nindent 8 }}
      affinity:
        {{- toYaml .statefulset.affinity | nindent 8 }}
      tolerations:
        {{- toYaml .statefulset.tolerations | nindent 8 }}
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
