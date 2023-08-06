
GOPATH=$(shell pwd)
SHELL := /bin/bash
PATH := bin:$(PATH)

setup:
   @GOPATH=$(GOPATH) go get "github.com/gorilla/mux"
   @GOPATH=$(GOPATH) go get "github.com/elazarl/go-bindata-assetfs"
   @GOPATH=$(GOPATH) go get github.com/jteeuwen/go-bindata/...
   @GOPATH=$(GOPATH) go get github.com/elazarl/go-bindata-assetfs/...
  
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
