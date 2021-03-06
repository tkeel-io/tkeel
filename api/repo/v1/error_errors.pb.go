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
var errRepoNotFound *errors.TError
var errListPlugin *errors.TError
var errInvalidArgument *errors.TError
var errInternalStore *errors.TError
var errInternalError *errors.TError
var errRepoExist *errors.TError
var errInstallerNotFound *errors.TError

func init() {
	errUnknown = errors.New(int(codes.Unknown), "io.tkeel.plugin.api.repo.v1.ERR_UNKNOWN", "未知类型")
	errors.Register(errUnknown)
	errRepoNotFound = errors.New(int(codes.NotFound), "io.tkeel.plugin.api.repo.v1.ERR_REPO_NOT_FOUND", "找不到REPO")
	errors.Register(errRepoNotFound)
	errListPlugin = errors.New(int(codes.Internal), "io.tkeel.plugin.api.repo.v1.ERR_LIST_PLUGIN", "获取REPO列表数据出错")
	errors.Register(errListPlugin)
	errInvalidArgument = errors.New(int(codes.InvalidArgument), "io.tkeel.plugin.api.repo.v1.ERR_INVALID_ARGUMENT", "请求参数无效")
	errors.Register(errInvalidArgument)
	errInternalStore = errors.New(int(codes.Internal), "io.tkeel.plugin.api.repo.v1.ERR_INTERNAL_STORE", "请求后端存储错误")
	errors.Register(errInternalStore)
	errInternalError = errors.New(int(codes.Internal), "io.tkeel.plugin.api.repo.v1.ERR_INTERNAL_ERROR", "内部错误")
	errors.Register(errInternalError)
	errRepoExist = errors.New(int(codes.InvalidArgument), "io.tkeel.plugin.api.repo.v1.ERR_REPO_EXIST", "REPO已存在")
	errors.Register(errRepoExist)
	errInstallerNotFound = errors.New(int(codes.NotFound), "io.tkeel.plugin.api.repo.v1.ERR_INSTALLER_NOT_FOUND", "INSTALLER不存在")
	errors.Register(errInstallerNotFound)
}

func ErrUnknown() errors.Error {
	return errUnknown
}

func ErrRepoNotFound() errors.Error {
	return errRepoNotFound
}

func ErrListPlugin() errors.Error {
	return errListPlugin
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

func ErrRepoExist() errors.Error {
	return errRepoExist
}

func ErrInstallerNotFound() errors.Error {
	return errInstallerNotFound
}
