include ../.env

.PHONY: clean
clean:
	find . -name \*vfsdata.go -exec rm -f {} \;

web/static/assets_vfsdata.go:
		GO111MODULE=on $(GO) generate ./web/static/...

.PHONY: generate
generate: web/static/assets_vfsdata.go
