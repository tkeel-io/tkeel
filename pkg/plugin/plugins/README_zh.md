# Plugins

<!-- TOC -->
- [概述](#概述)
- [设计背景](#设计背景)
- [实现](#实现)
- [存储](#存储)
  - [插件数据](#插件数据)
  - [路由数据](#路由数据)
- [API](#API)
- [启动命令](#启动命令)
<!-- TOC -->

## 概述

本文档介绍了**Keel**平台中和新插件**plugins**的设计背景及实现.

## 设计背景

插件机制中，对于平台中的插件需要统一管理，对插件注册和其扩展点进行管理等，所以需要一个系统插件来管理这个。

## 实现

利用`dapr`的**status**这个**component**来存储所有的插件数据，用跨应用的方式存储路由数据，可以动态修改路由状态，让**keel**插件可以根据插件状态进行阻拦，或实现其他功能。

利用`dapr`的**oauth2clientcredentials**这个**middleware**来添加`token`，`token`由**plugins**颁发并校验。

## 存储

**plugins**的存储主要分为两类，一类是**plugins**插件自己的管理的插件的数据，一类是需要共享给**keel**插件的路由数据。

### 插件数据

存储注册到平台的插件的相关信息和注册时间等。

```go
type Plugin struct {
	*openapi.IdentifyResp `json:",inline"`
	Secret                string `json:"secret"`
	RegisterTime          int64  `json:"register_time,omitempty"`
	StopTime              int64  `json:"stop_time,omitempty"`
}
```

其中`openapi.IdentifyResp`部分为注册的插件的`/v1/identify`接口返回值，`RegisterTime`为注册的时间戳，`StopTime`为停止时间戳(暂未使用)

### 路由数据

存储注册到平台的插件的扩展点的路由信息和插件的状态等。

```go
type PluginRoute struct {
	openapi.Pluginstatus
	Addons map[string]string
}
```

其中`openapi.Pluginstatus`为插件的状态信息，根据状态不同，**keel**对对应插件的请求进行不同处理，如拦截等。

`Addons`主要是插件注册到平台的扩展点的路由信息，当插件发起自身扩展点回调请求时，**keel**将根据此来路由到对应的插件上去。

## API

提供查询、删除、注册插件接口，使用restful风格API

  1. 查询

```base
GET /plugins?&id=xxx
```

response:

```go
struct {
	openapi.CommonResult `json:",inline"`
	Datas                []*Plugin `json:"data"`
}
```

将此结构转为json
  
  2. 删除

```bash
DELETE /plugins
-d {"id":["xxx","xx"]}
```

response:

```go
type CommonResult struct {
Ret int    `json:"ret"`
Msg string `json:"msg"`
}
```

将此结构转为json

  3. 注册

```bash
POST /plugins
-d {"id":"xxx","secret":"xxxx"}

secret 由系统管理员颁发
```

response:
```go
type CommonResult struct {	
    Ret int    `json:"ret"`	
    Msg string `json:"msg"`
}
```
将此结构转为json

  4. token生成

```bash
POST /oauth2/token
```

`dapr`边车的**oauth2clientcredentials**自动申请，自动添加`x-plugin-jwt`头信息。

response:
```go
type CommonResult struct {	
    Ret int    `json:"ret"`	
    Msg string `json:"msg"`
}
```
将此结构转为json

## 启动命令

---

替换`keel_oauth2_client.yaml`文件中的`PLUGIN_SECRET`、`PLUGIN_ID`、`PLUGIN_MANAGER_URL`

```bash
DAPR_HTTP_ADDRESS=127.0.0.1:3500  \
PLUGIN_ID=plugins \
dapr run --app-id plugins \
         --app-protocol http \
         --app-port 8080 \
         --dapr-http-port 3500 \
         --log-level debug \
         --config ./config/config.yaml \
         --components-path ./config/components \
         ./manager
```