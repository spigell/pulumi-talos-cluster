package hcloud

import (
	"path/filepath"
	"testing"

	hcloudsdk "github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	cloud "github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cloud/go"
	cluster "github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cluster/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mocks implements pulumi.MockResourceMonitor.
type Mocks struct {
	CallFunc        func(args pulumi.MockCallArgs) (resource.PropertyMap, error)
	NewResourceFunc func(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error)
}

func (m *Mocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	if m.CallFunc != nil {
		return m.CallFunc(args)
	}
	return resource.PropertyMap{}, nil
}

func (m *Mocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	if m.NewResourceFunc != nil {
		return m.NewResourceFunc(args)
	}
	return args.Name, args.Inputs, nil
}

func defaultMocks() *Mocks {
	return &Mocks{
		CallFunc: func(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
			if args.Token == "hcloud:index/getImage:getImage" {
				return resource.PropertyMap{
					"id": resource.NewStringProperty("12345678"),
				}, nil
			}
			return resource.PropertyMap{}, nil
		},
		NewResourceFunc: func(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
			switch args.TypeToken {
			case "hcloud:index/network:Network":
				return "100", args.Inputs, nil
			case "hcloud:index/networkSubnet:NetworkSubnet":
				return args.Name + "-subnet", args.Inputs, nil
			case "hcloud:index/primaryIp:PrimaryIp":
				addr := "1.2.3.4"
				if t, ok := args.Inputs["type"]; ok && t.IsString() && t.StringValue() == "ipv6" {
					addr = "2001:db8::1"
				}
				return "200", resource.PropertyMap{
					"ipAddress": resource.NewStringProperty(addr),
				}, nil
			case "hcloud:index/sshKey:SshKey":
				return "mock-ssh-key-id", args.Inputs, nil
			default:
				return args.Name + "_id", args.Inputs, nil
			}
		},
	}
}

func TestHetznerUpWithFixtures(t *testing.T) {
	tests := []struct {
		name           string
		fixture        string
		expectedIDs    []string
		expectNetworks bool
	}{
		{
			name:           "load-valid",
			fixture:        "load-valid.yaml",
			expectedIDs:    []string{"worker-0", "controlplane-0"},
			expectNetworks: true,
		},
		{
			name:           "load-minimal",
			fixture:        "load-minimal.yaml",
			expectedIDs:    []string{"simple-machine"},
			expectNetworks: true,
		},
		{
			name:           "validation-networks-present",
			fixture:        "validation-networks-present.yaml",
			expectedIDs:    []string{"worker-1"},
			expectNetworks: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixturePath := filepath.Join("..", "..", "..", "cluster", "fixtures", tt.fixture)
			testCluster, err := cluster.Load(fixturePath)
			require.NoError(t, err)

			machines := map[string]*cluster.Machine{}
			for _, m := range testCluster.Machines {
				machines[m.ID] = m
			}

			err = pulumi.RunErr(func(ctx *pulumi.Context) error {
				h, err := NewWithIPS(ctx, testCluster)
				require.NoError(t, err)

				deployed, err := h.Up()
				require.NoError(t, err)

				assert.Len(t, deployed.Servers, len(tt.expectedIDs))

				gotIDs := make([]string, 0, len(deployed.Servers))
				for _, s := range deployed.Servers {
					gotIDs = append(gotIDs, s.ID())
					assertServerMatchesMachine(t, s, machines, tt.expectNetworks)
				}

				assert.ElementsMatch(t, tt.expectedIDs, gotIDs)

				return nil
			}, pulumi.WithMocks("project", "stack", defaultMocks()))

			assert.NoError(t, err)
		})
	}
}

func assertServerMatchesMachine(
	t *testing.T,
	s cloud.Server,
	machines map[string]*cluster.Machine,
	expectNetworks bool,
) {
	t.Helper()

	server, ok := s.(*hetznerServer)
	require.True(t, ok, "server is not hetznerServer")

	m := machines[server.id]
	require.NotNil(t, m)

	expectedDC := m.Hcloud.Datacenter
	if dc, ok := server.args.Datacenter.(pulumi.String); ok {
		assert.Equal(t, expectedDC, string(dc))
	}

	expectedTalos := m.TalosInitialVersion
	assert.Contains(t, server.imageSelector, expectedTalos)

	if expectNetworks {
		if assert.NotNil(t, server.Args().Networks) {
			switch nets := server.Args().Networks.(type) {
			case hcloudsdk.ServerNetworkTypeArray:
				assert.NotEmpty(t, nets)
			case *hcloudsdk.ServerNetworkTypeArray:
				assert.NotEmpty(t, *nets)
			default:
				t.Fatalf("unexpected networks type %T", nets)
			}
		}
	} else {
		assert.Nil(t, server.Args().Networks)
	}
}
