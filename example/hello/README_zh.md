# Demo - echo plugin

这是一个插件的示例实现，对本插件的端口进行访问时将回复相同内容。

## Pre-requistes

- [Keel控制台及Keel的最小化k8s环境安装](https://github.com/tkeel-io/cli/blob/master/README_zh.md)
- [Helm安装](https://helm.sh/)


## Install keel-echo plugin by Helm

通过**Helm**安装**keel-echo**插件

```bash
cd deploy/chart/keel-echo
helm install -n keel-system keel-echo .
```


## Register keel-echo plugin

### Register keel-echo plugin by Keel CLI

1. 注册**keel-echo**插件
```bash
tkeel plugin register keel-echo
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
keel-echo  keel-system  True    Running   ACTIVE        1         0.0.1    2m   2021-10-05 11:25.19  
```

### Register keel-echo plugin by curl

1. 安装chart


2. 注册插件

```bash
curl http://${RUDDERADDR}/v1/plugin -d '{"id":"keel-echo","secret":"changeme"}'
```

输出应如下：
```bash
{"ret":0,"msg":"ok"}
```

3. 检查是否注册成功

```bash
curl http://${RUDDERADDR}/v1/plugin/get?id=keel-echo
```

输出应如下：
```bash
{
    "ret": 0,
    "msg": "ok",
    "data": {
        "ret": 0,
        "msg": "ok",
        "plugin_id": "keel-echo",
        "version": "v0.2.0",
        "tkeel_version": "v0.2.0",
        "secret": "changme",
        "register_time": 1633681152,
        "status": "ACTIVE",
        "register_addons": null
    }
}
```

## keel-echo output

访问端点**echo**。

查看**keel-echo**的日志。

```bash
echo

REQUEST_DATA
.
.
.

```