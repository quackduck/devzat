
GOPATH=$(shell pwd)
SHELL := /bin/bash
PATH := bin:$(PATH)

setup:
@GOPATH=$(GOPATH) go get "devzat/plugin@v0.0.0-00010101000000-000000000000"
@GOPATH=$(GOPATH) go get "github.com/TwiN/go-away@v1.6.10"
@GOPATH=$(GOPATH) go get "github.com/acarl005/stripansi@v0.0.0-20180116102854-5a71ef0e047d"
@GOPATH=$(GOPATH) go get "github.com/alecthomas/chroma@v0.10.0"
@GOPATH=$(GOPATH) go get "github.com/bwmarrin/discordgo@v0.27.2-0.20230704233747-e39e715086d2"
@GOPATH=$(GOPATH) go get "github.com/dghubble/go-twitter@v0.0.0-20221104224141-912508c3888b"
@GOPATH=$(GOPATH) go get "github.com/dghubble/oauth1@v0.7.2"
@GOPATH=$(GOPATH) go get "github.com/gliderlabs/ssh@v0.3.5"
@GOPATH=$(GOPATH) go get "github.com/jwalton/gchalk@v1.3.0"
@GOPATH=$(GOPATH) go get "github.com/leaanthony/go-ansi-parser@v1.6.1"
@GOPATH=$(GOPATH) go get "github.com/quackduck/go-term-markdown@v0.14.2"
@GOPATH=$(GOPATH) go get "github.com/quackduck/term@v0.0.0-20230512153006-5935fcd4d5e9"
@GOPATH=$(GOPATH) go get "github.com/shurcooL/tictactoe@v0.0.0-20210613024444-e573ff1376a3"
@GOPATH=$(GOPATH) go get "github.com/slack-go/slack@v0.12.2"
@GOPATH=$(GOPATH) go get "golang.org/x/image@v0.9.0"
@GOPATH=$(GOPATH) go get "google.golang.org/grpc@v1.56.2"
@GOPATH=$(GOPATH) go get "gopkg.in/yaml.v2@v2.4.0"
@GOPATH=$(GOPATH) go get "github.com/MichaelMure/go-term-text@v0.3.1"
@GOPATH=$(GOPATH) go get "github.com/anmitsu/go-shlex@v0.0.0-20200514113438-38f4b401e2be"
@GOPATH=$(GOPATH) go get "github.com/caarlos0/sshmarshal@v0.1.0"
@GOPATH=$(GOPATH) go get "github.com/cenkalti/backoff/v4@v4.2.1"
@GOPATH=$(GOPATH) go get "github.com/dghubble/sling@v1.4.1"
@GOPATH=$(GOPATH) go get "github.com/disintegration/imaging@v1.6.2"
@GOPATH=$(GOPATH) go get "github.com/dlclark/regexp2@v1.10.0"
@GOPATH=$(GOPATH) go get "github.com/eliukblau/pixterm@v1.3.1"
@GOPATH=$(GOPATH) go get "github.com/fatih/color@v1.15.0"
@GOPATH=$(GOPATH) go get "github.com/golang/protobuf@v1.5.3"
@GOPATH=$(GOPATH) go get "github.com/gomarkdown/markdown@v0.0.0-20230322041520-c84983bdbf2a"
@GOPATH=$(GOPATH) go get "github.com/google/go-querystring@v1.1.0"
@GOPATH=$(GOPATH) go get "github.com/gorilla/websocket@v1.5.0"
@GOPATH=$(GOPATH) go get "github.com/jwalton/go-supportscolor@v1.2.0"
@GOPATH=$(GOPATH) go get "github.com/kyokomi/emoji/v2@v2.2.12"
@GOPATH=$(GOPATH) go get "github.com/lucasb-eyer/go-colorful@v1.2.0"
@GOPATH=$(GOPATH) go get "github.com/mattn/go-colorable@v0.1.13"
@GOPATH=$(GOPATH) go get "github.com/mattn/go-isatty@v0.0.19"
@GOPATH=$(GOPATH) go get "github.com/mattn/go-runewidth@v0.0.14"
@GOPATH=$(GOPATH) go get "github.com/rivo/uniseg@v0.4.4"
@GOPATH=$(GOPATH) go get "golang.org/x/crypto@v0.11.0"
@GOPATH=$(GOPATH) go get "golang.org/x/net@v0.12.0"
@GOPATH=$(GOPATH) go get "golang.org/x/sys@v0.10.0"
@GOPATH=$(GOPATH) go get "golang.org/x/term@v0.10.0"
@GOPATH=$(GOPATH) go get "golang.org/x/text@v0.11.0"
@GOPATH=$(GOPATH) go get "google.golang.org/protobuf@v1.31.0"
@GOPATH=$(GOPATH) go get "google.golang.org/genproto/googleapis/rpc@v0.0.0-20230706204954-ccb25ca9f130"

  
build:
   @GOPATH=$(GOPATH) go build ./...
   @GOPATH=$(GOPATH) go install ./...

run:
   bin/main


#This runs setup, build, and launches the application
dockertest:
   make setup
   make build
   make run
