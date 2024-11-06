package resources

import (
	"fmt"

	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/siderolabs/talos/pkg/machinery/config/machine"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider"
)

var (
	ApplyResourceName            = provider.ApplyType()
	ApplyResourceMachinesKey     = "applyMachines"
	ApplyTypesMachinesPath       = provider.ProviderName + ":index:" + ApplyResourceMachinesKey
	ApplyTypesMachinesByTypePath = provider.ProviderName + ":index:" + "applyMachinesByType"
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
		ApplyResourceMachinesKey: {
			TypeSpec: schema.TypeSpec{
				Type: "object",
				Ref:  fmt.Sprintf("#types/%s", ApplyTypesMachinesByTypePath),
			},
			Description: "The machine configurations to apply.",
		},
		provider.ClusterResourceOutputsClientConfiguration: ClusterProperties()[provider.ClusterResourceOutputsClientConfiguration],
	}
}

func ApplyRequiredInputProperties() []string {
	return []string{ApplyResourceMachinesKey, provider.ClusterResourceOutputsClientConfiguration}
}

func ApplyTypes() map[string]schema.ComplexTypeSpec {
	types := make(map[string]schema.ComplexTypeSpec)

	types[ApplyTypesMachinesByTypePath] = schema.ComplexTypeSpec{
		ObjectTypeSpec: schema.ObjectTypeSpec{
			Type: "object",
			Properties: map[string]schema.PropertySpec{
				machine.TypeControlPlane.String(): {
					TypeSpec: schema.TypeSpec{
						Type:  "array",
						Items: &schema.TypeSpec{Type: "object", Ref: fmt.Sprintf("#types/%s", ApplyTypesMachinesPath)},
					},
				},
				machine.TypeInit.String(): {
					TypeSpec: schema.TypeSpec{
						Type: "object",
						Ref:  fmt.Sprintf("#types/%s", ApplyTypesMachinesPath),
					},
				},
				machine.TypeWorker.String(): {
					TypeSpec: schema.TypeSpec{
						Type:  "array",
						Items: &schema.TypeSpec{Type: "object", Ref: fmt.Sprintf("#types/%s", ApplyTypesMachinesPath)},
					},
				},
			},
			Required: []string{
				machine.TypeInit.String(),
			},
		},
	}

	types[ApplyTypesMachinesPath] = schema.ComplexTypeSpec{
		ObjectTypeSpec: schema.ObjectTypeSpec{
			Type: "object",
			Properties: map[string]schema.PropertySpec{
				BasicMachinesMachineIDKey: {
					TypeSpec: schema.TypeSpec{
						Type: "string",
					},
					Description: "ID or name of the machine.",
				},
				BasicResourceNodeKey: {
					TypeSpec: schema.TypeSpec{
						Type: "string",
					},
					Description: "The IP address of the node where configuration will be applied.",
				},
				"configuration": {
					TypeSpec: schema.TypeSpec{
						Type: "string",
					},
					Description: "Configuration settings for machines to apply. \n" +
						"This can be retrieved from the cluster resource.",
				},
				provider.ClusterResourceOutputsUserConfigPatches: {
					TypeSpec: schema.TypeSpec{
						Type: "string",
					},
					Description: "User-provided machine configuration to apply. \n" +
						"This can be retrieved from the cluster resource.",
				},
			},
			Required: []string{
				BasicMachinesMachineIDKey,
				BasicResourceNodeKey,
				"configuration",
			},
		},
	}
	return types
}
