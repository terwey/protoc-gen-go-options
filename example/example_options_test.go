package example

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/terwey/protoc-gen-go-options/example/identifier"
	"google.golang.org/protobuf/proto"
)

func ExampleNewOneofMessage() {
	msg := NewOneofMessage(
		WithText("example text"),
	)
	fmt.Printf("%v", msg)
	// Output: text:"example text"
}

func ExampleNewOneOfMessage2() {
	msg2 := NewOneofMessage(
		WithNumber(42),
	)
	fmt.Println(msg2)
	// Output: number:42
}

func ExampleApplyOneofMessageOptions() {
	msg := NewOneofMessage(
		WithText("example text"),
	)

	ApplyOneofMessageOptions(msg, WithText("foo bar"))
	fmt.Println(msg)
	// Output: text:"foo bar"
}

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
				}...),
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
		{
			name: "WithNewNestedForComplexMessage",
			opts: []ComplexMessageOption{
				WithNewNestedForComplexMessage(WithDescription("complex nested")),
			},
			want: &ComplexMessage{
				Nested: &NestedMessage{
					Description: proto.String("complex nested"),
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
				WithTags([]string{"tag1", "tag2"}...),
			},
			want: &RepeatedFieldsMessage{
				Tags: []string{"tag1", "tag2"},
			},
		},
		{
			name: "SetValues",
			opts: []RepeatedFieldsMessageOption{
				WithValues([]int32{1, 2, 3}...),
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
		{
			name: "WithNewBasicForNestedMessage",
			opts: []NestedMessageOption{
				WithNewBasicForNestedMessage(WithName("nested basic")),
			},
			want: &NestedMessage{
				Basic: &BasicMessage{
					Name: proto.String("nested basic"),
				},
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

func TestDuplicateFieldnameFoo(t *testing.T) {
	tests := []struct {
		name string
		opts []FooOption
		want *Foo
	}{
		{
			name: "Set ID",
			opts: []FooOption{
				WithIdForFoo(&identifier.Identifier{
					Id: proto.String("IdForFoo"),
				}),
			},
			want: &Foo{
				Id: &identifier.Identifier{Id: proto.String("IdForFoo")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewFoo(tt.opts...)
			if diff := cmp.Diff(got, tt.want, cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("NewBasicMessage() (-want +got):\n%s", diff)
			}

			existing := &Foo{}
			ApplyFooOptions(existing, tt.opts...)
			if diff := cmp.Diff(existing, tt.want, cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("ApplyBasicMessageOptions() (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDuplicateFieldnameBar(t *testing.T) {
	tests := []struct {
		name string
		opts []BarOption
		want *Bar
	}{
		{
			name: "Set ID",
			opts: []BarOption{
				WithIdForBar(&identifier.Identifier{
					Id: proto.String("IdForBar"),
				}),
			},
			want: &Bar{
				Id: &identifier.Identifier{
					Id: proto.String("IdForBar"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewBar(tt.opts...)
			if diff := cmp.Diff(got, tt.want, cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("NewBasicMessage() (-want +got):\n%s", diff)
			}

			existing := &Bar{}
			ApplyBarOptions(existing, tt.opts...)
			if diff := cmp.Diff(existing, tt.want, cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("ApplyBasicMessageOptions() (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFoo(t *testing.T) {
	tests := []struct {
		name string
		opts []FooOption
		want *Foo
	}{
		{
			name: "WithNewIdForFoo",
			opts: []FooOption{
				WithNewIdForFoo(),
			},
			want: &Foo{
				Id: identifier.NewIdentifier(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewFoo(tt.opts...)
			if diff := cmp.Diff(got, tt.want, cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("%s (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func TestBar(t *testing.T) {
	tests := []struct {
		name string
		opts []BarOption
		want *Bar
	}{
		{
			name: "WithNewIdForBar",
			opts: []BarOption{
				WithNewIdForBar(),
			},
			want: &Bar{
				Id: identifier.NewIdentifier(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewBar(tt.opts...)
			if diff := cmp.Diff(got, tt.want, cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("%s (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func TestSomeMessage(t *testing.T) {
	tests := []struct {
		name string
		opts []SomeMessageOption
		want *SomeMessage
	}{
		{
			name: "WithNewIdentifierForSomeMessage",
			opts: []SomeMessageOption{
				WithNewIdentifierForSomeMessage(),
			},
			want: &SomeMessage{
				Identifier: identifier.NewIdentifier(),
			},
		},
		{
			name: "WithIdentifier",
			opts: []SomeMessageOption{
				WithIdentifier(&identifier.Identifier{Id: proto.String("custom-id")}),
			},
			want: &SomeMessage{
				Identifier: &identifier.Identifier{Id: proto.String("custom-id")},
			},
		},
		{
			name: "WithInclude",
			opts: []SomeMessageOption{
				WithInclude([]*identifier.Identifier{
					{Id: proto.String("id1")},
					{Id: proto.String("id2")},
				}...),
			},
			want: &SomeMessage{
				Include: []*identifier.Identifier{
					{Id: proto.String("id1")},
					{Id: proto.String("id2")},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewSomeMessage(tt.opts...)
			if diff := cmp.Diff(got, tt.want, cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("%s (-want +got):\n%s", tt.name, diff)
			}

			existing := &SomeMessage{}
			ApplySomeMessageOptions(existing, tt.opts...)
			if diff := cmp.Diff(existing, tt.want, cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("ApplySomeMessageOptions %s (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func TestFooBarWithEnum(t *testing.T) {
	tests := []struct {
		name string
		opts []FooBarWithEnumOption
		want *FooBarWithEnum
	}{
		{
			name: "WithStatus",
			opts: []FooBarWithEnumOption{
				WithStatus(FooBarWithEnum_ACTIVE.Enum()),
			},
			want: &FooBarWithEnum{
				Status: FooBarWithEnum_ACTIVE.Enum(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewFooBarWithEnum(tt.opts...)
			if diff := cmp.Diff(got, tt.want, cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("%s (-want +got):\n%s", tt.name, diff)
			}

			existing := &FooBarWithEnum{}
			ApplyFooBarWithEnumOptions(existing, tt.opts...)
			if diff := cmp.Diff(existing, tt.want, cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("ApplyFooBarWithEnumOptions %s (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}

func TestIdentifier(t *testing.T) {
	tests := []struct {
		name string
		opts []identifier.IdentifierOption
		want *identifier.Identifier
	}{
		{
			name: "WithId",
			opts: []identifier.IdentifierOption{
				identifier.WithId("id-value"),
			},
			want: &identifier.Identifier{
				Id: proto.String("id-value"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := identifier.NewIdentifier(tt.opts...)
			if diff := cmp.Diff(got, tt.want, cmp.Comparer(proto.Equal)); diff != "" {
				t.Errorf("%s (-want +got):\n%s", tt.name, diff)
			}
		})
	}
}
