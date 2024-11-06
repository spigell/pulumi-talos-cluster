package resources

import (
	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider"
)

var (
	BasicTypesClientConfifgurationPath = provider.ProviderName + ":index:" + provider.ClusterResourceOutputsClientConfiguration
	BasicResourceNodeKey               = "node"
	BasicMachinesMachineIDKey          = "machineId"
)

func BasicTypes() map[string]schema.ComplexTypeSpec {
	ClientConfiguration := schema.ComplexTypeSpec{
		ObjectTypeSpec: schema.ObjectTypeSpec{
			Type: "object",
			Properties: map[string]schema.PropertySpec{
				provider.ClusterResourceOutputsClientConfigurationCAKey: {
					TypeSpec: schema.TypeSpec{
						Type: "string",
					},
					Description: "The Certificate Authority (CA) certificate used to verify connections to the Talos API server.",
				},
				provider.ClusterResourceOutputsClientConfigurationClientKey: {
					TypeSpec: schema.TypeSpec{
						Type: "string",
					},
					Description: "The private key for the client certificate, used for authenticating the client to the Talos API server.",
				},
				provider.ClusterResourceOutputsClientConfigurationClientCertificateKey: {
					TypeSpec: schema.TypeSpec{
						Type: "string",
					},
					Description: "The client certificate used to authenticate to the Talos API server.",
				},
			},
		},
	}
	types := make(map[string]schema.ComplexTypeSpec)

	types[BasicTypesClientConfifgurationPath] = ClientConfiguration

	return types
}
