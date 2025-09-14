package cluster

type Cluster struct {
	Name              string     `yaml:"name"`
	PrivateNetwork    string     `yaml:"privateNetwork"`
	PrivateSubnetwork string     `yaml:"privateSubnetwork"`
	KubernetesVersion string     `yaml:"kubernetesVersion"`
	Machines          []*Machine `yaml:"machines"`
}

type Machine struct {
	ID                  string   `yaml:"id"`
	Type                string   `yaml:"type"`
	ServerType          string   `yaml:"serverType"`
	Platform            string   `yaml:"platform"`
	TalosInitialVersion string   `yaml:"talosInitialVersion"`
	TalosImage          string   `yaml:"talosImage"`
	PrivateIP           string   `yaml:"privateIP"`
	Datacenter          string   `yaml:"datacenter"`
	ConfigPatches       []string `yaml:"configPatches"`
	Userdata            string   `yaml:"userdata"`
}
