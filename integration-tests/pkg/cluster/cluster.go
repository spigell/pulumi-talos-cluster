package cluster

type Cluster struct {
	Name              string
	PrivateNetwork    string
	PrivateSubnetwork string
	KubernetesVersion string
	TalosImage        string
	Machines          []*Machine
}

type Machine struct {
	ID         string
	Type       string
	ServerType string
	TalosImage string
	PrivateIP  string
	Datacenter string
}
