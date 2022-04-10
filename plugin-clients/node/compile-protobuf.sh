#!/bin/sh
#    "generate-grpc": "grpc_tools_node_protoc --js_out=import_style=es,binary:./src/generated --grpc_out=grpc_js:./src/generated --plugin=protoc-gen-grpc=`which grpc_tools_node_protoc_plugin` -I ../../plugin ../../plugin/plugin.proto && protoc --plugin=protoc-gen-ts=./node_modules/.bin/protoc-gen-ts --ts_out=./src/generated -I ../../plugin ../../plugin/plugin.proto",

# grpc_tools_node_protoc --js_out=import_style=es,binary:./src/generated --grpc_out=grpc_js:./src/generated --plugin=protoc-gen-grpc=$(which grpc_tools_node_protoc_plugin) -I ../../plugin ../../plugin/plugin.proto
protoc --plugin=protoc-gen-ts=./node_modules/.bin/protoc-gen-ts --ts_out=./src/generated -I ../../plugin ../../plugin/plugin.proto
protoc --descriptor_set_out=./src/generated/plugin-desc.pb -I ../../plugin ../../plugin/plugin.proto