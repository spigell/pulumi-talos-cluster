package applier

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier/talosctl"
)

// GenerateSecrets runs "talosctl gen secrets" into a generated workDir and returns the command resource, file contents, and workDir used.
func (a *Applier) generateSecrets(deps []pulumi.Resource) (pulumi.StringOutput, error) {
	stageName := "gen-secrets"
	home := generateWorkDirNameForTalosctl(a.name, stageName, "common")
	t := talosctl.New()

	cmd, err := t.RunCommand(a.ctx, fmt.Sprintf("%s:%s", a.name, stageName), &talosctl.Args{
		TalosConfig: pulumi.String(""),
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
func GenerateConfig(ctx *pulumi.Context, name, workDir, clusterName string, endpoint pulumi.StringInput, deps []pulumi.Resource, opts ...pulumi.ResourceOption) (pulumi.Resource, error) {
	t := talosctl.New()

	cmd, err := t.RunCommand(ctx, fmt.Sprintf("%s:gen-config", name), &talosctl.Args{
		TalosConfig: pulumi.String(""),
		PrepareDeps: deps,
		Dir:         workDir,
		CommandArgs: pulumi.Sprintf("gen config %s %s --with-secrets %s/secrets.yaml --output-dir %s/configs --force",
			clusterName, endpoint, workDir, workDir),
	}, opts...)
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

// Wrapper methods on Applier for reuse.

func workDirForCluster(stack, name string) string {
	return generateWorkDirNameForTalosctl(stack, "cluster-gen", name)
}
