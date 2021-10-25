# auth API列表
外部 http 调用 HEADER 信息需携带 Authorization 字段，值为登录成功返回的access token。

>   请求参数类型说明：
>
>   body: 参数通过 http body 传输。如：[POST] /api/v1/some    -d `{"xxx":"xxx"}`
>
>   query: 参数通过url地址携带查询参数传输。如：[GET] /api/v1/some?**condition=XXX**
>
>   patch: 参数通过url patch传输。如：[DELETE] /api/v1/some/**XXX**

## 租户
> 面向系统管理员

[租户创建](tenant_create.md)

[租户列表](tenant_list.md)

[租户删除](tenant_delete.md)

## 用户 

[用户登录](login.md)

[用户列表](user_list.md)

[用户创建](user_create.md)

[用户认证](oauth_authenticate.md)

[用户角色添加](user_role_enable.md)

[用户角色删除](user_role_disable.md)

## 角色

[角色创建](role_create.md)

[角色删除](role_delete.md)

[角色列表](role_list.md)

[角色权限添加](role_permission_create.md)

[角色权限删除](role_permission_delete.md)

[角色权限列表](role_permission_list.md)

## 实体

[实体Token创建](entity_token_create.md)

[实体Token解析](entity_token_parse.md)

## OAuth2.0

[oauth/authorize](oauth_authorize.md)

[oauth/token](oauth_token.md)

[oauth/authenticate](oauth_authenticate.md)

