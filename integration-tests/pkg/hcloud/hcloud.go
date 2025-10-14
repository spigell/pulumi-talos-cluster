package hcloud

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"strconv"
	"strings"

	"github.com/pulumi/pulumi-hcloud/sdk/go/hcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cluster"
	"golang.org/x/crypto/ssh"
)

const (
	defaultDatacenter          = "nbg1-dc3"
	defaultTalosInitialVersion = "v1.11.2"
	testImageSelector          = "os=talos"
)

type Hetzner struct {
	ctx               *pulumi.Context
	privateNetwork    string
	privateSubnetwork string
	clusterName       string

	Servers []*Server
}

type Server struct {
	args          *hcloud.ServerArgs
	privateIP     string
	arch          string
	imageSelector string

	ID string
	IP pulumi.StringOutput
}

type Deployed struct {
	Servers []*Server
	Deps    []pulumi.Resource
}

func NewWithIPS(ctx *pulumi.Context, cluster *cluster.Cluster) (*Hetzner, error) {
	servers := make([]*Server, 0)

	pubKey, err := generateECDSAPubKey()
	if err != nil {
		return nil, err
	}

	sshKey, err := hcloud.NewSshKey(ctx, "sshKey", &hcloud.SshKeyArgs{
		PublicKey: pulumi.String(pubKey),
	}, pulumi.IgnoreChanges([]string{"publicKey"}))
	if err != nil {
		return nil, err
	}

	for _, s := range cluster.Machines {
		if s.Datacenter == "" {
			s.Datacenter = defaultDatacenter
		}

		if s.TalosInitialVersion == "" {
			s.TalosInitialVersion = defaultTalosInitialVersion
		}

		ipv4, err := hcloud.NewPrimaryIp(ctx, fmt.Sprintf("%s-ipv4", s.ID), &hcloud.PrimaryIpArgs{
			Name:         pulumi.Sprintf("%s-%s-ipv4", cluster.Name, s.ID),
			Datacenter:   pulumi.String(s.Datacenter),
			Type:         pulumi.String("ipv4"),
			AssigneeType: pulumi.String("server"),
			AutoDelete:   pulumi.Bool(true),
		})
		if err != nil {
			return nil, err
		}

		ipv6, err := hcloud.NewPrimaryIp(ctx, fmt.Sprintf("%s-ipv6", s.ID), &hcloud.PrimaryIpArgs{
			Name:         pulumi.Sprintf("%s-%s-ipv6", cluster.Name, s.ID),
			Datacenter:   pulumi.String(s.Datacenter),
			Type:         pulumi.String("ipv6"),
			AssigneeType: pulumi.String("server"),
			AutoDelete:   pulumi.Bool(true),
		})
		if err != nil {
			return nil, err
		}

		arch := "x86"

		if strings.HasPrefix(s.ServerType, "cax") || strings.HasPrefix(s.ServerType, "cpx") {
			arch = "arm"
		}

		servers = append(servers, &Server{
			ID:            s.ID,
			privateIP:     s.PrivateIP,
			arch:          arch,
			imageSelector: fmt.Sprintf("%s,version=%s,variant=%s,arch=%s", testImageSelector, s.TalosInitialVersion, s.Platform, arch),
			args: &hcloud.ServerArgs{
				Name: pulumi.Sprintf("%s-%s", cluster.Name, s.ID),
				SshKeys: pulumi.StringArray{
					sshKey.ID(),
				},
				ServerType: pulumi.String(s.ServerType),
				Datacenter: pulumi.String(s.Datacenter),
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

	servers := make([]pulumi.Resource, 0)

	for _, s := range h.Servers {
		image, err := hcloud.GetImage(h.ctx, &hcloud.GetImageArgs{
			WithSelector:     pulumi.StringRef(s.imageSelector),
			MostRecent:       pulumi.BoolRef(true),
			WithArchitecture: pulumi.StringRef(s.arch),
		})
		if err != nil {
			return nil, fmt.Errorf("can't find an image with selector: %s", s.imageSelector)
		}
		s.args.Image = pulumi.Sprintf("%d", image.Id)
		// Define the server
		server, err := hcloud.NewServer(h.ctx, s.ID, s.args,
			pulumi.DependsOn(deps),
			pulumi.IgnoreChanges([]string{"sshKeys", "userData"}),
		)
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

// generatePrivateKey creates a RSA Private Key of specified byte size.
func generateECDSAPubKey() (string, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", err
	}

	publicRsaKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", err
	}

	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)

	return string(pubKeyBytes), nil
}
