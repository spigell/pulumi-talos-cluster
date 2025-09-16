package cluster

type Cluster struct {
	Name              string     `yaml:"name"`
	PrivateNetwork    string     `yaml:"privateNetwork"`
	PrivateSubnetwork string     `yaml:"privateSubnetwork"`
	KubernetesVersion string     `yaml:"kubernetesVersion"`
	SkipInitApply     bool       `yaml:"skipInitApply"`
	Machines          []*Machine `yaml:"machines"`
}

type Machine struct {
	ID                     string         `yaml:"id"`
	Type                   string         `yaml:"type"`
	Platform               string         `yaml:"platform"`
	TalosInitialVersion    string         `yaml:"talosInitialVersion"`
	TalosImage             string         `yaml:"talosImage"`
	PrivateIP              string         `yaml:"privateIP"`
	ConfigPatches          []string       `yaml:"configPatches"`
	ApplyConfigViaUserdata bool           `yaml:"apply-config-via-userdata"`
	Hcloud                 *HcloudMachine `yaml:"hcloud"`
}

type HcloudMachine struct {
	ServerType string `yaml:"serverType"`
	Datacenter string `yaml:"datacenter"`
}
