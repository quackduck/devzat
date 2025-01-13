module github.com/quackduck/devzat/devzatapi

go 1.22

toolchain go1.23.1

require (
	github.com/quackduck/devzat/plugin v0.0.0-20230715201334-cc16f25360de
	google.golang.org/grpc v1.69.4
)

require (
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1 // indirect
	google.golang.org/protobuf v1.36.2 // indirect
)

replace github.com/quackduck/devzat/plugin => ../plugin
