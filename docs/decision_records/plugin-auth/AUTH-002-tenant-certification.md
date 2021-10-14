# AUTH-002-tenant-certfication

## Status
Implemented

## Context
平台的租户系统是一个开放式的约定性质的租户系统，平台仅要求每个注册进来的插件需要维护每个插件自己的租户信息，所以需要平台来管理整个租户，包含创建、查询、删除等功能。

## Decision

### 多租户
 用户系统设计为多租户体系,租户由系统管理员创建，可以理解租户为一个用户组，租户内可创建用户，角色。
 ![多租户](../../images/img/auth/tenant.png)
### 基于角色权限控制
 租户内可设置角色做插件的权限控制，如下
 ![基于角色权限控制](../../images/img/auth/rbac.png)