.PHONY: gen

gen:
	mkdir -p ./gen/go/posts
	protoc --go_out=./gen/go/posts --go_opt=paths=source_relative \
	--go-grpc_out=./gen/go/posts --go-grpc_opt=paths=source_relative \
	./posts.proto