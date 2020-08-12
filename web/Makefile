include ../.env

.PHONY: build
build: dist

.PHONY: clean
clean:
	rm -rf dist

.PHONY: distclean
distclean: clean
	rm -rf node_modules

.PHONY: deps
	$(NPM) ci

.PHONY: dev
dev:
	$(PARCEL) watch index.html
	
dist: deps
	$(PARCEL) build index.html

