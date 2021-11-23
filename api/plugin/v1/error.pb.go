// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.1
// source: api/plugin/v1/error.proto

package v1

import (
	reflect "reflect"
	sync "sync"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// @plugins=protoc-gen-go-errors
// 错误
type Error int32

const (
	// 未知类型
	// @code=UNKNOWN
	Error_ERR_UNKNOWN Error = 0
	// 找不到 Plugin
	// @code=NOT_FOUND
	Error_ERR_PLUGIN_NOT_FOUND Error = 1
	// 找不到 Plugin Route
	// @code=NOT_FOUND
	Error_ERR_PLUGIN_ROUTE_NOT_FOUND Error = 2
	// 找不到 Plugin Route
	// @code=ALREADY_EXISTS
	Error_ERR_PLUGIN_ALREADY_EXISTS Error = 3
	// 获取 Plugin 列表数据出错
	// @code=INTERNAL
	Error_ERR_LIST_PLUGIN Error = 4
	// 请求参数无效
	// @code=INVALID_ARGUMENT
	Error_ERR_INVALID_ARGUMENT Error = 5
	// 请求 Plugin OPENAPI 错误
	// @code=INTERNAL
	Error_ERR_INTERNAL_QUERY_PLUGIN_OPENAPI Error = 6
	// 请求后端存储错误
	// @code=INTERNAL
	Error_ERR_INTERNAL_STORE Error = 7
	// 删除的插件被依赖
	// @code=INTERNAL
	Error_ERR_DELETE_PLUGIN_HAS_BEEN_DEPENDED Error = 8
)

// Enum value maps for Error.
var (
	Error_name = map[int32]string{
		0: "ERR_UNKNOWN",
		1: "ERR_PLUGIN_NOT_FOUND",
		2: "ERR_PLUGIN_ROUTE_NOT_FOUND",
		3: "ERR_PLUGIN_ALREADY_EXISTS",
		4: "ERR_LIST_PLUGIN",
		5: "ERR_INVALID_ARGUMENT",
		6: "ERR_INTERNAL_QUERY_PLUGIN_OPENAPI",
		7: "ERR_INTERNAL_STORE",
		8: "ERR_DELETE_PLUGIN_HAS_BEEN_DEPENDED",
	}
	Error_value = map[string]int32{
		"ERR_UNKNOWN":                         0,
		"ERR_PLUGIN_NOT_FOUND":                1,
		"ERR_PLUGIN_ROUTE_NOT_FOUND":          2,
		"ERR_PLUGIN_ALREADY_EXISTS":           3,
		"ERR_LIST_PLUGIN":                     4,
		"ERR_INVALID_ARGUMENT":                5,
		"ERR_INTERNAL_QUERY_PLUGIN_OPENAPI":   6,
		"ERR_INTERNAL_STORE":                  7,
		"ERR_DELETE_PLUGIN_HAS_BEEN_DEPENDED": 8,
	}
)

func (x Error) Enum() *Error {
	p := new(Error)
	*p = x
	return p
}

func (x Error) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Error) Descriptor() protoreflect.EnumDescriptor {
	return file_api_plugin_v1_error_proto_enumTypes[0].Descriptor()
}

func (Error) Type() protoreflect.EnumType {
	return &file_api_plugin_v1_error_proto_enumTypes[0]
}

func (x Error) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Error.Descriptor instead.
func (Error) EnumDescriptor() ([]byte, []int) {
	return file_api_plugin_v1_error_proto_rawDescGZIP(), []int{0}
}

var File_api_plugin_v1_error_proto protoreflect.FileDescriptor

