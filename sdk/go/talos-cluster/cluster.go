// Code generated by Pulumi SDK Generator DO NOT EDIT.
// *** WARNING: Do not edit by hand unless you're certain you know what you are doing! ***

package taloscluster

import (
	"context"
	"reflect"

	"errors"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-talos-cluster/sdk/go/talos-cluster/internal"
)

// Initialize a new Talos cluster:
// - Creates secrets
// - Generates machine configurations for all nodes
type Cluster struct {
	pulumi.ResourceState

	// Client configuration for bootstrapping and applying resources.
	ClientConfiguration ClientConfigurationPtrOutput `pulumi:"clientConfiguration"`
	// TO DO
	GeneratedConfigurations pulumi.StringMapOutput `pulumi:"generatedConfigurations"`
	// TO DO
	Machines ApplyMachinesPtrOutput `pulumi:"machines"`
}

// NewCluster registers a new resource with the given unique name, arguments, and options.
func NewCluster(ctx *pulumi.Context,
	name string, args *ClusterArgs, opts ...pulumi.ResourceOption) (*Cluster, error) {
	if args == nil {
		return nil, errors.New("missing one or more required arguments")
	}

	if args.ClusterEndpoint == nil {
		return nil, errors.New("invalid value for required argument 'ClusterEndpoint'")
	}
	if args.ClusterMachines == nil {
		return nil, errors.New("invalid value for required argument 'ClusterMachines'")
	}
	if args.KubernetesVersion == nil {
		args.KubernetesVersion = pulumi.StringPtr("v1.31.0")
	}
	if args.TalosVersionContract == nil {
		args.TalosVersionContract = pulumi.String("v1.8.2")
	}
	opts = internal.PkgResourceDefaultOpts(opts)
	var resource Cluster
	err := ctx.RegisterRemoteComponentResource("talos-cluster:index:Cluster", name, args, &resource, opts...)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

type clusterArgs struct {
	// Cluster endpoint, the Kubernetes API endpoint accessible by all nodes
	ClusterEndpoint string `pulumi:"clusterEndpoint"`
	// Configuration settings for machines
	ClusterMachines []ClusterMachines `pulumi:"clusterMachines"`
	// Name of the cluster
	ClusterName string `pulumi:"clusterName"`
	// Kubernetes version to install.
	// Default is v1.31.0.
	KubernetesVersion *string `pulumi:"kubernetesVersion"`
	// Version of Talos features used for configuration generation.
	// Do not confuse this with the talosImage property.
	// Used in NewSecrets() and GetConfigurationOutput() resources.
	// This property is immutable to prevent version conflicts across provider updates.
	// See issue: https://github.com/siderolabs/terraform-provider-talos/issues/168
	// The default value is based on gendata.VersionTag, current: v1.8.2.
	TalosVersionContract string `pulumi:"talosVersionContract"`
}

// The set of arguments for constructing a Cluster resource.
type ClusterArgs struct {
	// Cluster endpoint, the Kubernetes API endpoint accessible by all nodes
	ClusterEndpoint pulumi.StringInput
	// Configuration settings for machines
	ClusterMachines ClusterMachinesArrayInput
	// Name of the cluster
	ClusterName string
	// Kubernetes version to install.
	// Default is v1.31.0.
	KubernetesVersion pulumi.StringPtrInput
	// Version of Talos features used for configuration generation.
	// Do not confuse this with the talosImage property.
	// Used in NewSecrets() and GetConfigurationOutput() resources.
	// This property is immutable to prevent version conflicts across provider updates.
	// See issue: https://github.com/siderolabs/terraform-provider-talos/issues/168
	// The default value is based on gendata.VersionTag, current: v1.8.2.
	TalosVersionContract pulumi.StringInput
}

func (ClusterArgs) ElementType() reflect.Type {
	return reflect.TypeOf((*clusterArgs)(nil)).Elem()
}

type ClusterInput interface {
	pulumi.Input

	ToClusterOutput() ClusterOutput
	ToClusterOutputWithContext(ctx context.Context) ClusterOutput
}

func (*Cluster) ElementType() reflect.Type {
	return reflect.TypeOf((**Cluster)(nil)).Elem()
}

func (i *Cluster) ToClusterOutput() ClusterOutput {
	return i.ToClusterOutputWithContext(context.Background())
}

func (i *Cluster) ToClusterOutputWithContext(ctx context.Context) ClusterOutput {
	return pulumi.ToOutputWithContext(ctx, i).(ClusterOutput)
}

// ClusterArrayInput is an input type that accepts ClusterArray and ClusterArrayOutput values.
// You can construct a concrete instance of `ClusterArrayInput` via:
//
//	ClusterArray{ ClusterArgs{...} }
type ClusterArrayInput interface {
	pulumi.Input

	ToClusterArrayOutput() ClusterArrayOutput
	ToClusterArrayOutputWithContext(context.Context) ClusterArrayOutput
}

type ClusterArray []ClusterInput

func (ClusterArray) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*Cluster)(nil)).Elem()
}

