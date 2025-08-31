SHELL := bash

VERSION         ?= $(shell pulumictl get version)

PACK            := talos-cluster
PROJECT         := github.com/spigell/pulumi-${PACK}

PROVIDER        := pulumi-resource-${PACK}
CODEGEN         := pulumi-gen-${PACK}
VERSION_PATH    := provider/pkg/version.Version

WORKING_DIR     := $(shell pwd)
SCHEMA_PATH     := ${WORKING_DIR}/provider/cmd/${PROVIDER}/schema.json

GOPATH          := $(shell go env GOPATH)

generate:: gen_go_sdk gen_dotnet_sdk gen_nodejs_sdk gen_python_sdk

build:: build_provider build_dotnet_sdk build_nodejs_sdk build_python_sdk

install:: install_provider install_nodejs_sdk

# Lint
lint::
	cd provider && golangci-lint run
	cd tests && golangci-lint run

# Tests
unit_tests:: generate_schema
	cd provider && set -o pipefail ; go test $$(go list ./... | grep -v tests | grep -v crds/generated) | grep -v 'no test files'

integration_tests:: integration_tests_nodejs integration_tests_go
	cd integration-tests && go test -timeout=25m -v

integration_tests_go:: build_provider_tests gen_go_sdk
	cd integration-tests && go test -timeout=25m -v -run $(TEST)

integration_tests_nodejs:: build_nodejs_sdk build_provider_tests install_nodejs_sdk
	cd integration-tests && go test -timeout=25m -v -run $(TEST)

# Provider

generate_schema::
	cd provider/cmd/$(CODEGEN) && go run . schema ../$(PROVIDER)
	cd provider/cmd/${PROVIDER} && VERSION=${VERSION} SCHEMA=${SCHEMA_PATH} go generate main.go

build_provider:: generate_schema
	rm -rf ${WORKING_DIR}/bin/${PROVIDER}
	cd provider/cmd/${PROVIDER} && go build -o ${WORKING_DIR}/bin/${PROVIDER} -ldflags "-X ${PROJECT}/${VERSION_PATH}=${VERSION}" .

build_provider_tests:: generate_schema
	mkdir -p ${WORKING_DIR}/bin/test
	rm -rf ${WORKING_DIR}/bin/test/${PROVIDER}
	cd provider/cmd/${PROVIDER} && go build -o ${WORKING_DIR}/bin/test/${PROVIDER} -ldflags "-X ${PROJECT}/${VERSION_PATH}=${VERSION}" .

install_provider:: build_provider
	cp ${WORKING_DIR}/bin/${PROVIDER} ${GOPATH}/bin

start_delve:
	cd provider/cmd/${PROVIDER} && dlv debug --listen=:2345 --headless=true --api-version=2 --build-flags="-gcflags 'all=-N -l'" --accept-multiclient --api-version=2 --continue --output ${WORKING_DIR}/bin/${PROVIDER}

# Go SDK

gen_go_sdk::
	rm -rf sdk/go
	pulumi package gen-sdk ${SCHEMA_PATH} --language go


# .NET SDK

gen_dotnet_sdk::
	rm -rf sdk/dotnet
	pulumi package gen-sdk ${SCHEMA_PATH} --language dotnet

build_dotnet_sdk:: DOTNET_VERSION := $(shell pulumictl get version --language dotnet)
build_dotnet_sdk:: gen_dotnet_sdk
	cd sdk/dotnet/ && \
		echo "${DOTNET_VERSION}" >version.txt && \
		dotnet build /p:Version=${DOTNET_VERSION}

install_dotnet_sdk:: build_dotnet_sdk
	rm -rf ${WORKING_DIR}/nuget
	mkdir -p ${WORKING_DIR}/nuget
	find . -name '*.nupkg' -print -exec cp -p {} ${WORKING_DIR}/nuget \;


# Node.js SDK

gen_nodejs_sdk::
	rm -rf sdk/nodejs
	pulumi package gen-sdk ${SCHEMA_PATH} --language nodejs

build_nodejs_sdk:: VERSION := $(shell pulumictl get version --language javascript)
build_nodejs_sdk:: gen_nodejs_sdk
	cd sdk/nodejs/ && \
		yarn install && \
		yarn run tsc --version && \
		yarn run tsc && \
		cp ${WORKING_DIR}/README.md package.json yarn.lock ./bin/ && \
		sed -i -e "s/\$${VERSION}/$(VERSION)/g" ./bin/package.json

install_nodejs_sdk:: build_nodejs_sdk link_nodejs_sdk

link_nodejs_sdk::
	yarn link --cwd ${WORKING_DIR}/sdk/nodejs/bin



# Python SDK

gen_python_sdk::
	rm -rf sdk/python
	pulumi package gen-sdk ${SCHEMA_PATH} --language python
	cp ${WORKING_DIR}/README.md sdk/python

build_python_sdk:: PYPI_VERSION := $(shell pulumictl get version --language python)
build_python_sdk:: gen_python_sdk
	cd sdk/python/ && \
		echo "module fake_python_module // Exclude this directory from Go tools\n\ngo 1.17" > go.mod && \
		cp ../../README.md . && \
		rm -rf ./bin/ ../python.bin/ && cp -R . ../python.bin && mv ../python.bin ./bin && \
		python3 -m venv venv && \
		./venv/bin/python -m pip install build && \
		cd ./bin && \
		../venv/bin/python -m build .

## Empty build target for Go
build_go_sdk::
