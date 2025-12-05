package cluster

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Load reads a cluster specification from the given path and unmarshals it into a Cluster.
func Load(path string) (*Cluster, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	raw, err := parseToMap(data)
	if err != nil {
		return nil, err
	}
	if err := validateCluster(raw); err != nil {
		return nil, err
	}

	normalized, err := yaml.Marshal(raw)
	if err != nil {
		return nil, err
	}

	var c Cluster
	if err := yaml.Unmarshal(normalized, &c); err != nil {
		return nil, err
	}

	return &c, nil
}
