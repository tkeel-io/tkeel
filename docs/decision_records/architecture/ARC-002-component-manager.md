# ARC-002-component-manager

## Status
Proposed

## Context
TKeel 的 v0.2.0 版本将核心模块去除实现的 **OPENAPI**。直接以 **dapr** 的 **service** 模式调用。

Manager 组件需要提供管理平台的前端和对应 API。同时需要继承现有的 Plugins 中已实现的功能，需要重新设计及实现。

## Decision
Manager 提供一个管理平台，仅平台管理员登录。

管理平台需实现以下功能：

1. 可以操作和管理 TKeel 平台的插件。进行安装、卸载等操作。

2. 可以操作和管理 TKeel 平台的租户信息。进行注册、删除等操作。

### 插件管理（仅支持K8S）
在管理平台安装时需指定插件的 helm 仓库地址和 docker 镜像仓库地址。

平台会对两个仓库进行管理，用户可自行选择仓库中已有的插件（chart 包和对应的 docker image）进行安装，亦可将满足平台插件发布条件的自定义插件上传再进行安装。

安装过程则是将 chart 安装进平台环境中，并完成插件的注册流程。

### 租户管理
租户管理及对平台中的租户数据进行操作。

租户数据存储在 Auth 服务中，整个过程即调用 Auth 服务的对应 API。

### 代码设计
Manager 组件功能由 HTTP 服务提供，故选择 `gin` 框架快速实现对应的 API 接口。

仓库管理以 `repository` 的 GO 包实现，包含对 helm 和 docker image 对应的 SDK 的调用和插件上传的检查。

插件的安装和服务扩容等功能以调用 kubernetes 的 SDK 实现。