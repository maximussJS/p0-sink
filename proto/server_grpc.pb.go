// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: proto/server.proto

package serverpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	BlockStream_Blocks_FullMethodName      = "/dtq.bs.server.BlockStream/Blocks"
	BlockStream_SingleBlock_FullMethodName = "/dtq.bs.server.BlockStream/SingleBlock"
)

// BlockStreamClient is the client API for BlockStream service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BlockStreamClient interface {
	Blocks(ctx context.Context, in *BlocksRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[BlockWrapper], error)
	SingleBlock(ctx context.Context, in *SingleBlockRequest, opts ...grpc.CallOption) (*BlockWrapper, error)
}

type blockStreamClient struct {
	cc grpc.ClientConnInterface
}

func NewBlockStreamClient(cc grpc.ClientConnInterface) BlockStreamClient {
	return &blockStreamClient{cc}
}

func (c *blockStreamClient) Blocks(ctx context.Context, in *BlocksRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[BlockWrapper], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &BlockStream_ServiceDesc.Streams[0], BlockStream_Blocks_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[BlocksRequest, BlockWrapper]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type BlockStream_BlocksClient = grpc.ServerStreamingClient[BlockWrapper]

func (c *blockStreamClient) SingleBlock(ctx context.Context, in *SingleBlockRequest, opts ...grpc.CallOption) (*BlockWrapper, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(BlockWrapper)
	err := c.cc.Invoke(ctx, BlockStream_SingleBlock_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BlockStreamServer is the server API for BlockStream service.
// All implementations must embed UnimplementedBlockStreamServer
// for forward compatibility.
type BlockStreamServer interface {
	Blocks(*BlocksRequest, grpc.ServerStreamingServer[BlockWrapper]) error
	SingleBlock(context.Context, *SingleBlockRequest) (*BlockWrapper, error)
	mustEmbedUnimplementedBlockStreamServer()
}

// UnimplementedBlockStreamServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedBlockStreamServer struct{}

func (UnimplementedBlockStreamServer) Blocks(*BlocksRequest, grpc.ServerStreamingServer[BlockWrapper]) error {
	return status.Errorf(codes.Unimplemented, "method Blocks not implemented")
}
func (UnimplementedBlockStreamServer) SingleBlock(context.Context, *SingleBlockRequest) (*BlockWrapper, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SingleBlock not implemented")
}
func (UnimplementedBlockStreamServer) mustEmbedUnimplementedBlockStreamServer() {}
func (UnimplementedBlockStreamServer) testEmbeddedByValue()                     {}

// UnsafeBlockStreamServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BlockStreamServer will
// result in compilation errors.
type UnsafeBlockStreamServer interface {
	mustEmbedUnimplementedBlockStreamServer()
}

func RegisterBlockStreamServer(s grpc.ServiceRegistrar, srv BlockStreamServer) {
	// If the following call pancis, it indicates UnimplementedBlockStreamServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&BlockStream_ServiceDesc, srv)
}

func _BlockStream_Blocks_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(BlocksRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(BlockStreamServer).Blocks(m, &grpc.GenericServerStream[BlocksRequest, BlockWrapper]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type BlockStream_BlocksServer = grpc.ServerStreamingServer[BlockWrapper]

func _BlockStream_SingleBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SingleBlockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlockStreamServer).SingleBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BlockStream_SingleBlock_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlockStreamServer).SingleBlock(ctx, req.(*SingleBlockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// BlockStream_ServiceDesc is the grpc.ServiceDesc for BlockStream service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var BlockStream_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "dtq.bs.server.BlockStream",
	HandlerType: (*BlockStreamServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SingleBlock",
			Handler:    _BlockStream_SingleBlock_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Blocks",
			Handler:       _BlockStream_Blocks_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "proto/server.proto",
}

const (
	BlocksService_CurrentHead_FullMethodName = "/dtq.bs.server.BlocksService/CurrentHead"
)

// BlocksServiceClient is the client API for BlocksService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BlocksServiceClient interface {
	CurrentHead(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*BlockHead, error)
}

type blocksServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewBlocksServiceClient(cc grpc.ClientConnInterface) BlocksServiceClient {
	return &blocksServiceClient{cc}
}

func (c *blocksServiceClient) CurrentHead(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*BlockHead, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(BlockHead)
	err := c.cc.Invoke(ctx, BlocksService_CurrentHead_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BlocksServiceServer is the server API for BlocksService service.
// All implementations must embed UnimplementedBlocksServiceServer
// for forward compatibility.
type BlocksServiceServer interface {
	CurrentHead(context.Context, *emptypb.Empty) (*BlockHead, error)
	mustEmbedUnimplementedBlocksServiceServer()
}

// UnimplementedBlocksServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedBlocksServiceServer struct{}

func (UnimplementedBlocksServiceServer) CurrentHead(context.Context, *emptypb.Empty) (*BlockHead, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CurrentHead not implemented")
}
func (UnimplementedBlocksServiceServer) mustEmbedUnimplementedBlocksServiceServer() {}
func (UnimplementedBlocksServiceServer) testEmbeddedByValue()                       {}

// UnsafeBlocksServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BlocksServiceServer will
// result in compilation errors.
type UnsafeBlocksServiceServer interface {
	mustEmbedUnimplementedBlocksServiceServer()
}

func RegisterBlocksServiceServer(s grpc.ServiceRegistrar, srv BlocksServiceServer) {
	// If the following call pancis, it indicates UnimplementedBlocksServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&BlocksService_ServiceDesc, srv)
}

func _BlocksService_CurrentHead_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlocksServiceServer).CurrentHead(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BlocksService_CurrentHead_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlocksServiceServer).CurrentHead(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

// BlocksService_ServiceDesc is the grpc.ServiceDesc for BlocksService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var BlocksService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "dtq.bs.server.BlocksService",
	HandlerType: (*BlocksServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CurrentHead",
			Handler:    _BlocksService_CurrentHead_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/server.proto",
}
