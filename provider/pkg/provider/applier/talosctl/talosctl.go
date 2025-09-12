package talosctl

import (
	"encoding/base64"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	talosctlBinary     = "talosctl"
	talosctlConfigName = "talosctl.yaml"
)

var interpreter = []string{
	"/bin/bash",
	"-c",
}

// Talosctl wraps the talosctl binary and basic command configuration.
type Talosctl struct {
	Binary       string
	BasicCommand string
}

// Args groups arguments used to execute a talosctl command.
type Args struct {
	TalosConfig     pulumi.StringInput
	PrepareDeps     []pulumi.Resource
	Dir             string
	CommandArgs     pulumi.StringInput
	RetryCount      int
	Environment     pulumi.StringMap
	Triggers        pulumi.Array
	AdditionalFiles []ExtraFile
}

// ExtraFile describes an additional file to place alongside talosctl.yaml.
type ExtraFile struct {
	Name    string
	Content pulumi.StringInput
}

// New creates a Talosctl initialized with the default binary and config path.
func New() *Talosctl {
	return &Talosctl{
		Binary:       talosctlBinary,
		BasicCommand: fmt.Sprintf("%s --talosconfig %s", talosctlBinary, talosctlConfigName),
	}
}

// WithNodeIP adds `-n` and `-e` flags for the provided node IP address.
func (t *Talosctl) WithNodeIP(ip string) *Talosctl {
	t.BasicCommand = fmt.Sprintf("%s -n %s -e %s", t.BasicCommand, ip, ip)

	return t
}

// RunCommand executes a talosctl command as a Pulumi resource.
func (t *Talosctl) RunCommand(
	ctx *pulumi.Context,
	name string,
	a *Args,
	opts ...pulumi.ResourceOption,
) (pulumi.Resource, error) {
	createGated, env, err := t.prepareAndGate(ctx, a)
	if err != nil {
		return nil, err
	}

	main, err := local.NewCommand(ctx, name, &local.CommandArgs{
		Create: createGated.ApplyT(func(args string) string {
			return withBashRetry(
				fmt.Sprintf("%s %s", t.BasicCommand, args),
				fmt.Sprint(a.RetryCount+1),
			)
		}).(pulumi.StringOutput),
		Dir:         pulumi.String(a.Dir),
		Interpreter: pulumi.ToStringArray(interpreter),
		Environment: env,
		Triggers:    a.Triggers,
	}, opts...)
	if err != nil {
		return nil, err
	}

	// Hidden cleanup after the resource completes.
	_ = local.RunOutput(ctx, local.RunOutputArgs{
		Command:     pulumi.Sprintf(`rm -rf %q`, a.Dir),
		Interpreter: pulumi.ToStringArray(interpreter),
	}, pulumi.DependsOn([]pulumi.Resource{main}))

	return main, nil
}

// RunGetCommand executes a talosctl command and returns its standard output.
func (t *Talosctl) RunGetCommand(
	ctx *pulumi.Context,
	a *Args,
	deps []pulumi.Resource,
) (pulumi.StringOutput, error) {
	createGated, env, err := t.prepareAndGate(ctx, a)
	if err != nil {
		// Return a zero output with error
		return pulumi.StringOutput{}, err
	}

	// Compose main + inline cleanup (no resource to depend on)
	cmd := createGated.ApplyT(func(args string) string {
		main := withBashRetry(
			fmt.Sprintf("%s %s", t.BasicCommand, args),
			fmt.Sprint(a.RetryCount+1),
		)
		cleanup := fmt.Sprintf("rm -rf %s", a.Dir)
		return main + " && " + cleanup
	}).(pulumi.StringOutput)

	out := local.RunOutput(ctx, local.RunOutputArgs{
		Command:     cmd,
		Interpreter: pulumi.ToStringArray(interpreter),
		// Only log stderr since stdout can keep a sensitive data.
		Logging:     local.LoggingStderr,
		Environment: env,
		Dir:         pulumi.String(a.Dir),
	}, pulumi.DependsOn(deps))

	return out.Stdout(), nil
}

