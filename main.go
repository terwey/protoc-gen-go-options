package main

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
)

var logEnabled = false

type OptionFlag string

const (
	GO_OPTIONS_OPTIONLESS OptionFlag = "GO_OPTIONS_OPTIONLESS"
	GO_OPTIONS_SKIP_INIT  OptionFlag = "GO_OPTIONS_SKIP_INIT"
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

	g.P("// Code generated by protoc-gen-go-options. DO NOT EDIT.")
	g.P("// source: ", file.Proto.GetName())
	log(g, "log enabled")
	g.P("package ", file.GoPackageName)
	g.P("import (\"google.golang.org/protobuf/proto\")")

	collisionMap := detectCollisions(file.Messages)

	for _, message := range file.Messages {
		generateOptionsForMessage(g, message, collisionMap)
	}
}

func log(g *protogen.GeneratedFile, v ...any) {
	if logEnabled {
		v = append([]any{"// debug: "}, v...)
		g.P(v...)
	}
}

func detectCollisions(messages []*protogen.Message) map[string]int {
	collisionMap := make(map[string]int)
	for _, msg := range messages {
		collisionMap[msg.GoIdent.GoName]++
		for _, field := range msg.Fields {
			collisionMap[field.GoName]++
		}
	}
	return collisionMap
}

func optionFlagForMessage(message *protogen.Message, o OptionFlag) bool {
	return strings.Contains(message.Comments.Leading.String(), string(o))
}

func generateOptionsForMessage(g *protogen.GeneratedFile, message *protogen.Message, collisionMap map[string]int) {
	if message.Fields == nil {
		log(g, "skipping message because it has no fields: ", message.GoIdent.GoName)
		return
	}
	log(g, "generating options for message: ", message.GoIdent.GoName)

	// Declare the Option interface for this message
	g.P(fmt.Sprintf("// %sOption defines a functional option for %s.", message.GoIdent.GoName, message.GoIdent.GoName))
	g.P(fmt.Sprintf("type %sOption func(*%s)", message.GoIdent.GoName, message.GoIdent.GoName))
	g.P()

	if !optionFlagForMessage(message, GO_OPTIONS_SKIP_INIT) {
		constructorName := fmt.Sprintf("New%s", message.GoIdent.GoName)
		g.P(fmt.Sprintf("// %s creates a new %s.", constructorName, message.GoIdent.GoName))
		g.P(fmt.Sprintf("func %s(opts ...%s) *%s {", constructorName, qualifiedIdentForName(g, message.GoIdent, "", "Option"), message.GoIdent.GoName))
		g.P(fmt.Sprintf("\tm := &%s{}", message.GoIdent.GoName))
		g.P("\tfor _, opt := range opts {")
		g.P("\t\topt(m)")
		g.P("\t}")
		g.P("\treturn m")
		g.P("}")
		g.P()
	}

	if !optionFlagForMessage(message, GO_OPTIONS_OPTIONLESS) {
		// Generate ApplyMessageOptions function
		applyName := fmt.Sprintf("Apply%sOptions", message.GoIdent.GoName)
		g.P(fmt.Sprintf("// %s applies the provided options to an existing %s.", applyName, message.GoIdent.GoName))
		g.P(fmt.Sprintf("func %s(m *%s, opts ...%s) *%s {", applyName, message.GoIdent.GoName, qualifiedIdentForName(g, message.GoIdent, "", "Option"), g.QualifiedGoIdent(message.GoIdent)))
		g.P("\tfor _, opt := range opts {")
		g.P("\t\topt(m)")
		g.P("\t}")
		g.P("\treturn m")
		g.P("}")
		g.P()
	}

	generateFieldOptions(g, message, collisionMap)
	generateOneOfOptions(g, message)
}

func generateFieldOptions(g *protogen.GeneratedFile, message *protogen.Message, collisionMap map[string]int) {
	log(g, "generating field options for message: ", message.GoIdent.GoName)
	for _, field := range message.Fields {
		// Skip fields that belong to a oneof group
		if field.Oneof != nil && !field.Oneof.Desc.IsSynthetic() {
			continue
		}

		optionName := fmt.Sprintf("With%s", field.GoName)
		if collisionMap[field.GoName] > 1 {
			optionName = fmt.Sprintf("%sFor%s", optionName, message.GoIdent.GoName)
		}

		if field.Desc.IsMap() {
			generateMapFieldOption(g, message, field, optionName)
		} else if field.Desc.IsList() {
			generateRepeatedFieldOption(g, message, field, optionName)
		} else if field.Desc.Kind() == protoreflect.MessageKind {
			generateNestedFieldOption(g, message, field, fmt.Sprintf("WithNew%sFor%s", field.GoName, message.GoIdent.GoName))
			generateDirectNestedFieldOption(g, message, field, optionName)
		} else {
			generateScalarFieldOption(g, message, field, optionName)
		}
	}
}

