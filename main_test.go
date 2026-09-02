package main

import (
	"strings"
	"testing"

	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/pluginpb"
)

// generateOptionsFor runs the plugin over one synthetic file whose message
// holds the given message-typed fields and returns the generated source.
func generateOptionsFor(t *testing.T, fields ...*descriptorpb.FieldDescriptorProto) string {
	t.Helper()
	file := &descriptorpb.FileDescriptorProto{
		Name:       proto.String("test/day.proto"),
		Package:    proto.String("test"),
		Syntax:     proto.String("proto3"),
		Dependency: []string{"google/rpc/status.proto", "google/protobuf/timestamp.proto"},
		Options:    &descriptorpb.FileOptions{GoPackage: proto.String("example.com/test;test")},
		MessageType: []*descriptorpb.DescriptorProto{{
			Name:  proto.String("Day"),
			Field: fields,
		}},
	}
	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"test/day.proto"},
		ProtoFile: []*descriptorpb.FileDescriptorProto{
			protodesc.ToFileDescriptorProto(anypb.File_google_protobuf_any_proto),
			protodesc.ToFileDescriptorProto(status.File_google_rpc_status_proto),
			protodesc.ToFileDescriptorProto(timestamppb.File_google_protobuf_timestamp_proto),
			file,
		},
	}
	gen, err := protogen.Options{}.New(req)
	if err != nil {
		t.Fatalf("new plugin: %v", err)
	}
	if err := run(gen); err != nil {
		t.Fatalf("run: %v", err)
	}
	resp := gen.Response()
	if resp.Error != nil {
		t.Fatalf("plugin error: %s", resp.GetError())
	}
	for _, f := range resp.File {
		if strings.HasSuffix(f.GetName(), "day_options.go") {
			return f.GetContent()
		}
	}
	t.Fatalf("no day_options.go in response: %v", resp.File)
	return ""
}

func messageField(name string, number int32, typeName string) *descriptorpb.FieldDescriptorProto {
	return &descriptorpb.FieldDescriptorProto{
		Name:     proto.String(name),
		Number:   proto.Int32(number),
		Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
		Type:     descriptorpb.FieldDescriptorProto_TYPE_MESSAGE.Enum(),
		TypeName: proto.String(typeName),
		JsonName: proto.String(name),
	}
}

// A google.rpc.Status field gets the direct setter only: genproto carries no
// StatusOption or NewStatus, so a WithNew* helper would not compile.
func TestExternalMessageFieldGetsDirectSetterOnly(t *testing.T) {
	src := generateOptionsFor(t, messageField("failure", 1, ".google.rpc.Status"))

	if strings.Contains(src, "StatusOption") || strings.Contains(src, "NewStatus(") {
		t.Fatalf("generated helpers for an external package:\n%s", src)
	}
	if !strings.Contains(src, "func WithFailure(value *status.Status) DayOption") {
		t.Fatalf("direct setter missing:\n%s", src)
	}
}

// The well-known special cases stay ahead of the external check.
func TestTimestampFieldKeepsTimeConstructor(t *testing.T) {
	src := generateOptionsFor(t, messageField("at", 1, ".google.protobuf.Timestamp"))

	if !strings.Contains(src, "func WithNewAtForDay(v time.Time) DayOption") {
		t.Fatalf("timestamp constructor missing:\n%s", src)
	}
	if !strings.Contains(src, "func WithAt(value *timestamppb.Timestamp) DayOption") {
		t.Fatalf("timestamp direct setter missing:\n%s", src)
	}
}
