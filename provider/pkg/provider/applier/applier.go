package applier

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	//"runtime"
	"time"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumiverse/pulumi-talos/sdk/go/talos/client"
	"github.com/pulumiverse/pulumi-talos/sdk/go/talos/machine"
	"github.com/spigell/pulumi-talos-cluster/provider/pkg/provider/types"
)

type Applier struct {
	ctx                 *pulumi.Context
	name                string
	clientConfiguration *machine.ClientConfigurationArgs
	parent              pulumi.ResourceOption
	commnanInterpreter  pulumi.StringArray
	skipInitNode        bool

	etcdMembers int
	etcdReadyHook *pulumi.ResourceHook

	InitNode *InitNode
}

type InitNode struct {
	IP   string
	Name string
}

func New(ctx *pulumi.Context, name string, client *machine.ClientConfigurationArgs, parent pulumi.ResourceOption) *Applier {
	return &Applier{
		name:                name,
		ctx:                 ctx,
		parent:              parent,
		clientConfiguration: client,
		// 1 is default value, because we have at least one init node.
		etcdMembers: 1,
		commnanInterpreter: pulumi.StringArray{
			pulumi.String("/bin/bash"),
			pulumi.String("-c"),
		},
	}
}

func (a *Applier) WithSkipedInitApply(skip bool) *Applier {
	a.skipInitNode = skip

	return a
}

func (a *Applier) WithEtcdMembersCount(count int) *Applier {
	a.etcdMembers = count

	return a
}

func (a *Applier) Init(m *types.MachineInfo) ([]pulumi.Resource, error) {
	// The Init node is special. We need to init by ourselves.
	applied, err := a.initApply(m, nil)
	if err != nil {
		return nil, err
	}

	deps := []pulumi.Resource{applied}

	bootstrap, err := machine.NewBootstrap(a.ctx, fmt.Sprintf("%s:bootstrap:%s", a.name, m.MachineID), &machine.BootstrapArgs{
		ClientConfiguration: a.clientConfiguration,
		Node:                pulumi.String(m.NodeIP),
	}, a.parent,
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "1m", Update: "1m"}),
		pulumi.DependsOn(deps),
	)
	if err != nil {
		return nil, err
	}

	deps = append(deps, bootstrap)

	cli, err := a.cliApply(m, deps, "init")
	if err != nil {
		return deps, err
	}

	return append(deps, cli...), nil
}

func (a *Applier) InitControlplane(m *types.MachineInfo, deps []pulumi.Resource) ([]pulumi.Resource, error) {
	if !a.skipInitNode {
		applied, err := a.initApply(m, deps)
		if err != nil {
			return nil, err
		}
		deps = append(deps, applied)
	}

	return deps, nil
}

func (a *Applier) ApplyToControlplane(m *types.MachineInfo, deps []pulumi.Resource) ([]pulumi.Resource, error) {
	cli, err := a.cliApply(m, deps, "controlplane")
	if err != nil {
		return deps, err
	}

	return append(deps, cli...), nil
}

func (a *Applier) ApplyTo(m *types.MachineInfo, deps []pulumi.Resource) ([]pulumi.Resource, error) {
	if !a.skipInitNode {
		applied, err := a.initApply(m, deps)
		if err != nil {
			return nil, err
		}
		deps = append(deps, applied)
	}

	cli, err := a.cliApply(m, deps, "worker")
	if err != nil {
		return deps, err
	}

	return append(deps, cli...), nil
}

func (a *Applier) SetHooks() error {
	etcdReadyHook, err := a.ctx.RegisterResourceHook("health-check", a.etcdReady, nil)
	a.etcdReadyHook = etcdReadyHook

	return err
}

func (a *Applier) cliApply(m *types.MachineInfo, deps []pulumi.Resource, role string) ([]pulumi.Resource, error) {
	hooks := []*pulumi.ResourceHook{a.etcdReadyHook}

	cmd := a.talosctlUpgradeCMD(m, deps)

	opts := []pulumi.ResourceOption{
		a.parent,
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "10m", Update: "10m"}),
		pulumi.DependsOn(deps),
	}

	etcdMemberTarget := a.etcdMembers

	if role == "init" {
		etcdMemberTarget = 1
	}

	if role == "controlplane" || role == "init" {
	//if role == "none" {
		opts = append(opts, pulumi.ResourceHooks(&pulumi.ResourceHookBinding{
			BeforeCreate: hooks,
			BeforeUpdate: hooks,
		}))
	}

	set, err := local.NewCommand(a.ctx, fmt.Sprintf("%s:cli-set-talos-version:%s", a.name, m.MachineID), &local.CommandArgs{
		Create: cmd.Command,
		Environment: pulumi.StringMap{
			"NODE_IP":       pulumi.String(m.NodeIP),
			"TALOSCTL_HOME": pulumi.String(cmd.Home.Dir),
			"ETCD_MEMBER_TARGET": pulumi.String(fmt.Sprint(etcdMemberTarget)),
		},
		Triggers: pulumi.Array{
			pulumi.String(m.TalosImage),
		},
		Interpreter: a.commnanInterpreter,
	},
		opts...,
	)
	if err != nil {
		return nil, err
	}

	deps = append(deps, set)

	apply, err := local.NewCommand(a.ctx, fmt.Sprintf("%s:cli-apply:%s", a.name, m.MachineID), &local.CommandArgs{
		Create: a.talosctlApplyCMD(m, deps),
		Triggers: pulumi.Array{
			pulumi.String(m.UserConfigPatches),
			pulumi.String(m.ClusterEnpoint),
		},
		Interpreter: a.commnanInterpreter,
	}, a.parent,
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "90s", Update: "90s"}),
		pulumi.DependsOn(deps),
	)
	if err != nil {
		return deps, err
	}

	deps = append(deps, apply)

	return deps, nil
}

