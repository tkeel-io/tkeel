# PLUGINS-003-tenant-management-api-design

## Status
Accepted

## Context
平台的租户系统需要平台去通知对应插件对租户进行启用或停止。

## Decision
涉及到对插件通知，即调用插件的接口，且租户系统为平台固有系统，所以由 **Plugins** 去实现这个功能。

1. 租户的启用(/tenant-bind):
调用此接口，传入插件 ID 和租户的相关信息，平台主动调用对应插件的 `/v1/tenant/bind` 接口，等待返回成功。

如成功则将租户信息存储到对应的 `Plugin` 数据中，失败则返回具体原因。

```bash
POST /tenant-bind
```

request:
```go
struct {
	PluginID string `json:"plugin_id"`
	Version  string `json:"version"`
	TenantID string `json:"tenant_id"`
	Extra    []byte `json:"extra"`
}
```
其中 `Extra` 字段为插件自定义字段。

response:
```go
type CommonResult struct {
	Ret int    `json:"ret"`
	Msg string `json:"msg"`
}
```

2. 租户的停用(/tenant-delete):
调用此接口，传入插件 ID 和租户的相关信息，平台主动调用对应插件的 `/v1/tenant/delete` 接口，等待返回成功。

如成功则将租户信息存储到对应的 `Plugin` 数据中，失败则返回具体原因。

```bash
POST /tenant-delete
```

request:
```go
struct {
	PluginID string `json:"plugin_id"`
	Version  string `json:"version"`
	TenantID string `json:"tenant_id"`
	Extra    []byte `json:"extra"`
}
```
其中`Extra`字段为插件自定义字段。

response:
```go
type CommonResult struct {
	Ret int    `json:"ret"`
	Msg string `json:"msg"`
}
```

## Consequences
请求中携带的 `Extra` 字段用于扩展部分插件的一些自定义的需求。初始化用户的时候的额外参数等。