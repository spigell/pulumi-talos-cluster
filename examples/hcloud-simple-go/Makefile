
DEBUG_COMMAND ?= "pre"

pulumi-debug-provider:
	PULUMI_DEBUG_PROVIDERS="talos-cluster:$(shell ss -tulnp | grep 'pulumi-resource' | awk '{print $$5}' | cut -f 2 -d ":")" pulumi $(DEBUG_COMMAND) --logtostderr -d -v 9 2> /tmp/log.txt
