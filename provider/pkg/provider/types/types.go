package types

import (
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
)

type ClusterMachine struct {
	MachineID     string                `pulumi:"machineId"`
	MachineType   string                `pulumi:"machineType"`
	NodeIP        pulumi.StringPtrInput `pulumi:"nodeIp"`
	TalosImage    pulumi.StringPtrInput `pulumi:"talosImage"`
	ConfigPatches pulumi.StringPtrInput `pulumi:"configPatches"`
}

func (m *ClusterMachine) ToMachineInfoMap(clusterEndpoint pulumi.StringInput, k8sVer pulumi.StringInput, config pulumi.StringOutput) *pulumi.Map {
	return &pulumi.Map{
		MachineIDKey:         pulumi.String(m.MachineID),
		UserConfigPatchesKey: m.ConfigPatches.ToStringPtrOutput().Elem(),
		KubernetesVersionKey: k8sVer.ToStringPtrOutput().Elem(),
		NodeIPKey:            m.NodeIP.ToStringPtrOutput().Elem(),
		TalosImageKey:        m.TalosImage.ToStringPtrOutput().Elem(),
		ClusterEnpointKey:    clusterEndpoint,
		ConfigurationKey:     config,
	}
}

type MachineInfo struct {
	MachineID         string `pulumi:"machineId"`
	NodeIP            string `pulumi:"nodeIp"`
	ClusterEnpoint    string `pulumi:"clusterEndpoint"`
	UserConfigPatches string `pulumi:"userConfigPatches"`
	TalosImage        string `pulumi:"talosImage"`
	KubernetesVersion string `pulumi:"kubernetesVersion"`
	Configuration     string `pulumi:"configuration"`
}

func ParseMachineInfo(m map[string]any) *MachineInfo {
	return &MachineInfo{
		MachineID:         m[MachineIDKey].(string),
		NodeIP:            m[NodeIPKey].(string),
		ClusterEnpoint:    m[ClusterEnpointKey].(string),
		TalosImage:        m[TalosImageKey].(string),
		KubernetesVersion: m[KubernetesVersionKey].(string),
		UserConfigPatches: m[UserConfigPatchesKey].(string),
		Configuration:     m[ConfigurationKey].(string),
	}
}
