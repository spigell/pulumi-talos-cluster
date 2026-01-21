package applier

import (
	"context"
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	tmachine "github.com/siderolabs/talos/pkg/machinery/config/machine"
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

	genType := m.MachineType
	outType := genType
	if genType == tmachine.TypeInit.String() {
		// talosctl doesn't support init output type; use controlplane config and patch type to init.
		outType = tmachine.TypeControlPlane.String()
	}

	patches := m.ConfigPatches.ToStringArrayOutput().ApplyTWithContext(a.ctx.Context(), func(_ context.Context, p []string) (string, error) {
		if genType == tmachine.TypeInit.String() {
			p = append([]string{"machine:\n  type: init"}, p...)
		}
		return mergePatchesYAML(p)
	}).(pulumi.StringOutput)

	patchFlag := patches.ApplyT(func(p string) string {
		if strings.TrimSpace(p) != "" {
			return " --config-patch @patches.yaml"
		}
		return ""
	}).(pulumi.StringOutput)

	cmd, err := t.RunCommand(a.ctx, fmt.Sprintf("%s:%s:%s", a.name, stageName, m.MachineID), &talosctl.Args{
		Dir: home,
		AdditionalFiles: []talosctl.ExtraFile{
			{
				Name:    "secrets.yaml",
				Content: secrets,
			},
			{
				Name:    "patches.yaml",
				Content: patches,
			},
		},
		CommandArgs: pulumi.Sprintf("%s %s %s --install-image %s --kubernetes-version %s --talos-version %s --output-types %s --with-secrets secrets.yaml%s --output -",
			talosctlGenerateConfigArgs(),
			c.ClusterName,
			c.ClusterEndpoint,
			m.TalosImage.ToStringPtrOutput().Elem(),
			c.KubernetesVersion.ToStringOutput(),
			c.TalosVersionContract.ToStringOutput(),
			outType,
			patchFlag,
		),
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

func mergePatchesYAML(patches []string) (string, error) {
	merged := ""
	for i, patch := range patches {
		if strings.TrimSpace(patch) == "" {
			continue
		}

		if merged == "" {
			merged = patch
			continue
		}

		out, err := MergeYAML(merged, patch).Build()
		if err != nil {
			return "", fmt.Errorf("merge patch %d: %w", i+1, err)
		}
		merged = out
	}

	return merged, nil
}
