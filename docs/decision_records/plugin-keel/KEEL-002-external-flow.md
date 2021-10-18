# KEEL-002-external-flow

## Status
Implemented

## Context
在平台中需要对外部流量进行对应的转发，且对外部请求进行认证和授权处理，在 **Keel** 中预留外部请求的扩展点，可以通过其他的插件来实现这个地方的检查。

## Decision
借用平台的能力来实现的检查点的扩展。

**Keel** 的 `/v1/identify` 中的接口返回值如下：
```go
openapi.IdentifyResp{
	CommonResult: openapi.SuccessResult(),
	PluginID:     k.p.GetIdentifyResp().PluginID,
	Version:      k.p.GetIdentifyResp().Version,
	AddonsPoints: []*openapi.AddonsPoint{
		{
			AddonsPoint: "externalPreRouteCheck",
			Desc: `
			callback before external flow routing
			input request header and path
			output http statecode
			200   -- allow
			other -- deny
			`,
		},
	},
}
```

扩展点检查逻辑为：
```text
Check whether the endpoint correctly implements this callback:
For example, when the request header contains the "x-keel-check" field, 
the HTTP request header 200 is returned. When the field value is "True", 
the body is 
{	
	"msg":"ok",
	"ret":0
}, 
When it is False, the body is 
{
	"msg":"faild",
	"ret":-1
}. 
If it is not included, it will judge whether the request is valid.
```

当插件注册时及运行以上逻辑进行检查扩展点是否实现。

## Consequences
检查点需要根据AUTH的用户认证方式进行处理。