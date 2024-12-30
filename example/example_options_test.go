package example

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
)

func TestNewAndApply(t *testing.T) {
	tests := []struct {
		name string
		opts []BasicMessageOption
		want *BasicMessage
	}{
		{
			name: "SetNameAndAge",
			opts: []BasicMessageOption{
				WithName("test"),
				WithAge(30),
			},
			want: &BasicMessage{
				Name: proto.String("test"),
				Age:  proto.Int32(30),
			},
		},
		{
			name: "SetActive",
			opts: []BasicMessageOption{
				WithIsActive(true),
			},
			want: &BasicMessage{
				IsActive: proto.Bool(true),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewBasicMessage(tt.opts...)
			if diff := cmp.Diff(got, tt.want, cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("NewBasicMessage() (-want +got):\n%s", diff)
			}

			existing := &BasicMessage{}
			ApplyBasicMessageOptions(existing, tt.opts...)
			if diff := cmp.Diff(existing, tt.want, cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("ApplyBasicMessageOptions() (-want +got):\n%s", diff)
			}
		})
	}
}

func TestOneofMessage(t *testing.T) {
	tests := []struct {
		name string
		opts []OneofMessageOption
		want *OneofMessage
	}{
		{
			name: "SetText",
			opts: []OneofMessageOption{
				WithText("hello"),
			},
			want: &OneofMessage{
				Choice: &OneofMessage_Text{Text: "hello"},
			},
		},
		{
			name: "SetNumber",
			opts: []OneofMessageOption{
				WithNumber(42),
			},
			want: &OneofMessage{
				Choice: &OneofMessage_Number{Number: 42},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewOneofMessage(tt.opts...)
			if diff := cmp.Diff(got, tt.want, cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("NewOneofMessage() (-want +got):\n%s", diff)
			}

			existing := &OneofMessage{}
			ApplyOneofMessageOptions(existing, tt.opts...)
			if diff := cmp.Diff(existing, tt.want, cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("ApplyOneofMessageOptions() (-want +got):\n%s", diff)
			}
		})
	}
}

func TestComplexMessage(t *testing.T) {
	tests := []struct {
		name string
		opts []ComplexMessageOption
		want *ComplexMessage
	}{
		{
			name: "SetMetadata",
			opts: []ComplexMessageOption{
				WithMetadata(map[string]int32{"key1": 1, "key2": 2}),
			},
			want: &ComplexMessage{
				Metadata: map[string]int32{"key1": 1, "key2": 2},
			},
		},
		{
			name: "SetNestedList",
			opts: []ComplexMessageOption{
				WithNestedList([]*NestedMessage{
					{Description: proto.String("nested1")},
					{Description: proto.String("nested2")},
				}),
			},
			want: &ComplexMessage{
				NestedList: []*NestedMessage{
					{Description: proto.String("nested1")},
					{Description: proto.String("nested2")},
				},
			},
		},
		{
			name: "SetNested",
			opts: []ComplexMessageOption{
				WithNested(&NestedMessage{
					Description: proto.String("nested example"),
				}),
			},
			want: &ComplexMessage{
				Nested: &NestedMessage{
					Description: proto.String("nested example"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewComplexMessage(tt.opts...)
			if diff := cmp.Diff(got, tt.want, cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("NewComplexMessage() (-want +got):\n%s", diff)
			}

			existing := &ComplexMessage{}
			ApplyComplexMessageOptions(existing, tt.opts...)
			if diff := cmp.Diff(existing, tt.want, cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("ApplyComplexMessageOptions() (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRepeatedFieldsMessage(t *testing.T) {
	tests := []struct {
		name string
		opts []RepeatedFieldsMessageOption
		want *RepeatedFieldsMessage
	}{
		{
			name: "SetTags",
			opts: []RepeatedFieldsMessageOption{
				WithTags([]string{"tag1", "tag2"}),
			},
			want: &RepeatedFieldsMessage{
				Tags: []string{"tag1", "tag2"},
			},
		},
		{
			name: "SetValues",
			opts: []RepeatedFieldsMessageOption{
				WithValues([]int32{1, 2, 3}),
			},
			want: &RepeatedFieldsMessage{
				Values: []int32{1, 2, 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRepeatedFieldsMessage(tt.opts...)
			if diff := cmp.Diff(got, tt.want, cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("NewRepeatedFieldsMessage() (-want +got):\n%s", diff)
			}

			existing := &RepeatedFieldsMessage{}
			ApplyRepeatedFieldsMessageOptions(existing, tt.opts...)
			if diff := cmp.Diff(existing, tt.want, cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("ApplyRepeatedFieldsMessageOptions() (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNestedMessage(t *testing.T) {
	tests := []struct {
		name string
		opts []NestedMessageOption
		want *NestedMessage
	}{
		{
			name: "SetBasic",
			opts: []NestedMessageOption{
				WithBasic(&BasicMessage{
					Name: proto.String("basic"),
				}),
			},
			want: &NestedMessage{
				Basic: &BasicMessage{
					Name: proto.String("basic"),
				},
			},
		},
		{
			name: "SetDescription",
			opts: []NestedMessageOption{
				WithDescription("a nested message"),
			},
			want: &NestedMessage{
				Description: proto.String("a nested message"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewNestedMessage(tt.opts...)
			if diff := cmp.Diff(got, tt.want, cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("NewNestedMessage() (-want +got):\n%s", diff)
			}

			existing := &NestedMessage{}
			ApplyNestedMessageOptions(existing, tt.opts...)
			if diff := cmp.Diff(existing, tt.want, cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("ApplyNestedMessageOptions() (-want +got):\n%s", diff)
			}
		})
	}
}
