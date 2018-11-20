CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep rmdeps
	if test ! -d src; then mkdir src; fi
	if test ! -d src/github.com/whosonfirst/go-whosonfirst-meta; then mkdir -p src/github.com/whosonfirst/go-whosonfirst-meta/; fi
	cp  meta.go src/github.com/whosonfirst/go-whosonfirst-meta/
	cp -r build src/github.com/whosonfirst/go-whosonfirst-meta/
	cp -r meta src/github.com/whosonfirst/go-whosonfirst-meta/
	cp -r options src/github.com/whosonfirst/go-whosonfirst-meta/
	cp -r stats src/github.com/whosonfirst/go-whosonfirst-meta/
	cp -r vendor/* src/

rmdeps:
	if test -d src; then rm -rf src; fi

build:	fmt bin

dist-build:
	OS=darwin make dist-os
	OS=windows make dist-os
	OS=linux make dist-os

dist-os:
	mkdir -p dist/$(OS)
	GOOS=$(OS) GOPATH=$(GOPATH) GOARCH=386 go build -o dist/$(OS)/wof-build-metafiles cmd/wof-build-metafiles.go
	chmod +x dist/$(OS)/wof-build-metafiles
	cd dist/$(OS) && shasum -a 256 wof-build-metafiles > wof-build-metafiles.sha256

rmdist:
	if test -d dist; then rm -rf dist; fi

deps:   rmdeps
	@GOPATH=$(GOPATH) go get -u "github.com/facebookgo/atomicfile"
	@GOPATH=$(GOPATH) go get -u "github.com/tidwall/gjson"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-geojson-v2"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-index"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-csv"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-placetypes"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-repo"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-uri"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-log"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/warning"
	rm -rf src/github.com/whosonfirst/go-whosonfirst-geojson-v2/vendor/github.com/whosonfirst/warning
	rm -rf src/github.com/whosonfirst/go-whosonfirst-placetypes/vendor/github.com/whosonfirst/warning

vendor-deps: rmdeps deps
	if test ! -d vendor; then mkdir vendor; fi
	if test -d vendor; then rm -rf vendor; fi
	cp -r src vendor
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

bin: 	self
	@GOPATH=$(GOPATH) go build -o bin/wof-build-metafiles cmd/wof-build-metafiles.go
	@GOPATH=$(GOPATH) go build -o bin/wof-update-metafile cmd/wof-update-metafile.go
	@GOPATH=$(GOPATH) go build -o bin/wof-meta-prepare cmd/wof-meta-prepare.go
	@GOPATH=$(GOPATH) go build -o bin/wof-meta-stats cmd/wof-meta-stats.go

fmt:
	go fmt *.go
	go fmt build/*.go
	go fmt cmd/*.go
	go fmt meta/*.go
	go fmt options/*.go
	go fmt stats/*.go
