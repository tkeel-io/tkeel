# KEEL-003-api-version

## Status
Proposed

## Context
平台应该维护一个版本，插件要依赖于特定的版本才能正确注册且在平台中运行起来，平台的版本应该在外部调用平台的URL中体现出来，做成类似于 `http://${KEEL}/${VERSION}/${plugin_id}/${method}` 的格式对外提供服务，URL 版本格式需要统一。