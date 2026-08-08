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
	NodeIP       pulumi.StringInput
	TalosConfig  pulumi.StringInput // Wired-in configuration
}

// Args groups arguments used to execute a talosctl command.
type Args struct {
	TalosConfig      pulumi.StringInput
	PrepareDeps      []pulumi.Resource
	Dir              string
	CommandArgs      pulumi.StringInput
	RetryCount       int
	Environment      pulumi.StringMap
	Triggers         pulumi.Array
	UpdateOnChange   bool
	AdditionalFiles  []ExtraFile
	TryInsecureFirst bool
	// Logging controls what the Pulumi engine records from the command execution.
	// Defaults:
	//   - RunCommand: local.LoggingNone (keeps current behavior of streaming both stdout/stderr)
	//   - RunGetCommand: local.LoggingStderr (previous default)
	Logging local.Logging
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

// WithTalosConfig wires a default talosconfig into the Talosctl instance.
func (t *Talosctl) WithTalosConfig(config pulumi.StringInput) *Talosctl {
	t.TalosConfig = config

	return t
}

// WithNodeIP adds `-n` and `-e` flags for the provided node IP address.
func (t *Talosctl) WithNodeIP(ip string) *Talosctl {
	t.BasicCommand = fmt.Sprintf("%s -n %s -e %s", t.BasicCommand, ip, ip)

	return t
}

// WithNodeIPInput adds node flags without requiring the IP to be known during preview.
func (t *Talosctl) WithNodeIPInput(ip pulumi.StringInput) *Talosctl {
	t.NodeIP = ip

	return t
}

// RunCommand executes a talosctl command as a Pulumi resource.
func (t *Talosctl) RunCommand(
	ctx *pulumi.Context,
	name string,
	a *Args,
	opts ...pulumi.ResourceOption,
) (*local.Command, error) {
	logging := a.Logging
	if logging == "" {
		logging = local.LoggingNone
	}

	createGated, env, err := t.prepareAndGate(ctx, a)
	if err != nil {
		return nil, err
	}

	commandArgs := &local.CommandArgs{
		Create:      createGated,
		Dir:         pulumi.String(a.Dir),
		Interpreter: pulumi.ToStringArray(interpreter),
		Environment: env,
		Triggers:    a.Triggers,
		Logging:     logging,
	}
	if a.UpdateOnChange {
		commandArgs.Update = createGated.ToStringPtrOutput()
	}

	main, err := local.NewCommand(ctx, name, commandArgs, opts...)
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

// RunGetCommand executes a talosctl command as an invoke and returns its standard output.
func (t *Talosctl) RunGetCommand(
	ctx *pulumi.Context,
	a *Args,
) (pulumi.StringOutput, error) {
	logging := a.Logging
	if logging == "" {
		logging = local.LoggingStderr
	}

	createGated, env, err := t.prepareAndGate(ctx, a)
	if err != nil {
		// Return a zero output with error
		return pulumi.StringOutput{}, err
	}

	command := createGated.ApplyT(func(cmd string) string {
		if strings.TrimSpace(cmd) == "" {
			return ""
		}

		return fmt.Sprintf("trap 'rm -rf %q' EXIT\n%s", a.Dir, cmd)
	}).(pulumi.StringOutput)

	run := local.RunOutput(ctx, local.RunOutputArgs{
		Command:     command,
		Dir:         pulumi.String(a.Dir),
		Interpreter: pulumi.ToStringArray(interpreter),
		// Only log stderr since stdout can keep a sensitive data.
		Logging:     logging,
		Environment: env,
	})

	return run.Stdout(), nil
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

	if t.TalosConfig != nil {
		args.AdditionalFiles = append(args.AdditionalFiles, ExtraFile{Name: talosctlConfigName, Content: t.TalosConfig})
	}

	// Prepare: write talosctl.yaml + additional files
	prepared := pulumi.Unsecret(t.prepareAll(ctx, args)).(pulumi.BoolOutput)

	basicCommand := pulumi.StringInput(pulumi.String(t.BasicCommand))
	if t.NodeIP != nil {
		basicCommand = pulumi.Unsecret(
			pulumi.Sprintf("%s -n %s -e %s", t.BasicCommand, t.NodeIP, t.NodeIP),
		).(pulumi.StringOutput)
	}

	createGated = pulumi.All(prepared, args.CommandArgs, basicCommand).
		ApplyT(func(v []any) string {
			if !v[0].(bool) {
				return ""
			}

			cmdArgs := v[1].(string)
			resolvedBasicCommand := v[2].(string)
			hasConfig := t.TalosConfig != nil

			// Build base commands (no retries yet).
			secure := fmt.Sprintf("%s %s --talosconfig %s", resolvedBasicCommand, cmdArgs, talosctlConfigName)
			if !hasConfig {
				secure = fmt.Sprintf("%s %s", resolvedBasicCommand, cmdArgs)
			}
			insecure := fmt.Sprintf("%s %s --insecure", resolvedBasicCommand, cmdArgs)

			// Select auth pipeline.
			selected := secure
			if args.TryInsecureFirst && hasConfig {
				selected = fmt.Sprintf("(%s) || (%s)", insecure, secure)
			}

			// Wrap with retry if requested.
			if args.RetryCount <= 0 {
				return selected
			}

			return withBashRetry(selected, fmt.Sprint(args.RetryCount+1))
		}).(pulumi.StringOutput)

	return createGated, env, nil
}

// prepareAll builds one shell that writes talosctl.yaml (if provided) and any AdditionalFiles.
// It uses `local.RunOutput` as the prepare step, returning a BoolOutput.
func (t *Talosctl) prepareAll(ctx *pulumi.Context, args *Args) pulumi.BoolOutput {
	// Collect all file contents (including appended talosconfig) in order.
	inputs := make([]any, len(args.AdditionalFiles))
	for i, f := range args.AdditionalFiles {
		inputs[i] = f.Content
	}

	// Build the full shell command as a StringOutput from resolved inputs
	cmd := pulumi.All(inputs...).ApplyT(func(resolved []any) string {
		var b strings.Builder
		// Ensure TALOS_HOME exists, private perms
		fmt.Fprintf(&b, `mkdir -p %s && umask 077`, args.Dir)

		// 2) Additional files
		for i, ef := range args.AdditionalFiles {
			content := resolved[i].(string)
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
