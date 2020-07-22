#!/bin/sh

# This can't have any space in the newlines or the Mgoogle lines don't get read correctly
docker run --rm -it -u $(id -u):$(id -g) -v $PWD:/src:rw aphistic/protoc \
-I . -I /go/src -I /go/src/github.com/gogo/protobuf/protobuf \
--gogofast_out=\
Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,\
plugins=grpc:. \
softcopy.proto