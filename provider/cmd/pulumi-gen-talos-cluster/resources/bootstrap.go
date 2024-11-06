package resources

import (
	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider"
)

var BootstrapResourceName = provider.BootstrapType()

var Bootstrap = map[string]schema.ResourceSpec{
	BootstrapResourceName: {
		IsComponent: true,
		ObjectTypeSpec: schema.ObjectTypeSpec{
			Description: "Initialize a new talos cluster: creates etcd cluster. It must not depend on apply call",
			Properties:  BootstrapProperties(),
		},
		InputProperties: BootstrapInputProperties(),
		RequiredInputs:  BootstrapRequiredInputProperties(),
	},
}

func BootstrapProperties() map[string]schema.PropertySpec {
	return map[string]schema.PropertySpec{}
}

func BootstrapInputProperties() map[string]schema.PropertySpec {
	return map[string]schema.PropertySpec{
		BasicResourceNodeKey: {
			TypeSpec: schema.TypeSpec{
				Type: "string",
			},
			Description: "An IP address or fqdn for node bootstraping.",
		},
		provider.ClusterResourceOutputsClientConfiguration: ClusterProperties()[provider.ClusterResourceOutputsClientConfiguration],
	}
}

func BootstrapRequiredInputProperties() []string {
	return []string{
		provider.ClusterResourceOutputsClientConfiguration,
		BasicResourceNodeKey,
	}
}

func BootstrapTypes() map[string]schema.ComplexTypeSpec {
	types := make(map[string]schema.ComplexTypeSpec)

	return types
}
