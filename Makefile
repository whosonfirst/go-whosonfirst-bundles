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
	cp -r vendor/* src/

deps:   rmdeps
	@GOPATH=$(GOPATH) go get -u "github.com/facebookgo/atomicfile"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-csv"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-index"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-geojson-v2"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-sqlite-features"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-hash"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-log"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-meta"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-uri"
	mv src/github.com/whosonfirst/go-whosonfirst-sqlite-features/vendor/github.com/whosonfirst/go-whosonfirst-sqlite src/github.com/whosonfirst/
	rm -rf src/github.com/whosonfirst/go-whosonfirst-index/vendor/github.com/whosonfirst/go-whosonfirst-sqlite
	rm -rf src/github.com/whosonfirst/go-whosonfirst-meta/vendor/github.com/whosonfirst/go-whosonfirst-csv
	rm -rf src/github.com/whosonfirst/go-whosonfirst-meta/vendor/github.com/whosonfirst/go-whosonfirst-geojson-v2
	rm -rf src/github.com/whosonfirst/go-whosonfirst-meta/vendor/github.com/whosonfirst/go-whosonfirst-index
	rm -rf src/github.com/whosonfirst/go-whosonfirst-meta/vendor/github.com/whosonfirst/go-whosonfirst-log
	rm -rf src/github.com/whosonfirst/go-whosonfirst-meta/vendor/github.com/whosonfirst/go-whosonfirst-uri

vendor-deps: rmdeps deps
	if test ! -d vendor; then mkdir vendor; fi
	if test -d vendor; then rm -rf vendor; fi
	cp -r src vendor
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

fmt:
	go fmt *.go
	go fmt cmd/*.go

bin:	self
	@GOPATH=$(GOPATH) go build -o bin/wof-bundle cmd/wof-bundle.go

