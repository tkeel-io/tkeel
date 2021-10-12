[auth api列表](./README.md)

---
### 用户认证

###### 接口功能

>  token获取用户信息，可用来认证token

###### HTTP请求方式

> POST

###### URL

>  /authenticate

###### 支持格式

> JSON


###### 请求参数

> | 参数  | 必选 | 类型   | 说明      |
> | :---- | :--- | :----- | --------- |
> | token | ture | string | 认证Token |

###### 返回数据(data)

> | 返回字段  | 字段类型 | 说明   |
> | :-------- | :------- | :----- |
> | user_id   | string   | 用户ID |
> | tenant_id | string   | 租户ID |
> | name      | string   | 用户名 |
> | email     | string   | 邮箱   |

###### 接口示例

```
POST /auth/authenticate HTTP/1.1
Host: 192.168.123.2:30777
Content-Type: application/json
Content-Length: 494

{
 "token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJrZWVsIiwiZXhwIjoiMjAyMS0xMC0xMlQxNToyMDoxNS44MDg5Mjc4MDdaIiwiaWF0IjoiMjAyMS0xMC0xMlQwMzoyMDoxNS44MDg5Mjc4MDdaIiwiaXNzIjoibWFuYWdlciIsImp0aSI6IjgxM2Y0MTRmLWMyMTUtNDA3Mi04NmY2LTEwNDJkYzgzMjhhMCIsIm5iZiI6IjIwMjEtMTAtMTJUMDM6MjA6MTUuODA4OTI3ODA3WiIsInN1YiI6InVzZXIiLCJ0aWQiOiJlMWNjMmU5OC1jYmFjLTRiOWItOThmYi1kNTVmYzQ0Njg3ZTUiLCJ1aWQiOiIyMDliNDU2Yy0yNTE5LTQ2NTctOTJjYi0zNmI4MmIwMDdkYTQifQ.1_o4PXgp8nGz9UgZ0BsbfBUw-1vsES7_oH012FitHsg"
}
```

``` json
{
    "ret": 0,
    "msg": "ok",
    "data": {
        "user_id": "209b456c-2519-4657-92cb-36b82b007da4",
        "tenant_id": "e1cc2e98-cbac-4b9b-98fb-d55fc44687e5",
        "name": "tenant11Admin",
        "email": ""
    }
}
```