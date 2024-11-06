// Code generated by Pulumi SDK Generator DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package taloscluster

import (
	"context"
	"reflect"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-talos-cluster/sdk/go/talos-cluster/internal"
)

var _ = internal.GetEnvOrDefault

type ApplyMachines struct {
	// Configuration settings for machines to apply.
	// This can be retrieved from the cluster resource.
	Configuration string `pulumi:"configuration"`
	// ID or name of the machine.
	MachineId string `pulumi:"machineId"`
	// The IP address of the node where configuration will be applied.
	Node string `pulumi:"node"`
	// User-provided machine configuration to apply.
	// This can be retrieved from the cluster resource.
	UserConfigPatches *string `pulumi:"userConfigPatches"`
}

// ApplyMachinesInput is an input type that accepts ApplyMachinesArgs and ApplyMachinesOutput values.
// You can construct a concrete instance of `ApplyMachinesInput` via:
//
//	ApplyMachinesArgs{...}
type ApplyMachinesInput interface {
	pulumi.Input

	ToApplyMachinesOutput() ApplyMachinesOutput
	ToApplyMachinesOutputWithContext(context.Context) ApplyMachinesOutput
}

type ApplyMachinesArgs struct {
	// Configuration settings for machines to apply.
	// This can be retrieved from the cluster resource.
	Configuration pulumi.StringInput `pulumi:"configuration"`
	// ID or name of the machine.
	MachineId pulumi.StringInput `pulumi:"machineId"`
	// The IP address of the node where configuration will be applied.
	Node pulumi.StringInput `pulumi:"node"`
	// User-provided machine configuration to apply.
	// This can be retrieved from the cluster resource.
	UserConfigPatches pulumi.StringPtrInput `pulumi:"userConfigPatches"`
}

func (ApplyMachinesArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*ApplyMachines)(nil)).Elem()
}

func (i ApplyMachinesArgs) ToApplyMachinesOutput() ApplyMachinesOutput {
	return i.ToApplyMachinesOutputWithContext(context.Background())
}

func (i ApplyMachinesArgs) ToApplyMachinesOutputWithContext(ctx context.Context) ApplyMachinesOutput {
	return pulumi.ToOutputWithContext(ctx, i).(ApplyMachinesOutput)
}

// ApplyMachinesArrayInput is an input type that accepts ApplyMachinesArray and ApplyMachinesArrayOutput values.
// You can construct a concrete instance of `ApplyMachinesArrayInput` via:
//
//	ApplyMachinesArray{ ApplyMachinesArgs{...} }
type ApplyMachinesArrayInput interface {
	pulumi.Input

	ToApplyMachinesArrayOutput() ApplyMachinesArrayOutput
	ToApplyMachinesArrayOutputWithContext(context.Context) ApplyMachinesArrayOutput
}

type ApplyMachinesArray []ApplyMachinesInput

func (ApplyMachinesArray) ElementType() reflect.Type {
	return reflect.TypeOf((*[]ApplyMachines)(nil)).Elem()
}

func (i ApplyMachinesArray) ToApplyMachinesArrayOutput() ApplyMachinesArrayOutput {
	return i.ToApplyMachinesArrayOutputWithContext(context.Background())
}

func (i ApplyMachinesArray) ToApplyMachinesArrayOutputWithContext(ctx context.Context) ApplyMachinesArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(ApplyMachinesArrayOutput)
}

type ApplyMachinesOutput struct{ *pulumi.OutputState }

func (ApplyMachinesOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*ApplyMachines)(nil)).Elem()
}

func (o ApplyMachinesOutput) ToApplyMachinesOutput() ApplyMachinesOutput {
	return o
}

func (o ApplyMachinesOutput) ToApplyMachinesOutputWithContext(ctx context.Context) ApplyMachinesOutput {
	return o
}

// Configuration settings for machines to apply.
// This can be retrieved from the cluster resource.
func (o ApplyMachinesOutput) Configuration() pulumi.StringOutput {
	return o.ApplyT(func(v ApplyMachines) string { return v.Configuration }).(pulumi.StringOutput)
}

// ID or name of the machine.
func (o ApplyMachinesOutput) MachineId() pulumi.StringOutput {
	return o.ApplyT(func(v ApplyMachines) string { return v.MachineId }).(pulumi.StringOutput)
}

// The IP address of the node where configuration will be applied.
func (o ApplyMachinesOutput) Node() pulumi.StringOutput {
	return o.ApplyT(func(v ApplyMachines) string { return v.Node }).(pulumi.StringOutput)
}

