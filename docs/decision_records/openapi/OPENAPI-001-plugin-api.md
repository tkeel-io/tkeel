# OPENAPI-001-plugin-api

## Status
Implemented

## Context
为了管理平台中的插件，我们需要对插件进行一些必要的交互，即要求平台的插件必须实现一些接口，我们需要让插件在平台中进行注册，同时也需要获取插件当前的运行状态，来管理当前插件的访问是否可达和管理插件的生命周期。

## Decision
1. 在注册阶段，需要插件实现 `/${VERSION}/identify` 接口。这个接口是在平台收到某插件注册请求时，主动请求注册插件的这个接口，然后由注册方插件通过这个接口范围自己的信息，如*插件 ID *和*版本号*等信息，请求方法是 **GET**。

```bash
POST /${VERSION}/identify
```

```json
{
    "ret": 0,
    "msg": "ok",
    "plugin_id": "yunify-xxx", // 反域名格式插件名
    "version": "1.0",
    "addons_points": [ // 插件自身的扩展点及说明
        {
            "addons_point": "create_plugins",
            "desc": "注册插件成功后调用，返回值data内数据返回给调用方"
        },
        {
            "addons_point": "install_plugins",
            "desc": "安装插件成功后调用，返回值data内数据返回给调用方"
        },
        {
            "addons_point": "update_plugins",
            "desc": "安装插件成功后调用，返回值data内数据返回给调用方"
        },
        {
            "addons_point": "unstall_plugins",
            "desc": "卸载插件成功后调用，无返回值"
        }
    ],
    "main_plugins": [
        {
            "id": "xxxx", // 被扩展的插件的名字
            "version": "1.0", // 被扩展插件的版本
            "endpoints": [
                {
                    "addons_point": "xxxx", // 扩展点名称
                    "endpoint": "xxx" // 对应实现端点名称（回调路径） 此路径必须注册到平台才能被调用
                }
            ]
        }
    ] // 可选，如果是实现了其他插件的扩展点带上
    // 以下暂未实现相关功能
    //    "dependent_plugin":[
    //        {
    //		"name":"xx",
    //		"version":"1.0" // 默认最小以来版本
    //		"match_rule":"EQU "// EQU|NEQ|LSS|LEQ|GTR|GEQ
    //	    }
    //	] // 可选，如果有依赖插件需带上此
}
```
2. 在插件运行在平台中时，平台会不定期的去获取插件的状态信息，及以 **GET** 方法访问插件的 `/${VERSION}/status` 接口，类似于软件中的插件检查。

```bash
GET /v1/status
```

```json
{
    "ret": 0,
    "msg": "ok",
    "status": "ACTIVE"
    // 状态有下列几种 
    // ACTIVE   --> 正常运行
    // STARTING --> 启动中 程序正在启动
    // STOPPING --> 停止中 程序正在停止
    // FAILED   --> 错误 程序自身错误
}
```
3. 插件的扩展机制的实现，这个是在 `/${VERSION}/identify` 中声明自己有哪些扩展点，或者声明自己扩展了哪些插件的哪些扩展点，用于解决某些插件的一些扩展需求，让插件的扩展也可以借助平台的插件机制，可以自由替换和变更。为了完成这个

## Consequences
`/v1/identify` 接口的返回值中可选字段可省略，`version` 这个字段暂时未作逻辑处理。