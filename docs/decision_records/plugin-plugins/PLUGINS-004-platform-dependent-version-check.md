# PLUGINS-004-platform-dependent-version-check

## Status
Implemented

## Context
平台的开发是个迭代的过程，新的版本应该兼容老的版本的某些功能，所以对于接入平台的插件而言，需要指定依赖的平台版本，只有与指定的版本匹配才能在平台中正常运行。

## Decision
在注册流程中获取待注册插件的 `/v1/identify` 后需要检查是否与当前平台的版本一致，即检查是否小于等于 **Plugins** 插件的版本，如大于，则拒绝注册。

将依赖平台的版本信息写入到 **Plugin** 数据中，也写入到 **PluginRoute** 数据中，在对应结构中添加 `TkeelVersion` 字段存储相关信息。

修改后：

1. 新增 `pkg/keel.PlguinRoute` 中 `TkeelVersion` 字段。
2. 新增 `pkg/openapi.IdentifyResp` 中 `TkeelVersion` 字段。
3. 增加注册时检查。

```go
type PluginRoute struct {
	Status         openapi.PluginStatus `json:"status"`
	TkeelVersion   string               `json:"tkeel_version"`
	RegisterAddons map[string]string    `json:"register_addons,omitempty"`
}

// IdentifyResp response of /v1/identify.
type IdentifyResp struct {
	CommonResult `json:",inline"`
	PluginID     string         `json:"plugin_id"`
	Version      string         `json:"version"`
	TkeelVersion string         `json:"tkeel_version"`
	AddonsPoints []*AddonsPoint `json:"addons_points,omitempty"`
	MainPlugins  []*MainPlugin  `json:"main_plugins,omitempty"`
}
```

## Consequences
检查前两位版本，且平台版本向下兼容（平台版本大于或等于注册插件的以来版本）。