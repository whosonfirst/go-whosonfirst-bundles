CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

rmdeps:
	if test -d src; then rm -rf src; fi 

build:	rmdeps deps fmt bin

self:   prep
	if test -d src; then rm -rf src; fi
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-bundles/
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-bundles/compress
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-bundles/hash
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-bundles/prune
	cp *.go src/github.com/whosonfirst/go-whosonfirst-bundles/
	cp hash/*.go src/github.com/whosonfirst/go-whosonfirst-bundles/hash/
	cp compress/*.go src/github.com/whosonfirst/go-whosonfirst-bundles/compress/
	cp prune/*.go src/github.com/whosonfirst/go-whosonfirst-bundles/prune/
	cp -r vendor/src/* src/

deps:   rmdeps
	@GOPATH=$(GOPATH) go get -u "github.com/aws/aws-sdk-go"
	@GOPATH=$(GOPATH) go get -u "github.com/facebookgo/atomicfile"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-clone"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-hash"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-log"

vendor-deps: deps
	if test ! -d vendor; then mkdir vendor; fi
	if test -d vendor/src; then rm -rf vendor/src; fi
	cp -r src vendor/src
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

fmt:
	go fmt *.go
	go fmt compress/*.go
	go fmt cmd/*.go
	go fmt hash/*.go
	go fmt prune/*.go

bin:	self
	@GOPATH=$(GOPATH) go build -o bin/wof-bundle-metafiles cmd/wof-bundle-metafiles.go
	@GOPATH=$(GOPATH) go build -o bin/wof-bundles-prune cmd/wof-bundles-prune.go
