# Plugin

<!-- TOC -->
- [概述](#概述)
- [设计背景](#设计背景)
- [技术实现](#技术实现)
  - [插件访问平台](#插件访问平台)
  - [插件访问插件](#插件访问插件)
  - [平台访问插件](#平台访问插件)
  - [插件扩展点回调](#插件扩展点回调)
  - [HTTP接口规范](#HTTP接口规范)
    - [必选](#必选)
    - [可选](#可选)
<!-- TOC -->

## 概述

---

本文档描述了插件机制的设计背景和实现，以及它在**Keel**平台中如何运作的。

## 设计背景

---

为了实现**Keel**平台的高扩展及简易定制化的设计理念，采用了插件化的扩展机制，插件化的机制不仅仅可以满足功能的按需扩展，同时也可以简化开发人员定制化开发的难度，仅需根据不同的需求开发新的插件即可。

所以在**Keel**平台中，插件就是平台功能的扩展，当插件被注册到平台之后，插件就成为了平台的一部分。

**Keel**平台是在`dapr`的框架之上部署和实现的一个平台。插件也就是`dapr`中定义的一个`service`。

## 技术实现

---

通过定义一套**HTTP**接口的规范，需要接入**Keel**平台的插件必须按照规范来实现对应的**HTTP**接口。

当实现了对应接口的插件注册到平台时，平台按照特定的顺序对这几个接口进行访问，完成整个注册的流程。

**Keel**中拥有核心插件
  * **plugins**: 用于管理**Keel**平台中的所有插件。
  * **keel**: 用于注册和路由内部和外部请求，统一的外部流量入口。
  * **auth**: 提供用户管理及其他认证资源。

###  插件访问平台

---

插件访问平台时，需要访问平台的注册到`dapr`中的服务调用的形式来进行访问。

```bash
curl -X${VERB}
     -H "Authorization: token"\
     "http://${dapr_host}:${dapr_port}/v1.0/invoke/keel/method/${plugin_id}/${endpoint}"
```

其中`token`由登陆时平台提供，内部调用时由调用方提供，及平台提供。

###  插件访问插件

---

因为插件已被注册到平台中，故与插件访问平台逻辑相同，调用方式相同，访问的就是平台被某插件扩展的某个功能。

###  平台访问插件

---

平台主动访问插件时，一般为其余插件访问平台，平台将请求进行转发，或平台对插件的状态(`/v1/status`)、认证接口(`/v1/identify`)等进行访问。

###  插件扩展点回调

---

当插件需要访问自己注册的扩展点的回调时，发送如下请求

```bash
curl -X${VERB}
     -H "Authorization: token"\
     "http://${dapr_host}:${dapr_port}/v1.0/invoke/keel/method/addons/${addons_point}"

```

其中`VERB`为插件定义的回调的请求方法，`addons_point`为插件定义的扩展点的名称，由注册时的`addons`中声明。

### HTTP接口规范

---

**Keel**平台中插件需要实现的接口，及接口实现的规范。

#### 必选

---

必须实现的端点。

 - 程序认证

    ---

    程序启动时声明扩展了哪些平台功能。

    ```
    Get /v1/identify
    ```
    
    调用此端点时无参。
        
    返回值：

    最简版本

    ```json
    {
        "ret":0,
        "msg":"ok"
        "plugin_id":"yunify-xxx", // 组织名-插件名
        "version":"1.0"
    }
    ```
        
    与最简版本对比多出来的字段均为可选字段

    ```json
    {
        "ret":0,
        "msg":"ok",
        "plugin_id":"yunify-xxx", // 组织名-插件名
        "version":"1.0",
        "addons":[ // 插件自身的扩展点及说明 可选
            {
                "endpoint":"before-send-callback", // 注册到平台的扩展端点
                "desc":"注册插件成功后调用，返回值data内数据返回给调用方"
            },
            {
                "endpoint":"after-send-callback",
                "desc":"安装插件成功后调用，返回值data内数据返回给调用方"
            }
        ]
    }
    ```

    ```json
    {
        "ret":0,
        "msg":"ok",
        "plugin_id":"yunify-xxx", // 组织名-插件名
        "version":"1.0",
        "main_plugin": [
            {
                "id":"plugin", // 被扩展的插件的名字
                "version":"1.0", // 被扩展插件的版本
                "endpoints":[
                    {
                        "addons":"beforeSendCallBack", // 扩展点名称
                        "endpoint":"/before-send-callback" // 对应实现端点名称（回调路径） 此路径必须注册到平台才能被调用
                    }
                ]
            } 
        ]
    }
    ```

 - 状态检查

    ---

    不定期的对插件进行状态检查，如果状态错误，则将请求阻拦，直至正常.

    ```
    Get /v1/status
    ```

    调用此端点时无参。

    返回值：

    ```json
    {
        "ret":0,
        "msg":"ok",
        "status":"ACTIVE"
        // 状态有下列几种 
        // ACTIVE   --> 正常运行
        // STARTING --> 启动中 程序正在启动
        // STOPPING --> 停止中 程序正在停止
        // FAILED   --> 错误 程序自身错误
    }
    ```

#### 可选

---

以下端点为扩展端点，插件按需实现。

 - 扩展插件校验

    ---
        
    扩展插件是插件提供的扩展点，其他插件来实现，此时提供扩展点的插件需要检查是否实现了对应的扩展点。

    检查实现的接口是否满足扩展点需求，变量为请求中**field**

    ```bash
    curl -X${VERB}
        -H "Authorization: token"\
        "http://${dapr_host}:${dapr_port}/v1.0/invoke/keel/method/${id}/${endpoint}"
    ```

    ```
    Post /v1/addons/identify
    ```

    ```json
    {
        "plugin": {
            "id": "ddd-xxxx", // 请求插件的插件ID
            "version": "1.0" // 版本
        }, // 新增的插件的名称和版本
        "endpoint": [
            {
                "addons_point": "xxxx",
                "endpoint": "xxxx"
            }
        ] // 新增插件实现的端点和目标
    }
    ```
        
    ```json
    {
        "ret":0, // 通过返回值判定是否检验通过
        "msg":"ok"
    }
    ```