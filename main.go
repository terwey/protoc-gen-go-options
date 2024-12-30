package main

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	opts := protogen.Options{
		ParamFunc: func(name, value string) error {
			// Handle any custom plugin parameters here, if needed.
			return nil
		},
	}

	opts.Run(func(gen *protogen.Plugin) error {
		// Declare support for editions
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS)
		// if you also want to do FEATURE_PROTO3_OPTIONAL you can do the following
		// gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL | pluginpb.CodeGeneratorResponse_FEATURE_SUPPORTS_EDITIONS)

		// this is required to get it to work with editions, need a minimum and maximum edition
		gen.SupportedEditionsMinimum = descriptorpb.Edition_EDITION_PROTO2
		gen.SupportedEditionsMaximum = descriptorpb.Edition_EDITION_2023

		for _, file := range gen.Files {
			if !file.Generate {
				continue
			}
			generateFile(gen, file)
		}
		return nil
	})
}

func generateFile(gen *protogen.Plugin, file *protogen.File) {
	filename := file.GeneratedFilenamePrefix + "_options.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)

	// Write the package declaration
	g.P("package ", file.GoPackageName)
	g.P()

	// Add imports
	g.P("import (")
	g.P("\t\"google.golang.org/protobuf/proto\"")
	g.P(")")
	g.P()

	for _, message := range file.Messages {
		generateMessageOptions(g, message)
	}
}

func generateMessageOptions(g *protogen.GeneratedFile, message *protogen.Message) {
	messageName := message.GoIdent.GoName

	// Generate NewMessage function
	g.P(fmt.Sprintf("// New%s creates a new %s with the provided options.", messageName, messageName))
	g.P(fmt.Sprintf("func New%s(opts ...%sOption) *%s {", messageName, messageName, messageName))
	g.P(fmt.Sprintf("\tm := &%s{}", messageName))
	g.P("\tfor _, opt := range opts {")
	g.P("\t\topt(m)")
	g.P("\t}")
	g.P("\treturn m")
	g.P("}")
	g.P()

	// Generate option type
	g.P(fmt.Sprintf("// %sOption defines a functional option for %s.", messageName, messageName))
	g.P(fmt.Sprintf("type %sOption func(*%s)", messageName, messageName))
	g.P()

	// Generate option setters
	for _, field := range message.Fields {
		if field.Oneof != nil && !field.Oneof.Desc.IsSynthetic() {
			// Handle oneof fields
			generateOneofOption(g, message, field)
		} else if field.Desc.IsList() {
			// Handle repeated fields
			generateRepeatedOption(g, message, field)
		} else {
			// Handle regular fields
			generateRegularOption(g, message, field)
		}
	}
}

func generateRegularOption(g *protogen.GeneratedFile, message *protogen.Message, field *protogen.Field) {
	fieldName := field.GoName

	// Handle map fields
	if field.Desc.IsMap() {
		generateMapOption(g, message, field)
		return
	}

	// Handle other fields
	fieldType := getGoTypeFromKind(g, field)
	g.P(fmt.Sprintf("// With%s sets the %s field of %s.", fieldName, fieldName, message.GoIdent.GoName))
	g.P(fmt.Sprintf("func With%s(value %s) %sOption {", fieldName, fieldType, message.GoIdent.GoName))
	g.P(fmt.Sprintf("\treturn func(m *%s) {", message.GoIdent.GoName))
	if protoHelperFunc(field.Desc.Kind()) != "" {
		g.P(fmt.Sprintf("\t\tm.%s = proto.%s(value)", fieldName, protoHelperFunc(field.Desc.Kind())))
	} else {
		g.P(fmt.Sprintf("\t\tm.%s = value", fieldName))
	}
	g.P("\t}")
	g.P("}")
	g.P()
}

