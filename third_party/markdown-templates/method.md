---
title: "{{.OperationID}}"
description: '{{.Summary}}'
---
{{$definitions := .Definitions}}

调用该接口{{.Summary}}。

{{.Description}}

## Request


```
{{.Operation}} {{.Path}}
```

{{with $paths := FilterParameters .Parameters "path"}}

| Name | Located in | Type | Description | 
| ---- | ---------- | ----------- | ----------- | {{range $param := $paths}}
| {{$param.Name}} | path | {{$param.Type}} | {{$param.Description}} |  {{end}}{{end}}

{{with $queries := FilterParameters .Parameters "query"}}

###  Request Parameters

| Name | Located in | Type | Description |  Required |
| ---- | ---------- | ----------- | ----------- |  ---- |{{range $param := $queries}}
| {{$param.Name}} | query | {{$param.Type}} | {{$param.Description}} |  {{$param.Required}} |{{end}}{{end}}

{{with $bodies := FilterParameters .Parameters "body"}}

### Request Body

{{range $resp := $bodies}}
{{if eq $resp.Type  "array" }}   
| Description | Type | Schema |
| ----------- | ------ | ------ |
| {{$resp.Description}} | Array | [{{FilterSchema $resp.Schema.Items.Ref}}](#{{FilterSchema $resp.Items.Ref}}) |

#### {{FilterSchema $resp.Items.Ref}}

{{template "schema.md" CollectSchema $definitions  $resp.Items.Ref}}
{{else}} 
| Description | Type | Schema |
| ----------- | ------ | ------ |
| {{$resp.Description}} | Object | [{{FilterSchema $resp.Schema.Ref}}](#{{FilterSchema $resp.Schema.Ref}}) |

#### {{FilterSchema $resp.Schema.Ref}}

{{template "schema.md" CollectSchema $definitions  $resp.Schema.Ref}}
{{end}}
{{end}}{{end}}

## Response

{{range $code, $resp := .Responses}}

### Response  {{$code}}

{{if ne $resp.Schema.Items.Ref  "" }}   
| Code1 | Description | Type | Schema |
| ---- | ----------- | ------ | ------ |
| {{$code}} | {{$resp.Description}} | Array | [{{FilterSchema $resp.Schema.Items.Ref}}](#{{FilterSchema $resp.Schema.Items.Ref}}) |

#### {{FilterSchema $resp.Schema.Items.Ref}}

{{template "schema.md" CollectSchema $definitions  $resp.Schema.Items.Ref}}
{{else if ne $resp.Schema.Ref  "" }} 
| Code2 | Description | Type | Schema |
| ---- | ----------- | ------ | ------ |
| {{$code}} | {{$resp.Description}} | Object | [{{FilterSchema $resp.Schema.Ref}}](#{{FilterSchema $resp.Schema.Ref}}) |

#### {{FilterSchema $resp.Schema.Ref}}

{{template "schema.md" CollectSchema $definitions  $resp.Schema.Ref}}
{{else}}
| Code3 | Description | Type | 
| ---- | ----------- | ------ | 
| {{$code}} | {{$resp.Description}} | {{$resp.Schema}} |
{{end}} 
{{end}}

