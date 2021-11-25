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
var errPluginNotFound *errors.TError
var errListPlugin *errors.TError
var errInvalidArgument *errors.TError
var errInternalStore *errors.TError

func init() {
	errUnknown = errors.New(int(codes.Unknown), "repo.v1.ERR_UNKNOWN", Error_ERR_UNKNOWN.String())
	errors.Register(errUnknown)
	errPluginNotFound = errors.New(int(codes.NotFound), "repo.v1.ERR_PLUGIN_NOT_FOUND", Error_ERR_PLUGIN_NOT_FOUND.String())
	errors.Register(errPluginNotFound)
	errListPlugin = errors.New(int(codes.Internal), "repo.v1.ERR_LIST_PLUGIN", Error_ERR_LIST_PLUGIN.String())
	errors.Register(errListPlugin)
	errInvalidArgument = errors.New(int(codes.InvalidArgument), "repo.v1.ERR_INVALID_ARGUMENT", Error_ERR_INVALID_ARGUMENT.String())
	errors.Register(errInvalidArgument)
	errInternalStore = errors.New(int(codes.Internal), "repo.v1.ERR_INTERNAL_STORE", Error_ERR_INTERNAL_STORE.String())
	errors.Register(errInternalStore)
}

func ErrUnknown() errors.Error {
	return errUnknown
}

func ErrPluginNotFound() errors.Error {
	return errPluginNotFound
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