# PLUGINS-002-plugin-management-api-design

## Status
Implemented

## Context
**Plugins** 插件提供了管理整个平台插件的功能，需要将此功能实现多个接口，用户平台管理员或 [tkeel-cli](https://github.com/tkeel-io/cli#keel-cli) 直接调用。

## Decision
设计提供注册、删除、查询三类接口，其中查询仅支持对指定 **PLUGIN_ID** 进行查询和查询所有已注册插件。

1. 注册(/register):
调用此接口，传入需要注册的 **PLUGIN_ID** 和平台管理员管理的平台的 **secret**。

**Plugins** 去请求待注册插件的 `/v1/identify` 接口，获取具体插件信息，根据插件信息再进行扩展点和扩展插件等功能的检查。

检查完成后，更新 `AllRegisteredPluginsMap`，创建新插件的 `Plugin` 数据和 `PluginRoute` 数据。

```bash
POST /register
```

request:
```go
struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}
```

response:
```go
type CommonResult struct {
	Ret int    `json:"ret"`
	Msg string `json:"msg"`
}
```

2. 删除(/delete)
调用此接口，传入需要删除的插件的 **PLUGIN_ID**。

**Plugins** 读取 `AllRegisteredPluginsMap`，并删除对应数据，删除插件的 `Plugin` 数据和 `PluginRoute` 数据。

```bash
POST /delete
```

request:
```go
struct {
	ID string `json:"id"`
}
```
response:
```go
type CommonResult struct {
	Ret int    `json:"ret"`
	Msg string `json:"msg"`
}
```

3. 查询指定插件(/get)
调用此接口，传入查询插件的 **PLUGIN_ID**。

**Plugins** 读取 **statestore** 查看是否存在对应ID的 `Plugin` 数据，如存在则返回，不存在返回 *HTTP StateCode 400*。

```bash
Get /get?id=${ID}
```

response:
```go
struct {
	openapi.CommonResult `json:",inline"`
	Data                 *reqPlugin `json:"data"`
}
type CommonResult struct {
	Ret int    `json:"ret"`
	Msg string `json:"msg"`
}
type reqPlugin struct {
	*keel.Plugin      `json:",inline"`
	*keel.PluginRoute `json:",inline"`
}
```

4. 查询所有插件(/list)
调用此接口，无参数。

**Plugins** 读取 `AllRegisteredPluginsMap`，遍历获取 **statestore** 中的 `Plugin` 数据，并返回。

```bash
Get /list
```

response:
```go
struct {
	openapi.CommonResult `json:",inline"`
	Data                 []*reqPlugin `json:"data"`
}
type CommonResult struct {
	Ret int    `json:"ret"`
	Msg string `json:"msg"`
}
type reqPlugin struct {
	*keel.Plugin      `json:",inline"`
	*keel.PluginRoute `json:",inline"`
}
```

## Consequences
可以实现。