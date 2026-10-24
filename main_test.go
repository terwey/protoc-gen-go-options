package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"cloud.google.com/go/longrunning/autogen/longrunningpb"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/protobuf/cmd/protoc-gen-go/internal_gengo"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
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
			"google/longrunning/operations.proto",
			"other/thing.proto",
		},
		Options: &descriptorpb.FileOptions{GoPackage: proto.String("github.com/terwey/protoc-gen-go-options/compiletest/fixture;fixture")},
		MessageType: []*descriptorpb.DescriptorProto{
			{Name: proto.String("Inner"), Field: []*descriptorpb.FieldDescriptorProto{bytesField("blob", 1)}},
			{Name: proto.String("Fieldless")},
			{Name: proto.String("Product"), Field: []*descriptorpb.FieldDescriptorProto{bytesField("blob", 1)}},
			{Name: proto.String("ProductOption"), Field: []*descriptorpb.FieldDescriptorProto{bytesField("blob", 1)}},
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
				messageField("product", 9, ".fixture.Product"),
				messageField("choice", 10, ".fixture.ProductOption"),
				messageField("job", 11, ".google.longrunning.Operation"),
				messageField("thing", 12, ".other.Thing"),
			}},
		},
	}
}

// otherFile is a sibling package of the same Go module, generated in a run of
// its own as buf's per-directory strategy does.
func otherFile() *descriptorpb.FileDescriptorProto {
	return &descriptorpb.FileDescriptorProto{
		Name:    proto.String("other/thing.proto"),
		Package: proto.String("other"),
		Syntax:  proto.String("proto2"),
		Options: &descriptorpb.FileOptions{GoPackage: proto.String("github.com/terwey/protoc-gen-go-options/compiletest/other;other")},
		MessageType: []*descriptorpb.DescriptorProto{
			{Name: proto.String("Thing"), Field: []*descriptorpb.FieldDescriptorProto{bytesField("blob", 1)}},
		},
	}
}

// withImports returns the descriptors of files and every file they import,
// each after its imports, as a CodeGeneratorRequest lists them.
func withImports(files ...protoreflect.FileDescriptor) []*descriptorpb.FileDescriptorProto {
	var out []*descriptorpb.FileDescriptorProto
	seen := map[string]bool{}
	var add func(protoreflect.FileDescriptor)
	add = func(fd protoreflect.FileDescriptor) {
		if seen[fd.Path()] {
			return
		}
		seen[fd.Path()] = true
		imports := fd.Imports()
		for i := range imports.Len() {
			add(imports.Get(i).FileDescriptor)
		}
		out = append(out, protodesc.ToFileDescriptorProto(fd))
	}
	for _, fd := range files {
		add(fd)
	}
	return out
}

// generate runs protoc-gen-go and this plugin over the named file of the
// fixture set, with the given plugin parameter, and returns the generated
// sources by file name.
func generate(t *testing.T, file, parameter string) map[string]string {
	t.Helper()
	t.Cleanup(func() { optionsModule = "" })
	protoFiles := append(withImports(
		anypb.File_google_protobuf_any_proto,
		status.File_google_rpc_status_proto,
		date.File_google_type_date_proto,
		timestamppb.File_google_protobuf_timestamp_proto,
		durationpb.File_google_protobuf_duration_proto,
		fieldmaskpb.File_google_protobuf_field_mask_proto,
		longrunningpb.File_google_longrunning_operations_proto,
	), otherFile(), fixtureFile())
	req := &pluginpb.CodeGeneratorRequest{
		FileToGenerate: []string{file},
		Parameter:      proto.String(parameter),
		ProtoFile:      protoFiles,
	}
	gen, err := pluginOptions().New(req)
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
	return out
}

// generateFixture generates fixture/day.proto with no plugin parameter of its own.
func generateFixture(t *testing.T) map[string]string {
	t.Helper()
	out := generate(t, "fixture/day.proto", "paths=source_relative")
	if _, ok := out["day_options.go"]; !ok {
		t.Fatalf("no day_options.go in response")
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
		"type ProductOptionFunc func(*Product)",
		"type ProductOptionOption func(*ProductOption)",
		"func WithNewProductForDay(opts ...ProductOptionFunc) DayOption",
		"func WithNewChoiceForDay(opts ...ProductOptionOption) DayOption",
		"func WithJob(value *longrunningpb.Operation) DayOption",
		"func WithNewThingForDay(opts ...other.ThingOption) DayOption",
	} {
		if !strings.Contains(src, want) {
			t.Errorf("missing %q\n%s", want, src)
		}
	}
	for _, reject := range []string{"StatusOption", "WithNewFailureForDay", "WithNewFieldlessForDay", "WithNewNestedForDay", "OperationOption", "WithNewJobForDay"} {
		if strings.Contains(src, reject) {
			t.Errorf("emitted %q, which has no declaration to call\n%s", reject, src)
		}
	}
}

func TestOptionsModuleBoundsWhatCountsAsOurs(t *testing.T) {
	src := generate(t, "fixture/day.proto",
		"paths=source_relative,options_module=github.com/terwey/protoc-gen-go-options/compiletest/fixture")["day_options.go"]

	if strings.Contains(src, "WithNewThingForDay") {
		t.Errorf("emitted WithNewThingForDay for a package outside options_module\n%s", src)
	}
	if !strings.Contains(src, "func WithNewInnerForDay(opts ...InnerOption) DayOption") {
		t.Errorf("dropped WithNewInnerForDay inside options_module\n%s", src)
	}
}

// The generated options must type-check against protoc-gen-go's output for the
// same file; string checks cannot see a reference to a symbol that does not
// exist.
func TestGeneratedOptionsCompile(t *testing.T) {
	packages := map[string]map[string]string{
		"fixture": generateFixture(t),
		"other":   generate(t, "other/thing.proto", "paths=source_relative"),
	}
	// The directory matches the fixtures' go_package, so fixture imports other.
	dir := "compiletest"
	for pkg, files := range packages {
		pkgDir := filepath.Join(dir, pkg)
		if err := os.MkdirAll(pkgDir, 0o755); err != nil {
			t.Fatal(err)
		}
		for name, content := range files {
			if err := os.WriteFile(filepath.Join(pkgDir, name), []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	out, err := exec.Command("go", "build", "./"+dir+"/...").CombinedOutput()
	if err != nil {
		t.Fatalf("generated packages do not compile (kept in %s):\n%s", dir, out)
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
}
