package resources

import (
	"fmt"

	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
)

var (
	ApplyResourceName         = provider.ApplyType()
	ApplyTypesMachineInfoKey  = "machineInfo"
	ApplyTypesMachineInfoPath = provider.ProviderName + ":index:" + ApplyTypesMachineInfoKey
)

var Apply = map[string]schema.ResourceSpec{
	ApplyResourceName: {
		IsComponent: true,
		ObjectTypeSpec: schema.ObjectTypeSpec{
			Description: "Apply the configuration to nodes.",
			Properties:  ApplyProperties(),
		},
		InputProperties: ApplyInputProperties(),
		RequiredInputs:  ApplyRequiredInputProperties(),
	},
}

func ApplyProperties() map[string]schema.PropertySpec {
	return map[string]schema.PropertySpec{}
}

func ApplyInputProperties() map[string]schema.PropertySpec {
	return map[string]schema.PropertySpec{
		"applyMachines": {
			TypeSpec: schema.TypeSpec{
				Type: "object",
				Ref:  fmt.Sprintf("#types/%s", BasicMachinesByTypePath),
			},
			Description: "The machine configurations to apply.",
		},
		provider.ClusterResourceOutputsClientConfiguration: ClusterProperties()[provider.ClusterResourceOutputsClientConfiguration],
	}
}

func ApplyRequiredInputProperties() []string {
	return []string{"applyMachines", provider.ClusterResourceOutputsClientConfiguration}
}

func ApplyTypes() map[string]schema.ComplexTypeSpec {
	ty := make(map[string]schema.ComplexTypeSpec)

	ty[ApplyTypesMachineInfoPath] = schema.ComplexTypeSpec{
		ObjectTypeSpec: schema.ObjectTypeSpec{
			Type: "object",
			Properties: map[string]schema.PropertySpec{
				types.MachineIDKey: {
					TypeSpec: schema.TypeSpec{
						Type: "string",
					},
					Description: "ID or name of the machine.",
				},
				types.NodeIPKey: {
					TypeSpec: schema.TypeSpec{
						Type: "string",
					},
					Description: "The IP address of the node where configuration will be applied.",
				},
				types.ConfigurationKey: {
					TypeSpec: schema.TypeSpec{
						Type: "string",
					},
					Description: "Configuration settings for machines to apply. \n" +
						"This can be retrieved from the cluster resource.",
				},
				types.UserConfigPatchesKey: {
					TypeSpec: schema.TypeSpec{
						Type: "string",
					},
					Description: "User-provided machine configuration to apply. \n" +
						"This can be retrieved from the cluster resource.",
				},
				types.TalosImageKey: {
					TypeSpec: schema.TypeSpec{
						Type: "string",
					},
					Description: "TO DO",
				},
				types.KubernetesVersionKey: {
					TypeSpec: schema.TypeSpec{
						Type: "string",
					},
					Description: "TO DO",
				},
				types.ClusterEnpointKey: {
					TypeSpec: schema.TypeSpec{
						Type: "string",
					},
					Description: "cluster endpoint applied to node",
				},
			},
			Required: []string{
				types.MachineIDKey,
				types.NodeIPKey,
				types.ConfigurationKey,
			},
		},
	}
	return ty
}
