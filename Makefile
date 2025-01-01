install:
	go install

generate:
	protoc -Iexample --go_out=paths=source_relative:example/identifier --go-options_out=paths=source_relative:example/identifier example/identifier.proto
	protoc -Iexample --go_out=paths=source_relative:example --go-options_out=paths=source_relative:example example/example.proto

# generate:
# 	protoc -Iexample --go_out=paths=source_relative:identifier --go-options_out=paths=source_relative:example/identifier example/identifier.proto
# 	protoc -Iexample --go_out=paths=source_relative:example --go-options_out=paths=source_relative:example example/example.proto