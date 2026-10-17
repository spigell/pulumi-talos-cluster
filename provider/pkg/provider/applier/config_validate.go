package applier

import (
	"fmt"
	"strings"

	"github.com/siderolabs/talos/pkg/machinery/config/configloader"
	"github.com/siderolabs/talos/pkg/machinery/config/validation"
)

// configValidationMode validates generated machine configs without platform-specific requirements:
// no install disk is required and the node is not a container. This reports invalid or conflicting
// config patches (for example legacy machine.kubelet fields next to a KubeNodeConfig document)
// for every platform.
type configValidationMode struct{}

func (configValidationMode) String() string        { return "cloud" }
func (configValidationMode) RequiresInstall() bool { return false }
func (configValidationMode) InContainer() bool     { return false }

// validateMachineConfig loads and validates a machine config produced by `talosctl gen config`.
// A node that receives an invalid config never starts the Talos API, so failing here shows the
// problem with the patches instead of a connection refused error later.
func validateMachineConfig(machineID, raw string) error {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	cfg, err := configloader.NewFromBytes([]byte(raw))
	if err != nil {
		return fmt.Errorf("machine %s: generated machine config cannot be loaded, check its configPatches: %w", machineID, err)
	}

	// WithLocal matches `talosctl validate`: it skips the Kubernetes version compatibility checks, which compare
	// against the Talos version of this library instead of the Talos version running on the node.
	if _, err := cfg.ValidateAsClient(configValidationMode{}, validation.WithLocal()); err != nil {
		return fmt.Errorf("machine %s: generated machine config is invalid, check its configPatches: %w", machineID, err)
	}

	return nil
}
