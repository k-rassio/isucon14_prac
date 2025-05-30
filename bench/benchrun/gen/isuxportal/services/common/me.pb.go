// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        (unknown)
// source: isuxportal/services/common/me.proto

package common

import (
	resources "github.com/isucon/isucon14/bench/benchrun/gen/isuxportal/resources"
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

type GetCurrentSessionRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetCurrentSessionRequest) Reset() {
	*x = GetCurrentSessionRequest{}
	mi := &file_isuxportal_services_common_me_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetCurrentSessionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCurrentSessionRequest) ProtoMessage() {}

func (x *GetCurrentSessionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_isuxportal_services_common_me_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetCurrentSessionRequest.ProtoReflect.Descriptor instead.
func (*GetCurrentSessionRequest) Descriptor() ([]byte, []int) {
	return file_isuxportal_services_common_me_proto_rawDescGZIP(), []int{0}
}

type GetCurrentSessionResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Team                *resources.Team                 `protobuf:"bytes,1,opt,name=team,proto3" json:"team,omitempty"`
	Contestant          *resources.Contestant           `protobuf:"bytes,2,opt,name=contestant,proto3" json:"contestant,omitempty"`
	DiscordServerId     string                          `protobuf:"bytes,3,opt,name=discord_server_id,json=discordServerId,proto3" json:"discord_server_id,omitempty"`
	Contest             *resources.Contest              `protobuf:"bytes,4,opt,name=contest,proto3" json:"contest,omitempty"`
	ContestantInstances []*resources.ContestantInstance `protobuf:"bytes,5,rep,name=contestant_instances,json=contestantInstances,proto3" json:"contestant_instances,omitempty"`
	PushVapidKey        string                          `protobuf:"bytes,6,opt,name=push_vapid_key,json=pushVapidKey,proto3" json:"push_vapid_key,omitempty"`
}

func (x *GetCurrentSessionResponse) Reset() {
	*x = GetCurrentSessionResponse{}
	mi := &file_isuxportal_services_common_me_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetCurrentSessionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCurrentSessionResponse) ProtoMessage() {}

