package main

import (
	"fmt"
	"strings"

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

	// Write the comments that it is generated
	g.P("// Code generated by protoc-gen-go-options. DO NOT EDIT.")
	g.P("// source: ", file.Proto.GetName())
	g.P()

	// Write the package declaration
	g.P("package ", file.GoPackageName)
	g.P()

	// Add imports
	g.P("import (")
	g.P("\t\"google.golang.org/protobuf/proto\"")
	g.P(")")
	g.P()

	// Collect all field names to detect shared ones
	sharedFields := findSharedFieldNames(file.Messages)

	for _, message := range file.Messages {
		generateMessageOptions(g, message, sharedFields)
	}

	for _, message := range file.Messages {
		generateNewOptionsFunction(g, message, file, sharedFields)
	}
}

func generateMessageOptions(g *protogen.GeneratedFile, message *protogen.Message, sharedFields map[string]bool) {
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

	// Generate ApplyMessageOptions function
	g.P(fmt.Sprintf("// Apply%sOptions applies the provided options to an existing %s.", messageName, messageName))
	g.P(fmt.Sprintf("func Apply%sOptions(m *%s, opts ...%sOption) {", messageName, messageName, messageName))
	g.P("\tfor _, opt := range opts {")
	g.P("\t\topt(m)")
	g.P("\t}")
	g.P("}")
	g.P()

	// Generate option type
	g.P(fmt.Sprintf("// %sOption defines a functional option for %s.", messageName, messageName))
	g.P(fmt.Sprintf("type %sOption func(*%s)", messageName, messageName))
	g.P()

	// Generate option setters
	for _, field := range message.Fields {
		fieldName := field.GoName
		optionName := fmt.Sprintf("%s", fieldName)
		if sharedFields[fieldName] {
			optionName = fmt.Sprintf("%sFor%s", fieldName, message.GoIdent.GoName)
		}

		if field.Oneof != nil && !field.Oneof.Desc.IsSynthetic() {
			// Handle oneof fields
			generateOneofNewOptionWithName(g, message, field, optionName)
		} else if field.Desc.IsList() {
			// Handle repeated fields
			generateRepeatedOptionWithName(g, message, field, optionName)
		} else {
			// Handle regular fields
			generateRegularOptionWithName(g, message, field, optionName)
		}
	}
}

func generateNewOptionsFunction(g *protogen.GeneratedFile, message *protogen.Message, file *protogen.File, sharedFields map[string]bool) {
	messageName := message.GoIdent.GoName
	for _, field := range message.Fields {
		if field.Oneof != nil && !field.Oneof.Desc.IsSynthetic() {
			continue
		}
		if field.Desc.IsList() || field.Desc.IsMap() {
			continue
		}
		if field.Desc.Kind() == protoreflect.MessageKind {
			fieldName := field.GoName
			fieldType := g.QualifiedGoIdent(field.Message.GoIdent)
			fieldPackage := getPackageNameFromImportPath(field.Message.GoIdent.GoImportPath)
			currentPackage := string(file.GoPackageName)

			optionName := fieldName
			if sharedFields[fieldName] {
				optionName = fmt.Sprintf("%sFor%s", fieldName, message.GoIdent.GoName)
			}

			// Determine if the package name is needed
			var constructor string
			if fieldPackage == currentPackage {
				constructor = fmt.Sprintf("New%s", field.Message.GoIdent.GoName) // Same package: no prefix
			} else {
				constructor = fmt.Sprintf("%s.New%s", fieldPackage, field.Message.GoIdent.GoName) // Different package: include prefix
			}

			methodName := fmt.Sprintf("New%s", optionName)

			g.P(fmt.Sprintf("// %s creates a new %s and sets it to the %s field.", methodName, fieldType, fieldName))
			g.P(fmt.Sprintf("func %s() %sOption {", methodName, messageName))
			g.P(fmt.Sprintf("\treturn func(m *%s) {", messageName))
			g.P(fmt.Sprintf("\t\tm.%s = %s()", fieldName, constructor)) // Use conditional constructor
			g.P("\t}")
			g.P("}")
			g.P()
		}
	}
}

// Extract the package name from the import path
func getPackageNameFromImportPath(importPath protogen.GoImportPath) string {
	parts := strings.Split(string(importPath), "/")
	return parts[len(parts)-1]
}

