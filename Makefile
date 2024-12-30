install:
	go install

generate:
	protoc -Iexample --go_out=paths=source_relative:example --go-options_out=paths=source_relative:example example/example.proto