// User-provided machine configuration to apply.
// This can be retrieved from the cluster resource.
func (o ApplyMachinesOutput) UserConfigPatches() pulumi.StringPtrOutput {
	return o.ApplyT(func(v ApplyMachines) *string { return v.UserConfigPatches }).(pulumi.StringPtrOutput)
}

type ApplyMachinesArrayOutput struct{ *pulumi.OutputState }

func (ApplyMachinesArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]ApplyMachines)(nil)).Elem()
}

func (o ApplyMachinesArrayOutput) ToApplyMachinesArrayOutput() ApplyMachinesArrayOutput {
	return o
}

func (o ApplyMachinesArrayOutput) ToApplyMachinesArrayOutputWithContext(ctx context.Context) ApplyMachinesArrayOutput {
	return o
}

func (o ApplyMachinesArrayOutput) Index(i pulumi.IntInput) ApplyMachinesOutput {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) ApplyMachines {
		return vs[0].([]ApplyMachines)[vs[1].(int)]
	}).(ApplyMachinesOutput)
}

type ApplyMachinesByType struct {
	Controlplane []ApplyMachines `pulumi:"controlplane"`
	Init         ApplyMachines   `pulumi:"init"`
	Worker       []ApplyMachines `pulumi:"worker"`
}

// ApplyMachinesByTypeInput is an input type that accepts ApplyMachinesByTypeArgs and ApplyMachinesByTypeOutput values.
// You can construct a concrete instance of `ApplyMachinesByTypeInput` via:
//
//	ApplyMachinesByTypeArgs{...}
type ApplyMachinesByTypeInput interface {
	pulumi.Input

	ToApplyMachinesByTypeOutput() ApplyMachinesByTypeOutput
	ToApplyMachinesByTypeOutputWithContext(context.Context) ApplyMachinesByTypeOutput
}

type ApplyMachinesByTypeArgs struct {
	Controlplane ApplyMachinesArrayInput `pulumi:"controlplane"`
	Init         ApplyMachinesInput      `pulumi:"init"`
	Worker       ApplyMachinesArrayInput `pulumi:"worker"`
}

func (ApplyMachinesByTypeArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*ApplyMachinesByType)(nil)).Elem()
}

func (i ApplyMachinesByTypeArgs) ToApplyMachinesByTypeOutput() ApplyMachinesByTypeOutput {
	return i.ToApplyMachinesByTypeOutputWithContext(context.Background())
}

func (i ApplyMachinesByTypeArgs) ToApplyMachinesByTypeOutputWithContext(ctx context.Context) ApplyMachinesByTypeOutput {
	return pulumi.ToOutputWithContext(ctx, i).(ApplyMachinesByTypeOutput)
}

type ApplyMachinesByTypeOutput struct{ *pulumi.OutputState }

func (ApplyMachinesByTypeOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*ApplyMachinesByType)(nil)).Elem()
}

func (o ApplyMachinesByTypeOutput) ToApplyMachinesByTypeOutput() ApplyMachinesByTypeOutput {
	return o
}

func (o ApplyMachinesByTypeOutput) ToApplyMachinesByTypeOutputWithContext(ctx context.Context) ApplyMachinesByTypeOutput {
	return o
}

func (o ApplyMachinesByTypeOutput) Controlplane() ApplyMachinesArrayOutput {
	return o.ApplyT(func(v ApplyMachinesByType) []ApplyMachines { return v.Controlplane }).(ApplyMachinesArrayOutput)
}

func (o ApplyMachinesByTypeOutput) Init() ApplyMachinesOutput {
	return o.ApplyT(func(v ApplyMachinesByType) ApplyMachines { return v.Init }).(ApplyMachinesOutput)
}

func (o ApplyMachinesByTypeOutput) Worker() ApplyMachinesArrayOutput {
	return o.ApplyT(func(v ApplyMachinesByType) []ApplyMachines { return v.Worker }).(ApplyMachinesArrayOutput)
}

type ClientConfiguration struct {
	// The Certificate Authority (CA) certificate used to verify connections to the Talos API server.
	CaCertificate *string `pulumi:"caCertificate"`
	// The client certificate used to authenticate to the Talos API server.
	ClientCertificate *string `pulumi:"clientCertificate"`
	// The private key for the client certificate, used for authenticating the client to the Talos API server.
	ClientKey *string `pulumi:"clientKey"`
}