func (t *Talosctl) prepareAndGate(ctx *pulumi.Context, args *Args) (createGated pulumi.StringOutput, env pulumi.StringMap, err error) {
	if args.CommandArgs == nil {
		return pulumi.StringOutput{}, nil, fmt.Errorf("Args.CommandArgs is required")
	}

	if args.Dir == "" {
		return pulumi.StringOutput{}, nil, fmt.Errorf("Args.Dir is required")
	}

	env = args.Environment
	if env == nil {
		env = pulumi.StringMap{}
	}

	// Prepare: write talosctl.yaml + additional files
	prepared := t.prepareAll(ctx, args)

	createGated = pulumi.All(prepared, args.CommandArgs).
		ApplyT(func(v []any) string {
			if !v[0].(bool) {
				return ""
			}
			return v[1].(string)
		}).(pulumi.StringOutput)

	return createGated, env, nil
}

// prepareAll builds one shell that writes talosctl.yaml and any AdditionalFiles.
// It uses `local.RunOutput` as the prepare step, returning a BoolOutput.
func (t *Talosctl) prepareAll(ctx *pulumi.Context, args *Args) pulumi.BoolOutput {
	// Gather inputs: main config + each extra file content
	inputs := []any{args.TalosConfig}
	for _, f := range args.AdditionalFiles {
		inputs = append(inputs, f.Content)
	}

	// Build the full shell command as a StringOutput from resolved inputs
	cmd := pulumi.All(inputs...).ApplyT(func(resolved []any) string {
		// resolved[0] = main talosctl.yaml content
		main := resolved[0].(string)

		// 1) talosctl.yaml
		talosConfigPath := filepath.Join(args.Dir, talosctlConfigName)
		mainB64 := base64.StdEncoding.EncodeToString([]byte(main))
		var b strings.Builder
		// Ensure TALOS_HOME exists, private perms
		fmt.Fprintf(&b, `mkdir -p %s && umask 077; `, args.Dir)
		// Write talosctl.yaml
		fmt.Fprintf(&b, `printf %%s %q | base64 -d > %s && chmod 600 %s`,
			mainB64, talosConfigPath, talosConfigPath)

		// 2) Additional files
		for i, ef := range args.AdditionalFiles {
			content := resolved[1+i].(string)
			contentB64 := base64.StdEncoding.EncodeToString([]byte(content))

			finalPath := filepath.Join(args.Dir, ef.Name)
			fmt.Fprintf(&b, ` && printf %%s %q | base64 -d > %s && chmod 600 %s`,
				contentB64, finalPath, finalPath)
		}

		return b.String()
	}).(pulumi.StringOutput)

	// Execute prepare (single invoke). Error propagates via rejected output.
	run := local.RunOutput(ctx, local.RunOutputArgs{
		Command:     cmd,
		Environment: args.Environment,
		Interpreter: pulumi.ToStringArray(interpreter),
	}, pulumi.DependsOn(args.PrepareDeps))

	// Map stderr to success/failure, preserving errors.
	return run.Stderr().ApplyT(func(s string) (bool, error) {
		if s != "" {
			return false, fmt.Errorf("prepare error (stderr is not empty): %s", s)
		}
		return true, nil
	}).(pulumi.BoolOutput)
}

func withBashRetry(cmd, maxTries string) string {
	return fmt.Sprintf(strings.Join([]string{
		"n=0",
		"until [ $n -ge %[1]s ]",
		"do %s && break",
		"sleep 10",
		"n=$((n+1))",
		"done",
		// Exiting with 0 if command succeeded.
		// Otherwise exit with 10 exit code.
		"[ $n -ge %[1]s ] && exit 10 || true",
	}, " ; "), maxTries, cmd)
}