func (a *Applier) UpgradeK8S(ma []*types.MachineInfo, deps []pulumi.Resource) ([]pulumi.Resource, error) {
	k8s, err := local.NewCommand(a.ctx, fmt.Sprintf("%s:cli-set-k8s-version:%s", a.name, ma[0].MachineID), &local.CommandArgs{
		Create:      a.talosctlUpgradeK8SCMD(ma),
		Interpreter: a.commnanInterpreter,
		Triggers: pulumi.Array{
			pulumi.String(ma[0].KubernetesVersion),
		},
	}, a.parent,
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "20m", Update: "20m"}),
		pulumi.DependsOn(deps),
	)
	if err != nil {
		return deps, err
	}

	return append(deps, k8s), nil
}

func (a *Applier) etcdReady(args *pulumi.ResourceHookArgs) error {
	envObj := args.NewInputs["environment"].ObjectValue().Mappable()

	ip, ok := envObj["NODE_IP"].(string)
	if !ok || ip == "" {
		return fmt.Errorf("environment.NODE_IP is missing or not a string")
	}
	workDir, ok := envObj["TALOSCTL_HOME"].(string)
	if !ok || workDir == "" {
		return fmt.Errorf("environment.TALOSCTL_HOME is missing or not a string")
	}
	etcdMembersTarget, ok := envObj["ETCD_MEMBER_TARGET"].(string)
	if !ok || etcdMembersTarget == "" {
		return fmt.Errorf("environment.ETCD_MEMBER_TARGET is missing or not a string")
	}

	const (
		maxRetries    = 10
		healthTimeout = 10 * time.Second
		listTimeout   = 10 * time.Second
	)

	runTalosctl := func(timeout time.Duration, args ...string) ([]byte, error) {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		args = append(args, "--talosconfig", "./talosctl.yaml")
		cmd := exec.CommandContext(ctx, "talosctl", args...)
		a.ctx.Log.Debug(fmt.Sprintf("health: command: %v", cmd), nil)
		cmd.Dir = workDir
		return cmd.Output()
	}

	expected, err := strconv.ParseInt(etcdMembersTarget, 10, 0)  // final desired size (e.g., 3)
	if err != nil {
		//  TO DO: return
		//return 

		return err
	}
	consecutiveOK := 0         // require 2 consecutive matches to avoid flapping
	const okStreak = 2

	for attempt := 1; attempt <= maxRetries; attempt++ {
		backoff := time.Duration(attempt) * time.Second

		// 1) health
		if _, err := runTalosctl(healthTimeout, "-n", ip, "-e", ip, "etcd", "status"); err != nil {
			a.ctx.Log.Debug(fmt.Sprintf("etcd health attempt %d/%d failed: %v", attempt, maxRetries, err), nil)
			time.Sleep(backoff)
			continue
		}

		// 2) members (tabular, not JSON)
		out, err := runTalosctl(listTimeout, "-n", ip, "-e", ip, "etcd", "members")
		if err != nil {
			a.ctx.Log.Debug(fmt.Sprintf("etcd members attempt %d/%d failed: %v", attempt, maxRetries, err), nil)
			time.Sleep(backoff)
			continue
		}

		got, perr := countEtcdMembersFromTable(out)
		if perr != nil {
			a.ctx.Log.Debug(fmt.Sprintf("parse members attempt %d/%d failed: %v", attempt, maxRetries, perr), nil)
			time.Sleep(backoff)
			continue
		}

		match := (got == expected)

		if !match {
			consecutiveOK = 0
			a.ctx.Log.Debug(fmt.Sprintf("attempt %d/%d: expected %d, got %d",
				attempt, maxRetries, expected, got), nil)
			time.Sleep(backoff)
			continue
		}

		consecutiveOK++
		if consecutiveOK < okStreak {
			// keep looping to ensure stability
			a.ctx.Log.Debug(fmt.Sprintf("attempt %d/%d: success. waiting for consecutiveOK: %d/%d",
				attempt, maxRetries, consecutiveOK, okStreak), nil)
			time.Sleep(backoff / 2)
			continue
		}

		a.ctx.Log.Info(
			fmt.Sprintf("[INFO] talos-cluster: etcd health check passed. attempts %d/%d. etcdMembers: %d",
				attempt, maxRetries, got),
			nil,
		)

		a.ctx.Log.Info(fmt.Sprintf("[INFO] talos-cluster: etcd health check passed. attempts %d/%d made. etcdMember: %d", attempt, maxRetries, got), nil)
		return nil
	}

	return fmt.Errorf("etcd health check failed after %d attempts", maxRetries)
}

