package cluster

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func fixturePath(name string) string {
	return filepath.Join("..", "fixtures", name)
}

func TestValidateFixtures(t *testing.T) {
	cases := []struct {
		name     string
		file     string
		wantErr  bool
		contains string
	}{
		{name: "valid", file: "load-valid.yaml"},
		{name: "minimal", file: "load-minimal.yaml"},
		{name: "networks present in range", file: "validation-networks-present.yaml"},
		{name: "anchors allowed", file: "validation-anchors.yaml"},
		{name: "missing name", file: "validation-missing-name.yaml", wantErr: true, contains: "name"},
		{name: "missing machines", file: "validation-missing-machines.yaml", wantErr: true, contains: "machines"},
		{name: "empty machines", file: "validation-empty-machines.yaml", wantErr: true, contains: "machines"},
		{name: "missing id", file: "validation-missing-id.yaml", wantErr: true, contains: "missing properties: 'id'"},
		{name: "missing type", file: "validation-missing-type.yaml", wantErr: true, contains: "missing properties: 'type'"},
		{name: "missing platform", file: "validation-missing-platform.yaml", wantErr: false, contains: "platform"},
		{name: "missing private ip", file: "validation-missing-private-ip.yaml", wantErr: true, contains: "privateIP"},
		{name: "unsupported platform", file: "validation-unsupported-platform.yaml", wantErr: true, contains: "platform"},
		{name: "ip outside", file: "validation-ip-outside.yaml", wantErr: true, contains: "must be inside"},
		{name: "missing networks", file: "validation-missing-networks.yaml", wantErr: true, contains: "when 'usePrivateNetwork' is true"},
		{name: "single network", file: "validation-single-network.yaml", wantErr: true, contains: "when 'usePrivateNetwork' is true"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(fixturePath(tt.file))
			require.NoError(t, err)

			raw, err := parseToMap(data)
			require.NoError(t, err)

			err = validateCluster(raw)
			if tt.wantErr {
				require.Error(t, err)
				if tt.contains != "" {
					require.Contains(t, err.Error(), tt.contains)
				}
				return
			}
			require.NoError(t, err)
		})
	}
}
