# Charts

**tkeel-platform**核心插件及每个**plugin**的**dapr**配置文件的**chart**安装包。

## Quickstart

* Pre-requisites:
   * [安装kubernetes](https://kubernetes.io/)
   * [安装dapr](https://dapr.io/)

1. 安装中间件

   ```bash
   cd ./tkeel-middleware
   helm -n keel-test install tkeel-mid .
   ```

   Expected output:
   ```
   NAME: tkeel-platform
   LAST DEPLOYED: Tue Oct 12 16:37:28 2021
   NAMESPACE: keel-test
   STATUS: deployed
   REVISION: 1
   TEST SUITE: None
   ```

2. 安装核心插件
   
   安装核心插件时，需要先安装**tkeel-plugin-component**并指定对应的插件ID，再安装对应的插件。

   1. 安装plugins,keel,auth核心插件:

   `PLUGIN_ID`取值列表
   - plugins
   - keel
   - auth

   ```bash
   cd ../tkeel-plugin-component
   helm -n keel-test --set pluginID=${PLUGIN_ID} install tkeel-pc-${PLUGIN_ID} .
   ```
   
   Expected output:
   ```
   NAME: tkeel-pc-plugins
   LAST DEPLOYED: Tue Oct 12 16:59:12 2021
   NAMESPACE: keel-test
   STATUS: deployed
   REVISION: 1
   TEST SUITE: None
   ```

   ```bash
   cd ../${PLUGIN_ID}
   helm -n keel-test install tkeel-${PLUGIN_ID} .
   ```

   Expected output:
   ```
   NAME: tkeel-plugins
   LAST DEPLOYED: Tue Oct 12 17:03:10 2021
   NAMESPACE: keel-test
   STATUS: deployed
   REVISION: 1
   TEST SUITE: None
   ```


## Tkeel-plugin-component

每个**plugin**的对应**dapr**边车必须拥有的配置文件。

### Values

* pluginID: 生成的插件的ID。
* plugins: **tkeel-platform**中核心插件**plugins**的相关配置信息。
   * port: **plugins**插件的**service**暴露的端口。
* secret: **tkeel-platform**的认证密钥，平台管理员管理持有。

## Tkeel-middleware

**tkeel-platform**的核心插件中所使用到的中间件资源。

### Values

* redis: **redis**子**chart**中所需要覆盖的变量。

## Plugins

核心插件**plugins**的**chart**。

### Values

* daprConfig: **tkeel-platform**中所有插件的**chart**必须包含此**value**，用于填充`Deployment.spec.template.metadata.annotations.dapr.io/config`字段。
* middleware: 核心插件所需中间件相关信息。
* replicaCount: 部署时副本数量。
* appPort: 指定与**dapr**边车交互的port。
* secret: 平台管理员所管理的平台密钥。
* image: **plugins**插件对应的镜像信息。

## Keel

核心插件**Keel**的**chart**。

### Values

* daprConfig: **tkeel-platform**中所有插件的**chart**必须包含此**value**，用于填充`Deployment.spec.template.metadata.annotations.dapr.io/config`字段。
* middleware: 核心插件所需中间件相关信息。
* replicaCount: 部署时副本数量。
* appPort: 指定与**dapr**边车交互的port。
* nodePort: 指定**keel**插件对应的**service**暴露的节点端口。
* image: **plugins**插件对应的镜像信息。

## Auth

核心插件**Auth**的**chart**。

### Values

* daprConfig: **tkeel-platform**中所有插件的**chart**必须包含此**value**，用于填充`Deployment.spec.template.metadata.annotations.dapr.io/config`字段。
* middleware: 核心插件所需中间件相关信息。
* replicaCount: 部署时副本数量。
* appPort: 指定与**dapr**边车交互的port。
* image: **plugins**插件对应的镜像信息。