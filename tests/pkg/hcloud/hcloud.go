package hcloud

import (
	"fmt"
	"strconv"

	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-talos-cluster/tests/pkg/cluster"
)

const (
	sshPubKey         = "ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBJgyB/EEX/fsSICjyHha9Pnt1IM7brDsFelakF1hTNdKjA+qdvojKWSNraGN81ewf4nxexV6E6e5fEeyr2IIcAQ="
	location          = "nbg1"
	datacenter        = "nbg1-dc3"
	testImageSelector = "os=talos,testing=true"
)

type Hetzner struct {
	ctx               *pulumi.Context
	privateNetwork    string
	privateSubnetwork string
	clusterName       string

	Servers []*Server
}

type Server struct {
	args      *hcloud.ServerArgs
	privateIP string

	ID string
	IP pulumi.StringOutput
}

type Deployed struct {
	Servers []*Server
	Deps    []pulumi.Resource
}

func NewWithIPS(ctx *pulumi.Context, cluster *cluster.Cluster) (*Hetzner, error) {
	servers := make([]*Server, 0)

	sshKey, err := hcloud.NewSshKey(ctx, "sshKey", &hcloud.SshKeyArgs{
		PublicKey: pulumi.String(sshPubKey),
	})
	if err != nil {
		return nil, err
	}

	for _, s := range cluster.Machines {
		ipv4, err := hcloud.NewPrimaryIp(ctx, fmt.Sprintf("%s-ipv4", s.ID), &hcloud.PrimaryIpArgs{
			Name:         pulumi.Sprintf("%s-%s-ipv4", cluster.Name, s.ID),
			Datacenter:   pulumi.String(datacenter),
			Type:         pulumi.String("ipv4"),
			AssigneeType: pulumi.String("server"),
			AutoDelete:   pulumi.Bool(true),
		})
		if err != nil {
			return nil, err
		}

		ipv6, err := hcloud.NewPrimaryIp(ctx, fmt.Sprintf("%s-ipv6", s.ID), &hcloud.PrimaryIpArgs{
			Name:         pulumi.Sprintf("%s-%s-ipv6", cluster.Name, s.ID),
			Datacenter:   pulumi.String(datacenter),
			Type:         pulumi.String("ipv6"),
			AssigneeType: pulumi.String("server"),
			AutoDelete:   pulumi.Bool(true),
		})
		if err != nil {
			return nil, err
		}

		servers = append(servers, &Server{
			ID:        s.ID,
			privateIP: s.PrivateIP,
			args: &hcloud.ServerArgs{
				Name: pulumi.Sprintf("%s-%s", cluster.Name, s.ID),
				SshKeys: pulumi.StringArray{
					sshKey.ID(),
				},
				ServerType: pulumi.String(s.ServerType),
				Datacenter: pulumi.String(datacenter),
				PublicNets: hcloud.ServerPublicNetArray{
					&hcloud.ServerPublicNetArgs{
						//nolint: gocritic // this is the only way to convert string to int
						Ipv4: ipv4.ID().ApplyT(func(id string) (int, error) {
							return strconv.Atoi(id)
						}).(pulumi.IntOutput),
						//nolint: gocritic // this is the only way to convert string to int
						Ipv6: ipv6.ID().ApplyT(func(id string) (int, error) {
							return strconv.Atoi(id)
						}).(pulumi.IntOutput),
					},
				},
			},
			IP: ipv4.IpAddress,
		},
		)
	}

	return &Hetzner{
		ctx:               ctx,
		Servers:           servers,
		privateNetwork:    cluster.PrivateNetwork,
		privateSubnetwork: cluster.PrivateSubnetwork,
		clusterName:       cluster.Name,
	}, nil
}

func (h *Hetzner) Up() (*Deployed, error) {
	deps := make([]pulumi.Resource, 0)

	if h.privateNetwork != "" {
		network, err := hcloud.NewNetwork(h.ctx, "private-network", &hcloud.NetworkArgs{
			Name:    pulumi.Sprintf("private-network-%s", h.clusterName),
			IpRange: pulumi.String(h.privateNetwork), // Define the CIDR block for the network
		})
		if err != nil {
			return nil, fmt.Errorf("error creating network: %w", err)
		}

		//nolint: gocritic // this is the only way to convert string to int
		convertedNetID := network.ID().ApplyT(func(id string) (int, error) {
			return strconv.Atoi(id)
		}).(pulumi.IntOutput)

		// Add a subnet to the private network
		subnet, err := hcloud.NewNetworkSubnet(h.ctx, "private-subnet", &hcloud.NetworkSubnetArgs{
			NetworkId:   convertedNetID,
			Type:        pulumi.String("server"),
			NetworkZone: pulumi.String("eu-central"),        // Adjust based on your preferred region
			IpRange:     pulumi.String(h.privateSubnetwork), // Define the CIDR block for the subnet
		})
		if err != nil {
			return nil, fmt.Errorf("error creating subnet: %w", err)
		}

		for _, s := range h.Servers {
			s.args.Networks = &hcloud.ServerNetworkTypeArray{
				&hcloud.ServerNetworkTypeArgs{
					NetworkId: convertedNetID,
					Ip:        pulumi.String(s.privateIP),
				},
			}
		}

		deps = append(deps, subnet)
	}

	image, err := hcloud.GetImage(h.ctx, &hcloud.GetImageArgs{
		WithSelector:     pulumi.StringRef(testImageSelector),
		MostRecent:       pulumi.BoolRef(true),
		WithArchitecture: pulumi.StringRef("x86"),
	})
	if err != nil {
		return nil, fmt.Errorf("can't find an image")
	}

	servers := make([]pulumi.Resource, 0)

	for _, s := range h.Servers {
		s.args.Image = pulumi.Sprintf("%d", image.Id)
		// Define the server
		server, err := hcloud.NewServer(h.ctx, s.ID, s.args, pulumi.DependsOn(deps))
		if err != nil {
			return nil, fmt.Errorf("error creating server: %w", err)
		}

		servers = append(servers, server)
	}

	return &Deployed{
		Servers: h.Servers,
		Deps:    servers,
	}, nil
}

func (s *Server) Args() *hcloud.ServerArgs {
	return s.args
}

func (s *Server) WithUserdata(userdata pulumi.StringOutput) *Server {
	s.args.UserData = userdata

	return s
}
