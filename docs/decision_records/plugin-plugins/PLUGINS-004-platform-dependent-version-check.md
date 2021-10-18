# PLUGINS-004-platform-dependent-version-check

## Status
Proposed

## Context
平台的开发是个迭代的过程，新的版本应该兼容老的版本的某些功能，所以对于接入平台的插件而言，需要指定依赖的平台版本，只有与指定的版本匹配才能在平台中正常运行。

## Decision
在注册流程中获取待注册插件的 `/v1/identify` 后需要检查是否与当前平台的版本一致，即检查是否小于等于 **Plugins** 插件的版本，如大于，则拒绝注册。

将依赖平台的版本信息写入到 **Plugin** 数据中，也写入到 **PluginRoute** 数据中，在对应结构中添加 `TkeelVersion` 字段存储相关信息。