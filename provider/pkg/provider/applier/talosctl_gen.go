package applier

import (
	"context"
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	clientconfig "github.com/siderolabs/talos/pkg/machinery/client/config"
	tmachine "github.com/siderolabs/talos/pkg/machinery/config/machine"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier/talosctl"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
)

var talosctlGenerateBaseArgs = strings.Join([]string{
	"gen config",
	"--force",
	"--with-docs=false --with-examples=false",
}, " ")

// GenerateSecrets runs "talosctl gen secrets" into a generated workDir and returns the command resource, file contents, and workDir used.
func (a *Applier) generateSecrets() (pulumi.StringOutput, error) {
	stageName := "gen-secrets"
	t := talosctl.New()
	home := generateWorkDirNameForTalosctl(a.name, stageName, "common")

	cmd, err := t.RunCommand(a.ctx, fmt.Sprintf("%s:%s", a.name, stageName), &talosctl.Args{
		Dir:         home,
		CommandArgs: pulumi.String("gen secrets --force -o -"),
	}, []pulumi.ResourceOption{
		a.parent,
	}...)
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	return cmd.(*local.Command).Stdout, nil
}

// GenerateConfig runs "talosctl gen config" into workDir/configs and returns the command resource.
func (a *Applier) generateMachineConfig(c *types.Cluster, m *types.ClusterMachine, secrets pulumi.StringOutput) (pulumi.Resource, error) {
	stageName := "gen-machine-config"
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

	return t.RunCommand(a.ctx, fmt.Sprintf("%s:%s:%s", a.name, stageName, m.MachineID), &talosctl.Args{
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
			talosctlGenerateBaseArgs,
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
}

func (a *Applier) generateTalosconfig(c *types.Cluster, secrets pulumi.StringOutput) (pulumi.Resource, error) {
	stageName := "gen-talosconfig"
	home := generateWorkDirNameForTalosctl(a.name, stageName, "")
	t := talosctl.New()

	return t.RunCommand(a.ctx, fmt.Sprintf("%s:%s", c.ClusterName, stageName), &talosctl.Args{
		Dir: home,
		AdditionalFiles: []talosctl.ExtraFile{
			{
				Name:    "secrets.yaml",
				Content: secrets,
			},
		},
		CommandArgs: pulumi.Sprintf("%s %s %s --output-types talosconfig --with-secrets secrets.yaml --output -",
			talosctlGenerateBaseArgs,
			c.ClusterName,
			c.ClusterEndpoint,
		),
	})
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

func ExtractTalosconfigCreds(raw, clusterName string) (ca, key, cert string, err error) {
	cfg, err := clientconfig.FromString(raw)
	if err != nil {
		return "", "", "", fmt.Errorf("parse talosconfig: %w", err)
	}

	if len(cfg.Contexts) == 0 {
		return "", "", "", fmt.Errorf("talosconfig contexts missing")
	}

	ctx := cfg.Contexts[cfg.Context]
	if ctx == nil {
		if ctx = cfg.Contexts[clusterName]; ctx == nil {
			for _, candidate := range cfg.Contexts {
				ctx = candidate
				break
			}
		}
	}

	if ctx == nil {
		return "", "", "", fmt.Errorf("talosconfig contexts missing")
	}

	caVal := strings.TrimSpace(ctx.CA)
	keyVal := strings.TrimSpace(ctx.Key)
	certVal := strings.TrimSpace(ctx.Crt)

	if caVal == "" || keyVal == "" || certVal == "" {
		return "", "", "", fmt.Errorf("talosconfig missing client credentials")
	}

	return caVal, keyVal, certVal, nil
}
