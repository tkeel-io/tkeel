# ARC-003-component-keel

## Status
Proposed

## Context
TKeel 的 v0.2.0 版本将核心模块去除实现的 **OPENAPI**。直接以 **dapr** 的 **service** 模式调用。

Keel 组件需要提供用户使用平台的前端和对应 API。同时需要实现动态路由和基于 URL 的对插件的权限管理功能。

## Decision
