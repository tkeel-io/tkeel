[auth api列表](./README.md)

---
### 解析实体Token

###### 接口功能

> 解析实体Token

###### HTTP请求方式

> POST

###### URL

>  /token/parse

###### 支持格式

> JSON

###### 请求参数

> | 参数         | 必选 | 类型 | 说明      |
> | :----------- | :--- | :--- | --------- |
> | entity_token | ture | body | 实体Token |

###### 返回数据(data)

> | 返回字段    | 字段类型 | 说明     |
> | :---------- | :------- | :------- |
> | user_id     | string   | 用户ID   |
> | tenant_id   | string   | 租户ID   |
> | token_id    | string   | TokenID  |
> | entity_id   | string   | 实体ID   |
> | entity_type | string   | 实体类型 |

###### 接口示例

```
POST /auth/token/parse HTTP/1.1
Host: 192.168.123.2:30777
Content-Type: application/json
Content-Length: 678

{
   "entity_token":"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJrZWVsIiwiZWlkIjoiZGV2aWNlaWQxMjMiLCJleHAiOiIyMDIyLTEwLTEyVDAzOjQzOjQ2LjQ4NjQzNDQzNloiLCJpYXQiOiIyMDIxLTEwLTEyVDAzOjQzOjQ2LjQ4NjQzNDQzNloiLCJpc3MiOiJtYW5hZ2VyIiwianRpIjoiMTNhMTU1ZDEtNDUyYS00YzY5LThhZmMtMjg2YjE0NWUwM2M2IiwibmJmIjoiMjAyMS0xMC0xMlQwMzo0Mzo0Ni40ODY0MzQ0MzZaIiwic3ViIjoiZW50aXR5IiwidGlkIjoiZTFjYzJlOTgtY2JhYy00YjliLTk4ZmItZDU1ZmM0NDY4N2U1IiwidHlwIjoiZGV2aWNlIiwidWlkIjoiMjA5YjQ1NmMtMjUxOS00NjU3LTkyY2ItMzZiODJiMDA3ZGE0In0.mxIMTlEZH51ysA9gxDevoSBFWBDPI18Y8zORZC8-WkKAH1XNJTgLE72q6vAGIkItXlgSH3ElXaGDps_HPxQzvtTjrxxPc1s2dfh1AZTIKErKvGDrK489ZY3FO3ui8doPgLmRHbZHtQGTUyDyHzYsGEbp7NmQbsj32fx6AwJVSL0"
}
```

``` json
{
    "ret": 0,
    "msg": "ok",
    "data": {
        "user_id": "209b456c-2519-4657-92cb-36b82b007da4",
        "tenant_id": "e1cc2e98-cbac-4b9b-98fb-d55fc44687e5",
        "token_id": "13a155d1-452a-4c69-8afc-286b145e03c6",
        "entity_type": "device",
        "entity_id": "deviceid123"
    }
}
```