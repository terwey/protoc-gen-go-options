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
fmt.Println(msg) // Output: &OneofMessage{Choice: &OneofMessage_Text{Text: "example text"}}

msg2 := NewOneofMessage(
    WithNumber(42),
)
fmt.Println(msg2) // Output: &OneofMessage{Choice: &OneofMessage_Number{Number: 42}}
```

For more examples, refer to the [`example`](./example) directory.

## License

This project is licensed under the [MIT License](LICENSE).