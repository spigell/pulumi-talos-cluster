SHELL := bash
PROG ?= testdata/programs/hcloud-simple-go
TAG ?= $(shell git describe --tags --abbrev=0)

GH_TEST ?= k3s-private-non-ha-simple

include ../Makefile

dev-go-testprog: clean
	echo $(PROG) > .current_test_dir
	@cd $(PROG) && \
	go mod edit -replace=github.com/spigell/pulumi-talos-cluster/sdk=../../../../sdk
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
	yarn unlink --cwd ../sdk/nodejs/bin || true
	rm -rf .current_test_dir

github-run:
	gh workflow run --ref $$(git rev-parse --abbrev-ref HEAD) -f example=$(GH_EXAMPLE) main-test-examples.yaml
	sleep 10
	watch gh run view $$(gh run list --workflow=main-test-examples.yaml -b $$(git rev-parse --abbrev-ref HEAD) -L 1 --json databaseId | jq .[0].databaseId -r) -v
