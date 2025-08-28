FROM golang:1.24-alpine

RUN apk add --no-cache \
    protobuf \
    protobuf-dev \
    make \
    git \
    gcc \
    musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY posts.proto ./

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

RUN mkdir -p ./gen/go/posts

RUN protoc \
    --go_out=./gen/go/posts \
    --go_opt=paths=source_relative \
    --go-grpc_out=./gen/go/posts \
    --go-grpc_opt=paths=source_relative \
    ./posts.proto

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/api/

EXPOSE 50051
CMD ["./main"]