package resources

import (
	"fmt"

	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/siderolabs/talos/pkg/machinery/config/machine"
	"github.com/siderolabs/talos/pkg/machinery/gendata"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
)

var (
	ClusterResourceName                 = provider.ClusterType()
	ClusterTypesMachinesTypesPath       = provider.ProviderName + ":index:" + "machineTypes"
	ClusterTypesMachinesKey             = "clusterMachines"
	ClusterTypesMachinesPath            = provider.ProviderName + ":index:" + ClusterTypesMachinesKey
	ClusterTypesClusterNameKey          = "clusterName"
	ClusterTypesTalosVersionContractKey = "talosVersionContract"
	ClusterTypesMachinesMachineTypeKey  = "machineType"
)

var Cluster = map[string]schema.ResourceSpec{
	ClusterResourceName: {
		IsComponent: true,
		ObjectTypeSpec: schema.ObjectTypeSpec{
			Description: "Initialize a new Talos cluster: \n" +
				"- Creates secrets \n" +
				"- Generates machine configurations for all nodes",
			Properties: ClusterProperties(),
			Required:   ClusterRequiredProperties(),
		},
		InputProperties: ClusterInputProperties(),
		RequiredInputs:  ClusterRequiredInputProperties(),
	},
}

func ClusterTypes() map[string]schema.ComplexTypeSpec {
	ty := make(map[string]schema.ComplexTypeSpec)

	ty[ClusterTypesMachinesTypesPath] = schema.ComplexTypeSpec{
		ObjectTypeSpec: schema.ObjectTypeSpec{
			Type:        "string",
			Description: "Allowed machine types",
			Plain: []string{
				machine.TypeControlPlane.String(),
				machine.TypeWorker.String(),
				machine.TypeInit.String(),
			},
		},
		Enum: []schema.EnumValueSpec{
			{
				Value: machine.TypeControlPlane.String(),
			},
			{
				Value: machine.TypeWorker.String(),
			},
			{
				Value: machine.TypeInit.String(),
			},
		},
	}

	ty[ClusterTypesMachinesPath] = schema.ComplexTypeSpec{
		ObjectTypeSpec: schema.ObjectTypeSpec{
			Type: "object",
			Properties: map[string]schema.PropertySpec{
				types.MachineIDKey: {
					TypeSpec: schema.TypeSpec{
						Type:  "string",
						Plain: true,
					},
					Description: "ID or name of the machine.",
				},
				ClusterTypesMachinesMachineTypeKey: {
					TypeSpec: schema.TypeSpec{
						Type:  "enum",
						Plain: true,
						Ref:   fmt.Sprintf("#types/%s", ClusterTypesMachinesTypesPath),
					},
					Description: "Type of the machine.",
				},
				types.NodeIPKey: {
					TypeSpec: schema.TypeSpec{
						Type: "string",
					},
					Description: "The IP address of the node where configuration will be applied.",
				},
				types.TalosImageKey: {
					TypeSpec: schema.TypeSpec{
						Type: "string",
					},
					Description: fmt.Sprintf("Talos OS installation image. \n"+
						"Used in the `install` configuration and set via CLI. \n"+
						"The default is generated based on the Talos machinery version, current: %s.", provider.GenerateDefaultInstallerImage()),
					Default: provider.GenerateDefaultInstallerImage(),
				},
				"configPatches": {
					TypeSpec: schema.TypeSpec{
						Type: "array",
						Items: &schema.TypeSpec{
							Type: "string",
						},
					},
					Description: "User-provided machine configuration to apply. \n" +
						"Must be a valid array of YAML strings. \n" +
						"For structure, see https://www.talos.dev/latest/reference/configuration/v1alpha1/config/",
				},
			},
			Required: []string{
				ClusterTypesMachinesMachineTypeKey,
				types.MachineIDKey,
				types.NodeIPKey,
			},
		},
	}

	return ty
}

func ClusterProperties() map[string]schema.PropertySpec {
	return map[string]schema.PropertySpec{
		provider.ClusterResourceOutputsClientConfiguration: {
			TypeSpec: schema.TypeSpec{
				Type: "object",
				Ref:  fmt.Sprintf("#types/%s", BasicClientConfifgurationPath),
			},
			Description: "Client configuration for bootstrapping and applying resources.",
		},
		provider.ClusterResourceOutputsGeneratedConfigurations: {
			TypeSpec: schema.TypeSpec{
				Type: "object",
			},
			Description: "Generated machine configuration YAML keyed by machine ID.",
		},
		provider.ClusterResourceOutputsMachines: {
			TypeSpec: schema.TypeSpec{
				Type: "object",
				Ref:  fmt.Sprintf("#types/%s", BasicMachinesByTypePath),
			},
			Description: "Machine information grouped by machine type.",
		},
	}
}

func ClusterRequiredProperties() []string {
	return []string{
		provider.ClusterResourceOutputsMachines,
		provider.ClusterResourceOutputsGeneratedConfigurations,
		provider.ClusterResourceOutputsClientConfiguration,
	}
}

func ClusterInputProperties() map[string]schema.PropertySpec {
	return map[string]schema.PropertySpec{
		types.ClusterEnpointKey: {
			TypeSpec: schema.TypeSpec{
				Type: "string",
			},
			Description: "Cluster endpoint, the Kubernetes API endpoint accessible by all nodes",
		},
		ClusterTypesClusterNameKey: {
			TypeSpec: schema.TypeSpec{
				Type:  "string",
				Plain: true,
			},
			Description: "Name of the cluster",
		},
		types.KubernetesVersionKey: {
			TypeSpec: schema.TypeSpec{
				Type: "string",
			},
			Description: fmt.Sprintf("Kubernetes version to install. \n"+
				"Default is %s.", provider.DefaultK8SVersion),
			Default: provider.DefaultK8SVersion,
		},
		ClusterTypesTalosVersionContractKey: {
			TypeSpec: schema.TypeSpec{
				Type: "string",
			},
			Description: fmt.Sprintf("Version of Talos features used for configuration generation. \n"+
				"Do not confuse this with the talosImage property. \n"+
				"Used in NewSecrets() and GetConfigurationOutput() resources. \n"+
				"This property is immutable to prevent version conflicts across provider updates. \n"+
				"See issue: https://github.com/siderolabs/terraform-provider-talos/issues/168 \n"+
				"The default value is based on gendata.VersionTag, current: %s.", gendata.VersionTag),
			Default: gendata.VersionTag,
		},
		ClusterTypesMachinesKey: {
			TypeSpec: schema.TypeSpec{
				Type: "array",
				Items: &schema.TypeSpec{
					Type: "object",
					Ref:  fmt.Sprintf("#types/%s", ClusterTypesMachinesPath),
				},
			},
			Description: "Configuration settings for machines",
		},
	}
}

func ClusterRequiredInputProperties() []string {
	return []string{
		ClusterTypesClusterNameKey,
		types.ClusterEnpointKey,
		ClusterTypesMachinesKey,
	}
}