func (i ClusterArray) ToClusterArrayOutput() ClusterArrayOutput {
	return i.ToClusterArrayOutputWithContext(context.Background())
}

func (i ClusterArray) ToClusterArrayOutputWithContext(ctx context.Context) ClusterArrayOutput {
	return pulumi.ToOutputWithContext(ctx, i).(ClusterArrayOutput)
}

// ClusterMapInput is an input type that accepts ClusterMap and ClusterMapOutput values.
// You can construct a concrete instance of `ClusterMapInput` via:
//
//	ClusterMap{ "key": ClusterArgs{...} }
type ClusterMapInput interface {
	pulumi.Input

	ToClusterMapOutput() ClusterMapOutput
	ToClusterMapOutputWithContext(context.Context) ClusterMapOutput
}

type ClusterMap map[string]ClusterInput

func (ClusterMap) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*Cluster)(nil)).Elem()
}

func (i ClusterMap) ToClusterMapOutput() ClusterMapOutput {
	return i.ToClusterMapOutputWithContext(context.Background())
}

func (i ClusterMap) ToClusterMapOutputWithContext(ctx context.Context) ClusterMapOutput {
	return pulumi.ToOutputWithContext(ctx, i).(ClusterMapOutput)
}

type ClusterOutput struct{ *pulumi.OutputState }

func (ClusterOutput) ElementType() reflect.Type {
	return reflect.TypeOf((**Cluster)(nil)).Elem()
}

func (o ClusterOutput) ToClusterOutput() ClusterOutput {
	return o
}

func (o ClusterOutput) ToClusterOutputWithContext(ctx context.Context) ClusterOutput {
	return o
}

// Client configuration for bootstrapping and applying resources.
func (o ClusterOutput) ClientConfiguration() ClientConfigurationPtrOutput {
	return o.ApplyT(func(v *Cluster) ClientConfigurationPtrOutput { return v.ClientConfiguration }).(ClientConfigurationPtrOutput)
}

// TO DO
func (o ClusterOutput) GeneratedConfigurations() pulumi.StringMapOutput {
	return o.ApplyT(func(v *Cluster) pulumi.StringMapOutput { return v.GeneratedConfigurations }).(pulumi.StringMapOutput)
}

// TO DO
func (o ClusterOutput) Machines() ApplyMachinesPtrOutput {
	return o.ApplyT(func(v *Cluster) ApplyMachinesPtrOutput { return v.Machines }).(ApplyMachinesPtrOutput)
}

type ClusterArrayOutput struct{ *pulumi.OutputState }

func (ClusterArrayOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*[]*Cluster)(nil)).Elem()
}

func (o ClusterArrayOutput) ToClusterArrayOutput() ClusterArrayOutput {
	return o
}

func (o ClusterArrayOutput) ToClusterArrayOutputWithContext(ctx context.Context) ClusterArrayOutput {
	return o
}

func (o ClusterArrayOutput) Index(i pulumi.IntInput) ClusterOutput {
	return pulumi.All(o, i).ApplyT(func(vs []interface{}) *Cluster {
		return vs[0].([]*Cluster)[vs[1].(int)]
	}).(ClusterOutput)
}

type ClusterMapOutput struct{ *pulumi.OutputState }

func (ClusterMapOutput) ElementType() reflect.Type {
	return reflect.TypeOf((*map[string]*Cluster)(nil)).Elem()
}

func (o ClusterMapOutput) ToClusterMapOutput() ClusterMapOutput {
	return o
}

func (o ClusterMapOutput) ToClusterMapOutputWithContext(ctx context.Context) ClusterMapOutput {
	return o
}

func (o ClusterMapOutput) MapIndex(k pulumi.StringInput) ClusterOutput {
	return pulumi.All(o, k).ApplyT(func(vs []interface{}) *Cluster {
		return vs[0].(map[string]*Cluster)[vs[1].(string)]
	}).(ClusterOutput)
}

func init() {
	pulumi.RegisterInputType(reflect.TypeOf((*ClusterInput)(nil)).Elem(), &Cluster{})
	pulumi.RegisterInputType(reflect.TypeOf((*ClusterArrayInput)(nil)).Elem(), ClusterArray{})
	pulumi.RegisterInputType(reflect.TypeOf((*ClusterMapInput)(nil)).Elem(), ClusterMap{})
	pulumi.RegisterOutputType(ClusterOutput{})
	pulumi.RegisterOutputType(ClusterArrayOutput{})
	pulumi.RegisterOutputType(ClusterMapOutput{})
}
