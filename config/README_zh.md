# Config

本文档介绍了`keel`核心搭建起来所需要的配置文件，包含`dapr`和**k8s**中所需要的配置文件。

## 必要条件

1. k8s的**namespace**下面必须初始化过**dapr**环境(运行过`dapr init`)。
2. 必须拥有可访问的`redis`的`StatefulSet`，将访问路径修改至example/components/中文件的对应位置。

## 安装插件

```bash
## 安装manager
kubectl -n ${namespace} -f example/components/
kubectl -n ${namespace} -f plugin/manager/
kubectl -n ${namespace} -f plugin/manager/components/
kubectl -n ${namespace} -f plugin/manager/k8s/

## 查看kmanager中对应容器日志，如发生错误 
## error register plugin manager: error json Unmarshal(): unexpected end of JSON input
## 则进入容器运行以下命令
wget --post-data '{"id":"kmanager","secret":"zhu88jie"}' http://127.0.0.1:8080/plugins
## 或访问对应service的外网暴露的port
wget --post-data '{"id":"${costom_plugin}","secret":"zhu88jie"}' http://${node_ip}:${port}/plugins 
## 返回OK则安装完成
```

```bash
## 安装gateway
##如以成功安装manager则省略下一句
##---------
kubectl -n ${namespace} -f example/components/
##---------
kubectl -n ${namespace} -f plugin/gateway/
kubectl -n ${namespace} -f plugin/gateway/components/
kubectl -n ${namespace} -f plugin/gateway/k8s/

## 进入manager容器运行以下命令 注册gateway
wget --post-data '{"id":"keel","secret":"zhu88jie"}' http://127.0.0.1:8080/plugins
## 或访问对应service的外网暴露的port
wget --post-data '{"id":"${costom_plugin}","secret":"zhu88jie"}' http://${node_ip}:${port}/plugins 
```

安装自定义插件时，拷贝plugin下任一目录，递归修改其下所有配置文件，将插件名和对应文件名修改。

`config.yaml`:
* `metadata.name`
* `spec.httpPipeLine.handlers[0].name`

`components/*_oauth2.yaml`:
* `metadata.name`
* `spec.metadata[0].value`

`k8s/deployment.yaml`:
* `metadata.name`
* `metadata.labels.app`
* `spec.selector.matchLabels.app`
* `spec.template.metadata.labels.app`
* `spec.template.metadata.annotations.dapr.io/app-id`
* `spec.template.spec.containers[0].name`
* `spec.template.spec.containers[0].image`

`k8s/service.yaml`:
* `metadata.name`
* `spec.selector.app`

```bash
## 安装自定义插件
##如以成功安装manager则省略下一句
##---------
kubectl -n ${namespace} -f example/components/
##---------
kubectl -n ${namespace} -f plugin/${costom_plugin}/
kubectl -n ${namespace} -f plugin/${costom_plugin}/components/
kubectl -n ${namespace} -f plugin/${costom_plugin}/k8s/

## 进入manager容器运行以下命令
wget --post-data '{"id":"${costom_plugin}","secret":"zhu88jie"}' http://127.0.0.1:8080/plugins 
## 或访问对应service的外网暴露的port
wget --post-data '{"id":"${costom_plugin}","secret":"zhu88jie"}' http://${node_ip}:${port}/plugins 
```
