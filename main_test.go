package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/protobuf/cmd/protoc-gen-go/internal_gengo"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/pluginpb"
)

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

func bytesField(name string, number int32) *descriptorpb.FieldDescriptorProto {
	return &descriptorpb.FieldDescriptorProto{
		Name:     proto.String(name),
		Number:   proto.Int32(number),
		Label:    descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL.Enum(),
		Type:     descriptorpb.FieldDescriptorProto_TYPE_BYTES.Enum(),
		JsonName: proto.String(name),
	}
}

// fixtureFile holds every field shape generateNestedFieldOption decides on:
// well-known constructors, genproto messages, a fieldless local message, a
// nested local message, and an ordinary local message.
func fixtureFile() *descriptorpb.FileDescriptorProto {
	return &descriptorpb.FileDescriptorProto{
		Name:    proto.String("fixture/day.proto"),
		Package: proto.String("fixture"),
		Syntax:  proto.String("proto2"),
		Dependency: []string{
			"google/rpc/status.proto",
			"google/type/date.proto",
			"google/protobuf/timestamp.proto",
			"google/protobuf/duration.proto",
			"google/protobuf/field_mask.proto",
		},
		Options: &descriptorpb.FileOptions{GoPackage: proto.String("github.com/terwey/protoc-gen-go-options/compiletest/fixture;fixture")},
		MessageType: []*descriptorpb.DescriptorProto{
			{Name: proto.String("Inner"), Field: []*descriptorpb.FieldDescriptorProto{bytesField("blob", 1)}},
			{Name: proto.String("Fieldless")},
			{Name: proto.String("Outer"), NestedType: []*descriptorpb.DescriptorProto{
				{Name: proto.String("Nested"), Field: []*descriptorpb.FieldDescriptorProto{bytesField("blob", 1)}},
			}},
			{Name: proto.String("Day"), Field: []*descriptorpb.FieldDescriptorProto{
				messageField("failure", 1, ".google.rpc.Status"),
				messageField("on", 2, ".google.type.Date"),
				messageField("at", 3, ".google.protobuf.Timestamp"),
				messageField("took", 4, ".google.protobuf.Duration"),
				messageField("mask", 5, ".google.protobuf.FieldMask"),
				messageField("inner", 6, ".fixture.Inner"),
				messageField("fieldless", 7, ".fixture.Fieldless"),
				messageField("nested", 8, ".fixture.Outer.Nested"),
			}},
		},
	}
}

// generateFixture runs protoc-gen-go and this plugin over fixtureFile and
// returns the generated sources by file name.
func generateFixture(t *testing.T) map[string]string {
	t.Helper()
	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{"fixture/day.proto"},
		Parameter:      proto.String("paths=source_relative"),
		ProtoFile: []*descriptorpb.FileDescriptorProto{
			protodesc.ToFileDescriptorProto(anypb.File_google_protobuf_any_proto),
			protodesc.ToFileDescriptorProto(status.File_google_rpc_status_proto),
			protodesc.ToFileDescriptorProto(date.File_google_type_date_proto),
			protodesc.ToFileDescriptorProto(timestamppb.File_google_protobuf_timestamp_proto),
			protodesc.ToFileDescriptorProto(durationpb.File_google_protobuf_duration_proto),
			protodesc.ToFileDescriptorProto(fieldmaskpb.File_google_protobuf_field_mask_proto),
			fixtureFile(),
		},
	}
	gen, err := protogen.Options{}.New(req)
	if err != nil {
		t.Fatalf("new plugin: %v", err)
	}
	for _, f := range gen.Files {
		if f.Generate {
			internal_gengo.GenerateFile(gen, f)
		}
	}
	if err := run(gen); err != nil {
		t.Fatalf("run: %v", err)
	}
	resp := gen.Response()
	if resp.Error != nil {
		t.Fatalf("plugin error: %s", resp.GetError())
	}
	out := map[string]string{}
	for _, f := range resp.File {
		out[filepath.Base(f.GetName())] = f.GetContent()
	}
	if _, ok := out["day_options.go"]; !ok {
		t.Fatalf("no day_options.go in response: %v", resp.File)
	}
	return out
}

func TestNestedOptionHelpersMatchWhatExists(t *testing.T) {
	src := generateFixture(t)["day_options.go"]

	for _, want := range []string{
		"func WithNewInnerForDay(opts ...InnerOption) DayOption",
		"func WithNewOnForDay(v time.Time) DayOption",
		"func WithNewAtForDay(v time.Time) DayOption",
		"func WithNewTookForDay(v time.Duration) DayOption",
		"func WithNewMaskForDay(paths ...string) DayOption",
		"func WithFailure(value *status.Status) DayOption",
		"func WithFieldless(value *Fieldless) DayOption",
		"func WithNested(value *Outer_Nested) DayOption",
	} {
		if !strings.Contains(src, want) {
			t.Errorf("missing %q\n%s", want, src)
		}
	}
	for _, reject := range []string{"StatusOption", "WithNewFailureForDay", "WithNewFieldlessForDay", "WithNewNestedForDay"} {
		if strings.Contains(src, reject) {
			t.Errorf("emitted %q, which has no declaration to call\n%s", reject, src)
		}
	}
}

// The generated options must type-check against protoc-gen-go's output for the
// same file; string checks cannot see a reference to a symbol that does not
// exist.
func TestGeneratedOptionsCompile(t *testing.T) {
	files := generateFixture(t)
	dir, err := os.MkdirTemp(".", "compiletest-")
	if err != nil {
		t.Fatal(err)
	}
	pkgDir := filepath.Join(dir, "fixture")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(pkgDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.Command("go", "build", "./"+pkgDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generated package does not compile (kept in %s):\n%s", dir, out)
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
}
