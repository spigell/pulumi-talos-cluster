package applier_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	machineapi "github.com/siderolabs/talos/pkg/machinery/api/machine"
	"github.com/siderolabs/talos/pkg/machinery/compatibility"
	"github.com/siderolabs/talos/pkg/machinery/gendata"
	cluster "github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cluster/go"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
	"github.com/stretchr/testify/require"
)

// TestTestProgramConfigsAreValid renders the machine config of every machine in the hcloud test programs
// with the provider and real talosctl, using the provider's default version contract, and validates it
// with `talosctl validate`. An invalid config would boot a node that never starts the Talos API, so this
// catches broken config patches before an hcloud integration test creates servers.
func TestTestProgramConfigsAreValid(t *testing.T) {
	specs, err := filepath.Glob(filepath.Join("..", "testdata", "programs", "*", "cluster.yaml"))
	require.NoError(t, err)
	require.NotEmpty(t, specs)

	for _, spec := range specs {
		program := filepath.Base(filepath.Dir(spec))
		clu, err := cluster.Load(spec)
		require.NoError(t, err, program)

		for _, m := range clu.Machines {
			t.Run(program+"/"+m.ID, func(t *testing.T) {
				t.Setenv("PULUMI_MOCK_RESOURCES", "1")

				// The node checks this itself; the provider and `talosctl validate` skip it (local validation).
				k8s, err := compatibility.ParseKubernetesVersion(strings.TrimPrefix(clu.KubernetesVersion, "v"))
				require.NoError(t, err)
				talos, err := compatibility.ParseTalosVersion(&machineapi.VersionInfo{Tag: m.TalosImage[strings.LastIndex(m.TalosImage, ":")+1:]})
				require.NoError(t, err)
				require.NoError(t, k8s.SupportedWith(talos), "kubernetesVersion %s with talosImage %s", clu.KubernetesVersion, m.TalosImage)

				rendered := renderMachineConfig(t, clu, m)
				require.NotEmpty(t, rendered)

				file := filepath.Join(t.TempDir(), "machineconfig.yaml")
				require.NoError(t, os.WriteFile(file, []byte(rendered), 0o600))

				// #nosec G204 -- test runs talosctl on a file it wrote.
				out, err := exec.CommandContext(t.Context(), "talosctl", "validate", "--mode", "cloud", "--config", file).CombinedOutput()
				require.NoError(t, err, "talosctl validate: %s", out)
			})
		}
	}
}

func renderMachineConfig(t *testing.T, clu *cluster.Cluster, m *cluster.Machine) string {
	t.Helper()

	// RunErr does not wait for this ApplyT callback, so the rendered config is passed through a channel.
	rendered := make(chan string, 1)
	mock := &ProxyMock{t: t}
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		app, err := applier.New(ctx, clu.Name, pulumi.StringMap{}, nil)
		if err != nil {
			return err
		}
		app.WithHooks(false)

		secrets, err := app.GenerateSecrets()
		if err != nil {
			return err
		}

		patches := pulumi.StringArray{}
		for _, p := range m.ConfigPatches {
			patches = append(patches, pulumi.String(p))
		}

		configuration, err := app.GenerateMachineConfig(&types.Cluster{
			ClusterName:          clu.Name,
			ClusterEndpoint:      pulumi.String("https://10.10.10.5:6443"),
			KubernetesVersion:    pulumi.String(clu.KubernetesVersion),
			TalosVersionContract: pulumi.String(gendata.VersionTag),
		}, &types.ClusterMachine{
			MachineID:     m.ID,
			MachineType:   m.Type,
			ConfigPatches: patches,
			TalosImage:    pulumi.StringPtr(m.TalosImage),
		}, secrets)
		if err != nil {
			return err
		}

		configuration.ApplyT(func(raw string) string {
			rendered <- raw
			return raw
		})

		return nil
	}, pulumi.WithMocks("project", "stack", mock))
	require.NoError(t, err)

	select {
	case raw := <-rendered:
		return raw
	case <-time.After(30 * time.Second):
		// Pulumi mocks do not surface the provider's validation error here; `pulumi up` does.
		t.Fatal("no rendered machine config: the provider rejected it as invalid (check the configPatches) or talosctl gen config failed, see the log above")
		return ""
	}
}