func generateRegularOptionWithName(g *protogen.GeneratedFile, message *protogen.Message, field *protogen.Field, optionName string) {
	fieldName := field.GoName

	optionName = "With" + optionName

	// Handle map fields
	if field.Desc.IsMap() {
		generateMapOptionWithName(g, message, field, optionName)
		return
	}

	// Handle other fields
	fieldType := getGoTypeFromKind(g, field)
	g.P(fmt.Sprintf("// %s sets the %s field of %s.", optionName, fieldName, message.GoIdent.GoName))
	g.P(fmt.Sprintf("func %s(value %s) %sOption {", optionName, fieldType, message.GoIdent.GoName))
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

func generateMapOptionWithName(g *protogen.GeneratedFile, message *protogen.Message, field *protogen.Field, optionName string) {
	keyType := getBaseGoType(g, field.Message.Fields[0])
	valueType := getBaseGoType(g, field.Message.Fields[1])
	mapType := fmt.Sprintf("map[%s]%s", keyType, valueType)
	fieldName := field.GoName
	g.P(fmt.Sprintf("// %s sets the %s field of %s.", optionName, fieldName, message.GoIdent.GoName))
	g.P(fmt.Sprintf("func %s(value %s) %sOption {", optionName, mapType, message.GoIdent.GoName))
	g.P(fmt.Sprintf("\treturn func(m *%s) {", message.GoIdent.GoName))
	g.P(fmt.Sprintf("\t\tm.%s = value", fieldName))
	g.P("\t}")
	g.P("}")
	g.P()
}

func generateRepeatedOptionWithName(g *protogen.GeneratedFile, message *protogen.Message, field *protogen.Field, optionName string) {
	optionName = "With" + optionName
	fieldType := "..." + getBaseGoType(g, field)
	fieldName := field.GoName
	g.P(fmt.Sprintf("// %s sets the %s field of %s.", optionName, fieldName, message.GoIdent.GoName))
	g.P(fmt.Sprintf("func %s(values %s) %sOption {", optionName, fieldType, message.GoIdent.GoName))
	g.P(fmt.Sprintf("\treturn func(m *%s) {", message.GoIdent.GoName))
	g.P(fmt.Sprintf("\t\tm.%s = values", fieldName))
	g.P("\t}")
	g.P("}")
	g.P()
}

func generateOneofNewOptionWithName(g *protogen.GeneratedFile, message *protogen.Message, field *protogen.Field, optionName string) {
	fieldName := field.GoName
	wrapperType := fmt.Sprintf("%s_%s", message.GoIdent.GoName, fieldName)

	generateOneofOptionWithName(g, message, field, "With"+optionName)

	optionName = "WithNew" + optionName

	g.P(fmt.Sprintf("// %s sets the %s OneOf field of %s.", optionName, fieldName, field.Oneof.GoName))
	if supportsOptions(field) {
		// g.P(fmt.Sprintf("func %s(opts ...%sOption) %sOption {", optionName, field.Message.GoIdent.GoName, message.GoIdent.GoName))
		g.P(fmt.Sprintf("func %s(opts ...%sOption) %sOption {", optionName, g.QualifiedGoIdent(field.Message.GoIdent), message.GoIdent.GoName))

	} else {
		g.P(fmt.Sprintf("func %s(value %s) %sOption {", optionName, getGoTypeFromKind(g, field), message.GoIdent.GoName))
	}
	g.P(fmt.Sprintf("\treturn func(m *%s) {", message.GoIdent.GoName))
	if supportsOptions(field) {
		g.P(fmt.Sprintf("\t\tm.%s = &%s{", field.Oneof.GoName, wrapperType))
		fieldPackage := getPackageNameFromImportPath(field.Message.GoIdent.GoImportPath)
		g.P(fmt.Sprintf("\t\t\t%s: %s.New%s(opts...),", fieldName, fieldPackage, field.Message.GoIdent.GoName))
		g.P("\t\t}")
	} else {
		g.P(fmt.Sprintf("\t\tm.%s = &%s{", field.Oneof.GoName, wrapperType))
		g.P(fmt.Sprintf("\t\t\t%s: value,", fieldName))
		g.P("\t\t}")
	}
	g.P("\t}")
	g.P("}")
	g.P()
}

func generateOneofOptionWithName(g *protogen.GeneratedFile, message *protogen.Message, field *protogen.Field, optionName string) {
	fieldName := field.GoName
	wrapperType := message.GoIdent.GoName + "_" + field.GoName
	g.P(fmt.Sprintf("// %s sets the %s field of %s.", optionName, fieldName, message.GoIdent.GoName))
	g.P(fmt.Sprintf("func %s(value %s) %sOption {", optionName, getGoTypeFromKind(g, field), message.GoIdent.GoName))
	g.P(fmt.Sprintf("\treturn func(m *%s) {", message.GoIdent.GoName))
	g.P(fmt.Sprintf("\t\tm.%s = &%s{", field.Oneof.GoName, wrapperType))
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

func findSharedFieldNames(messages []*protogen.Message) map[string]bool {
	fieldCount := make(map[string]int)
	sharedFields := make(map[string]bool)

	// Count field names
	for _, message := range messages {
		for _, field := range message.Fields {
			fieldName := field.GoName
			fieldCount[fieldName]++
		}
	}

	// Mark shared fields
	for name, count := range fieldCount {
		if count > 1 {
			sharedFields[name] = true
		}
	}

	return sharedFields
}

func supportsOptions(field *protogen.Field) bool {
	return field.Desc.Kind() == protoreflect.MessageKind
}
