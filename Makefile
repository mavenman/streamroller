.DEFAULT_GOAL := build
VERSION := 1.0.0
BINDATA_TAG := v3.0.5
GOVENDOR_TAG := v1.0.8
LINTER_TAG := v1.0.3

# Creates binary
build: gogenerate
	go build -ldflags="-X main.version=$(VERSION)" -o streamroller *.go

build-linux: gogenerate
	gox -os="linux" -arch="amd64" -output="streamroller"

# Gets govendor if not found and installs all dependencies
deps:
	@if [ "$$(which govendor)" = "" ]; then \
		go get -v -u github.com/kardianos/govendor; \
		cd $$GOPATH/src/github.com/kardianos/govendor;\
		git checkout tags/$(GOVENDOR_TAG);\
		go install;\
	fi
	govendor sync

# Creates binarys for all available systems in gox and then zips/tars for distribution.
dist: gogenerate
	which gox && echo "" || go get github.com/mitchellh/gox
	rm -rf tmp dist
	gox -os="linux windows freebsd" -osarch="darwin/amd64" -output='tmp/{{.OS}}-{{.Arch}}-$(VERSION)/{{.Dir}}' -ldflags="-X main.version=$(VERSION)"
	mkdir dist

	# Build for Windows
	@for i in $$(find ./tmp -type f -name "streamroller.exe" | awk -F'/' '{print $$3}'); \
	do \
	  zip -j "dist/streamroller-$$i.zip" "./tmp/$$i/streamroller.exe"; \
	done

	# Build for everything else
	@for i in $$(find ./tmp -type f -not -name "streamroller.exe" | awk -F'/' '{print $$3}'); \
	do \
	  chmod +x "./tmp/$$i/streamroller"; \
	  tar -zcvf "dist/streamroller-$$i.tar.gz" --directory="./tmp/$$i" "./streamroller"; \
	done

	rm -rf tmp

gogenerate: setup-bindata setup-easyjson
	go generate

# Builds and installs binary. Mainly used from people wanting to install from source.
install: deps gogenerate
	go install -ldflags="-X main.version=$(VERSION)" *.go

# Setups go-bindata
setup-bindata:
	@if [ "$$(which go-bindata)" = "" ]; then \
		go get -u -v github.com/jteeuwen/go-bindata; \
		cd $$GOPATH/src/github.com/jteeuwen/go-bindata;\
		git checkout tags/$(BINDATA_TAG);\
		cd go-bindata;\
		go install;\
	fi

# Setups easyjson
setup-easyjson:
	@if [ "$$(which easyjson)" = "" ]; then \
		go get -u -v github.com/mailru/easyjson; \
	fi

# Setups linter configuration for tests
setup-linter:
	@if [ "$$(which gometalinter)" = "" ]; then \
		go get -u -v github.com/alecthomas/gometalinter; \
		cd $$GOPATH/src/github.com/alecthomas/gometalinter;\
		git checkout tags/$(LINTER_TAG);\
		go install;\
		gometalinter --install;\
	fi

# Runs tests
test: setup-linter gogenerate
	gometalinter --vendor --fast --errors --dupl-threshold=100 --cyclo-over=25 --min-occurrences=5 --disable=gas --disable=gotype ./...
