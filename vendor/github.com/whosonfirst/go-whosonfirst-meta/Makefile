dist-build:
	OS=darwin make dist-os
	OS=windows make dist-os
	OS=linux make dist-os

dist-os:
	@echo "build tools for $(OS)"
	mkdir -p dist/$(OS)
	GOOS=$(OS) GOARCH=386 go build -mod vendor -o dist/$(OS)/wof-build-metafiles cmd/wof-build-metafiles/main.go
	chmod +x dist/$(OS)/wof-build-metafiles
	cd dist/$(OS) && shasum -a 256 wof-build-metafiles > wof-build-metafiles.sha256

rmdist:
	if test -d dist; then rm -rf dist; fi

tools:
	go build -mod vendor -o bin/wof-build-metafiles cmd/wof-build-metafiles/main.go
	go build -mod vendor -o bin/wof-update-metafile cmd/wof-update-metafile/main.go
	go build -mod vendor -o bin/wof-meta-prepare cmd/wof-meta-prepare/main.go
	go build -mod vendor -o bin/wof-meta-stats cmd/wof-meta-stats/main.go

fmt:
	go fmt *.go
	go fmt build/*.go
	go fmt cmd/wof-build-metafiles/*.go
	go fmt cmd/wof-update-metafile/*.go
	go fmt cmd/wof-meta-prepare/*.go
	go fmt cmd/wof-meta-stats/*.go
	go fmt meta/*.go
	go fmt options/*.go
	go fmt stats/*.go
