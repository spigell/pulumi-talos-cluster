package resources

import (
	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/spigell/pulumi-talos-cluster/pkg/provider"
)

var (
	ClusterResourceName = provider.ProviderName + provider.ClusterType()
	ClusterResourceNodeType = "index:nodes"
)

var nodesOutputs = schema.ComplexTypeSpec{
	ObjectTypeSpec: schema.ObjectTypeSpec{
		Type: "object",
		Properties: map[string]schema.PropertySpec{
			"nodes": {},
		},
	},
}


func ClusterProperties() map[string]schema.PropertySpec {
	return map[string]schema.PropertySpec{
		"usedata": {
			TypeSpec: schema.TypeSpec{
				Type: "object",
			},
		},
	}
}

func ClusterInputProperties() map[string]schema.PropertySpec {
	return map[string]schema.PropertySpec{
		"machines": {
			TypeSpec: schema.TypeSpec{
				Type: "object",
				Ref:  "#types/" + ClusterResourceNodeType,
			},
			Description: "Configuration for machines",
		},
	}
}

func ClusterTypes() (map[string]schema.ComplexTypeSpec) {
	types := make(map[string]schema.ComplexTypeSpec)

	types[ClusterResourceNodeType] = nodesOutputs

	return types
}
