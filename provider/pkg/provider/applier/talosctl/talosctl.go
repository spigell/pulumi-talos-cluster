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

// CatFile produces a StringOutput with the contents of a file in Dir.
func (t *Talosctl) CatFile(ctx *pulumi.Context, dir, filename string, deps []pulumi.Resource) pulumi.StringOutput {
	out := local.RunOutput(ctx, local.RunOutputArgs{
		Command:     pulumi.Sprintf("cat %s", filepath.Join(dir, filename)),
		Interpreter: pulumi.ToStringArray(interpreter),
	}, pulumi.DependsOn(deps))
	return out.Stdout()
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
		BasicCommand: talosctlBinary,
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
		Create:      createGated,
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

	out := local.RunOutput(ctx, local.RunOutputArgs{
		Command:     createGated,
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

	var (
		useTalosconfig pulumi.BoolOutput
		talosConfig    pulumi.StringInput
	)

	if args.TalosConfig == nil {
		// No talosconfig provided; skip writing/flag entirely.
		useTalosconfig = pulumi.Bool(false).ToBoolOutput()
		talosConfig = pulumi.String("")
	} else {
		talosConfig = args.TalosConfig
		useTalosconfig = pulumi.StringInput(args.TalosConfig).ToStringPtrOutput().ApplyT(func(v *string) bool {
			return v != nil
		}).(pulumi.BoolOutput)
	}

	// Prepare: write talosctl.yaml + additional files
	prepared := t.prepareAll(ctx, args, talosConfig, useTalosconfig)

	createGated = pulumi.All(prepared, args.CommandArgs, useTalosconfig).
		ApplyT(func(v []any) string {
			if !v[0].(bool) {
				return ""
			}

			cmdArgs := v[1].(string)
			withConfig := v[2].(bool)

			base := t.BasicCommand
			if withConfig {
				base = fmt.Sprintf("%s --talosconfig %s", base, talosctlConfigName)
			}

			return withBashRetry(
				fmt.Sprintf("%s %s", base, cmdArgs),
				fmt.Sprint(args.RetryCount+1),
			)
		}).(pulumi.StringOutput)

	return createGated, env, nil
}

// prepareAll builds one shell that writes talosctl.yaml (if provided) and any AdditionalFiles.
// It uses `local.RunOutput` as the prepare step, returning a BoolOutput.
func (t *Talosctl) prepareAll(ctx *pulumi.Context, args *Args, talosConfig pulumi.StringInput, useTalosconfig pulumi.BoolOutput) pulumi.BoolOutput {
	// Gather inputs: main config + each extra file content
	inputs := []any{talosConfig, useTalosconfig}
	for _, f := range args.AdditionalFiles {
		inputs = append(inputs, f.Content)
	}

	// Build the full shell command as a StringOutput from resolved inputs
	cmd := pulumi.All(inputs...).ApplyT(func(resolved []any) string {
		// resolved[0] = main talosctl.yaml content
		main := resolved[0].(string)
		useConfig := resolved[1].(bool)

		// 1) talosctl.yaml
		talosConfigPath := filepath.Join(args.Dir, talosctlConfigName)
		var b strings.Builder
		// Ensure TALOS_HOME exists, private perms
		fmt.Fprintf(&b, `mkdir -p %s && umask 077`, args.Dir)

		// Write talosctl.yaml only if provided
		if useConfig {
			mainB64 := base64.StdEncoding.EncodeToString([]byte(main))
			fmt.Fprintf(&b, ` && printf %%s %q | base64 -d > %s && chmod 600 %s`,
				mainB64, talosConfigPath, talosConfigPath)
		}

		// 2) Additional files
		for i, ef := range args.AdditionalFiles {
			content := resolved[2+i].(string)
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
