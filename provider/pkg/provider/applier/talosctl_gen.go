package applier

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/siderolabs/crypto/x509"
	clientconfig "github.com/siderolabs/talos/pkg/machinery/client/config"
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
	stageName := "cli-gen-secrets"
	t := talosctl.New()
	home := generateWorkDirNameForTalosctl(a.name, stageName, "common")

	cmd, err := t.RunCommand(a.ctx, fmt.Sprintf("%s:%s", a.name, stageName), &talosctl.Args{
		Dir:         home,
		CommandArgs: pulumi.String("gen secrets --force -o -"),
	}, []pulumi.ResourceOption{
		a.parent,
		pulumi.IgnoreChanges([]string{ignoreCreateChange}),
	}...)
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	return cmd.Stdout, nil
}

// GenerateConfig runs "talosctl gen config" into workDir/configs and returns the command resource.
func (a *Applier) generateMachineConfig(c *types.Cluster, m *types.ClusterMachine, secrets pulumi.StringOutput) (pulumi.Resource, error) {
	stageName := "cli-gen-machine-config"
	home := generateWorkDirNameForTalosctl(a.name, stageName, m.MachineID)
	t := talosctl.New()

	genType := m.MachineType
	outType := genType
	if genType == "init" {
		// talosctl doesn't support init output type; use controlplane config and patch type to init.
		outType = "controlplane"
	}

	patches := m.ConfigPatches.ToStringArrayOutput().ApplyTWithContext(a.ctx.Context(), func(_ context.Context, p []string) (string, error) {
		if genType == "init" {
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
		UpdateOnChange: true,
	}, []pulumi.ResourceOption{
		a.parent,
		pulumi.IgnoreChanges([]string{ignoreCreateChange}),
	}...)
}

func (a *Applier) generateTalosconfig(c *types.Cluster, secrets pulumi.StringOutput) (pulumi.Resource, error) {
	stageName := "cli-gen-talosconfig"
	home := generateWorkDirNameForTalosctl(a.name, stageName, "")
	t := talosctl.New()

	return t.RunCommand(a.ctx, fmt.Sprintf("%s:%s", a.name, stageName), &talosctl.Args{
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
	}, []pulumi.ResourceOption{
		a.parent,
		pulumi.IgnoreChanges([]string{ignoreCreateChange}),
	}...)
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

func (a *Applier) buildTalosConfig(endpoints pulumi.StringArrayInput, nodes pulumi.StringArrayInput) pulumi.StringOutput {
	return pulumi.All(a.clientConfiguration, endpoints, nodes).ApplyT(func(values []any) (string, error) {
		rawCfg := values[0].(map[string]string)
		resolvedEndpoints := values[1].([]string)
		resolvedNodes := values[2].([]string)
		ca := decodeMaybeBase64(rawCfg["caCertificate"])
		key := decodeMaybeBase64(rawCfg["clientKey"])
		cert := decodeMaybeBase64(rawCfg["clientCertificate"])

		cfg := clientconfig.NewConfig(
			a.name,
			resolvedEndpoints,
			[]byte(ca),
			&x509.PEMEncodedCertificateAndKey{
				Crt: []byte(cert),
				Key: []byte(key),
			},
		)

		if ctx, ok := cfg.Contexts[a.name]; ok {
			ctx.Nodes = resolvedNodes
		}

		out, err := cfg.Bytes()
		if err != nil {
			return "", fmt.Errorf("marshal talosconfig: %w", err)
		}

		return string(out), nil
	}).(pulumi.StringOutput)
}

// decodeMaybeBase64 returns the original string unless it is valid base64, in which case it returns the decoded bytes as a string.
func decodeMaybeBase64(s string) string {
	if s == "" {
		return s
	}

	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return s
	}

	return string(decoded)
}