func (x *GetCurrentSessionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_isuxportal_services_common_me_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetCurrentSessionResponse.ProtoReflect.Descriptor instead.
func (*GetCurrentSessionResponse) Descriptor() ([]byte, []int) {
	return file_isuxportal_services_common_me_proto_rawDescGZIP(), []int{1}
}

func (x *GetCurrentSessionResponse) GetTeam() *resources.Team {
	if x != nil {
		return x.Team
	}
	return nil
}

func (x *GetCurrentSessionResponse) GetContestant() *resources.Contestant {
	if x != nil {
		return x.Contestant
	}
	return nil
}

func (x *GetCurrentSessionResponse) GetDiscordServerId() string {
	if x != nil {
		return x.DiscordServerId
	}
	return ""
}

func (x *GetCurrentSessionResponse) GetContest() *resources.Contest {
	if x != nil {
		return x.Contest
	}
	return nil
}

func (x *GetCurrentSessionResponse) GetContestantInstances() []*resources.ContestantInstance {
	if x != nil {
		return x.ContestantInstances
	}
	return nil
}

func (x *GetCurrentSessionResponse) GetPushVapidKey() string {
	if x != nil {
		return x.PushVapidKey
	}
	return ""
}

var File_isuxportal_services_common_me_proto protoreflect.FileDescriptor

var file_isuxportal_services_common_me_proto_rawDesc = []byte{
	0x0a, 0x23, 0x69, 0x73, 0x75, 0x78, 0x70, 0x6f, 0x72, 0x74, 0x61, 0x6c, 0x2f, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x73, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x6d, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x20, 0x69, 0x73, 0x75, 0x78, 0x70, 0x6f, 0x72, 0x74, 0x61,
	0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73,
	0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x1a, 0x1f, 0x69, 0x73, 0x75, 0x78, 0x70, 0x6f, 0x72,
	0x74, 0x61, 0x6c, 0x2f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x2f, 0x74, 0x65,
	0x61, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x25, 0x69, 0x73, 0x75, 0x78, 0x70, 0x6f,
	0x72, 0x74, 0x61, 0x6c, 0x2f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x2f, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x73, 0x74, 0x61, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x2e, 0x69, 0x73, 0x75, 0x78, 0x70, 0x6f, 0x72, 0x74, 0x61, 0x6c, 0x2f, 0x72, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x73, 0x74, 0x61, 0x6e, 0x74,
	0x5f, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x22, 0x69, 0x73, 0x75, 0x78, 0x70, 0x6f, 0x72, 0x74, 0x61, 0x6c, 0x2f, 0x72, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x1a, 0x0a, 0x18, 0x47, 0x65, 0x74, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e,
	0x74, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22,
	0x8d, 0x03, 0x0a, 0x19, 0x47, 0x65, 0x74, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x53, 0x65,
	0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x34, 0x0a,
	0x04, 0x74, 0x65, 0x61, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x69, 0x73,
	0x75, 0x78, 0x70, 0x6f, 0x72, 0x74, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x72,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x2e, 0x54, 0x65, 0x61, 0x6d, 0x52, 0x04, 0x74,
	0x65, 0x61, 0x6d, 0x12, 0x46, 0x0a, 0x0a, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x73, 0x74, 0x61, 0x6e,
	0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x69, 0x73, 0x75, 0x78, 0x70, 0x6f,
	0x72, 0x74, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x72, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x73, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x73, 0x74, 0x61, 0x6e, 0x74, 0x52,
	0x0a, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x73, 0x74, 0x61, 0x6e, 0x74, 0x12, 0x2a, 0x0a, 0x11, 0x64,
	0x69, 0x73, 0x63, 0x6f, 0x72, 0x64, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x5f, 0x69, 0x64,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x64, 0x69, 0x73, 0x63, 0x6f, 0x72, 0x64, 0x53,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x49, 0x64, 0x12, 0x3d, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65,
	0x73, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x23, 0x2e, 0x69, 0x73, 0x75, 0x78, 0x70,
	0x6f, 0x72, 0x74, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x72, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x73, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x73, 0x74, 0x52, 0x07, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x73, 0x74, 0x12, 0x61, 0x0a, 0x14, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x73,
	0x74, 0x61, 0x6e, 0x74, 0x5f, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x18, 0x05,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x2e, 0x2e, 0x69, 0x73, 0x75, 0x78, 0x70, 0x6f, 0x72, 0x74, 0x61,
	0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x73, 0x2e, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x73, 0x74, 0x61, 0x6e, 0x74, 0x49, 0x6e, 0x73, 0x74,
	0x61, 0x6e, 0x63, 0x65, 0x52, 0x13, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x73, 0x74, 0x61, 0x6e, 0x74,
	0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x73, 0x12, 0x24, 0x0a, 0x0e, 0x70, 0x75, 0x73,
	0x68, 0x5f, 0x76, 0x61, 0x70, 0x69, 0x64, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x06, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0c, 0x70, 0x75, 0x73, 0x68, 0x56, 0x61, 0x70, 0x69, 0x64, 0x4b, 0x65, 0x79, 0x42,
	0x9d, 0x02, 0x0a, 0x24, 0x63, 0x6f, 0x6d, 0x2e, 0x69, 0x73, 0x75, 0x78, 0x70, 0x6f, 0x72, 0x74,
	0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x73, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x42, 0x07, 0x4d, 0x65, 0x50, 0x72, 0x6f, 0x74,
	0x6f, 0x50, 0x01, 0x5a, 0x48, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x69, 0x73, 0x75, 0x63, 0x6f, 0x6e, 0x2f, 0x69, 0x73, 0x75, 0x63, 0x6f, 0x6e, 0x31, 0x34, 0x2f,
	0x62, 0x65, 0x6e, 0x63, 0x68, 0x2f, 0x62, 0x65, 0x6e, 0x63, 0x68, 0x72, 0x75, 0x6e, 0x2f, 0x67,
	0x65, 0x6e, 0x2f, 0x69, 0x73, 0x75, 0x78, 0x70, 0x6f, 0x72, 0x74, 0x61, 0x6c, 0x2f, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0xa2, 0x02, 0x04,
	0x49, 0x50, 0x53, 0x43, 0xaa, 0x02, 0x20, 0x49, 0x73, 0x75, 0x78, 0x70, 0x6f, 0x72, 0x74, 0x61,
	0x6c, 0x2e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73,
	0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0xca, 0x02, 0x20, 0x49, 0x73, 0x75, 0x78, 0x70, 0x6f,
	0x72, 0x74, 0x61, 0x6c, 0x5c, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x5c, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x73, 0x5c, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0xe2, 0x02, 0x2c, 0x49, 0x73, 0x75,
	0x78, 0x70, 0x6f, 0x72, 0x74, 0x61, 0x6c, 0x5c, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x5c, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x5c, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x5c, 0x47, 0x50,
	0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x23, 0x49, 0x73, 0x75, 0x78,
	0x70, 0x6f, 0x72, 0x74, 0x61, 0x6c, 0x3a, 0x3a, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x3a, 0x3a, 0x53,
	0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x73, 0x3a, 0x3a, 0x43, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_isuxportal_services_common_me_proto_rawDescOnce sync.Once
	file_isuxportal_services_common_me_proto_rawDescData = file_isuxportal_services_common_me_proto_rawDesc
)

func file_isuxportal_services_common_me_proto_rawDescGZIP() []byte {
	file_isuxportal_services_common_me_proto_rawDescOnce.Do(func() {
		file_isuxportal_services_common_me_proto_rawDescData = protoimpl.X.CompressGZIP(file_isuxportal_services_common_me_proto_rawDescData)
	})
	return file_isuxportal_services_common_me_proto_rawDescData
}

var file_isuxportal_services_common_me_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_isuxportal_services_common_me_proto_goTypes = []any{
	(*GetCurrentSessionRequest)(nil),     // 0: isuxportal.proto.services.common.GetCurrentSessionRequest
	(*GetCurrentSessionResponse)(nil),    // 1: isuxportal.proto.services.common.GetCurrentSessionResponse
	(*resources.Team)(nil),               // 2: isuxportal.proto.resources.Team
	(*resources.Contestant)(nil),         // 3: isuxportal.proto.resources.Contestant
	(*resources.Contest)(nil),            // 4: isuxportal.proto.resources.Contest
	(*resources.ContestantInstance)(nil), // 5: isuxportal.proto.resources.ContestantInstance
}
var file_isuxportal_services_common_me_proto_depIdxs = []int32{
	2, // 0: isuxportal.proto.services.common.GetCurrentSessionResponse.team:type_name -> isuxportal.proto.resources.Team
	3, // 1: isuxportal.proto.services.common.GetCurrentSessionResponse.contestant:type_name -> isuxportal.proto.resources.Contestant
	4, // 2: isuxportal.proto.services.common.GetCurrentSessionResponse.contest:type_name -> isuxportal.proto.resources.Contest
	5, // 3: isuxportal.proto.services.common.GetCurrentSessionResponse.contestant_instances:type_name -> isuxportal.proto.resources.ContestantInstance
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_isuxportal_services_common_me_proto_init() }
func file_isuxportal_services_common_me_proto_init() {
	if File_isuxportal_services_common_me_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_isuxportal_services_common_me_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_isuxportal_services_common_me_proto_goTypes,
		DependencyIndexes: file_isuxportal_services_common_me_proto_depIdxs,
		MessageInfos:      file_isuxportal_services_common_me_proto_msgTypes,
	}.Build()
	File_isuxportal_services_common_me_proto = out.File
	file_isuxportal_services_common_me_proto_rawDesc = nil
	file_isuxportal_services_common_me_proto_goTypes = nil
	file_isuxportal_services_common_me_proto_depIdxs = nil
}