// ClientConfigurationInput is an input type that accepts ClientConfigurationArgs and ClientConfigurationOutput values.
// You can construct a concrete instance of `ClientConfigurationInput` via:
//
//	ClientConfigurationArgs{...}
type ClientConfigurationInput interface {
	pulumi.Input

	ToClientConfigurationOutput() ClientConfigurationOutput
	ToClientConfigurationOutputWithContext(context.Context) ClientConfigurationOutput
}

type ClientConfigurationArgs struct {
	// The Certificate Authority (CA) certificate used to verify connections to the Talos API server.
	CaCertificate pulumi.StringPtrInput `pulumi:"caCertificate"`
	// The client certificate used to authenticate to the Talos API server.
	ClientCertificate pulumi.StringPtrInput `pulumi:"clientCertificate"`
	// The private key for the client certificate, used for authenticating the client to the Talos API server.
	ClientKey pulumi.StringPtrInput `pulumi:"clientKey"`
}

func (ClientConfigurationArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*ClientConfiguration)(nil)).Elem()
}

func (i ClientConfigurationArgs) ToClientConfigurationOutput() ClientConfigurationOutput {
	return i.ToClientConfigurationOutputWithContext(context.Background())
}

func (i ClientConfigurationArgs) ToClientConfigurationOutputWithContext(ctx context.Context) ClientConfigurationOutput {
	return pulumi.ToOutputWithContext(ctx, i).(ClientConfigurationOutput)
}

type ClientConfigurationOutput struct{ *pulumi.OutputState }

func (ClientConfigurationOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*ClientConfiguration)(nil)).Elem()
}

func (o ClientConfigurationOutput) ToClientConfigurationOutput() ClientConfigurationOutput {
	return o
}

func (o ClientConfigurationOutput) ToClientConfigurationOutputWithContext(ctx context.Context) ClientConfigurationOutput {
	return o
}

// The Certificate Authority (CA) certificate used to verify connections to the Talos API server.
func (o ClientConfigurationOutput) CaCertificate() pulumi.StringPtrOutput {
	return o.ApplyT(func(v ClientConfiguration) *string { return v.CaCertificate }).(pulumi.StringPtrOutput)
}

// The client certificate used to authenticate to the Talos API server.
func (o ClientConfigurationOutput) ClientCertificate() pulumi.StringPtrOutput {
	return o.ApplyT(func(v ClientConfiguration) *string { return v.ClientCertificate }).(pulumi.StringPtrOutput)
}

// The private key for the client certificate, used for authenticating the client to the Talos API server.
func (o ClientConfigurationOutput) ClientKey() pulumi.StringPtrOutput {
	return o.ApplyT(func(v ClientConfiguration) *string { return v.ClientKey }).(pulumi.StringPtrOutput)
}

type ClientConfigurationPtrOutput struct{ *pulumi.OutputState }

func (ClientConfigurationPtrOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**ClientConfiguration)(nil)).Elem()
}

func (o ClientConfigurationPtrOutput) ToClientConfigurationPtrOutput() ClientConfigurationPtrOutput {
	return o
}

func (o ClientConfigurationPtrOutput) ToClientConfigurationPtrOutputWithContext(ctx context.Context) ClientConfigurationPtrOutput {
	return o
}

func (o ClientConfigurationPtrOutput) Elem() ClientConfigurationOutput {
	return o.ApplyT(func(v *ClientConfiguration) ClientConfiguration {
		if v != nil {
			return *v
		}
		var ret ClientConfiguration
		return ret
	}).(ClientConfigurationOutput)
}

// The Certificate Authority (CA) certificate used to verify connections to the Talos API server.
func (o ClientConfigurationPtrOutput) CaCertificate() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *ClientConfiguration) *string {
		if v == nil {
			return nil
		}
		return v.CaCertificate
	}).(pulumi.StringPtrOutput)
}

// The client certificate used to authenticate to the Talos API server.
func (o ClientConfigurationPtrOutput) ClientCertificate() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *ClientConfiguration) *string {
		if v == nil {
			return nil
		}
		return v.ClientCertificate
	}).(pulumi.StringPtrOutput)
}

// The private key for the client certificate, used for authenticating the client to the Talos API server.
func (o ClientConfigurationPtrOutput) ClientKey() pulumi.StringPtrOutput {
	return o.ApplyT(func(v *ClientConfiguration) *string {
		if v == nil {
			return nil
		}
		return v.ClientKey
	}).(pulumi.StringPtrOutput)
}

