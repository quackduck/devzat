module github.com/quackduck/devzat/devzatapi

go 1.20

require (
	devzat/plugin v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.56.1
)

require (
	github.com/golang/protobuf v1.5.3 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)

replace devzat/plugin => ../plugin