package applier

import (
	"fmt"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	tmachine "github.com/siderolabs/talos/pkg/machinery/config/machine"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier/hooks"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/applier/talosctl"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
)

func (a *Applier) upgrade(m *types.MachineInfo, role tmachine.Type, deps []pulumi.Resource) (pulumi.Resource, error) {
	opts := []pulumi.ResourceOption{
		a.parent,
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "10m", Update: "10m"}),
		pulumi.DependsOn(deps),
	}

	etcdMemberTarget := a.etcdMembers

	if role == tmachine.TypeInit {
		etcdMemberTarget = 1
	}

	if (role == tmachine.TypeInit || role == tmachine.TypeControlPlane) && a.opts.etcdHookEnabled {
		// Populate upgrade hooks lazily and reuse if already registered.
		if _, ok := a.hooks[hookStageUpgrade]; !ok {
			h, err := a.ctx.RegisterResourceHook("health-check", hooks.EtcdReadyHook(a.ctx.Log), nil)
			if err != nil {
				return nil, err
			}
			a.hooks[hookStageUpgrade] = []*pulumi.ResourceHook{h}
		}

		if hooksForStage, ok := a.hooks[hookStageUpgrade]; ok && len(hooksForStage) > 0 {
			opts = append(opts, pulumi.ResourceHooks(&pulumi.ResourceHookBinding{
				BeforeCreate: hooksForStage,
				BeforeUpdate: hooksForStage,
			}))
		}
	}

	stageName := "cli-upgrade"
	home := generateWorkDirNameForTalosctl(a.name, stageName, m.MachineID)
	t := talosctl.New().
		WithNodeIPInput(m.NodeIP).
		WithTalosConfig(a.TalosconfigForNode(m.NodeIP))
	upgradeArgs := m.TalosImage.ToStringOutput().ApplyT(func(image string) (string, error) {
		return talosctlUpgradeArgs(image, m.MachineID)
	}).(pulumi.StringOutput)

	return t.RunCommand(a.ctx, fmt.Sprintf("%s:%s:%s", a.name, stageName, m.MachineID), &talosctl.Args{
		PrepareDeps: deps,
		Dir:         home,
		CommandArgs: upgradeArgs,
		RetryCount:  10,
		Environment: pulumi.StringMap{
			"NODE_IP":            m.NodeIP,
			"TALOSCTL_HOME":      pulumi.String(home),
			"ETCD_MEMBER_TARGET": pulumi.String(fmt.Sprint(etcdMemberTarget)),
			"TALOS_TARGET_IMAGE": m.TalosImage,
		},
		UpdateOnChange: true,
		UpdateSkipIf:   upgradeSkipIfNodeRunsTargetImage,
	}, opts...)
}

// upgradeSkipIfNodeRunsTargetImage skips a re-run of talosctl upgrade (and its reboot) when the node
// already runs TALOS_TARGET_IMAGE: the installed machine config points at the image, the running Talos
// tag matches the image tag and, for Image Factory images, the running schematic matches.
// Anything that cannot be read or does not match keeps the upgrade.
const upgradeSkipIfNodeRunsTargetImage = `target="$TALOS_TARGET_IMAGE"
ref="${target##*/}"
ref="${ref%%@*}"
case "$ref" in *:*) ;; *) exit 1 ;; esac
tag="${ref##*:}"
name="${ref%%:*}"
installed=$(talosctl_node get machineconfig v1alpha1 -o yaml 2>/dev/null | awk '
  !f && $1 == "install:" && NF == 1 { f = 1; match($0, /^ */); indent = RLENGTH; next }
  f { match($0, /^ */); if ($0 !~ /^ *$/ && RLENGTH <= indent) { f = 0 } else if ($1 == "image:") { v = $2; gsub(/"/, "", v); print v; exit } }')
[ -n "$installed" ] && [ "$installed" = "$target" ] || exit 1
running=$(talosctl_node version 2>/dev/null | awk '/^Server:/ { s = 1 } s && $1 == "Tag:" { print $2; exit }')
[ "$running" = "$tag" ] || exit 1
if [[ "$name" =~ ^[0-9a-f]{64}$ ]]; then
  schematic=$(talosctl_node get extensions -o yaml 2>/dev/null | awk '$1 == "name:" && $2 == "schematic" { f = 1; next } f && $1 == "version:" { print $2; exit }')
  [ "$schematic" = "$name" ] || exit 1
fi
echo "node $NODE_IP already runs $target; skipping talosctl upgrade"`

func talosctlUpgradeArgs(image, machineID string) (string, error) {
	img := strings.TrimSpace(image)
	if img == "" {
		return "", fmt.Errorf("talos image is required for machine %s", machineID)
	}

	// --drain defaults to true since talosctl v1.13 and requires a kubeconfig
	// from the target node, which workers cannot serve during provisioning.
	base := fmt.Sprintf("upgrade --debug --drain=false --image %s", img)

	return base, nil
}
