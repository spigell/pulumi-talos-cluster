package resources

import (
	"fmt"

	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/siderolabs/talos/pkg/machinery/config/machine"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider"
)

var (
	BasicClientConfifgurationPath = provider.ProviderName + ":index:" + provider.ClusterResourceOutputsClientConfiguration
	BasicMachinesByTypePath       = provider.ProviderName + ":index:" + "applyMachines"
	BasicMachineTopologyPath      = provider.ProviderName + ":index:" + provider.ClusterResourceOutputsMachineTopology
)

func BasicTypes() map[string]schema.ComplexTypeSpec {
	types := make(map[string]schema.ComplexTypeSpec)

	types[BasicMachinesByTypePath] = schema.ComplexTypeSpec{
		ObjectTypeSpec: schema.ObjectTypeSpec{
			Type: typeObject,
			Properties: map[string]schema.PropertySpec{
				machine.TypeControlPlane.String(): {
					TypeSpec: schema.TypeSpec{
						Type:  typeArray,
						Items: &schema.TypeSpec{Type: typeObject, Ref: fmt.Sprintf("#types/%s", ApplyTypesMachineInfoPath)},
					},
				},
				machine.TypeInit.String(): {
					TypeSpec: schema.TypeSpec{
						Type:  typeArray,
						Items: &schema.TypeSpec{Type: typeObject, Ref: fmt.Sprintf("#types/%s", ApplyTypesMachineInfoPath)},
					},
				},
				machine.TypeWorker.String(): {
					TypeSpec: schema.TypeSpec{
						Type:  typeArray,
						Items: &schema.TypeSpec{Type: typeObject, Ref: fmt.Sprintf("#types/%s", ApplyTypesMachineInfoPath)},
					},
				},
			},
			Required: []string{
				machine.TypeInit.String(),
			},
		},
	}

	types[BasicMachineTopologyPath] = schema.ComplexTypeSpec{
		ObjectTypeSpec: schema.ObjectTypeSpec{
			Type: typeObject,
			Properties: map[string]schema.PropertySpec{
				machine.TypeControlPlane.String(): {
					TypeSpec: schema.TypeSpec{Type: typeArray, Items: &schema.TypeSpec{Type: typeString}},
				},
				machine.TypeInit.String(): {
					TypeSpec: schema.TypeSpec{Type: typeArray, Items: &schema.TypeSpec{Type: typeString}},
				},
				machine.TypeWorker.String(): {
					TypeSpec: schema.TypeSpec{Type: typeArray, Items: &schema.TypeSpec{Type: typeString}},
				},
			},
			Required: []string{
				machine.TypeInit.String(),
				machine.TypeControlPlane.String(),
				machine.TypeWorker.String(),
			},
		},
	}

	types[BasicClientConfifgurationPath] = schema.ComplexTypeSpec{
		ObjectTypeSpec: schema.ObjectTypeSpec{
			Type: typeObject,
			Properties: map[string]schema.PropertySpec{
				provider.ClusterResourceOutputsClientConfigurationCAKey: {
					TypeSpec: schema.TypeSpec{
						Type: typeString,
					},
					Description: "The Certificate Authority (CA) certificate used to verify connections to the Talos API server.",
				},
				provider.ClusterResourceOutputsClientConfigurationClientKey: {
					TypeSpec: schema.TypeSpec{
						Type: typeString,
					},
					Description: "The private key for the client certificate, used for authenticating the client to the Talos API server.",
				},
				provider.ClusterResourceOutputsClientConfigurationClientCertificateKey: {
					TypeSpec: schema.TypeSpec{
						Type: typeString,
					},
					Description: "The client certificate used to authenticate to the Talos API server.",
				},
			},
		},
	}

	return types
}
