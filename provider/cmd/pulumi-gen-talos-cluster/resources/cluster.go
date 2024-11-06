package resources

import (
	"fmt"

	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/siderolabs/talos/pkg/machinery/config/machine"
	"github.com/siderolabs/talos/pkg/machinery/gendata"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider"
)

var (
	ClusterResourceName                 = provider.ClusterType()
	ClusterTypesMachinesTypesPath       = provider.ProviderName + ":index:" + "machineTypes"
	ClusterTypesMachinesKey             = "clusterMachines"
	ClusterTypesMachinesPath            = provider.ProviderName + ":index:" + ClusterTypesMachinesKey
	ClusterTypesClusterNameKey          = "clusterName"
	ClusterTypesClusterEndpointKey      = "clusterEndpoint"
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
		},
		InputProperties: ClusterInputProperties(),
		RequiredInputs:  ClusterRequiredInputProperties(),
	},
}

func ClusterTypes() map[string]schema.ComplexTypeSpec {
	types := make(map[string]schema.ComplexTypeSpec)

	types[ClusterTypesMachinesTypesPath] = schema.ComplexTypeSpec{
		ObjectTypeSpec: schema.ObjectTypeSpec{
			Type:        "string",
			Description: "Allowed machine types",
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

	types[ClusterTypesMachinesPath] = schema.ComplexTypeSpec{
		ObjectTypeSpec: schema.ObjectTypeSpec{
			Type: "object",
			Properties: map[string]schema.PropertySpec{
				BasicMachinesMachineIDKey: {
					TypeSpec: schema.TypeSpec{
						Type:  "string",
						Plain: true,
					},
					Description: "ID or name of the machine.",
				},
				"talosImage": {
					TypeSpec: schema.TypeSpec{
						Type: "string",
					},
					Description: fmt.Sprintf("Talos OS installation image. \n"+
						"Used in the `install` configuration and set via CLI. \n"+
						"The default is generated based on the Talos machinery version, current: %s.", provider.GenerateDefaultInstallerImage()),
					Default: provider.GenerateDefaultInstallerImage(),
				},
				ClusterTypesMachinesMachineTypeKey: {
					TypeSpec: schema.TypeSpec{
						Type:  "enum",
						Ref:   fmt.Sprintf("#types/%s", ClusterTypesMachinesTypesPath),
						Plain: true,
					},
					Description: "Type of the machine.",
				},
				"kubernetesVersion": {
					TypeSpec: schema.TypeSpec{
						Type: "string",
					},
					Description: fmt.Sprintf("Kubernetes version to install. \n"+
						"Default is %s.", provider.DefaultK8SVersion),
					Default: provider.DefaultK8SVersion,
				},
				"configPatches": {
					TypeSpec: schema.TypeSpec{
						Type: "string",
					},
					Description: "User-provided machine configuration to apply. \n" +
						"Must be a valid YAML string. \n" +
						"For structure, see https://www.talos.dev/latest/reference/configuration/v1alpha1/config/",
				},
			},
			Required: []string{
				ClusterTypesMachinesMachineTypeKey,
				BasicMachinesMachineIDKey,
			},
		},
	}

	return types
}

func ClusterProperties() map[string]schema.PropertySpec {
	return map[string]schema.PropertySpec{
		provider.ClusterResourceOutputsInitMachineConfiguration: {
			TypeSpec: schema.TypeSpec{
				Type: "string",
			},
			Description: "The generated machine configurations for the init node. \n" +
				"This is an unstructured string, but it is valid YAML.",
		},
		provider.ClusterResourceOutputsWorkerMachineConfigurations: {
			TypeSpec: schema.TypeSpec{
				Type: "object",
			},
			Description: "The map of generated machine configurations for workers. \n" +
				"This is an unstructured string, but it is valid YAML.",
		},
		provider.ClusterResourceOutputsControlplaneMachineConfigurations: {
			TypeSpec: schema.TypeSpec{
				Type: "object",
			},
			Description: "The map of generated machine configurations for controlplanes. \n" +
				"This is an unstructured string, but it is valid YAML.",
		},
		provider.ClusterResourceOutputsUserConfigPatches: {
			TypeSpec: schema.TypeSpec{
				Type: "object",
			},
			Description: "Map of user-provided machine configuration patches. \n" +
				"Can be used in the apply resource.",
		},
		provider.ClusterResourceOutputsClientConfiguration: {
			TypeSpec: schema.TypeSpec{
				Type: "object",
				Ref:  fmt.Sprintf("#types/%s", BasicTypesClientConfifgurationPath),
			},
			Description: "Client configuration for bootstrapping and applying resources.",
		},
	}
}

func ClusterInputProperties() map[string]schema.PropertySpec {
	return map[string]schema.PropertySpec{
		ClusterTypesClusterEndpointKey: {
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
		ClusterTypesTalosVersionContractKey,
		ClusterTypesClusterNameKey,
		ClusterTypesClusterEndpointKey,
		ClusterTypesMachinesKey,
	}
}
