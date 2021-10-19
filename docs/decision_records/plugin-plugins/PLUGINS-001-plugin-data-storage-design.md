# PLUGINS-001-plugin-data-storage-design

## Status
Implemented

## Context
**Plugins** 管理插件的同时自己也是平台的一个插件，根据平台的插件机制，插件的数据是由插件自己来管理的，所以 **Plugins** 需要管理平台中已注册的插件信息。

根据 **OPENAPI** 的规范可以拿到一个插件在注册时需要传递到平台的所有信息，所以注册时应该将此信息存储，还需要存储一个能够快速获取所有已注册插件的ID的表，同时根据 **Keel** 插件的路由机制，还需要存储一个拥有插件当前状态和插件扩展点的路由表。

## Decision
根据上述需求， **Plugins** 管理的数据可以分为两个大类，一个是仅自己这个插件需要的，另一个是和 **Keel** 插件共享的数据。

在 **dapr** 中选取 **statestore** 这个 **component** 来存储数据，又根据 **dapr** 的特性需要同时创建两个 **statestore** 的 **component** ,一个存储非共享数据，一个存储共享数据。

**statestore** 的好处就是存储以 `KV` 进行查询或存储，同时支持事务操作。

对于事务级别，根据业务场景选择乐观锁机制，且对于路由表数据， **Keel** 插件仅读。

### Statestore配置
存储后端暂时选择的是 **redis**，**dapr** 的屏蔽资源层的特性可以在后期对代码无修改的情况下选择其他存储后端。

* public-store: 用于存放路由表数据，可跨插件共享。
```yaml
apiVersion: dapr.io/v1alpha1
kind: Component
metadata:
  name: {{ printf "%s-redis-public-store" $.Chart.Name . }}
  namespace: {{ $.Release.Namespace | quote }}
spec:
  type: state.redis
  version: v1
  metadata:
  - name: redisHost
    value: {{ printf "%s-redis:6379" $.Chart.Name }}
  - name: redisPassword
    secretKeyRef:
      key: redis-password
      name: {{ printf "%s-redis" $.Chart.Name }}
  - name: keyPrefix
    value: none
```
* private-store: 用于存放插件自身私有的数据。
```yaml
apiVersion: dapr.io/v1alpha1
kind: Component
metadata:
  name: {{ printf "%s-redis-private-store" $.Chart.Name . }}
  namespace: {{ $.Release.Namespace | quote }}
spec:
  type: state.redis
  version: v1
  metadata:
  - name: redisHost
    value: {{ printf "%s-redis:6379" $.Chart.Name }}
  - name: redisPassword
    secretKeyRef:
      key: redis-password
      name: {{ printf "%s-redis" $.Chart.Name }}
```

根据 **dapr** 提供的能力，可以直接用同一个 **redis** 作为实际存储的后端，节约资源。

### 存储数据的数据结构

* Plugin(插件数据):
```go
type Tenant struct {
	TenantID   string      `json:"tenant_id"`
	TenantName string      `json:"tenant_name"`
	ExpireTime int64       `json:"expire_time"`
	Extra      interface{} `json:"extra,omitempty"`
}

type Plugin struct {
	*openapi.IdentifyResp `json:",inline"`
	Secret                string    `json:"secret"`
	RegisterTime          int64     `json:"register_time,omitempty"`
	ActiveTenant          []*Tenant `json:"active_tenant,omitempty"`
}
```

存储的 `Key` 为；`plugin_${PLUGIN_ID}`。

`openapi.IdentifyResp` 就是 **OPENAPI** 中 `/v1/identify` 接口的返回值的结构， `Plugin.ActiveTenant` 保存的是所有启用了此插件的租户的信息。

* AllRegisteredPluginsMap(所有已注册插件Map):
```go
type AllRegisteredPluginsMap map[string]string
```

存储的 `Key` 为；`all_registered_plugin`。

`map[string]string` 中 `Key` 为已注册的插件ID， `Value` 暂定为 `true` 占位符，后期可扩展。

* scrape_state
存储的 `Key` 为；`scrape_state`。

一个 **BOOLEN** 值，分布式锁，用于解决定获取插件状态这个功能时分布式事务问题。

* PluginRoute(插件路由表):
```go
type PluginRoute struct {
	Status         openapi.PluginStatus `json:"status"`
	RegisterAddons map[string]string    `json:"register_addons,omitempty"`
}
```

存储的 `Key` 为：`plugin_route_${PLUGIN_ID}`。

`PluginRoute.Status` 为最后一次获取插件状态时，插件的状态信息，根据此状态来放行或拦截请求此插件的流量。

`PluginRoute.RegisterAddons` 是插件注册时声明的扩展点的回调路径。 `Key` 为注册时的扩展点名称， `Value` 为 `${PLUGIN_ID}/${ENDPOINT}`。

## Consequences
数据格式大多为 **OPENAPI** 规范中定义的数据格式，随 **OPENAPI** 版本变化。