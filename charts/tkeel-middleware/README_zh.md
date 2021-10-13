# keel-paas

核心插件和**redis**的安装**chart**

```yaml
# Default values for keel-paas.
# This is a YAML-formatted file.
# Declare variables to be passed into your templates.

replicaCount: 1

global:
  pluginPort: 8082
  pluginSecret: changeme
  privateStore: tkeel-redis-private-store
  publicStore: tkeel-redis-public-store

plugin-components:
  plugins:
  - name: plugins    
  - name: keel
  - name: auth

nameOverride: ""
fullnameOverride: ""

redis:
  architecture: standalone

redisPrivate:
  name: tkeel-redis-private-store
  addr: tkeel-redis-master:6379
  type: private
  passSecret: tkeel-redis
redisPublic:
  name: tkeel-redis-public-store
  addr: tkeel-redis-master:6379
  type: public
  passSecret: tkeel-redis
```

```bash
wget http://127.0.0.1:8080/register --post-data '{"id":"keel","secret":"changeme"}'
```