package applier

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/siderolabs/talos/pkg/machinery/config/types/v1alpha1"
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

func (a *Applier) talosctlUpgradeCMD(m *ApplyMachine) pulumi.StringOutput {
	return pulumi.All(a.basicClient().TalosConfig(), m.Node, m.Configuration).ApplyT(func(args []any) (string, error) {
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
		}, " && "), talosctl.BasicCommand, ip, config.MachineConfig.Install().Image()))

		return command, nil
	}).(pulumi.StringOutput)
}

func (a *Applier) talosctlApplyCMD(m *ApplyMachine) pulumi.StringOutput {
	return pulumi.All(a.basicClient().TalosConfig(), m.UserConfigPatches, m.Node, m.Configuration).ApplyT(func(args []any) (string, error) {
		talosConfig := args[0].(string)
		userPatches := args[1].(string)
		ip := args[2].(string)
		machineConfig := args[3].(string)

		talosctl := a.NewTalosctl()

		if err := talosctl.prepare(talosConfig); err != nil {
			return "", fmt.Errorf("failed to prepare temp home for talos cli: %w", err)
		}

		config, err := mergeYAML(machineConfig, userPatches)
		if err != nil {
			return "", fmt.Errorf("failed merge yaml strings: %w", err)
		}

		configPath := filepath.Join(talosctl.Home.Dir, fmt.Sprintf("machineconfig-%s.yaml", m.MachineID))
		if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
			return "", fmt.Errorf("failed to write machine config: %w", err)
		}

		command := withBashRetry(fmt.Sprintf(strings.Join([]string{
			"%[1]s apply-config -n %[2]s -e %[2]s -f %s",
		}, " && "), talosctl.BasicCommand, ip, configPath))

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

func withBashRetry(cmd string) string {
	return fmt.Sprintf(strings.Join([]string{
		"n=0",
		"until [ $n -ge 5 ]",
		"do %s && break",
		"sleep 10",
		"n=$((n+1))",
		"done",
	}, " ; "), cmd)
}
