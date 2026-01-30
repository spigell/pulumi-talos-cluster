package applier_test

import (
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier/talosctl"
	"github.com/stretchr/testify/assert"
)

// Ensure TryInsecureFirst emits a secure->insecure fallback when talosconfig is present.
func TestApplyConfigUsesInsecureFallback(t *testing.T) {
	mock := &ProxyMock{t: t}

	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		tctl := talosctl.New().
			WithTalosConfig(pulumi.String(minimalTalosconfig))

		dir := filepath.Join(t.TempDir(), "talosctl-home")

		// Command doesn't need to contact a node; `--help` returns immediately.
		_, runErr := tctl.RunCommand(ctx, "test-apply:try-insecure-first", &talosctl.Args{
			Dir:              dir,
			CommandArgs:      pulumi.String("--help"),
			TryInsecureFirst: true,
		})

		return runErr
	}, pulumi.WithMocks("project", "stack", mock))

	assert.NoError(t, err)
	assert.NotEmpty(t, mock.lastCreate)
	assert.Contains(t, mock.lastCreate, ") || (")
	assert.Contains(t, mock.lastCreate, "talosctl --help --insecure")
	assert.Contains(t, mock.lastCreate, "talosctl --help --talosconfig talosctl.yaml")
}

const minimalTalosconfig = `
context: test
contexts:
  test:
    endpoints: []
    nodes: []
    ca: Zgo=
    crt: Zgo=
    key: Zgo=
`