func generateMapOption(g *protogen.GeneratedFile, message *protogen.Message, field *protogen.Field) {
	fieldName := field.GoName
	keyType := getBaseGoType(g, field.Message.Fields[0])
	valueType := getBaseGoType(g, field.Message.Fields[1])
	mapType := fmt.Sprintf("map[%s]%s", keyType, valueType)

	g.P(fmt.Sprintf("// With%s sets the %s field of %s.", fieldName, fieldName, message.GoIdent.GoName))
	g.P(fmt.Sprintf("func With%s(value %s) %sOption {", fieldName, mapType, message.GoIdent.GoName))
	g.P(fmt.Sprintf("\treturn func(m *%s) {", message.GoIdent.GoName))
	g.P(fmt.Sprintf("\t\tm.%s = value", fieldName))
	g.P("\t}")
	g.P("}")
	g.P()
	return
}

func generateRepeatedOption(g *protogen.GeneratedFile, message *protogen.Message, field *protogen.Field) {
	fieldName := field.GoName
	fieldType := getGoTypeFromKind(g, field)

	g.P(fmt.Sprintf("// With%s sets the %s field of %s.", fieldName, fieldName, message.GoIdent.GoName))
	g.P(fmt.Sprintf("func With%s(values %s) %sOption {", fieldName, fieldType, message.GoIdent.GoName))
	g.P(fmt.Sprintf("\treturn func(m *%s) {", message.GoIdent.GoName))
	g.P(fmt.Sprintf("\t\tm.%s = values", fieldName))
	g.P("\t}")
	g.P("}")
	g.P()
}

func generateOneofOption(g *protogen.GeneratedFile, message *protogen.Message, field *protogen.Field) {
	fieldName := field.GoName
	wrapperType := message.GoIdent.GoName + "_" + field.GoName

	g.P(fmt.Sprintf("// With%s sets the %s field of %s.", fieldName, fieldName, message.GoIdent.GoName))
	g.P(fmt.Sprintf("func With%s(value %s) %sOption {", fieldName, getGoTypeFromKind(g, field), message.GoIdent.GoName))
	g.P(fmt.Sprintf("\treturn func(m *%s) {", message.GoIdent.GoName))
	g.P(fmt.Sprintf("\t\tm.Choice = &%s{", wrapperType))
	g.P(fmt.Sprintf("\t\t\t%s: value,", fieldName))
	g.P("\t\t}")
	g.P("\t}")
	g.P("}")
	g.P()
}

func getGoTypeFromKind(g *protogen.GeneratedFile, field *protogen.Field) string {
	// Handle repeated fields
	if field.Desc.IsList() {
		return "[]" + getBaseGoType(g, field)
	}

	// Handle scalar and complex types directly
	return getBaseGoType(g, field)
}

func getBaseGoType(g *protogen.GeneratedFile, field *protogen.Field) string {
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		return "bool"
	case protoreflect.EnumKind:
		return g.QualifiedGoIdent(field.Enum.GoIdent)
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "uint32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "uint64"
	case protoreflect.FloatKind:
		return "float32"
	case protoreflect.DoubleKind:
		return "float64"
	case protoreflect.StringKind:
		return "string"
	case protoreflect.BytesKind:
		return "[]byte"
	case protoreflect.MessageKind:
		return "*" + g.QualifiedGoIdent(field.Message.GoIdent)
	default:
		return "interface{}" // Fallback for unsupported kinds
	}
}

// protoHelperFunc returns the name of the proto helper function for the given kind.
// e.g. for a *string we return proto.String
func protoHelperFunc(kind protoreflect.Kind) string {
	switch kind {
	case protoreflect.BoolKind:
		return "Bool"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "Int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return "Uint32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return "Int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "Uint64"
	case protoreflect.FloatKind:
		return "Float32"
	case protoreflect.DoubleKind:
		return "Float64"
	case protoreflect.StringKind:
		return "String"
	default:
		// Bytes and messages do not use proto helper functions
		return ""
	}
}
