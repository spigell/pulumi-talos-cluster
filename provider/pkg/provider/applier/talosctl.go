package applier

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/siderolabs/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
	"gopkg.in/yaml.v3"
)

const (
	TalosctlConfigName = "talosctl.yaml"
)

type Talosctl struct {
	ctx          *pulumi.Context
	Binary       string
	BasicCommand string
	Home         *TalosctlHome
}


type TalosctlHome struct {
	Dir string
}

func (a *Applier) NewTalosctl(ctx *pulumi.Context, name string) *Talosctl {
	binary := "talosctl"
	home := filepath.Join(os.TempDir(), fmt.Sprintf("talos-home-%s-step-%s", a.name, name))

	return &Talosctl{
		ctx:          ctx,
		Binary:       binary,
		BasicCommand: fmt.Sprintf("%s --talosconfig %s/%s", binary, home, TalosctlConfigName),
		Home: &TalosctlHome{
			Dir: home,
		},
	}
}


// MachineConfig represents the parsed YAML structure.
type MachineConfig struct {
	Spec string
}

// getCurrentMachineConfig retrieves current machineconfig fron running cluster.
// BEWARE: this function should be used with caution. Do not call unprovised nodes!
func (t *Talosctl) getCurrentMachineConfig(node string, deps []pulumi.Resource) (*v1alpha1.Config, error) {
	command := withBashRetryAndHiddenStdErr(fmt.Sprintf("%s get machineconfig v1alpha1 -n %[2]s -e %[2]s -oyaml",
		t.BasicCommand, node,
	))
	cmd, err := local.Run(t.ctx, &local.RunArgs{
		Command: command,
	}, pulumi.DependsOn(deps))
	if err != nil {
		return nil, fmt.Errorf("error executing command: %w, cmd: %+v", err, cmd)
	}

	output := cmd.Stdout

	var config MachineConfig
	if err := yaml.Unmarshal([]byte(output), &config); err != nil {
		return nil, fmt.Errorf("error parsing YAML output: %w", err)
	}

	var spec v1alpha1.Config
	if err := yaml.Unmarshal([]byte(config.Spec), &spec); err != nil {
		return nil, fmt.Errorf("error parsing YAML spec string: %w", err)
	}

	return &spec, nil
}


func (t *Talosctl) prepare(config string) error {
	err := os.MkdirAll(t.Home.Dir, 0o700)
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}

	talosConfigPath := filepath.Join(t.Home.Dir, TalosctlConfigName)
	if err := os.WriteFile(talosConfigPath, []byte(config), 0o600); err != nil {
		return fmt.Errorf("failed to write talosconfig: %w", err)
	}

	return nil
}

// talosctlFastReboot returns command talosctl command for rebooting node.
// It doesn't wait for good node status.
func (a *Applier) talosctlFastReboot(m *types.MachineInfo) pulumi.StringOutput {
	return pulumi.All(a.basicClient().TalosConfig()).ApplyT(func(args []any) (string, error) {
		talosConfig := args[0].(string)

		name := "reboot"

		talosctl := a.NewTalosctl(a.ctx, name+"-"+m.MachineID)
		if err := talosctl.prepare(talosConfig); err != nil {
			return "", fmt.Errorf("failed to prepare temp home for talos cli: %w", err)
		}

		// Do not wait for succesfull reboot.
		talosctlFlags := "--wait --debug --timeout=20s"

		command := talosctl.withCleanCommand(withBashRetry(fmt.Sprintf(strings.Join([]string{
			"%[1]s %[2]s -n %[3]s -e %[3]s %s",
			// Talosctl exit with code 1 if timeout exceeded.
			// This code is allowed.
			"[ $? == 1 ] && true",
			// Do not retry this command.
		}, " ; "), talosctl.BasicCommand, name, m.NodeIP, talosctlFlags), "1"))

		return command, nil
	}).(pulumi.StringOutput)
}

func (a *Applier) talosctlUpgradeK8SCMD(ma []*types.MachineInfo) pulumi.StringOutput {
	return pulumi.All(a.basicClient().TalosConfig()).ApplyT(func(args []any) (string, error) {
		talosConfig := args[0].(string)

		name := "upgrade-k8s"

		talosctl := a.NewTalosctl(a.ctx, name+"-"+ma[0].MachineID)
		if err := talosctl.prepare(talosConfig); err != nil {
			return "", fmt.Errorf("failed to prepare temp home for talos cli: %w", err)
		}

		talosctlFlags := "--with-docs=false --with-examples=false"

		ips := make([]string, 0)
		for _, m := range ma {
			ips = append(ips, m.NodeIP)
		}

		command := talosctl.withCleanCommand(withBashRetry(fmt.Sprintf(strings.Join([]string{
			"%[1]s %[2]s -n %[3]s -e %[3]s --to %s %s",
		}, " && "), talosctl.BasicCommand, name, strings.Join(ips, " -e "), ma[0].KubernetesVersion, talosctlFlags), "2"))

		return command, nil
	}).(pulumi.StringOutput)
}

func withBashRetry(cmd string, retryCount string) string {
	return fmt.Sprintf(strings.Join([]string{
		"n=0",
		"until [ $n -ge %[1]s ]",
		"do %s && break",
		"sleep 10",
		"n=$((n+1))",
		"done",
		// Exiting with 0 if command succeeded.
		// Otherwise exit with 10 exit code.
		"[ $n -ge %[1]s ] && exit 10 || true",
	}, " ; "), retryCount, cmd)
}

func withBashRetryAndHiddenStdErr(cmd string) string {
	return fmt.Sprintf(strings.Join([]string{
		"n=0",
		"until [ $n -ge 5 ]",
		"do %s 2>/dev/null && break",
		"sleep 10",
		"n=$((n+1))",
		"done",
		// Exiting with 0 if command succeeded.
		// Otherwise exit with 10 exit code.
		"[ $n -ge 5 ] && exit 10 || true",
	}, " ; "), cmd)
}

func (t *Talosctl) withCleanCommand(cmd string) string {
	return fmt.Sprintf(strings.Join([]string{
		"%s",
		"rm -rfv %s",
	}, " ; "), cmd, t.Home.Dir)
}