func generateOneOfOptions(g *protogen.GeneratedFile, message *protogen.Message) {
	log(g, "generating oneof options for message: ", message.GoIdent.GoName)
	for _, oneof := range message.Oneofs {
		if oneof.Desc.IsSynthetic() {
			continue
		}

		for _, field := range oneof.Fields {
			fieldWrapperType := fmt.Sprintf("%s_%s", message.GoIdent.GoName, field.GoName)
			optionName := fmt.Sprintf("With%s", field.GoName)
			g.P(fmt.Sprintf("// %s sets the %s oneof field to %s.", optionName, oneof.GoName, field.GoName))
			if field.Desc.IsList() {
				log(g, "oneof field is a list")
				g.P(fmt.Sprintf("func %s(value ...%s) %s {", optionName, determineFieldType(g, field), qualifiedIdentForName(g, message.GoIdent, "", "Option")))
			} else {
				log(g, "oneof field is not a list")
				g.P(fmt.Sprintf("func %s(value %s) %s {", optionName, determineFieldType(g, field), qualifiedIdentForName(g, message.GoIdent, "", "Option")))
			}
			g.P(fmt.Sprintf("\treturn func(m *%s) {", message.GoIdent.GoName))
			// Assign the wrapper struct for oneof fields
			if field.Desc.IsList() {
				log(g, "oneof field is a list")
				g.P(fmt.Sprintf("\t\tm.%s = &%s{\n\t\t\t%s: value,\n\t\t}", oneof.GoName, fieldWrapperType, field.GoName))
			} else {
				log(g, "oneof field is not a list")
				g.P(fmt.Sprintf("\t\tm.%s = &%s{\n\t\t\t%s: value,\n\t\t}", oneof.GoName, fieldWrapperType, field.GoName))
			}
			g.P("\t}")
			g.P("}")
			g.P()
		}
	}
}

func generateNestedFieldOption(g *protogen.GeneratedFile, message *protogen.Message, field *protogen.Field, optionName string) {
	// we need to check if the message field is optionless
	optionless := optionFlagForMessage(field.Message, GO_OPTIONS_OPTIONLESS)
	log(g, "generating nested field option for message: ", message.GoIdent.GoName, ", optionless: ", optionless)
	g.P(fmt.Sprintf("// %s sets the %s field with a new instance.", optionName, field.GoName))
	if optionless {
		g.P(fmt.Sprintf("func %s() %s {", optionName, qualifiedIdentForName(g, message.GoIdent, "", "Option")))
	} else {
		g.P(fmt.Sprintf("func %s(opts ...%s) %s {", optionName, qualifiedIdentForName(g, field.Message.GoIdent, "", "Option"), qualifiedIdentForName(g, message.GoIdent, "", "Option")))
	}
	g.P(fmt.Sprintf("\treturn func(m *%s) {", message.GoIdent.GoName))
	if optionless {
		g.P(fmt.Sprintf("\t\tm.%s = %s()", field.GoName, qualifiedIdentForName(g, field.Message.GoIdent, "New", "")))
	} else {
		g.P(fmt.Sprintf("\t\tm.%s = %s(opts...)", field.GoName, qualifiedIdentForName(g, field.Message.GoIdent, "New", "")))
	}
	g.P("\t}")
	g.P("}")
	g.P()
}

// qualifiedIdentForName takes the protogen.GoIdent and will return a qualified Go identifier with the given prefix and suffix
// examples: prefix "New" for GoIdent "Foo" in an external package "foo" will return "foo.NewFoo"
// example: suffix "Option" for GoIdent "Bar" in the same package will return "BarOption"
func qualifiedIdentForName(g *protogen.GeneratedFile, ident protogen.GoIdent, prefix string, suffix string) string {
	log(g, "qualifying identifier for name: ", ident.GoName)
	return g.QualifiedGoIdent(ident.GoImportPath.Ident(prefix + ident.GoName + suffix))
}

