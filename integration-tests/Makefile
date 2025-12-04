SHELL := bash
PROG ?= testdata/programs/hcloud-js
TAG ?= $(shell git describe --tags --abbrev=0)
PYTHON ?= python3
PY_VENV ?= .venv
PY_VENV_BIN := $(PY_VENV)/bin

TEST ?= ./...

GH_TEST ?= TestHcloudClusterGo

generate_schema: 
	$(MAKE) -C .. generate_schema

dev-ts-testprog: clean
	cd .. && $(MAKE) install_nodejs_sdk
	cd $(PROG) && \
	yarn link "@spigell/pulumi-talos-cluster" && \
	yarn install
	@echo "Now you can create stack for test project in the testdata directory"

clean: export PROG=$(shell cat .current_test_dir || true)
clean:
	yarn unlink --cwd dist || true
	yarn unlink --cwd ../sdk/nodejs/bin || true
	rm -rf .current_test_dir

framework-unit-tests-ts:
	yarn install && yarn test

ts-lint:
	yarn install && yarn lint

framework-unit-tests-go:
	go test ./pkg/cluster/go ./pkg/cloud/go/...

framework-unit-tests-py:
	test -d $(PY_VENV) || $(PYTHON) -m venv $(PY_VENV)
	$(PY_VENV_BIN)/python -m pip install -r requirements.txt
	$(PY_VENV_BIN)/python -m pytest pkg/cluster/python

install_nodejs_integration_tests_infra:
	yarn install
	rm -rf dist
	yarn build
	cp package.json ./dist -v
	yarn link --cwd dist

integration_tests: integration_tests_go integration_tests_nodejs integration_tests_python

integration_tests_go: clean
	$(MAKE) -C .. build_provider_tests gen_go_sdk
	go test -timeout=25m -v -run $(TEST)

integration_tests_nodejs: clean install_nodejs_integration_tests_infra
	$(MAKE) -C .. build_nodejs_sdk build_provider_tests install_nodejs_sdk
	go test -timeout=25m -v -run $(TEST)

integration_tests_python: clean
	$(MAKE) -C .. build_python_sdk build_provider_tests
	go test -timeout=25m -v -run $(TEST)


github-run-test: export TEST=$(GH_TEST)
github-run-test:
	gh workflow run --ref $$(git rev-parse --abbrev-ref HEAD) -f test=$(TEST) main-run-integration-tests.yaml
	sleep 10
	watch gh run view $$(gh run list --workflow=main-run-integration-tests.yaml -b $$(git rev-parse --abbrev-ref HEAD) -L 1 --json databaseId | jq .[0].databaseId -r) -v
