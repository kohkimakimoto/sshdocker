.PHONY:help dev dist packaging fmt test testv deps deps_update website
.DEFAULT_GOAL := help

# This is a magic code to output help message at default
# see https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

dev: ## Build dev binary
	@bash -c $(CURDIR)/build/scripts/dev.sh

dist: ## Build dist binaries
	@bash -c $(CURDIR)/build/scripts/dist.sh

packaging: ## Create packages (now support RPM only)
	@bash -c $(CURDIR)/build/scripts/packaging.sh

clean: ## Clean the built binaries.
	@bash -c $(CURDIR)/build/scripts/clean.sh

fmt: ## Run `go fmt`
	go fmt $$(go list ./... | grep -v vendor)

test: ## Run tests
	go test -cover -timeout=360s $$(go list ./... | grep -v vendor)

testv: ## Run tests with verbose outputting.
	go test -cover -timeout=360s -v $$(go list ./... | grep -v vendor)

deps: ## Install dependences.
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/mitchellh/gox
	go get -u github.com/axw/gocov/gocov
	go get -u gopkg.in/matm/v1/gocov-html
	dep ensure

resetdeps: ## reset dependences.
	rm -rf Gopkg.*
	rm -rf vendor
	dep init
	dep ensure

website: ## Build webside resources.
	cd website && make deps && make
