// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.12
// source: pkg/grpc/ctrlserver/ctrlserver.proto

package rpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// MeshCtrlServerClient is the client API for MeshCtrlServer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MeshCtrlServerClient interface {
	GetMesh(ctx context.Context, in *GetMeshRequest, opts ...grpc.CallOption) (*GetMeshReply, error)
	JoinMesh(ctx context.Context, in *JoinMeshRequest, opts ...grpc.CallOption) (*JoinMeshReply, error)
}

type meshCtrlServerClient struct {
	cc grpc.ClientConnInterface
}

func NewMeshCtrlServerClient(cc grpc.ClientConnInterface) MeshCtrlServerClient {
	return &meshCtrlServerClient{cc}
}

func (c *meshCtrlServerClient) GetMesh(ctx context.Context, in *GetMeshRequest, opts ...grpc.CallOption) (*GetMeshReply, error) {
	out := new(GetMeshReply)
	err := c.cc.Invoke(ctx, "/rpctypes.MeshCtrlServer/GetMesh", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *meshCtrlServerClient) JoinMesh(ctx context.Context, in *JoinMeshRequest, opts ...grpc.CallOption) (*JoinMeshReply, error) {
	out := new(JoinMeshReply)
	err := c.cc.Invoke(ctx, "/rpctypes.MeshCtrlServer/JoinMesh", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MeshCtrlServerServer is the server API for MeshCtrlServer service.
// All implementations must embed UnimplementedMeshCtrlServerServer
// for forward compatibility
type MeshCtrlServerServer interface {
	GetMesh(context.Context, *GetMeshRequest) (*GetMeshReply, error)
	JoinMesh(context.Context, *JoinMeshRequest) (*JoinMeshReply, error)
	mustEmbedUnimplementedMeshCtrlServerServer()
}

// UnimplementedMeshCtrlServerServer must be embedded to have forward compatible implementations.
type UnimplementedMeshCtrlServerServer struct {
}

func (UnimplementedMeshCtrlServerServer) GetMesh(context.Context, *GetMeshRequest) (*GetMeshReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMesh not implemented")
}
func (UnimplementedMeshCtrlServerServer) JoinMesh(context.Context, *JoinMeshRequest) (*JoinMeshReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method JoinMesh not implemented")
}
func (UnimplementedMeshCtrlServerServer) mustEmbedUnimplementedMeshCtrlServerServer() {}

// UnsafeMeshCtrlServerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MeshCtrlServerServer will
// result in compilation errors.
type UnsafeMeshCtrlServerServer interface {
	mustEmbedUnimplementedMeshCtrlServerServer()
}

func RegisterMeshCtrlServerServer(s grpc.ServiceRegistrar, srv MeshCtrlServerServer) {
	s.RegisterService(&MeshCtrlServer_ServiceDesc, srv)
}

func _MeshCtrlServer_GetMesh_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetMeshRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MeshCtrlServerServer).GetMesh(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpctypes.MeshCtrlServer/GetMesh",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MeshCtrlServerServer).GetMesh(ctx, req.(*GetMeshRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MeshCtrlServer_JoinMesh_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(JoinMeshRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MeshCtrlServerServer).JoinMesh(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpctypes.MeshCtrlServer/JoinMesh",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MeshCtrlServerServer).JoinMesh(ctx, req.(*JoinMeshRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// MeshCtrlServer_ServiceDesc is the grpc.ServiceDesc for MeshCtrlServer service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MeshCtrlServer_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "rpctypes.MeshCtrlServer",
	HandlerType: (*MeshCtrlServerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetMesh",
			Handler:    _MeshCtrlServer_GetMesh_Handler,
		},
		{
			MethodName: "JoinMesh",
			Handler:    _MeshCtrlServer_JoinMesh_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/grpc/ctrlserver/ctrlserver.proto",
}
