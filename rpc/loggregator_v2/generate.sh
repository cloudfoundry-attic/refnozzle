#!/bin/bash

set -ex

install_dependencies() {
    go get github.com/golang/protobuf/proto
    go get github.com/golang/protobuf/protoc-gen-go
}

copy_proto_files() {
    cp "$GOPATH/src/github.com/cloudfoundry/loggregator-api/v2/envelope.proto" "$1/"
    cp "$GOPATH/src/github.com/cloudfoundry/loggregator-api/v2/egress.proto" "$1/"
}

generate_from_proto() {
    protoc "$1/envelope.proto" "$1/egress.proto" \
        --go_out=plugins=grpc:"$PWD" --proto_path="$1/"
}

cleanup() {
    rm -r "$1"
}

main() {
    mkdir -p tmp
    tmp_dir=tmp

    install_dependencies
    copy_proto_files "$tmp_dir"
    generate_from_proto "$tmp_dir"
    cleanup "$tmp_dir"
}
main
