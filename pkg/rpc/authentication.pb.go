// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.12
// source: pkg/grpc/ctrlserver/authentication.proto

package rpc

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type JoinAuthMeshRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MeshId       string `protobuf:"bytes,1,opt,name=meshId,proto3" json:"meshId,omitempty"`
	SharedSecret string `protobuf:"bytes,2,opt,name=sharedSecret,proto3" json:"sharedSecret,omitempty"`
}

func (x *JoinAuthMeshRequest) Reset() {
	*x = JoinAuthMeshRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_ctrlserver_authentication_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *JoinAuthMeshRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JoinAuthMeshRequest) ProtoMessage() {}

func (x *JoinAuthMeshRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_ctrlserver_authentication_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JoinAuthMeshRequest.ProtoReflect.Descriptor instead.
func (*JoinAuthMeshRequest) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_ctrlserver_authentication_proto_rawDescGZIP(), []int{0}
}

func (x *JoinAuthMeshRequest) GetMeshId() string {
	if x != nil {
		return x.MeshId
	}
	return ""
}

func (x *JoinAuthMeshRequest) GetSharedSecret() string {
	if x != nil {
		return x.SharedSecret
	}
	return ""
}

type JoinAuthMeshReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success bool    `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Token   *string `protobuf:"bytes,2,opt,name=token,proto3,oneof" json:"token,omitempty"`
}

func (x *JoinAuthMeshReply) Reset() {
	*x = JoinAuthMeshReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_ctrlserver_authentication_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *JoinAuthMeshReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JoinAuthMeshReply) ProtoMessage() {}

func (x *JoinAuthMeshReply) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_ctrlserver_authentication_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JoinAuthMeshReply.ProtoReflect.Descriptor instead.
func (*JoinAuthMeshReply) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_ctrlserver_authentication_proto_rawDescGZIP(), []int{1}
}

func (x *JoinAuthMeshReply) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *JoinAuthMeshReply) GetToken() string {
	if x != nil && x.Token != nil {
		return *x.Token
	}
	return ""
}

var File_pkg_grpc_ctrlserver_authentication_proto protoreflect.FileDescriptor

var file_pkg_grpc_ctrlserver_authentication_proto_rawDesc = []byte{
	0x0a, 0x28, 0x70, 0x6b, 0x67, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x63, 0x74, 0x72, 0x6c, 0x73,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x72, 0x70, 0x63, 0x74,
	0x79, 0x70, 0x65, 0x73, 0x22, 0x51, 0x0a, 0x13, 0x4a, 0x6f, 0x69, 0x6e, 0x41, 0x75, 0x74, 0x68,
	0x4d, 0x65, 0x73, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x6d,
	0x65, 0x73, 0x68, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x65, 0x73,
	0x68, 0x49, 0x64, 0x12, 0x22, 0x0a, 0x0c, 0x73, 0x68, 0x61, 0x72, 0x65, 0x64, 0x53, 0x65, 0x63,
	0x72, 0x65, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x73, 0x68, 0x61, 0x72, 0x65,
	0x64, 0x53, 0x65, 0x63, 0x72, 0x65, 0x74, 0x22, 0x52, 0x0a, 0x11, 0x4a, 0x6f, 0x69, 0x6e, 0x41,
	0x75, 0x74, 0x68, 0x4d, 0x65, 0x73, 0x68, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x18, 0x0a, 0x07,
	0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73,
	0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12, 0x19, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x88, 0x01,
	0x01, 0x42, 0x08, 0x0a, 0x06, 0x5f, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x32, 0x5a, 0x0a, 0x0e, 0x41,
	0x75, 0x74, 0x68, 0x65, 0x6e, 0x74, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x48, 0x0a,
	0x08, 0x4a, 0x6f, 0x69, 0x6e, 0x4d, 0x65, 0x73, 0x68, 0x12, 0x1d, 0x2e, 0x72, 0x70, 0x63, 0x74,
	0x79, 0x70, 0x65, 0x73, 0x2e, 0x4a, 0x6f, 0x69, 0x6e, 0x41, 0x75, 0x74, 0x68, 0x4d, 0x65, 0x73,
	0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x72, 0x70, 0x63, 0x74, 0x79,
	0x70, 0x65, 0x73, 0x2e, 0x4a, 0x6f, 0x69, 0x6e, 0x41, 0x75, 0x74, 0x68, 0x4d, 0x65, 0x73, 0x68,
	0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x42, 0x09, 0x5a, 0x07, 0x70, 0x6b, 0x67, 0x2f, 0x72,
	0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_grpc_ctrlserver_authentication_proto_rawDescOnce sync.Once
	file_pkg_grpc_ctrlserver_authentication_proto_rawDescData = file_pkg_grpc_ctrlserver_authentication_proto_rawDesc
)

func file_pkg_grpc_ctrlserver_authentication_proto_rawDescGZIP() []byte {
	file_pkg_grpc_ctrlserver_authentication_proto_rawDescOnce.Do(func() {
		file_pkg_grpc_ctrlserver_authentication_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_grpc_ctrlserver_authentication_proto_rawDescData)
	})
	return file_pkg_grpc_ctrlserver_authentication_proto_rawDescData
}

var file_pkg_grpc_ctrlserver_authentication_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_pkg_grpc_ctrlserver_authentication_proto_goTypes = []interface{}{
	(*JoinAuthMeshRequest)(nil), // 0: rpctypes.JoinAuthMeshRequest
	(*JoinAuthMeshReply)(nil),   // 1: rpctypes.JoinAuthMeshReply
}
var file_pkg_grpc_ctrlserver_authentication_proto_depIdxs = []int32{
	0, // 0: rpctypes.Authentication.JoinMesh:input_type -> rpctypes.JoinAuthMeshRequest
	1, // 1: rpctypes.Authentication.JoinMesh:output_type -> rpctypes.JoinAuthMeshReply
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_pkg_grpc_ctrlserver_authentication_proto_init() }
func file_pkg_grpc_ctrlserver_authentication_proto_init() {
	if File_pkg_grpc_ctrlserver_authentication_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_grpc_ctrlserver_authentication_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*JoinAuthMeshRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_grpc_ctrlserver_authentication_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*JoinAuthMeshReply); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_pkg_grpc_ctrlserver_authentication_proto_msgTypes[1].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pkg_grpc_ctrlserver_authentication_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pkg_grpc_ctrlserver_authentication_proto_goTypes,
		DependencyIndexes: file_pkg_grpc_ctrlserver_authentication_proto_depIdxs,
		MessageInfos:      file_pkg_grpc_ctrlserver_authentication_proto_msgTypes,
	}.Build()
	File_pkg_grpc_ctrlserver_authentication_proto = out.File
	file_pkg_grpc_ctrlserver_authentication_proto_rawDesc = nil
	file_pkg_grpc_ctrlserver_authentication_proto_goTypes = nil
	file_pkg_grpc_ctrlserver_authentication_proto_depIdxs = nil
}
