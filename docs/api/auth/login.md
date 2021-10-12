---

### 用户登录
###### 接口功能
> 用户登录获取 token

###### HTTP请求方式
> POST
###### URL
>  /user/login
###### 支持格式
> JSON


###### 请求参数
> |参数|必选|类型|说明|
|:-----  |:-------|:-----|-----                               |
|username    |ture    |string|用户名                          |
|password    |true    |string   |密码|

###### 返回数据(data)
> |返回字段|字段类型|说明                              |
|:-----   |:------|:-----------------------------   |
|token   |string    |用户身份令牌   |

###### 接口示例

```
POST /auth/user/login HTTP/1.1
Host: 192.168.123.2:30777
Content-Type: application/json
Content-Length: 61
{
    "username":"tenant11Admin",
    "password":"admin"
}
```

``` json
{
    "ret": 0,
    "msg": "ok",
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJrZWVsIiwiZXhwIjoiMjAyMS0xMC0xMlQxNToyMDoxNS44MDg5Mjc4MDdaIiwiaWF0IjoiMjAyMS0xMC0xMlQwMzoyMDoxNS44MDg5Mjc4MDdaIiwiaXNzIjoibWFuYWdlciIsImp0aSI6IjgxM2Y0MTRmLWMyMTUtNDA3Mi04NmY2LTEwNDJkYzgzMjhhMCIsIm5iZiI6IjIwMjEtMTAtMTJUMDM6MjA6MTUuODA4OTI3ODA3WiIsInN1YiI6InVzZXIiLCJ0aWQiOiJlMWNjMmU5OC1jYmFjLTRiOWItOThmYi1kNTVmYzQ0Njg3ZTUiLCJ1aWQiOiIyMDliNDU2Yy0yNTE5LTQ2NTctOTJjYi0zNmI4MmIwMDdkYTQifQ.1_o4PXgp8nGz9UgZ0BsbfBUw-1vsES7_oH012FitHsg"
    }
}
```