// countEtcdMembersFromTable parses `talosctl etcd members` tabular output.
// It ignores the first non-empty header line and counts subsequent non-empty lines.
func countEtcdMembersFromTable(out []byte) (int64, error) {
	sc := bufio.NewScanner(bytes.NewReader(out))
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)

	lines := make([]string, 0, 8)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	if err := sc.Err(); err != nil {
		return 0, fmt.Errorf("scan output: %w", err)
	}
	if len(lines) == 0 {
		return 0, fmt.Errorf("no output from talosctl etcd members")
	}

	// First non-empty line is the header. Everything after is a member row.
	memberRows := lines[1:]
	count := 0
	for _, row := range memberRows {
		// Be safe: skip accidental separator lines, etc.
		// Real rows should have multiple columns when split by fields.
		if len(strings.Fields(row)) >= 3 {
			count++
		}
	}

	return int64(count), nil
}

func (a *Applier) initApply(m *types.MachineInfo, deps []pulumi.Resource) (pulumi.Resource, error) {

	apply, err := machine.NewConfigurationApply(a.ctx, fmt.Sprintf("%s:initial-apply:%s", a.name, m.MachineID), &machine.ConfigurationApplyArgs{
		Node:                      pulumi.String(m.NodeIP),
		MachineConfigurationInput: pulumi.String(m.Configuration),
		// Staged is not supported in maintenance.
		// NoReboot can lead to failures.
		ApplyMode: pulumi.String("reboot"),
		OnDestroy: &machine.ConfigurationApplyOnDestroyArgs{
			Graceful: pulumi.Bool(true),
			Reboot:   pulumi.Bool(false),
			Reset:    pulumi.Bool(false),
		},
		Timeouts: &machine.TimeoutArgs{
			Create: pulumi.String("1m"),
			Update: pulumi.String("1m"),
		},
		ClientConfiguration: a.clientConfiguration,
	}, a.parent,
		// Ignore changes to machineConfigurationInput to prevent unnecessary updates since there will be an additional apply via cli.
		// Generated configuration has a contract with immutable talos version and sometimes new options can be skipped.
		pulumi.IgnoreChanges([]string{"machineConfigurationInput", "applyMode"}),
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "1m", Update: "1m"}),
		pulumi.DependsOn(deps),
	)
	if err != nil {
		return nil, err
	}

	deps = append(deps, apply)

	return local.NewCommand(a.ctx, fmt.Sprintf("%s:reboot:%s", a.name, m.MachineID), &local.CommandArgs{
		Create:      a.talosctlFastReboot(m),
		Interpreter: a.commnanInterpreter,
	}, a.parent,
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "5m", Update: "5m"}),
		pulumi.DependsOn(deps),
	)
}

func (a *Applier) basicClient() client.GetConfigurationResultOutput {
	return client.GetConfigurationOutput(a.ctx, client.GetConfigurationOutputArgs{
		ClusterName: pulumi.String(a.name),
		ClientConfiguration: &client.GetConfigurationClientConfigurationArgs{
			CaCertificate:     a.clientConfiguration.CaCertificate,
			ClientKey:         a.clientConfiguration.ClientKey,
			ClientCertificate: a.clientConfiguration.ClientCertificate,
		},
	})
}

func (a *Applier) NewTalosconfig(endpoints []string, nodes []string) client.GetConfigurationResultOutput {
	return client.GetConfigurationOutput(a.ctx, client.GetConfigurationOutputArgs{
		ClusterName: pulumi.String(a.name),
		Endpoints:   pulumi.ToStringArray(endpoints),
		Nodes:       pulumi.ToStringArray(nodes),
		ClientConfiguration: &client.GetConfigurationClientConfigurationArgs{
			CaCertificate:     a.clientConfiguration.CaCertificate,
			ClientKey:         a.clientConfiguration.ClientKey,
			ClientCertificate: a.clientConfiguration.ClientCertificate,
		},
	})
}
