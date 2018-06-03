CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

rmdeps:
	if test -d src; then rm -rf src; fi 

build:	rmdeps deps fmt bin

self:   prep
	if test -d src; then rm -rf src; fi
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-bundles
	cp *.go src/github.com/whosonfirst/go-whosonfirst-bundles/
	cp -r hash src/github.com/whosonfirst/go-whosonfirst-bundles/
	cp -r compress src/github.com/whosonfirst/go-whosonfirst-bundles/
	cp -r vendor/* src/

deps:   rmdeps
	@GOPATH=$(GOPATH) go get -u "github.com/facebookgo/atomicfile"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-index"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-hash"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-log"

vendor-deps: rmdeps deps
	if test ! -d vendor; then mkdir vendor; fi
	if test -d vendor; then rm -rf vendor; fi
	cp -r src vendor
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

fmt:
	go fmt *.go
	go fmt compress/*.go
	go fmt cmd/*.go
	go fmt hash/*.go

bin:	self
	@GOPATH=$(GOPATH) go build -o bin/wof-bundle-metafiles cmd/wof-bundle-metafiles.go

