SHELL := bash

# An explicit release version must not be replaced by a language-specific
# pulumictl value (which may include +dirty while the tree is being rebuilt).
ifeq ($(origin VERSION), undefined)
VERSION         := $(shell command -v pulumictl >/dev/null 2>&1 && pulumictl get version)
DOTNET_VERSION  = $(shell pulumictl get version --language dotnet)
NODEJS_VERSION  = $(shell pulumictl get version --language javascript)
PYPI_VERSION    = $(shell pulumictl get version --language python)
else
ifeq ($(strip $(VERSION)),)
$(error explicit VERSION must not be empty)
endif
SDK_VERSION     := $(patsubst v%,%,$(VERSION))
DOTNET_VERSION  := $(SDK_VERSION)
NODEJS_VERSION  := v$(SDK_VERSION)
PYPI_VERSION    = $(shell python3 scripts/python-version.py '$(SDK_VERSION)')
endif

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

# Tests
unit_tests:: generate_schema
	cd provider && set -o pipefail ; go test $$(go list ./... | grep -v tests | grep -v crds/generated) | grep -v 'no test files'

# Provider

generate_schema::
	cd provider/cmd/$(CODEGEN) && VERSION=${VERSION} go run . schema ../$(PROVIDER)
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
	cd provider/cmd/${PROVIDER} && dlv debug --check-go-version=false --listen=:2345 --headless=true --api-version=2 --build-flags="-gcflags 'all=-N -l'" --accept-multiclient --continue --output ${WORKING_DIR}/bin/${PROVIDER}

# Go SDK

gen_go_sdk::
	rm -rf sdk/go
	pulumi package gen-sdk ${SCHEMA_PATH} --language go


# .NET SDK

gen_dotnet_sdk::
	rm -rf sdk/dotnet
	pulumi package gen-sdk ${SCHEMA_PATH} --language dotnet
	# Pulumi's generator references logo.png even though this package has no icon.
	sed -i '/<PackageIcon>logo.png<\/PackageIcon>/d; /<None Include="logo.png">/,/<\/None>/d' sdk/dotnet/Pulumi.TalosCluster.csproj

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

build_nodejs_sdk:: gen_nodejs_sdk
	cd sdk/nodejs/ && \
		yarn install && \
		yarn run tsc --version && \
		yarn run tsc && \
		cp ${WORKING_DIR}/README.md package.json yarn.lock ./bin/ && \
		sed -i -e "s/\$${VERSION}/$(NODEJS_VERSION)/g" ./bin/package.json && \
		node -e 'const actual = require("./bin/package.json").version; const expected = process.argv[1]; if (actual.includes("+dirty") || actual !== expected) { console.error("Node SDK version " + actual + " does not match " + expected + " or is dirty"); process.exit(1); }' '$(NODEJS_VERSION)'

install_nodejs_sdk:: build_nodejs_sdk link_nodejs_sdk

link_nodejs_sdk::
	yarn link --cwd ${WORKING_DIR}/sdk/nodejs/bin



# Python SDK

gen_python_sdk::
	rm -rf sdk/python
	pulumi package gen-sdk ${SCHEMA_PATH} --language python
	cp ${WORKING_DIR}/README.md sdk/python

build_python_sdk:: gen_python_sdk
	cd sdk/python/ && \
		test -n "$(PYPI_VERSION)" && \
		sed -i -E 's/^  version = "[^"]*"/  version = "$(PYPI_VERSION)"/' pyproject.toml && \
		grep -Fx '  version = "$(PYPI_VERSION)"' pyproject.toml && \
		echo "module fake_python_module // Exclude this directory from Go tools\n\ngo 1.17" > go.mod && \
		cp ../../README.md . && \
		rm -rf ./bin/ ../python.bin/ && cp -R . ../python.bin && mv ../python.bin ./bin && \
		python3 -m venv venv && \
		./venv/bin/python -m pip install build && \
		cd ./bin && \
		../venv/bin/python -m build .

## Empty build target for Go
build_go_sdk::
