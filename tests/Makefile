SHELL := bash
PROG ?= testdata/programs/hcloud-go
TAG ?= $(shell git describe --tags --abbrev=0)

GH_TEST ?= TestHcloudClusterGo

dev-go-testprog: clean
	echo $(PROG) > .current_test_dir
	@cd $(PROG) && \
	go mod edit -replace=github.com/spigell/pulumi-talos-cluster/sdk=../../../../sdk && \
	go mod edit -replace=github.com/spigell/pulumi-talos-cluster/tests=../../../../tests
	@echo "Now you can create stack for test project in test-project directory"

# sed -i '/\@spigell\/pulumi-talos-cluster/d' package.json && \

dev-ts-testprog: clean
	cd .. && $(MAKE) install_nodejs_sdk
	cd $(PROG) && \
	yarn link "@spigell/pulumi-talos-cluster" && \
	yarn install
	@echo "Now you can create stack for test project in test-project directory"

clean: export PROG=$(shell cat .current_test_dir || true)
clean:
	cd $(PROG) && go mod edit -dropreplace=github.com/spigell/pulumi-talos-cluster/sdk || true
	cd $(PROG) && go mod edit -dropreplace=github.com/spigell/pulumi-talos-cluster/tests || true
	yarn unlink --cwd ../sdk/nodejs/bin || true
	rm -rf .current_test_dir


github-run-test: export TEST=$(GH_TEST)
github-run-test:
	gh workflow run --ref $$(git rev-parse --abbrev-ref HEAD) -f test=$(TEST) main-run-integration-tests.yaml
	sleep 10
	watch gh run view $$(gh run list --workflow=main-run-integration-tests.yaml -b $$(git rev-parse --abbrev-ref HEAD) -L 1 --json databaseId | jq .[0].databaseId -r) -v
