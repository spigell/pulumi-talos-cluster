package talosctl

import (
	"strings"
	"testing"
)

func TestTalosctlDefaultsUseLocalBinary(t *testing.T) {
	talos := New()

	if talos.Binary != "talosctl" {
		t.Fatalf("expected default binary to be talosctl, got %q", talos.Binary)
	}

	if !strings.Contains(talos.BasicCommand, "talosctl --talosconfig talosctl.yaml") {
		t.Fatalf("expected basic command to use talosctl config, got %q", talos.BasicCommand)
	}

	if strings.Contains(talos.BasicCommand, "pulumiverse") {
		t.Fatalf("unexpected pulumiverse reference in basic command: %q", talos.BasicCommand)
	}
}

func TestTalosctlWithNodeIPExtendsCommand(t *testing.T) {
	ip := "1.2.3.4"
	talos := New().WithNodeIP(ip)

	expectedFragment := " -n 1.2.3.4 -e 1.2.3.4"
	if !strings.Contains(talos.BasicCommand, expectedFragment) {
		t.Fatalf("expected node flags %q in basic command, got %q", expectedFragment, talos.BasicCommand)
	}

	if !strings.HasPrefix(talos.BasicCommand, talos.Binary) {
		t.Fatalf("expected basic command to start with binary %q, got %q", talos.Binary, talos.BasicCommand)
	}
}
