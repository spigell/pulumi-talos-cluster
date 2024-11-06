package provider

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	// "github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/provider"
	"github.com/pulumiverse/pulumi-talos/sdk/go/talos/machine"
	"github.com/siderolabs/talos/pkg/machinery/config/types/v1alpha1"
	"gopkg.in/yaml.v3"
)

type Cluster struct {
	pulumi.ResourceState
	ClusterArgs

}

func ClusterType() string {
	return ProviderName + ":index:Cluster"
}

type ClusterArgs struct {
	Userdata pulumi.MapOutput `pulumi:"userdata"`
}

func construct(ctx *pulumi.Context, c *Cluster, name string,
	args *ClusterArgs, inputs provider.ConstructInputs, opts ...pulumi.ResourceOption,
) (*provider.ConstructResult, error) {
	// Blit the inputs onto the arguments struct.
	if err := inputs.CopyTo(args); err != nil {
		return nil, errors.Wrap(err, "setting args")
	}

	// Register our component resource.
	if err := ctx.RegisterComponentResource(ClusterType(), name, c, opts...); err != nil {
		return nil, err
	}

	secrets, err := machine.NewSecrets(ctx, "secrets", &machine.SecretsArgs{})
	if err != nil {
		return nil, err
	}

	t := true

	configStruct := v1alpha1.Config{
		MachineConfig: &v1alpha1.MachineConfig{
			MachineInstall: &v1alpha1.InstallConfig{
				InstallDisk: "/dev/sda",
			},
		},
		ClusterConfig: &v1alpha1.ClusterConfig{
			AllowSchedulingOnControlPlanes: &t,
		},
	}

	configUnstruct := map[string]interface{}{
		"machine": map[string]interface{}{
			"install": map[string]interface{}{
				"disk": "/dev/sda",
			},
		},
		"cluster": map[string]interface{}{
			"allowSchedulingOnControlPlanes": true,
		},
	}

	yamlUnstruct, err := yaml.Marshal(configUnstruct)
	yamlStruct, err := yaml.Marshal(configStruct)

	fmt.Println(yamlUnstruct)

	configuration := machine.GetConfigurationOutput(ctx, machine.GetConfigurationOutputArgs{
		ClusterName:     pulumi.String("exampleCluster"),
		MachineType:     pulumi.String("controlplane"),
		KubernetesVersion: pulumi.String("v1.30.0"),
		TalosVersion: pulumi.String("v1.8.0"),
		ConfigPatches: pulumi.StringArray{
			pulumi.String(string(yamlStruct)),
		},
		MachineSecrets:  secrets.ToSecretsOutput().MachineSecrets(),
	}, nil)

	if err != nil {
		return nil, err
	}

	if err := ctx.RegisterResourceOutputs(c, pulumi.Map{
		"userdata":               configuration,
	}); err != nil {
		return nil, err
	}

	return provider.NewConstructResult(c)
}

func getPulumiKey(state pulumi.StringOutput, key string) pulumi.AnyOutput {
	return state.ApplyT(func(keys string) (any, error) {
		var c map[string]any

		decoded, err := base64.StdEncoding.DecodeString(keys)
		if err != nil {
			return "", nil
		}

		err = json.Unmarshal(decoded, &c)
		if err != nil {
			return "", err
		}

		return c[key], nil
	}).(pulumi.AnyOutput)
}
