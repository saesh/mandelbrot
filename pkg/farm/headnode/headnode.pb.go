// Code generated by protoc-gen-go. DO NOT EDIT.
// source: headnode.proto

package headnode

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type RegisterRequest struct {
	Ip                   string   `protobuf:"bytes,1,opt,name=ip,proto3" json:"ip,omitempty"`
	Port                 int32    `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
	Hostname             string   `protobuf:"bytes,3,opt,name=hostname,proto3" json:"hostname,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RegisterRequest) Reset()         { *m = RegisterRequest{} }
func (m *RegisterRequest) String() string { return proto.CompactTextString(m) }
func (*RegisterRequest) ProtoMessage()    {}
func (*RegisterRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_headnode_e366389ea81e107d, []int{0}
}
func (m *RegisterRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RegisterRequest.Unmarshal(m, b)
}
func (m *RegisterRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RegisterRequest.Marshal(b, m, deterministic)
}
func (dst *RegisterRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RegisterRequest.Merge(dst, src)
}
func (m *RegisterRequest) XXX_Size() int {
	return xxx_messageInfo_RegisterRequest.Size(m)
}
func (m *RegisterRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_RegisterRequest.DiscardUnknown(m)
}

var xxx_messageInfo_RegisterRequest proto.InternalMessageInfo

func (m *RegisterRequest) GetIp() string {
	if m != nil {
		return m.Ip
	}
	return ""
}

func (m *RegisterRequest) GetPort() int32 {
	if m != nil {
		return m.Port
	}
	return 0
}

func (m *RegisterRequest) GetHostname() string {
	if m != nil {
		return m.Hostname
	}
	return ""
}

type Void struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Void) Reset()         { *m = Void{} }
func (m *Void) String() string { return proto.CompactTextString(m) }
func (*Void) ProtoMessage()    {}
func (*Void) Descriptor() ([]byte, []int) {
	return fileDescriptor_headnode_e366389ea81e107d, []int{1}
}
func (m *Void) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Void.Unmarshal(m, b)
}
func (m *Void) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Void.Marshal(b, m, deterministic)
}
func (dst *Void) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Void.Merge(dst, src)
}
func (m *Void) XXX_Size() int {
	return xxx_messageInfo_Void.Size(m)
}
func (m *Void) XXX_DiscardUnknown() {
	xxx_messageInfo_Void.DiscardUnknown(m)
}

var xxx_messageInfo_Void proto.InternalMessageInfo

func init() {
	proto.RegisterType((*RegisterRequest)(nil), "headnode.RegisterRequest")
	proto.RegisterType((*Void)(nil), "headnode.Void")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// HeadNodeClient is the client API for HeadNode service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type HeadNodeClient interface {
	Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*Void, error)
}

type headNodeClient struct {
	cc *grpc.ClientConn
}

func NewHeadNodeClient(cc *grpc.ClientConn) HeadNodeClient {
	return &headNodeClient{cc}
}

func (c *headNodeClient) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*Void, error) {
	out := new(Void)
	err := c.cc.Invoke(ctx, "/headnode.HeadNode/Register", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// HeadNodeServer is the server API for HeadNode service.
type HeadNodeServer interface {
	Register(context.Context, *RegisterRequest) (*Void, error)
}

func RegisterHeadNodeServer(s *grpc.Server, srv HeadNodeServer) {
	s.RegisterService(&_HeadNode_serviceDesc, srv)
}

func _HeadNode_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HeadNodeServer).Register(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/headnode.HeadNode/Register",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HeadNodeServer).Register(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _HeadNode_serviceDesc = grpc.ServiceDesc{
	ServiceName: "headnode.HeadNode",
	HandlerType: (*HeadNodeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Register",
			Handler:    _HeadNode_Register_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "headnode.proto",
}

func init() { proto.RegisterFile("headnode.proto", fileDescriptor_headnode_e366389ea81e107d) }

var fileDescriptor_headnode_e366389ea81e107d = []byte{
	// 160 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0xcb, 0x48, 0x4d, 0x4c,
	0xc9, 0xcb, 0x4f, 0x49, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x80, 0xf1, 0x95, 0x02,
	0xb9, 0xf8, 0x83, 0x52, 0xd3, 0x33, 0x8b, 0x4b, 0x52, 0x8b, 0x82, 0x52, 0x0b, 0x4b, 0x53, 0x8b,
	0x4b, 0x84, 0xf8, 0xb8, 0x98, 0x32, 0x0b, 0x24, 0x18, 0x15, 0x18, 0x35, 0x38, 0x83, 0x98, 0x32,
	0x0b, 0x84, 0x84, 0xb8, 0x58, 0x0a, 0xf2, 0x8b, 0x4a, 0x24, 0x98, 0x14, 0x18, 0x35, 0x58, 0x83,
	0xc0, 0x6c, 0x21, 0x29, 0x2e, 0x8e, 0x8c, 0xfc, 0xe2, 0x92, 0xbc, 0xc4, 0xdc, 0x54, 0x09, 0x66,
	0xb0, 0x4a, 0x38, 0x5f, 0x89, 0x8d, 0x8b, 0x25, 0x2c, 0x3f, 0x33, 0xc5, 0xc8, 0x99, 0x8b, 0xc3,
	0x23, 0x35, 0x31, 0xc5, 0x2f, 0x3f, 0x25, 0x55, 0xc8, 0x9c, 0x8b, 0x03, 0x66, 0x8d, 0x90, 0xa4,
	0x1e, 0xdc, 0x35, 0x68, 0x56, 0x4b, 0xf1, 0x21, 0xa4, 0x40, 0x46, 0x28, 0x31, 0x24, 0xb1, 0x81,
	0x1d, 0x6c, 0x0c, 0x08, 0x00, 0x00, 0xff, 0xff, 0x85, 0xcc, 0xaf, 0x02, 0xc2, 0x00, 0x00, 0x00,
}
