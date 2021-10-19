# OPENAPI-001-plugin-api

## Status
Accepted

## Context
平台的租户机制其实仅仅是平台维护了一套租户的权限体系，仅保存了租户的账号之类的信息，因为对于平台而言，各种数据均有插件自己维护，所以租户之间的数据安全完全由插件自己维护起来，平台仅提供一个称为租户 ID 的唯一标识。

## Decision
为了实现上述目标，就通过 **OPENAPI** 来规定一套用于租户体系的交互接口，因为分配租户是平台进行分配的，所以平台可以查询到每个插件具体分配了哪些租户，故插件无需实现查询接口（后面根据具体情况可以增加）。

1. 给某租户启用插件时，平台将调用被启动插件的 `/${VERSION}/tenant/bind` 接口，并传入具体的租户信息，如ID和插件自定义信息等，插件应对传入的信息进行正确处理，如：初始化新的租户数据，允许新的租户请求等操作。

```bash
POST /${VERSION}/tenant/bind
```

request:
```json
{
	"tenant_id":"a",
	"tenant_name":"a_com",
	"expire_time":1653208631 // 当前版本未使用
    "extra":JSON_OBJ, // 插件自定义字段，json结构。
}
```
response:
```json
{
	"ret":0, // 或小于0
	"msg":"ok" // 额外信息
}
```

2. 给某租户停用插件时，平台将调用被停用插件的 `/${VERSION}/tenant/delete` 接口，并传入具体的租户信息，如 ID 和插件自定义信息等，插件应对传入的信息进行正确处理，如：删除或保留租户的数据，阻拦租户的请求等操作。

```bash
POST /${VERSION}/tenant/delete
```

request:
```json
{
	"tenant_id":"a",
	"tenant_name":"a_com",
    "extra":JSON_OBJ, // 插件自定义字段，json结构。
}
```
response:
```json
{
	"ret":0, // 或小于0
	"msg":"ok" // 额外信息
}
```

## Consequences
租户的请求暂定如此，如有更改将随 **OPENAPI** 版本更改。