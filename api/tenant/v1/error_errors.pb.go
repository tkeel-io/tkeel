// Code generated by protoc-gen-go-errors. DO NOT EDIT.

package v1

import (
	errors "github.com/tkeel-io/kit/errors"
	codes "google.golang.org/grpc/codes"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the ego package it is being compiled against.
const _ = errors.SupportPackageIsVersion1

var errUnknown *errors.TError
var errTenantAlreadyExisted *errors.TError
var errListTenant *errors.TError
var errInvalidArgument *errors.TError
var errInternalStore *errors.TError
var errInternalError *errors.TError
var errStoreCreatTenant *errors.TError
var errAlreadyExistedUser *errors.TError
var errResourceNotFound *errors.TError
var errStoreCreatAdmin *errors.TError
var errStoreCreatAdminRole *errors.TError

func init() {
	errUnknown = errors.New(int(codes.Unknown), "io.tkeel.security.api.tenant.v1.ERR_UNKNOWN", "未知类型")
	errors.Register(errUnknown)
	errTenantAlreadyExisted = errors.New(int(codes.Internal), "io.tkeel.security.api.tenant.v1.ERR_TENANT_ALREADY_EXISTED", "已存在的租户")
	errors.Register(errTenantAlreadyExisted)
	errListTenant = errors.New(int(codes.Internal), "io.tkeel.security.api.tenant.v1.ERR_LIST_TENANT", "获取租户列表数据出错")
	errors.Register(errListTenant)
	errInvalidArgument = errors.New(int(codes.InvalidArgument), "io.tkeel.security.api.tenant.v1.ERR_INVALID_ARGUMENT", "请求参数无效")
	errors.Register(errInvalidArgument)
	errInternalStore = errors.New(int(codes.Internal), "io.tkeel.security.api.tenant.v1.ERR_INTERNAL_STORE", "请求后端存储错误")
	errors.Register(errInternalStore)
	errInternalError = errors.New(int(codes.Internal), "io.tkeel.security.api.tenant.v1.ERR_INTERNAL_ERROR", "内部错误")
	errors.Register(errInternalError)
	errStoreCreatTenant = errors.New(int(codes.Internal), "io.tkeel.security.api.tenant.v1.ERR_STORE_CREAT_TENANT", "创建租户错误")
	errors.Register(errStoreCreatTenant)
	errAlreadyExistedUser = errors.New(int(codes.Internal), "io.tkeel.security.api.tenant.v1.ERR_ALREADY_EXISTED_USER_", "创建已存在的用户")
	errors.Register(errAlreadyExistedUser)
	errResourceNotFound = errors.New(int(codes.Internal), "io.tkeel.security.api.tenant.v1.ERR_RESOURCE_NOT_FOUND", "资源不存在")
	errors.Register(errResourceNotFound)
	errStoreCreatAdmin = errors.New(int(codes.Internal), "io.tkeel.security.api.tenant.v1.ERR_STORE_CREAT_ADMIN", "创建租户管理员用户错误")
	errors.Register(errStoreCreatAdmin)
	errStoreCreatAdminRole = errors.New(int(codes.Internal), "io.tkeel.security.api.tenant.v1.ERR_STORE_CREAT_ADMIN_ROLE", "创建租户管理员角色错误")
	errors.Register(errStoreCreatAdminRole)
}

func ErrUnknown() errors.Error {
	return errUnknown
}

func ErrTenantAlreadyExisted() errors.Error {
	return errTenantAlreadyExisted
}

func ErrListTenant() errors.Error {
	return errListTenant
}

func ErrInvalidArgument() errors.Error {
	return errInvalidArgument
}

func ErrInternalStore() errors.Error {
	return errInternalStore
}

func ErrInternalError() errors.Error {
	return errInternalError
}

func ErrStoreCreatTenant() errors.Error {
	return errStoreCreatTenant
}

func ErrAlreadyExistedUser() errors.Error {
	return errAlreadyExistedUser
}

func ErrResourceNotFound() errors.Error {
	return errResourceNotFound
}

func ErrStoreCreatAdmin() errors.Error {
	return errStoreCreatAdmin
}

func ErrStoreCreatAdminRole() errors.Error {
	return errStoreCreatAdminRole
}
