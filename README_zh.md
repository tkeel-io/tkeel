<h1 align="center"> tKeel</h1>
<h5 align="center"> 新一代物联网开放平台：简单易用，马上起飞</h5>
<div align="center">

[![Go Report Card](https://goreportcard.com/badge/github.com/tkeel-io/tkeel)](https://goreportcard.com/report/github.com/tkeel-io/tkeel)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/tkeel-io/tkeel)
![GitHub](https://img.shields.io/github/license/tkeel-io/tkeel?style=plastic)
[![GoDoc](https://godoc.org/github.com/tkeel-io/tkeel?status.png)](http://godoc.org/github.com/tkeel-io/tkeel)
</div>

<div align="center">

![img.png](docs/images/img/system.png)
</div>

tKeel 是一套强壮可复用的物联网平台，能帮助您快速构建解决方案。

整体架构采用 **微服务** 模式，提供 *可拔插架构* 以及 *稳固可靠且快速响应* 的 **数据平面**。

解决了构建高性能、从设备到应用，模块化接入的关键难题。 

[English](README.md)

## 🏃🏻‍♀️ 让我们开始吧

* 通过 [CLI](https://github.com/tkeel-io/cli#) 工具快速安装 tKeel 平台
* [示例代码](https://github.com/tkeel-io/tkeel/tree/main/example) 会帮助您快速了解如何使用我们的 tKeel 物联网开放平台。

[官网文档]() 会有详细介绍，从安装到使用细节。


## 🪜 架构设计

或许您对我们这个 [tKeel 物联网开放平台]() 非常感兴趣，容许我为您简单介绍一下。

<div align="center">

![img.png](docs/images/img/layer.png)


<i data-selectable-paragraph="">架构图</i>
</div>

- ### **Resource**
数据存储支持，可以是任意您使用的数据库。
- ### **Core**
是整个平台的数据核心，提供了一些 **数据组织** 形式以及 **处理方式**。
 - 提供了 *时序数据*、_属性数据_、*关系* 等不同形式组织的数据，来构建便于开发和理解的对象。
 - 通过 **快照** 和 **订阅**（Event数据）的方式解决数据交互的问题。

 - ### **Service** 
提供应用 *可拔插能力* 以及其他一些 *核心功能（消息路由、租户管理、权限控制）*。
- ### **Interface**
通过封装，向 **Application** 提供简易的 *工具* 以及友好的 *接口*。
- ### **Application**
各种不同量级的应用，可以是您现有平台的一切服务。
- 现有平台仅需通过调用 **Interface** 提供的 API 即可使用拿到需要的数据。

## ✅ 我们有充分的理由

### 🥛 保持简单 
`tKeel` 平台总结了多年来在物联网中遇到的挑战与一些通用问题，只为解决物联网开发中的痛点。

`tKeel` 善于处理分布式系统中的数据流向与抽象问题，屏蔽了底层的复杂性，向外提供了**更简单**，**面向开发者更友好的抽象** ，帮助用户 *快速构建* 物联网解决方案。

### ⛓️ 小, 但也很强大
`tKeel` 物联网开放平台基于 *微服务架构* 构建，非厂商绑定且提供 **可扩展**、**可靠** 和 **高性能** 的高效开发方法。

借助 [Dapr](https://dapr.io) 的强大能力通过 [sidecar](https://docs.dapr.io/concepts/dapr-services/sidecar/) 形式向用户提供 *更简单的* 抽象，通过 `HTTP` 以及 `gRPC` 交互。

该平台使您的代码可以忽视托管环境，让设备对接的应用/插件（Plugin） **高度可移植** ，**不受编程语言限制**，可以让开发者使用自己喜欢的技术栈进行得心应手的开发。

### 🔌 插件化
插件（Plugin）的实现基于 **OpenAPI 约定**，通过云原生的方式让部署简捷轻便。

通过插件机制可以方便的复用您或他人公开的插件。

我们提供了一个 [官方插件仓库]() 让开发者可以随意挑选应对自己需求场景的插件应用。当然，如果您能将您的插件公开，有着相同需求场景的开发者将会感激不尽。

通过 [插件指南]() 您会发现实现插件是一件非常简单的事情。

### 📊 专注于数据
tKeel 物联网开放平台通过数据中心（[tKeel-io/Core](https://github.com/tkeel-io/core )）定义了 **数据实体**。 
对真实世界的物体（things）进行了模拟、抽象，
您可以定义 **关系映射**，通过平台强大的能力获得更多、更快、更简单的数据提炼。

在 **数据实体** 的设计下，我们可以将这种抽象设计适配 *消息*、_测点_、*对象* 以及 *关系* 等，平台提供 **多层次** 和 **多纬度** 的数据服务。

配置 **关系映射** 让您无需记忆复杂的 `消息 Topic` 以及 `消息格式` ，因为我们提供了高性能的数据处理解决方案。

## 🛣️ 路线
我们规划了一条 [时间线](https://github.com/tkeel-io/tkeel/issues/30) 为项目做更多的支持。

## 💬 一起点亮世界
如果您有任何的建议和想法，欢迎您随时开启一个 [Issue](https://github.com/tkeel-io/keel/issues )，期待我们可以一起交流，让世界更美好。

同时 **非常感谢** 您的 `反馈` 与 `建议` ！

[社区文档](docs/development/README.md) 将会带领您了解如何开始为 tKeel 贡献。

### 🙌 贡献一己之力

[开发指南](docs/development/developing-tkeel.md) 向您解释了如何配置您的开发环境。

我们有这样一份希望项目参与者遵守的 [行为准则](docs/community/code-of-conduct.md)。请阅读全文，以便您了解哪些行为会被容忍，哪些行为不会被容忍。

### 🌟 联系我们
提出您可能有的任何问题，我们将确保尽快答复！

| 平台 | 链接 |
|:---|----|
|email| tkeel@yunify.com|
|微博| [@tkeel]()|

## 🏘️ 仓库

| 仓库 | 描述 |
|:-----|:------------|
| [tKeel](https://github.com/tkeel-io/tkeel) | 如您所见，包含了平台的代码和平台概览|
| [CLI](https://github.com/tkeel-io/cli) | tKeel CLI 是用于各种 tKeel 相关任务的主要工具 |
| [Helm](https://github.com/tkeel-io/helm-charts) | tKeel 对应的 Helm charts |
| [Core](https://github.com/tkeel-io/core) | tKeel 的数据中心 |

