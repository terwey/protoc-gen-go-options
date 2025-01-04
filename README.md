# `protoc-gen-go-options`

`protoc-gen-go-options` is a `protoc` plugin that generates Go code to simplify the creation of Protocol Buffer messages using the **options pattern**. This plugin helps you write clean and expressive code when working with Protobuf messages in Go.

## Installation

Install the plugin directly via `go install`:
```bash
go install github.com/terwey/protoc-gen-go-options@latest
```

Ensure the binary is in your `PATH` so `protoc` can find it.

## Example

### Protobuf Input

For this message definition:
```proto
syntax = "proto3";

message OneofMessage {
  oneof choice {
    string text = 1;
    int32 number = 2;
  }
}
```

### Generated Output

The plugin generates helper functions for setting fields:
```go
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
```

### Usage in Go

Construct Protobuf messages using functional options:
```go
msg := NewOneofMessage(
	WithText("example text"),
)
fmt.Printf("%v", msg)
// Output: text:"example text"

msg2 := NewOneofMessage(
	WithNumber(42),
)
fmt.Println(msg2)
// Output: number:42

ApplyOneofMessageOptions(msg, WithText("foo bar"))
fmt.Println(msg)
// Output: text:"foo bar"
```

For more examples, refer to the [`example`](./example) directory.

### Special Comments for Custom Behavior

The plugin recognizes the following special comments in your `.proto` files to customize the generated code:

#### `GO_OPTIONS_OPTIONLESS`
  Add this comment to a message to skip generating options for it. This is useful for messages where options are unnecessary.

  ```proto
  // GO_OPTIONS_OPTIONLESS
  message ExampleMessage {
      string field = 1;
  }
  ```

#### `GO_OPTIONS_SKIP_INIT`
  Add this comment to a message to skip generating the default constructor (`New[Message]`). This is useful if you prefer to handle initialization manually.

  ```proto
  // GO_OPTIONS_SKIP_INIT
  message ExampleMessage {
      string field = 1;
  }
  ```

### `GO_OPTIONS_JSON_PERSISTENT`

The `GO_OPTIONS_JSON_PERSISTENT` option enables the generated code to support JSON persistence for fields by generating JSON-specific helper methods.

For the following proto definition:

```proto
message JsonExample {
  // GO_OPTIONS_JSON_PERSISTENT
  BasicMessage basic = 1;
}
```

The generated Go code includes:

```go
// GetBasicAsJSON returns the Basic field as a JSON byte slice.
func (m *JsonExample) GetBasicAsJSON() ([]byte, error) {
  out, err := json.Marshal(m.Basic)
  if err != nil {
    return nil, fmt.Errorf("failed to marshal Basic field: %w", "%s", err)
  }
  return out, nil
}

// SetBasicFromJSON sets the Basic field from a JSON byte slice.
func (m *JsonExample) SetBasicFromJSON(v []byte) error {
  return json.Unmarshal(v, &m.Basic)
}
```

## License

This project is licensed under the [MIT License](LICENSE).