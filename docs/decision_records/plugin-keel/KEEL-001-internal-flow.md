# KEEL-001-internal-flow

## Status
Implemented

## Context
在平台中，对于内部流量需要进行反向代理。内部流量就是对于已经注册到平台的插件之间的访问请求和插件访问自己的扩展点的请求。**Keel** 需要获取注册的插件的具体的路由表和状态。

## Decision
使用 `go http` 包实现一个 **HTTP** 服务器，设计了一套路由规则。

1. 插件间访问
访问平台已注册的插件中提供的方法及访问 `http://${YOUR_DAPR}/v1.0/invoke/keel/method/${target_plugin_id}/${target_plugin_method}`。

**keel** 将对 `method` 后的 *path* 解析，将第一个 `/` 之前的解析为注册的插件 ID，`/` 之后的解析为插件自己实现的方法的路径。

对于上述请求 *URL*，**Keel** 将会转发到 `http://${KEEL_DAPR}/v1.0/invoke/${target_plugin_id}/method/${target_plugin_method}`。

2. 插件访问自己的扩展点
当插件自己注册了扩展点时，在逻辑中需要回调时访问 `http://${YOURE_DAPR}/v1.0/invoke/keel/method/addons/${your_addons_name}` 时将会指向注册了你这个扩展点的插件的对应的端点。如： `http://${KEEL_DAPR}/v1.0/invoke/${target_plugin_id}/method/${target_plugin_method}`。

插件访问 **Keel** 时，通过 **dapr** 的边车中的中间件 `middleware.http.oauth2clientcredentials` 将 **PLUGIN_ID** 写进请求的 **header** 中。

对应 **component** 的 **helm** 的 *chart* 文件：
```yaml
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
  {{- if (eq .Values.pluginID "plugins") }}
    value: "http://127.0.0.1:{{ .Values.pluginsPort }}/oauth2/token"
  {{- else }}
    value: "http://plugins:{{ .Values.pluginsPort }}/oauth2/token"
  {{- end }}
  - name: headerName
    value: "x-plugin-jwt"
  - name: authStyle
    value: 0
```

**Plugins** 中实现了一个 `/oauth2/token` 端点，将检查传入的 `clientSecret` 是否与注册时传入的 `secret` 匹配，且是否注册。

如插件未注册或不匹配，则返回无效，插件的 **dapr** 边车将拒绝此条请求。

## Consequences
与 **Plugins** 的存储放在公共状态存储里。