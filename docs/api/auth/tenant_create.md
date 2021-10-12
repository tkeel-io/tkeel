[auth api列表](./README.md)

---
### 创建租户
###### 接口功能
> 创建租户 [系统管理员]

###### HTTP请求方式
> POST
###### URL
>  /tenant/create
###### 支持格式
> JSON


###### 请求参数
> |参数|必选|类型|说明|
> |:-----  |:-------|:-----|-----                               |
> |title    |ture    |string|租户标题/租户名                          |
> |email    |    |string   |邮箱|
> |phone | |string |联系电话|
> |country | |string |国家|
> |city | |string |城市|
> |address | |string |地址|

###### 返回数据(data)
> |返回字段|字段类型|说明                              |
> |:-----   |:------|:-----------------------------   |
> |tenant_id   |string    |租户ID   |
> |title |string |租户标题/租户名 |
> |created_time |int64 |创建时间 |
> |tenant_admin |json |租户管理员,详细见下表格TenantAdmin |

| TenantAdmin  |        |                      |
| ------------ | ------ | -------------------- |
| id           | string | 租户管理员ID(用户ID) |
| name         | string | 用户名               |
| password     | string | 用户密码             |
| tenant_id    | stirng | 租户ID               |
| email        | string | 用户邮箱             |
| created_time | int64  | 创建时间             |



###### 接口示例

```
POST /auth/tenant/create HTTP/1.1
Content-Type: application/json
Content-Length: 59

{
   "title":"tenant11",
   "email":"tenant1@tkeel.io"
}
```

``` json
{
    "ret": 0,
    "msg": "ok",
    "data": {
        "tenant_id": "e1cc2e98-cbac-4b9b-98fb-d55fc44687e5",
        "title": "tenant11",
        "created_time": 1634008150,
        "tenant_admin": {
            "id": "209b456c-2519-4657-92cb-36b82b007da4",
            "name": "tenant11Admin",
            "password": "admin",
            "tenant_id": "e1cc2e98-cbac-4b9b-98fb-d55fc44687e5",
            "email": "",
            "create_time": 1634008150
        }
    }
}
```