# Gateway

<!-- TOC -->
- [概述](#概述)
- [设计背景](#设计背景)
- [实现](#实现)
- [存储](#存储)
  - [路由数据](#路由数据)
- [启动命令](#启动命令)
<!-- TOC -->

## 概述

-----

本文档介绍了**Keel**平台中和新插件**keel**的设计背景及实现.

为了让`dapr`的调用更加具象化，故将此插件命名为**keel**，它实现的就是平台的一个**gateway**的功能的插件。

## 设计背景

-----

为了实现内部流量的治理和外部流量的统一入口，同时为了让平台更加整体化，访问时用户仅需访问平台即可访问对应的插件，故需要一个`http`网关服务插件。

## 实现

-----

利用`dapr`的**status**这个**component**,用跨应用的方式存储路由数据,**keel**跨应用读取已注册的插件的路由信息，**plugins**插件可以动态修改路由状态，让**keel**插件可以根据插件状态进行阻拦。

## 存储

-----

**keel**仅读取跨应用的**status**中的路由数据，不修改。


### 路由数据

-----

注册到平台的插件的扩展点的路由信息和插件的状态等。

```go
type PluginRoute struct {
	openapi.Pluginstatus
	Addons map[string]string
}
```

其中`openapi.Pluginstatus`为插件的状态信息，根据状态不同，**keel**对对应插件的请求进行不同处理，如拦截等。

`Addons`主要是插件注册到平台的扩展点的路由信息，当插件发起自身扩展点回调请求时，**keel**将根据此来路由到对应的插件上去。

## 启动命令

-----

启动时，必须指定插件名和`dapr`边车中**appid**为**keel**。

替换`keel_oauth2_client.yaml`文件中的`PLUGIN_SECRET`、`PLUGIN_ID`、`PLUGIN_MANAGER_URL`

```bash
DAPR_HTTP_ADDRESS=127.0.0.1:3501  \
PLUGIN_HTTP_PORT=8081 \
PLUGIN_ID=keel \
dapr run --app-id keel \
         --app-protocol http \
         --app-port 8081 \
         --dapr-http-port 3501 \
         --log-level debug \
         --config ./config/config.yaml \
         --components-path ./config/components/ \
         ./gateway
```