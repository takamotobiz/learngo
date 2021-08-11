GO ?= go
GOFMT ?= gofmt "-s"
PACKAGES ?= $(shell $(GO) list ./...)
VETPACKAGES ?= $(shell $(GO) list ./... | grep -v /examples/)
GOFILES := $(shell find . -name "*.go")
TESTFOLDER := $(shell $(GO) list ./... | grep -E 'osmpbfparser-go$$|cmd' | grep -v examples)
TESTTAGS ?= ""
PROJ = osmpbfparser-go
ZONE ?= asia/taiwan

##@ Show
.PHONY: count-line

count-line:  ## Count *.go line in project
	    find . -name '*.go' | xargs wc -l

##@ Run
.PHONY: run-example

run-example:  ## Run ./cmd/example/main.go
	go run ./cmd/example/main.go

##@ Test
.PHONY: test install-richgo

install-richgo:  ## Install richgo
	go get -u github.com/kyoh86/richgo

test:  ## Run test
	echo "mode: count" > coverage.out
	for d in $(TESTFOLDER); do \
		$(GO) test -tags $(TESTTAGS) -v -covermode=count -coverprofile=profile.out $$d | richgo testfilter > tmp.out; \
		cat tmp.out; \
		if grep -q "^--- FAIL" tmp.out; then \
			rm tmp.out; \
			exit 1; \
		elif grep -q "build failed" tmp.out; then \
			rm tmp.out; \
			exit 1; \
		elif grep -q "setup failed" tmp.out; then \
			rm tmp.out; \
			exit 1; \
		fi; \
		if [ -f profile.out ]; then \
			cat profile.out | grep -v "mode:" >> coverage.out; \
			rm profile.out; \
		fi; \
	done

integration-test:  ## Run intergration test
	$(MAKE) TESTTAGS=integration test

integration-geojson-test:  ## Run geojson intergration test
	$(MAKE) TESTTAGS=geojson_integration test

##@ lint
.PHONY: linter-run install-lint

install-lint:  ## Install golangci-lint to ./bin
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.27.0

linter-run:  ## Run linter for all
	./bin/golangci-lint run ./...

##@ OSM
.PHONY: download-pbf

downloasd-pbf:  ## (ZONE=asia/taiwan) Download osm pbf file. Use ZONE variable to control which area to download. See https://download.geofabrik.de/
	wget http://download.geofabrik.de/${ZONE}-latest.osm.pbf -P ./static


##@ Help

.PHONY: help

help:  ## Display this help
	    @awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-0-9]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

