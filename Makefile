GO ?= go
GOFMT ?= gofmt
GO_FILES ?= $$(find . -name '*.go' | grep -v vendor | grep -v /examples/)
GOLANG_CI_LINT ?= golangci-lint
GO_IMPORTS ?= goimports
GO_IMPORTS_LOCAL ?= github.com/ZupIT/horusec-engine
HORUSEC ?= horusec
GO_FUMPT ?= gofumpt
GO_GCI ?= gci
ADDLICENSE ?= addlicense
GO_LIST_TO_TEST ?= $$(go list ./... | grep -v /text/examples/)

lint:
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOLANG_CI_LINT) run -v --timeout=5m -c .golangci.yml ./...

coverage:
	curl -fsSL https://raw.githubusercontent.com/ZupIT/horusec-devkit/main/scripts/coverage.sh | bash -s 66.3 .

test:
	$(GO) clean -testcache
	$(GO) test -v $(GO_LIST_TO_TEST) -race -timeout=5m -parallel=1 -failfast -short

format: install-format-dependencies
	$(GOFMT) -s -l -w $(GO_FILES)
	$(GO_IMPORTS) -w -local $(GO_IMPORTS_LOCAL) $(GO_FILES)
	$(GO_FUMPT) -l -w $(GO_FILES)
	$(GO_GCI) -w -local $(GO_IMPORTS_LOCAL) $(GO_FILES)

install-format-dependencies:
	$(GO) install golang.org/x/tools/cmd/goimports@latest
	$(GO) install mvdan.cc/gofumpt@latest
	$(GO) install github.com/daixiang0/gci@v0.2.9

security:
    ifeq (, $(shell which $(HORUSEC)))
		curl -fsSL https://raw.githubusercontent.com/ZupIT/horusec/master/deployments/scripts/install.sh | bash -s latest
		$(HORUSEC) start -p="./" -e="true"
    else
		$(HORUSEC) start -p="./" -e="true"
    endif

license:
	$(GO) install github.com/google/addlicense@latest
	@$(ADDLICENSE) -check -f ./copyright.txt $(shell find -regex '.*\.\(go\|js\|ts\|yml\|yaml\|sh\|dockerfile\)')

license-fix:
	$(GO) install github.com/google/addlicense@latest
	@$(ADDLICENSE) -f ./copyright.txt $(shell find -regex '.*\.\(go\|js\|ts\|yml\|yaml\|sh\|dockerfile\)')

pipeline: format license-fix lint test coverage security
