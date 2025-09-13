package cluster

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

// Load reads a cluster specification from the given path and unmarshals it into a Cluster.
func Load(path string) (*Cluster, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Cluster
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
