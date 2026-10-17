package examples

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/testing/integration"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/stretchr/testify/assert"
)

func TestHcloudClusterGo(t *testing.T) {
	test := getGoBaseOptions(t).
		With(integration.ProgramTestOptions{
			RunUpdateTest: false,
			Dir:           filepath.Join(getTestPrograms(t), "hcloud-go"),
			EditDirs: []integration.EditDir{
				{
					// Adds hcloud-go-worker-2 and a config patch on hcloud-go-worker-1.
					Dir:                    filepath.Join(getTestPrograms(t), "hcloud-go-edit"),
					Additive:               true,
					ExtraRuntimeValidation: validateAddWorkerAndPatchEdit,
				},
			},
		})

	programTest(t, &test)
}

// validateAddWorkerAndPatchEdit checks how the CLI stages react to a new node and to a config change on an
// existing node: no command is replaced, the changed node only re-applies its config (no talosctl upgrade,
// so no reboot), the untouched control plane is not upgraded, and the new node is created.
func validateAddWorkerAndPatchEdit(t *testing.T, stack integration.RuntimeValidationStackInfo) {
	ops := commandOps(stack.Events)
	for name, op := range ops {
		t.Logf("edit step: %s %s", op, name)
		switch op {
		case apitype.OpReplace, apitype.OpCreateReplacement, apitype.OpDeleteReplaced:
			t.Errorf("command %s was replaced (%s); CLI stages must update in place", name, op)
		}
	}

	assertCommandOp(t, ops, "cli-upgrade:hcloud-go-controlplane-1", apitype.OpSame)
	assertCommandOp(t, ops, "cli-upgrade:hcloud-go-worker-1", apitype.OpSame)
	assertCommandOp(t, ops, "cli-apply-config:hcloud-go-worker-1", apitype.OpUpdate)
	assertCommandOp(t, ops, "cli-upgrade:hcloud-go-worker-2", apitype.OpCreate)
	assertCommandOp(t, ops, "cli-apply-config:hcloud-go-worker-2", apitype.OpCreate)
}

// commandOps returns the operation of every command:local:Command in the update, keyed by resource name.
func commandOps(events []apitype.EngineEvent) map[string]apitype.OpType {
	ops := map[string]apitype.OpType{}
	for _, e := range events {
		var md *apitype.StepEventMetadata
		switch {
		case e.ResourcePreEvent != nil && !e.ResourcePreEvent.Planning:
			md = &e.ResourcePreEvent.Metadata
		case e.ResOutputsEvent != nil && !e.ResOutputsEvent.Planning:
			md = &e.ResOutputsEvent.Metadata
		default:
			continue
		}
		if md.Type != "command:local:Command" {
			continue
		}
		name := md.URN[strings.LastIndex(md.URN, "::")+2:]
		// A replacement emits several steps; keep the one that shows it.
		if prev, ok := ops[name]; ok && prev != apitype.OpSame && prev != apitype.OpUpdate && prev != apitype.OpCreate {
			continue
		}
		ops[name] = md.Op
	}

	return ops
}

func assertCommandOp(t *testing.T, ops map[string]apitype.OpType, suffix string, want apitype.OpType) {
	t.Helper()
	for name, op := range ops {
		if strings.HasSuffix(name, ":"+suffix) {
			assert.Equal(t, want, op, "operation for %s", name)
			return
		}
	}
	t.Errorf("no command %s in the edit step", suffix)
}

func TestHcloudHAClusterGo(t *testing.T) {
	test := getGoBaseOptions(t).
		With(integration.ProgramTestOptions{
			RunUpdateTest: false,
			Dir:           filepath.Join(getTestPrograms(t), "hcloud-ha-go"),
		})

	programTest(t, &test)
}

func getGoBaseOptions(t *testing.T) integration.ProgramTestOptions {
	base := getBaseOptions(t)
	goBase := base.With(integration.ProgramTestOptions{
		Dependencies: []string{
			// Path to GO SDK
			getCwd(t) + "../../sdk",
			getCwd(t) + "../../integration-tests",
		},
	})

	return goBase
}
