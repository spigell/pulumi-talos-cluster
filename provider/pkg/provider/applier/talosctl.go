package applier

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/siderolabs/talos/pkg/machinery/config/types/v1alpha1"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
	"gopkg.in/yaml.v3"
)

const (
	TalosctlConfigName = "talosctl.yaml"
)

type Talosctl struct {
	Binary       string
	BasicCommand string
	Home         *TalosctlHome
}

type TalosctlHome struct {
	Dir string
}

func (a *Applier) NewTalosctl() *Talosctl {
	binary := "talosctl"
	home := filepath.Join(os.TempDir(), fmt.Sprintf("talos-home-for-%s", a.name))

	return &Talosctl{
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
func (t *Talosctl) getCurrentMachineConfig(node string) (*v1alpha1.Config, error) {
	command := withBashRetryAndHiddenStdErr(fmt.Sprintf("%s get machineconfig -n %[2]s -e %[2]s -oyaml",
		t.BasicCommand, node,
	))
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error executing command: %w, output: %s", err, string(output))
	}

	var config MachineConfig
	if err := yaml.Unmarshal(output, &config); err != nil {
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

func (a *Applier) talosctlUpgradeCMD(m *types.MachineInfo) pulumi.StringOutput {
	return pulumi.All(a.basicClient().TalosConfig(), m.NodeIP, m.Configuration).ApplyT(func(args []any) (string, error) {
		talosConfig := args[0].(string)
		ip := args[1].(string)
		machineConfig := args[2].(string)
		var config v1alpha1.Config

		err := yaml.Unmarshal([]byte(machineConfig), &config)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal config from string: %w", err)
		}

		talosctl := a.NewTalosctl()
		if err := talosctl.prepare(talosConfig); err != nil {
			return "", fmt.Errorf("failed to prepare temp home for talos cli: %w", err)
		}

		command := withBashRetry(fmt.Sprintf(strings.Join([]string{
			"%[1]s upgrade --debug -n %[2]s -e %[2]s --image %s",
		}, " && "), talosctl.BasicCommand, ip, config.MachineConfig.Install().Image()), "5")

		return command, nil
	}).(pulumi.StringOutput)
}

// talosctlFastReboot returns command talosctl command for rebooting node.
// It doesn't wait for good node status.
func (a *Applier) talosctlFastReboot(m *types.MachineInfo) pulumi.StringOutput {
	return pulumi.All(a.basicClient().TalosConfig()).ApplyT(func(args []any) (string, error) {
		talosConfig := args[0].(string)

		talosctl := a.NewTalosctl()
		if err := talosctl.prepare(talosConfig); err != nil {
			return "", fmt.Errorf("failed to prepare temp home for talos cli: %w", err)
		}

		// Do not wait for succesfull reboot.
		talosctlFlags := "--wait --debug --timeout=20s"

		command := withBashRetry(fmt.Sprintf(strings.Join([]string{
			"%[1]s reboot -n %[2]s -e %[2]s %s",
			// Talosctl exit with code 1 if timeout exceeded.
			// This code is allowed.
			"[ $? == 1 ] && exit 0",
			// Do not retry this command.
		}, " ; "), talosctl.BasicCommand, m.NodeIP, talosctlFlags), "1")

		return command, nil
	}).(pulumi.StringOutput)
}

func (a *Applier) talosctlUpgradeK8SCMD(ma []*types.MachineInfo) pulumi.StringOutput {
	return pulumi.All(a.basicClient().TalosConfig()).ApplyT(func(args []any) (string, error) {
		talosConfig := args[0].(string)

		talosctl := a.NewTalosctl()
		if err := talosctl.prepare(talosConfig); err != nil {
			return "", fmt.Errorf("failed to prepare temp home for talos cli: %w", err)
		}

		talosctlFlags := "--with-docs=false --with-examples=false"

		ips := make([]string, 0)
		for _, m := range ma {
			ips = append(ips, m.NodeIP)
		}

		command := withBashRetry(fmt.Sprintf(strings.Join([]string{
			"%[1]s upgrade-k8s -n %[2]s -e %[2]s --to %s %s",
		}, " && "), talosctl.BasicCommand, strings.Join(ips, " -e "), ma[0].KubernetesVersion, talosctlFlags), "2")

		return command, nil
	}).(pulumi.StringOutput)
}

func mergeYAML(yaml1, yaml2 string) (string, error) {
	var data1, data2 map[string]interface{}

	// Unmarshal first YAML string
	if err := yaml.Unmarshal([]byte(yaml1), &data1); err != nil {
		return "", fmt.Errorf("failed to parse yaml1: %w", err)
	}

	// Unmarshal second YAML string
	if err := yaml.Unmarshal([]byte(yaml2), &data2); err != nil {
		return "", fmt.Errorf("failed to parse yaml2: %w", err)
	}

	// Merge data2 into data1
	mergedData := mergeMaps(data1, data2)

	// Marshal merged data back to YAML
	mergedYAML, err := yaml.Marshal(mergedData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal merged YAML: %w", err)
	}

	return string(mergedYAML), nil
}

// mergeMaps merges map2 into map1 recursively, with map2 overwriting map1's values for duplicate keys.
func mergeMaps(map1, map2 map[string]interface{}) map[string]interface{} {
	for k, v := range map2 {
		if vMap, ok := v.(map[string]interface{}); ok {
			// Handle nested maps by recursive merging
			if map1[k] == nil {
				map1[k] = vMap
			} else if map1Map, ok := map1[k].(map[string]interface{}); ok {
				map1[k] = mergeMaps(map1Map, vMap)
			} else {
				map1[k] = vMap
			}
		} else {
			// For non-map values, map2 overwrites map1
			map1[k] = v
		}
	}
	return map1
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
		"[ $n -ge %[1]s ] && exit 10 || exit 0",
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
		"[ $n -ge 5 ] && exit 10 || exit 0",
	}, " ; "), cmd)
}
