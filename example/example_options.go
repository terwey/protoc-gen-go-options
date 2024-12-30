package example

import (
	"google.golang.org/protobuf/proto"
)

// NewBasicMessage creates a new BasicMessage with the provided options.
func NewBasicMessage(opts ...BasicMessageOption) *BasicMessage {
	m := &BasicMessage{}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// BasicMessageOption defines a functional option for BasicMessage.
type BasicMessageOption func(*BasicMessage)

// WithName sets the Name field of BasicMessage.
func WithName(value string) BasicMessageOption {
	return func(m *BasicMessage) {
		m.Name = proto.String(value)
	}
}

// WithAge sets the Age field of BasicMessage.
func WithAge(value int32) BasicMessageOption {
	return func(m *BasicMessage) {
		m.Age = proto.Int32(value)
	}
}

// WithIsActive sets the IsActive field of BasicMessage.
func WithIsActive(value bool) BasicMessageOption {
	return func(m *BasicMessage) {
		m.IsActive = proto.Bool(value)
	}
}

// NewRepeatedFieldsMessage creates a new RepeatedFieldsMessage with the provided options.
func NewRepeatedFieldsMessage(opts ...RepeatedFieldsMessageOption) *RepeatedFieldsMessage {
	m := &RepeatedFieldsMessage{}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// RepeatedFieldsMessageOption defines a functional option for RepeatedFieldsMessage.
type RepeatedFieldsMessageOption func(*RepeatedFieldsMessage)

// WithTags sets the Tags field of RepeatedFieldsMessage.
func WithTags(values []string) RepeatedFieldsMessageOption {
	return func(m *RepeatedFieldsMessage) {
		m.Tags = values
	}
}

// WithValues sets the Values field of RepeatedFieldsMessage.
func WithValues(values []int32) RepeatedFieldsMessageOption {
	return func(m *RepeatedFieldsMessage) {
		m.Values = values
	}
}

// NewNestedMessage creates a new NestedMessage with the provided options.
func NewNestedMessage(opts ...NestedMessageOption) *NestedMessage {
	m := &NestedMessage{}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// NestedMessageOption defines a functional option for NestedMessage.
type NestedMessageOption func(*NestedMessage)

// WithBasic sets the Basic field of NestedMessage.
func WithBasic(value *BasicMessage) NestedMessageOption {
	return func(m *NestedMessage) {
		m.Basic = value
	}
}

// WithDescription sets the Description field of NestedMessage.
func WithDescription(value string) NestedMessageOption {
	return func(m *NestedMessage) {
		m.Description = proto.String(value)
	}
}

// NewOneofMessage creates a new OneofMessage with the provided options.
func NewOneofMessage(opts ...OneofMessageOption) *OneofMessage {
	m := &OneofMessage{}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// OneofMessageOption defines a functional option for OneofMessage.
type OneofMessageOption func(*OneofMessage)

// WithText sets the Text field of OneofMessage.
func WithText(value string) OneofMessageOption {
	return func(m *OneofMessage) {
		m.Choice = &OneofMessage_Text{
			Text: value,
		}
	}
}

// WithNumber sets the Number field of OneofMessage.
func WithNumber(value int32) OneofMessageOption {
	return func(m *OneofMessage) {
		m.Choice = &OneofMessage_Number{
			Number: value,
		}
	}
}

// NewComplexMessage creates a new ComplexMessage with the provided options.
func NewComplexMessage(opts ...ComplexMessageOption) *ComplexMessage {
	m := &ComplexMessage{}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// ComplexMessageOption defines a functional option for ComplexMessage.
type ComplexMessageOption func(*ComplexMessage)

// WithNested sets the Nested field of ComplexMessage.
func WithNested(value *NestedMessage) ComplexMessageOption {
	return func(m *ComplexMessage) {
		m.Nested = value
	}
}

// WithNestedList sets the NestedList field of ComplexMessage.
func WithNestedList(values []*NestedMessage) ComplexMessageOption {
	return func(m *ComplexMessage) {
		m.NestedList = values
	}
}

// WithMetadata sets the Metadata field of ComplexMessage.
func WithMetadata(value map[string]int32) ComplexMessageOption {
	return func(m *ComplexMessage) {
		m.Metadata = value
	}
}
