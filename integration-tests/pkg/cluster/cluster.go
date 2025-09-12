package cluster

type Cluster struct {
	Name              string
	PrivateNetwork    string
	PrivateSubnetwork string
	KubernetesVersion string
	Machines          []*Machine
}

type Machine struct {
	ID                  string
	Type                string
	ServerType          string
	Platform            string
	TalosInitialVersion string
	TalosImage          string
	PrivateIP           string
	Datacenter          string
}