var file_api_plugin_v1_error_proto_rawDesc = []byte{
	0x0a, 0x19, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2f, 0x76, 0x31, 0x2f,
	0x65, 0x72, 0x72, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x70, 0x6c, 0x75,
	0x67, 0x69, 0x6e, 0x2e, 0x76, 0x31, 0x2a, 0x88, 0x02, 0x0a, 0x05, 0x45, 0x72, 0x72, 0x6f, 0x72,
	0x12, 0x0f, 0x0a, 0x0b, 0x45, 0x52, 0x52, 0x5f, 0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10,
	0x00, 0x12, 0x18, 0x0a, 0x14, 0x45, 0x52, 0x52, 0x5f, 0x50, 0x4c, 0x55, 0x47, 0x49, 0x4e, 0x5f,
	0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x10, 0x01, 0x12, 0x1e, 0x0a, 0x1a, 0x45,
	0x52, 0x52, 0x5f, 0x50, 0x4c, 0x55, 0x47, 0x49, 0x4e, 0x5f, 0x52, 0x4f, 0x55, 0x54, 0x45, 0x5f,
	0x4e, 0x4f, 0x54, 0x5f, 0x46, 0x4f, 0x55, 0x4e, 0x44, 0x10, 0x02, 0x12, 0x1d, 0x0a, 0x19, 0x45,
	0x52, 0x52, 0x5f, 0x50, 0x4c, 0x55, 0x47, 0x49, 0x4e, 0x5f, 0x41, 0x4c, 0x52, 0x45, 0x41, 0x44,
	0x59, 0x5f, 0x45, 0x58, 0x49, 0x53, 0x54, 0x53, 0x10, 0x03, 0x12, 0x13, 0x0a, 0x0f, 0x45, 0x52,
	0x52, 0x5f, 0x4c, 0x49, 0x53, 0x54, 0x5f, 0x50, 0x4c, 0x55, 0x47, 0x49, 0x4e, 0x10, 0x04, 0x12,
	0x18, 0x0a, 0x14, 0x45, 0x52, 0x52, 0x5f, 0x49, 0x4e, 0x56, 0x41, 0x4c, 0x49, 0x44, 0x5f, 0x41,
	0x52, 0x47, 0x55, 0x4d, 0x45, 0x4e, 0x54, 0x10, 0x05, 0x12, 0x25, 0x0a, 0x21, 0x45, 0x52, 0x52,
	0x5f, 0x49, 0x4e, 0x54, 0x45, 0x52, 0x4e, 0x41, 0x4c, 0x5f, 0x51, 0x55, 0x45, 0x52, 0x59, 0x5f,
	0x50, 0x4c, 0x55, 0x47, 0x49, 0x4e, 0x5f, 0x4f, 0x50, 0x45, 0x4e, 0x41, 0x50, 0x49, 0x10, 0x06,
	0x12, 0x16, 0x0a, 0x12, 0x45, 0x52, 0x52, 0x5f, 0x49, 0x4e, 0x54, 0x45, 0x52, 0x4e, 0x41, 0x4c,
	0x5f, 0x53, 0x54, 0x4f, 0x52, 0x45, 0x10, 0x07, 0x12, 0x27, 0x0a, 0x23, 0x45, 0x52, 0x52, 0x5f,
	0x44, 0x45, 0x4c, 0x45, 0x54, 0x45, 0x5f, 0x50, 0x4c, 0x55, 0x47, 0x49, 0x4e, 0x5f, 0x48, 0x41,
	0x53, 0x5f, 0x42, 0x45, 0x45, 0x4e, 0x5f, 0x44, 0x45, 0x50, 0x45, 0x4e, 0x44, 0x45, 0x44, 0x10,
	0x08, 0x42, 0x5e, 0x0a, 0x1e, 0x64, 0x65, 0x76, 0x2e, 0x74, 0x6b, 0x65, 0x65, 0x6c, 0x2e, 0x72,
	0x75, 0x64, 0x64, 0x65, 0x72, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e,
	0x2e, 0x76, 0x31, 0x42, 0x0e, 0x4f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x50, 0x72, 0x6f, 0x74,
	0x6f, 0x56, 0x31, 0x50, 0x01, 0x5a, 0x2a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x74, 0x6b, 0x65, 0x65, 0x6c, 0x2d, 0x69, 0x6f, 0x2f, 0x74, 0x6b, 0x65, 0x65, 0x6c,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2f, 0x76, 0x31, 0x3b, 0x76,
	0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_plugin_v1_error_proto_rawDescOnce sync.Once
	file_api_plugin_v1_error_proto_rawDescData = file_api_plugin_v1_error_proto_rawDesc
)

func file_api_plugin_v1_error_proto_rawDescGZIP() []byte {
	file_api_plugin_v1_error_proto_rawDescOnce.Do(func() {
		file_api_plugin_v1_error_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_plugin_v1_error_proto_rawDescData)
	})
	return file_api_plugin_v1_error_proto_rawDescData
}

var file_api_plugin_v1_error_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_api_plugin_v1_error_proto_goTypes = []interface{}{
	(Error)(0), // 0: plugin.v1.Error
}
var file_api_plugin_v1_error_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_api_plugin_v1_error_proto_init() }
func file_api_plugin_v1_error_proto_init() {
	if File_api_plugin_v1_error_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_plugin_v1_error_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_plugin_v1_error_proto_goTypes,
		DependencyIndexes: file_api_plugin_v1_error_proto_depIdxs,
		EnumInfos:         file_api_plugin_v1_error_proto_enumTypes,
	}.Build()
	File_api_plugin_v1_error_proto = out.File
	file_api_plugin_v1_error_proto_rawDesc = nil
	file_api_plugin_v1_error_proto_goTypes = nil
	file_api_plugin_v1_error_proto_depIdxs = nil
}