type ClusterMachines struct {
	// User-provided machine configuration to apply.
	// Must be a valid YAML string.
	// For structure, see https://www.talos.dev/latest/reference/configuration/v1alpha1/config/
	ConfigPatches *string `pulumi:"configPatches"`
	// Kubernetes version to install.
	// Default is v1.31.0.
	KubernetesVersion *string `pulumi:"kubernetesVersion"`
	// ID or name of the machine.
	MachineId string `pulumi:"machineId"`
	// Type of the machine.
	MachineType MachineTypes `pulumi:"machineType"`
	// Talos OS installation image.
	// Used in the `install` configuration and set via CLI.
	// The default is generated based on the Talos machinery version, current: ghcr.io/siderolabs/installer:v1.8.2.
	TalosImage *string `pulumi:"talosImage"`
}

// Defaults sets the appropriate defaults for ClusterMachines
func (val *ClusterMachines) Defaults() *ClusterMachines {
	if val == nil {
		return nil
	}
	tmp := *val
	if tmp.KubernetesVersion == nil {
		kubernetesVersion_ := "v1.31.0"
		tmp.KubernetesVersion = &kubernetesVersion_
	}
	if tmp.TalosImage == nil {
		talosImage_ := "ghcr.io/siderolabs/installer:v1.8.2"
		tmp.TalosImage = &talosImage_
	}
	return &tmp
}

// ClusterMachinesInput is an input type that accepts ClusterMachinesArgs and ClusterMachinesOutput values.
// You can construct a concrete instance of `ClusterMachinesInput` via:
//
//	ClusterMachinesArgs{...}
type ClusterMachinesInput interface {
	pulumi.Input

	ToClusterMachinesOutput() ClusterMachinesOutput
	ToClusterMachinesOutputWithContext(context.Context) ClusterMachinesOutput
}

type ClusterMachinesArgs struct {
	// User-provided machine configuration to apply.
	// Must be a valid YAML string.
	// For structure, see https://www.talos.dev/latest/reference/configuration/v1alpha1/config/
	ConfigPatches pulumi.StringPtrInput `pulumi:"configPatches"`
	// Kubernetes version to install.
	// Default is v1.31.0.
	KubernetesVersion pulumi.StringPtrInput `pulumi:"kubernetesVersion"`
	// ID or name of the machine.
	MachineId string `pulumi:"machineId"`
	// Type of the machine.
	MachineType MachineTypes `pulumi:"machineType"`
	// Talos OS installation image.
	// Used in the `install` configuration and set via CLI.
	// The default is generated based on the Talos machinery version, current: ghcr.io/siderolabs/installer:v1.8.2.
	TalosImage pulumi.StringPtrInput `pulumi:"talosImage"`
}

// Defaults sets the appropriate defaults for ClusterMachinesArgs
func (val *ClusterMachinesArgs) Defaults() *ClusterMachinesArgs {
	if val == nil {
		return nil
	}
	tmp := *val
	if tmp.KubernetesVersion == nil {
		tmp.KubernetesVersion = pulumi.StringPtr("v1.31.0")
	}
	if tmp.TalosImage == nil {
		tmp.TalosImage = pulumi.StringPtr("ghcr.io/siderolabs/installer:v1.8.2")
	}
	return &tmp
}
func (ClusterMachinesArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*ClusterMachines)(nil)).Elem()
}

func (i ClusterMachinesArgs) ToClusterMachinesOutput() ClusterMachinesOutput {
	return i.ToClusterMachinesOutputWithContext(context.Background())
}

func (i ClusterMachinesArgs) ToClusterMachinesOutputWithContext(ctx context.Context) ClusterMachinesOutput {
	return pulumi.ToOutputWithContext(ctx, i).(ClusterMachinesOutput)
}

// ClusterMachinesArrayInput is an input type that accepts ClusterMachinesArray and ClusterMachinesArrayOutput values.
// You can construct a concrete instance of `ClusterMachinesArrayInput` via:
//
//	ClusterMachinesArray{ ClusterMachinesArgs{...} }
type ClusterMachinesArrayInput interface {
	pulumi.Input

	ToClusterMachinesArrayOutput() ClusterMachinesArrayOutput
	ToClusterMachinesArrayOutputWithContext(context.Context) ClusterMachinesArrayOutput
}

type ClusterMachinesArray []ClusterMachinesInput

func (ClusterMachinesArray) ElementType() reflect.Type {
	return reflect.TypeOf((*[]ClusterMachines)(nil)).Elem()
}

