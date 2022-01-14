// Code generated by protoc-gen-go-errors. DO NOT EDIT.

package v1

import (
	errors "github.com/tkeel-io/kit/errors"
	codes "google.golang.org/grpc/codes"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the ego package it is being compiled against.
const _ = errors.SupportPackageIsVersion1

var oauthErrUnknown *errors.TError
var oauthErrInvalidRequest *errors.TError
var oauthErrUnauthorizedClient *errors.TError
var oauthErrAccessDenied *errors.TError
var oauthErrUnsupportedResponseType *errors.TError
var oauthErrInvalidScope *errors.TError
var oauthErrServerError *errors.TError
var oauthErrInvalidClient *errors.TError
var oauthErrInvalidGrant *errors.TError
var oauthErrUnsupportedGrantType *errors.TError

func init() {
	oauthErrUnknown = errors.New(int(codes.Unknown), "api.security_oauth.v1.OAUTH_ERR_UNKNOWN", Error_OAUTH_ERR_UNKNOWN.String())
	errors.Register(oauthErrUnknown)
	oauthErrInvalidRequest = errors.New(int(codes.InvalidArgument), "api.security_oauth.v1.OAUTH_ERR_INVALID_REQUEST", Error_OAUTH_ERR_INVALID_REQUEST.String())
	errors.Register(oauthErrInvalidRequest)
	oauthErrUnauthorizedClient = errors.New(int(codes.InvalidArgument), "api.security_oauth.v1.OAUTH_ERR_UNAUTHORIZED_CLIENT", Error_OAUTH_ERR_UNAUTHORIZED_CLIENT.String())
	errors.Register(oauthErrUnauthorizedClient)
	oauthErrAccessDenied = errors.New(int(codes.InvalidArgument), "api.security_oauth.v1.OAUTH_ERR_ACCESS_DENIED", Error_OAUTH_ERR_ACCESS_DENIED.String())
	errors.Register(oauthErrAccessDenied)
	oauthErrUnsupportedResponseType = errors.New(int(codes.InvalidArgument), "api.security_oauth.v1.OAUTH_ERR_UNSUPPORTED_RESPONSE_TYPE", Error_OAUTH_ERR_UNSUPPORTED_RESPONSE_TYPE.String())
	errors.Register(oauthErrUnsupportedResponseType)
	oauthErrInvalidScope = errors.New(int(codes.InvalidArgument), "api.security_oauth.v1.OAUTH_ERR_INVALID_SCOPE", Error_OAUTH_ERR_INVALID_SCOPE.String())
	errors.Register(oauthErrInvalidScope)
	oauthErrServerError = errors.New(int(codes.Internal), "api.security_oauth.v1.OAUTH_ERR_SERVER_ERROR", Error_OAUTH_ERR_SERVER_ERROR.String())
	errors.Register(oauthErrServerError)
	oauthErrInvalidClient = errors.New(int(codes.PermissionDenied), "api.security_oauth.v1.OAUTH_ERR_INVALID_CLIENT", Error_OAUTH_ERR_INVALID_CLIENT.String())
	errors.Register(oauthErrInvalidClient)
	oauthErrInvalidGrant = errors.New(int(codes.Internal), "api.security_oauth.v1.OAUTH_ERR_INVALID_GRANT", Error_OAUTH_ERR_INVALID_GRANT.String())
	errors.Register(oauthErrInvalidGrant)
	oauthErrUnsupportedGrantType = errors.New(int(codes.InvalidArgument), "api.security_oauth.v1.OAUTH_ERR_UNSUPPORTED_GRANT_TYPE", Error_OAUTH_ERR_UNSUPPORTED_GRANT_TYPE.String())
	errors.Register(oauthErrUnsupportedGrantType)
}

func OauthErrUnknown() errors.Error {
	return oauthErrUnknown
}

func OauthErrInvalidRequest() errors.Error {
	return oauthErrInvalidRequest
}

func OauthErrUnauthorizedClient() errors.Error {
	return oauthErrUnauthorizedClient
}

func OauthErrAccessDenied() errors.Error {
	return oauthErrAccessDenied
}

func OauthErrUnsupportedResponseType() errors.Error {
	return oauthErrUnsupportedResponseType
}

func OauthErrInvalidScope() errors.Error {
	return oauthErrInvalidScope
}

func OauthErrServerError() errors.Error {
	return oauthErrServerError
}

func OauthErrInvalidClient() errors.Error {
	return oauthErrInvalidClient
}

func OauthErrInvalidGrant() errors.Error {
	return oauthErrInvalidGrant
}

func OauthErrUnsupportedGrantType() errors.Error {
	return oauthErrUnsupportedGrantType
}