# Demo - Check extension before keel routing

这是一个关于插件扩展点的示例实现，当Keel插件对外部请求进行路由时调用扩展点的回调来检查此次请求是否合法。

当demo插件被安装后，每当Keel收到外部流量请求时均会请求demo插件的echo端点，echo端点默认允许所有流量通过，并将Keel调用扩展点时的内容输出到标准输出中。

**Keel**插件`externalPreRouteCheck`扩展点
```
callback before external flow routing
input request header and path
output http statecode
200   -- allow
other -- deny

外部流量路由前的回调
输入请求头和路径
输出 http 状态码 
200   -- 允许
other -- 拒绝
```

## Pre-requistes

- [Keel控制台及Keel的最小化k8s环境安装](https://github.com/tkeel-io/cli/blob/master/README_zh.md)
- [Helm安装](https://helm.sh/)


## Install demo-echo plugin by Helm

通过**Helm**安装**demo-echo**插件

```bash
cd deploy/chart/demo-echo
helm install -n keel-system demo-echo .
```


## Register demo-echo plugin

### Register demo-echo plugin by Keel CLI

1. 注册**demo-echo**插件
```bash
tkeel plugin register demo-echo
```
2. 检查状态
```bash
tkeel plugin list
```

输出应如下：
```bash
plugin list              
NAME       NAMESPACE    HEALTHY  STATUS    PLUGINSTATUS  REPLICAS  VERSION  AGE  CREATED              
auth       keel-system  True     Running   ACTIVE        1         0.0.1    37m  2021-10-07 16:07.00  
plugins    keel-system  True     Running   ACTIVE        1         0.0.1    37m  2021-10-07 16:07.00  
keel       keel-system  True     Running   ACTIVE        1         0.0.1    37m  2021-10-07 16:07.00
demo-echo  keel-system  True    Running   ACTIVE        1         0.0.1    2m   2021-10-05 11:25.19  
```

### Register demo-echo plugin by curl

1. 安装chart


2. 注册插件

```bash
curl http://${KEELADDR}/plugins/register -d '{"id":"demo-echo","secret":"changeme"}'
```

输出应如下：
```bash
{"ret":0,"msg":"ok"}
```

3. 检查是否注册成功

```bash
curl http://${KEELADDR}/plugins/get?id=demo-echo
```

输出应如下：
```bash
{
    "ret": 0,
    "msg": "ok",
    "data": {
        "ret": 0,
        "msg": "ok",
        "plugin_id": "demo-echo",
        "version": "0.0.1",
        "main_plugins": [
            {
                "id": "keel",
                "version": "1.0",
                "endpoints": [
                    {
                        "addons_point": "externalPreRouteCheck",
                        "endpoint": "echo"
                    }
                ]
            }
        ],
        "secret": "changme",
        "register_time": 1633681152,
        "status": "ACTIVE",
        "register_addons": null
    }
}
```

## Demo-echo output

每当有外部请求进入**Keel**时就会访问回调端点**echo**。

查看**demo-echo**的日志。

```bash
echo

Check the request header information,

return a status code 200,body:

{

"msg":"ok", // costom msg

"ret":0, // must be zero

}

If invalid, return a status code other than 200 or return body:

{

"msg":"faild", // costom msg

"ret":-1, // negative

}
```