func (i ClusterMachinesArray) ToClusterMachinesArrayOutput() ClusterMachinesArrayOutput {
	return i.ToClusterMachinesArrayOutputWithContext(context.Background())
}

func (i ClusterMachinesArray) ToClusterMachinesArrayOutputWithContext(ctx context.Context) ClusterMachinesArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(ClusterMachinesArrayOutput)
}

type ClusterMachinesOutput struct{ *pulumi.OutputState }

func (ClusterMachinesOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*ClusterMachines)(nil)).Elem()
}

func (o ClusterMachinesOutput) ToClusterMachinesOutput() ClusterMachinesOutput {
	return o
}

func (o ClusterMachinesOutput) ToClusterMachinesOutputWithContext(ctx context.Context) ClusterMachinesOutput {
	return o
}

// User-provided machine configuration to apply.
// Must be a valid YAML string.
// For structure, see https://www.talos.dev/latest/reference/configuration/v1alpha1/config/
func (o ClusterMachinesOutput) ConfigPatches() pulumi.StringPtrOutput {
	return o.ApplyT(func(v ClusterMachines) *string { return v.ConfigPatches }).(pulumi.StringPtrOutput)
}

// Kubernetes version to install.
// Default is v1.31.0.
func (o ClusterMachinesOutput) KubernetesVersion() pulumi.StringPtrOutput {
	return o.ApplyT(func(v ClusterMachines) *string { return v.KubernetesVersion }).(pulumi.StringPtrOutput)
}

// ID or name of the machine.
func (o ClusterMachinesOutput) MachineId() pulumi.StringOutput {
	return o.ApplyT(func(v ClusterMachines) string { return v.MachineId }).(pulumi.StringOutput)
}

// Type of the machine.
func (o ClusterMachinesOutput) MachineType() MachineTypesOutput {
	return o.ApplyT(func(v ClusterMachines) MachineTypes { return v.MachineType }).(MachineTypesOutput)
}

// Talos OS installation image.
// Used in the `install` configuration and set via CLI.
// The default is generated based on the Talos machinery version, current: ghcr.io/siderolabs/installer:v1.8.2.
func (o ClusterMachinesOutput) TalosImage() pulumi.StringPtrOutput {
	return o.ApplyT(func(v ClusterMachines) *string { return v.TalosImage }).(pulumi.StringPtrOutput)
}

type ClusterMachinesArrayOutput struct{ *pulumi.OutputState }

func (ClusterMachinesArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]ClusterMachines)(nil)).Elem()
}

func (o ClusterMachinesArrayOutput) ToClusterMachinesArrayOutput() ClusterMachinesArrayOutput {
	return o
}

func (o ClusterMachinesArrayOutput) ToClusterMachinesArrayOutputWithContext(ctx context.Context) ClusterMachinesArrayOutput {
	return o
}

func (o ClusterMachinesArrayOutput) Index(i pulumi.IntInput) ClusterMachinesOutput {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) ClusterMachines {
		return vs[0].([]ClusterMachines)[vs[1].(int)]
	}).(ClusterMachinesOutput)
}

func init() {
	pulumi.RegisterInputType(reflect.TypeOf((*ApplyMachinesInput)(nil)).Elem(), ApplyMachinesArgs{})
	pulumi.RegisterInputType(reflect.TypeOf((*ApplyMachinesArrayInput)(nil)).Elem(), ApplyMachinesArray{})
	pulumi.RegisterInputType(reflect.TypeOf((*ApplyMachinesByTypeInput)(nil)).Elem(), ApplyMachinesByTypeArgs{})
	pulumi.RegisterInputType(reflect.TypeOf((*ClientConfigurationInput)(nil)).Elem(), ClientConfigurationArgs{})
	pulumi.RegisterInputType(reflect.TypeOf((*ClusterMachinesInput)(nil)).Elem(), ClusterMachinesArgs{})
	pulumi.RegisterInputType(reflect.TypeOf((*ClusterMachinesArrayInput)(nil)).Elem(), ClusterMachinesArray{})
	pulumi.RegisterOutputType(ApplyMachinesOutput{})
	pulumi.RegisterOutputType(ApplyMachinesArrayOutput{})
	pulumi.RegisterOutputType(ApplyMachinesByTypeOutput{})
	pulumi.RegisterOutputType(ClientConfigurationOutput{})
	pulumi.RegisterOutputType(ClientConfigurationPtrOutput{})
	pulumi.RegisterOutputType(ClusterMachinesOutput{})
	pulumi.RegisterOutputType(ClusterMachinesArrayOutput{})
}
