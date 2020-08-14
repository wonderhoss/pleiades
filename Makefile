include .env

BINARY := pleiades
VERSION := $(shell git describe --always --dirty --tags 2>/dev/null || echo "undefined")
ECHO := echo

.NOTPARALLEL:

.PHONY: all
all: test build

.PHONY: release
release: clean lint vet test web generate $(BINARY)

.PHONY: build
build: clean web generate $(BINARY)

.PHONY: web
web: web/dist

.PHONY: generate
generate:
	make -C pkg generate

.PHONY: clean
clean:
	rm -f $(BINARY)
	rm -rf events
	rm -f .pleiades_resumeID
	make -C web clean
	make -C pkg clean

.PHONY: distclean
distclean: clean
	make -C web distclean
	rm -f .env
	rm -f dump.rdb

# Run go fmt against code
.PHONY: fmt
fmt:
	$(GO) fmt ./pkg/... ./cmd/...

# Run go vet against code
.PHONY: vet
vet:
	$(GO) vet -tags dev -composites=false ./pkg/... ./cmd/...

.PHONY: lint
lint:
	@ $(ECHO) "\033[36mLinting code\033[0m"
	$(LINTER) run --disable-all --build-tags dev \
                --exclude-use-default=false \
                --enable=govet \
                --enable=ineffassign \
                --enable=deadcode \
                --enable=golint \
                --enable=goconst \
                --enable=gofmt \
                --enable=goimports \
                --skip-dirs=pkg/client/ \
                --deadline=120s \
                --tests ./...
	@ $(ECHO)

.PHONY: check
check: fmt lint vet test

.PHONY: test
test:
	@ $(ECHO) "\033[36mRunning test suite in Ginkgo\033[0m"
	$(GINKGO) -v -p -race -randomizeAllSpecs ./pkg/... ./cmd/...
	@ $(ECHO)

.PHONY: dev
dev:
	make -C web dev &
	GO111MODULE=on $(GO) run -tags dev github.com/gargath/pleiades/cmd frontend

# Build binary
$(BINARY): fmt vet generate
	GO111MODULE=on CGO_ENABLED=0 $(GO) build -o $(BINARY) -ldflags="-X main.VERSION=${VERSION}" github.com/gargath/pleiades/cmd

web/dist:
	make -C web build
