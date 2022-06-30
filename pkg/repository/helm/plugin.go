package helm

const (
	PluginConfig = `apiVersion: dapr.io/v1alpha1
kind: Configuration
metadata:
  name: {{ .Values.pluginID }}
  namespace: {{ .Release.Namespace | quote }}
spec:
  accessControl:
    trustDomain: "tkeel"
    {{- if (eq .Values.pluginID "keel") }}
    defaultAction: allow
    {{- else }}
    defaultAction: deny
    policies:
    - appId: rudder
      defaultAction: allow
      trustDomain: 'tkeel'
      namespace: {{ .Release.Namespace | quote }}
    - appId: keel
      defaultAction: allow
      trustDomain: 'tkeel'
      namespace: {{ .Release.Namespace | quote }}
    {{- end }}
{{- if (ne .Values.pluginID "keel") }}
  httpPipeline:
    handlers:
    - name: {{ .Values.pluginID }}-oauth2-client
      type: middleware.http.oauth2clientcredentials
{{- end -}}`
	PluginOAuth2 = `{{- if (ne .Values.pluginID "keel") -}}
apiVersion: dapr.io/v1alpha1
kind: Component
metadata:
  name: {{ .Values.pluginID }}-oauth2-client
  namespace: {{ .Release.Namespace | quote }}
spec:
  type: middleware.http.oauth2clientcredentials
  version: v1
  metadata:
  - name: clientId
    value: {{ .Values.pluginID | quote }}
  - name: clientSecret
    value: {{ .Values.secret | quote }}
  - name: scopes
    value: "http://{{ .Values.pluginID }}.com"
  - name: tokenURL
  {{- if (eq .Values.pluginID "rudder") }}
    value: "http://127.0.0.1:{{ .Values.rudderPort }}/v1/oauth2/plugin"
  {{- else }}
    value: "http://rudder:{{ .Values.rudderPort }}/v1/oauth2/plugin"
  {{- end }}
  - name: headerName
    value: "x-plugin-jwt"
  - name: authStyle
    value: 1
{{- end -}}`
)
