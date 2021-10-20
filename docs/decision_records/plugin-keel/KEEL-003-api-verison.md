# KEEL-003-api-version

## Status
Implemented

## Context
平台应该维护一个版本，插件要依赖于特定的版本才能正确注册且在平台中运行起来，平台的版本应该在外部调用平台的URL中体现出来。

## Decision
1. 外部流量
参考 [OPENAPI 添加平台版本](../openapi/OPENAPI-003-add-platform-version.md#命名规范) 中提到的命名规范，在当前外部访问时，暴露的平台 API 的路径的基础上添加上版本信息。

`http://${KEEL}/${plugin_id}/${method}` => `http://${KEEL}/${TKEEL_VERSION}/${plugin_id}/${method}`

TKEEL_VERSION 取平台版本的前两位，并在开头添加小写字母「v」。

检查 API 中对应的 ${TKEEL_VERSION} 是否大于或等于 被调用插件依赖的平台版本。

~~STAGE 取每个版本的各个不同阶段的 API，如：Alpha，Beta，Release 这三种，其中 Release 可省略。~~

```bash
TKEEL_VERSION=v1.0
curl http://${KEEL}/${TKEEL_VERSION}/${plugin_id}/${method}
```

2. 内部流量
插件与插件间调用时，调用方的平台依赖版本必须大于等于被调用方的以来版本。

同理，扩展点调用时也应满足上述条件。

## Consequences
用第三位版本代替，当前没有 Stage 的需求。