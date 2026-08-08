// revive:disable:var-naming
package types

// revive:enable:var-naming

import (
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	MachineIDKey         = "machineId"
	NodeIPKey            = "nodeIp"
	TalosImageKey        = "talosImage"
	UserConfigPatchesKey = "userConfigPatches"
	ConfigurationKey     = "configuration"
	KubernetesVersionKey = "kubernetesVersion"
	ClusterEnpointKey    = "clusterEndpoint"
	KubeconfigKey        = "kubeconfig"
	TalosconfigKey       = "talosconfig"
)

type Cluster struct {
	ClusterName          string             `pulumi:"clusterName"`
	TalosVersionContract pulumi.StringInput `pulumi:"talosVersionContract"`
	ClusterEndpoint      pulumi.StringInput `pulumi:"clusterEndpoint"`
	KubernetesVersion    pulumi.StringInput `pulumi:"kubernetesVersion"`

	ClusterMachines []*ClusterMachine `pulumi:"clusterMachines"`
}

type ClusterMachine struct {
	MachineID     string                  `pulumi:"machineId"`
	MachineType   string                  `pulumi:"machineType"`
	NodeIP        pulumi.StringPtrInput   `pulumi:"nodeIp"`
	TalosImage    pulumi.StringPtrInput   `pulumi:"talosImage"`
	ConfigPatches pulumi.StringArrayInput `pulumi:"configPatches"`
}

type ClientConfigurationArgs struct {
	CaCertificate     pulumi.StringInput `pulumi:"caCertificate"`
	ClientKey         pulumi.StringInput `pulumi:"clientKey"`
	ClientCertificate pulumi.StringInput `pulumi:"clientCertificate"`
}

func (m *ClusterMachine) ToMachineInfoMap(clusterEndpoint pulumi.StringInput, k8sVer pulumi.StringInput, config pulumi.StringOutput) *pulumi.Map {
	return &pulumi.Map{
		MachineIDKey: pulumi.String(m.MachineID),
		UserConfigPatchesKey: m.ConfigPatches.ToStringArrayOutput().
			ApplyT(func(arr []string) string {
				return strings.Join(arr, "\n---\n")
			}).(pulumi.StringOutput),
		KubernetesVersionKey: k8sVer.ToStringPtrOutput().Elem(),
		NodeIPKey:            m.NodeIP.ToStringPtrOutput().Elem(),
		TalosImageKey:        m.TalosImage.ToStringPtrOutput().Elem(),
		ClusterEnpointKey:    clusterEndpoint,
		ConfigurationKey:     config,
	}
}

type MachineInfo struct {
	MachineID         string             `pulumi:"machineId"`
	NodeIP            pulumi.StringInput `pulumi:"nodeIp"`
	ClusterEnpoint    pulumi.StringInput `pulumi:"clusterEndpoint"`
	UserConfigPatches pulumi.StringInput `pulumi:"userConfigPatches"`
	TalosImage        pulumi.StringInput `pulumi:"talosImage"`
	KubernetesVersion pulumi.StringInput `pulumi:"kubernetesVersion"`
	Configuration     pulumi.StringInput `pulumi:"configuration"`
}

func ParseMachineInfo(m map[string]any) *MachineInfo {
	return &MachineInfo{
		MachineID:         m[MachineIDKey].(string),
		NodeIP:            pulumi.String(m[NodeIPKey].(string)),
		ClusterEnpoint:    pulumi.String(m[ClusterEnpointKey].(string)),
		TalosImage:        pulumi.String(m[TalosImageKey].(string)),
		KubernetesVersion: pulumi.String(m[KubernetesVersionKey].(string)),
		UserConfigPatches: pulumi.String(m[UserConfigPatchesKey].(string)),
		Configuration:     pulumi.String(m[ConfigurationKey].(string)),
	}
}
