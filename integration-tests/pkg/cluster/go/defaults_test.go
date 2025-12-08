package cluster

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSchemaDefaults(t *testing.T) {
	val, err := schemaDefault("properties", "kubernetesVersion", "default")
	require.NoError(t, err)
	require.NotNil(t, val)

	val, err = schemaDefault("properties", "machineDefaults", "properties", "hcloud", "properties", "serverType", "default")
	require.NoError(t, err)
	require.NotNil(t, val)

	val, err = schemaDefault("properties", "machineDefaults", "properties", "hcloud", "properties", "datacenter", "default")
	require.NoError(t, err)
	require.NotNil(t, val)

	val, err = schemaDefault("properties", "machines", "items", "properties", "talosImage", "default")
	require.NoError(t, err)
	require.NotNil(t, val)
}

func TestDefaultsAreApplied(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "fixtures", "load-defaults.yaml"))
	require.NoError(t, err)

	raw, err := parseToMap(data)
	require.NoError(t, err)

	err = validateCluster(raw)
	require.NoError(t, err)

	kubeVersion, err := schemaDefault("properties", "kubernetesVersion", "default")
	require.NoError(t, err)
	serverType, err := schemaDefault("properties", "machineDefaults", "properties", "hcloud", "properties", "serverType", "default")
	require.NoError(t, err)
	datacenter, err := schemaDefault("properties", "machineDefaults", "properties", "hcloud", "properties", "datacenter", "default")
	require.NoError(t, err)
	talosImage, err := schemaDefault("properties", "machines", "items", "properties", "talosImage", "default")
	require.NoError(t, err)

	require.Equal(t, kubeVersion, raw["kubernetesVersion"])

	machines, ok := raw["machines"].([]any)
	require.True(t, ok)
	require.Len(t, machines, 1)

	machine := machines[0].(map[string]any)
	hcloud := machine["hcloud"].(map[string]any)
	require.Equal(t, serverType, hcloud["serverType"])
	require.Equal(t, datacenter, hcloud["datacenter"])
	require.Equal(t, talosImage, machine["talosImage"])
}