func generateDirectNestedFieldOption(g *protogen.GeneratedFile, message *protogen.Message, field *protogen.Field, optionName string) {
	log(g, "generating direct nested field option for message: ", message.GoIdent.GoName)
	g.P(fmt.Sprintf("// %s sets the %s field directly.", optionName, field.GoName))
	g.P(fmt.Sprintf("func %s(value *%s) %s {", optionName, g.QualifiedGoIdent(field.Message.GoIdent), qualifiedIdentForName(g, message.GoIdent, "", "Option")))
	g.P(fmt.Sprintf("\treturn func(m *%s) {", message.GoIdent.GoName))
	g.P(fmt.Sprintf("\t\tm.%s = value", field.GoName))
	g.P("\t}")
	g.P("}")
	g.P()
}

func generateScalarFieldOption(g *protogen.GeneratedFile, message *protogen.Message, field *protogen.Field, optionName string) {
	log(g, "generating scalar field option for message: ", message.GoIdent.GoName)
	fieldType := determineFieldType(g, field)
	g.P(fmt.Sprintf("// %s sets the %s field.", optionName, field.GoName))
	if field.Desc.IsList() {
		log(g, "field is a list")
		g.P(fmt.Sprintf("func %s(value ...%s) %s {", optionName, fieldType, qualifiedIdentForName(g, message.GoIdent, "", "Option")))
	} else if field.Desc.Kind() == protoreflect.EnumKind {
		log(g, "field is an enum")
		g.P(fmt.Sprintf("func %s(value *%s) %s {", optionName, fieldType, qualifiedIdentForName(g, message.GoIdent, "", "Option")))
	} else {
		log(g, "field is not a list")
		g.P(fmt.Sprintf("func %s(value %s) %s {", optionName, fieldType, qualifiedIdentForName(g, message.GoIdent, "", "Option")))
	}
	g.P(fmt.Sprintf("\treturn func(m *%s) {", g.QualifiedGoIdent(message.GoIdent)))
	if field.Desc.IsList() {
		log(g, "field is a list")
		g.P(fmt.Sprintf("\t\tm.%s = value", field.GoName))
	} else if protoHelperFunc(field.Desc.Kind()) != "" {
		log(g, "field is a scalar")
		g.P(fmt.Sprintf("\t\tm.%s = proto.%s(value)", field.GoName, protoHelperFunc(field.Desc.Kind())))
	} else if field.Desc.Kind() == protoreflect.EnumKind {
		log(g, "field is an enum")
		g.P(fmt.Sprintf("\t\tm.%s = value", field.GoName))
	} else {
		log(g, "field is an interface")
		g.P(fmt.Sprintf("\t\tm.%s = value", field.GoName))
	}
	g.P("\t}")
	g.P("}")
	g.P()
}

func generateRepeatedFieldOption(g *protogen.GeneratedFile, message *protogen.Message, field *protogen.Field, optionName string) {
	log(g, "generating repeated field option for message: ", message.GoIdent.GoName)
	elementType := determineFieldType(g, field)
	g.P(fmt.Sprintf("// %s sets the %s field.", optionName, field.GoName))
	g.P(fmt.Sprintf("func %s(values ...%s) %s {", optionName, elementType, qualifiedIdentForName(g, message.GoIdent, "", "Option")))
	g.P(fmt.Sprintf("\treturn func(m *%s) {", message.GoIdent.GoName))
	g.P(fmt.Sprintf("\t\tm.%s = values", field.GoName))
	g.P("\t}")
	g.P("}")
	g.P()
}

func generateMapFieldOption(g *protogen.GeneratedFile, message *protogen.Message, field *protogen.Field, optionName string) {
	log(g, "generating map field option for message: ", message.GoIdent.GoName)
	keyType := determineFieldType(g, field.Message.Fields[0])
	valueType := determineFieldType(g, field.Message.Fields[1])
	g.P(fmt.Sprintf("// %s sets the %s field.", optionName, field.GoName))
	g.P(fmt.Sprintf("func %s(value map[%s]%s) %s {", optionName, keyType, valueType, qualifiedIdentForName(g, message.GoIdent, "", "Option")))
	g.P(fmt.Sprintf("\treturn func(m *%s) {", message.GoIdent.GoName))
	g.P(fmt.Sprintf("\t\tm.%s = value", field.GoName))
	g.P("\t}")
	g.P("}")
	g.P()
}

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
		return ""
	}
}

func determineFieldType(g *protogen.GeneratedFile, field *protogen.Field) string {
	log(g, "determining field type for field: ", field.GoName)
	log(g, "field kind: ", field.Desc.Kind())
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		return "bool"
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
	case protoreflect.EnumKind:
		return g.QualifiedGoIdent(field.Enum.GoIdent)
	default:
		return "interface{}"
	}
}
