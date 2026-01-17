package applier

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier/talosctl"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
)

// GenerateSecrets runs "talosctl gen secrets" into a generated workDir and returns the command resource, file contents, and workDir used.
func (a *Applier) generateSecrets(deps []pulumi.Resource) (pulumi.StringOutput, error) {
	stageName := "gen-secrets"
	t := talosctl.New()
	home := generateWorkDirNameForTalosctl(a.name, stageName, "common")

	cmd, err := t.RunCommand(a.ctx, fmt.Sprintf("%s:%s", a.name, stageName), &talosctl.Args{
		Dir:         home,
		CommandArgs: pulumi.String(talosctlGenerateSecretsArgs()),
	}, []pulumi.ResourceOption{
		a.parent,
		pulumi.DependsOn(deps),
	}...)
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	return cmd.(*local.Command).Stdout, nil
}

func talosctlGenerateSecretsArgs() string {
	return strings.Join([]string{
		"gen secrets --force -o -",
	}, " ; ")
}

// GenerateConfig runs "talosctl gen config" into workDir/configs and returns the command resource.
func (a *Applier) generateConfig(c *types.Cluster, m *types.ClusterMachine, secrets pulumi.StringOutput) (pulumi.Resource, error) {
	stageName := "gen-config"
	home := generateWorkDirNameForTalosctl(a.name, stageName, m.MachineID)
	t := talosctl.New()

	cmd, err := t.RunCommand(a.ctx, fmt.Sprintf("%s:%s:%s", a.name, stageName, m.MachineID), &talosctl.Args{
		Dir: home,
		AdditionalFiles: []talosctl.ExtraFile{
			{
				Name:    "secrets.yaml",
				Content: secrets,
			},
			{
				Name: "patches.yaml",
				Content: m.ConfigPatches.ToStringArrayOutput().ApplyT(func(p []string) string {
					return strings.Join(p, "\n---\n")
				}).(pulumi.StringOutput),
			},
		},
		CommandArgs: pulumi.Sprintf("%s %s %s --with-secrets secrets.yaml --config-patch @patches.yaml --output-dir -",
			talosctlGenerateConfigArgs(), c.ClusterName, c.ClusterEndpoint),
	}, []pulumi.ResourceOption{
		a.parent,
	}...)
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

func talosctlGenerateConfigArgs() string {
	return strings.Join([]string{
		"gen config",
		"--force",
		"--with-docs=false --with-examples=false",
	}, " ")
}

// Wrapper methods on Applier for reuse.

func workDirForCluster(stack, name string) string {
	return generateWorkDirNameForTalosctl(stack, name, "common")
}
