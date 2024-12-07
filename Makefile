BUILDPREFIX ?= build

name := driver

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

TZ ?= UTC

project_url := "https://github.com/zhughes3/go-accelerate"

run: clean build
	@set -a && source ./cmd/driver/.env && set +a && ./build/darwin/bin/driver

build: $(BUILDPREFIX)/$(GOOS)/version.txt $(BUILDPREFIX)/$(GOOS)/bin/$(name)

clean:
	@rm -Rf $(BUILDPREFIX)

$(BUILDPREFIX)/$(GOOS)/bin/$(name):
	$(eval version := $(shell cat $(BUILDPREFIX)/$(GOOS)/version.txt))
	$(eval compiler := $(shell go version))
	$(eval buildtime := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ"))
	$(eval commit := $(shell git rev-parse --short HEAD))
	$(eval LDFLAGS += -X 'github.com/zhughes3/go-accelerate/pkg/app.version=$(version)' \
		-X 'github.com/zhughes3/go-accelerate/pkg/app.compiler=$(compiler)' \
		-X 'github.com/zhughes3/go-accelerate/pkg/app.buildtime=$(buildtime)' \
		-X 'github.com/zhughes3/go-accelerate/pkg/app.commit=$(commit)')

	@mkdir -p $(BUILDPREFIX)/$(GOOS)/bin && \
	CGO_ENABLED=0 GOARCH=$(GOARCH) GOOS=$(GOOS) go build -ldflags "$(LDFLAGS)" -o $(BUILDPREFIX)/$(GOOS)/bin/$(name) ./cmd/driver

$(BUILDPREFIX)/$(GOOS)/version.txt:
	@git fetch --tags
	@mkdir -p $(BUILDPREFIX)/$(GOOS)
	@git describe --tags --always --dirty --match "v[0-9]*.[0-9]*.[0-9]*" > $(BUILDPREFIX)/$(GOOS)/version.txt

.PHONY: clean build run
