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
	"github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cloud"
	"github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cluster"
	"golang.org/x/crypto/ssh"
)

const (
	defaultDatacenter          = "nbg1-dc3"
	defaultTalosInitialVersion = "v1.10.3"
	testImageSelector          = "os=talos"
)

type Hetzner struct {
	ctx               *pulumi.Context
	privateNetwork    string
	privateSubnetwork string
	clusterName       string

	servers []*Server
}

type Server struct {
	args          *hcloud.ServerArgs
	privateIP     string
	arch          string
	imageSelector string

	id string
	ip pulumi.StringOutput
}

func NewWithIPS(ctx *pulumi.Context, cluster *cluster.Cluster) (*Hetzner, error) {
	servers := make([]*Server, 0, len(cluster.Machines))

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

	for _, machine := range cluster.Machines {
		server, err := newServer(ctx, cluster, machine, sshKey)
		if err != nil {
			return nil, err
		}

		servers = append(servers, server)
	}

	return &Hetzner{
		ctx:               ctx,
		servers:           servers,
		privateNetwork:    cluster.PrivateNetwork,
		privateSubnetwork: cluster.PrivateSubnetwork,
		clusterName:       cluster.Name,
	}, nil
}

func (h *Hetzner) Servers() []cloud.Server {
	result := make([]cloud.Server, len(h.servers))
	for i, s := range h.servers {
		result[i] = s
	}
	return result
}

func (s *Server) Args() *hcloud.ServerArgs {
	return s.args
}

func (s *Server) ID() string {
	return s.id
}

func (s *Server) IP() pulumi.StringOutput {
	return s.ip
}

func (s *Server) WithUserdata(userdata pulumi.StringOutput) cloud.Server {
	s.args.UserData = userdata

	return s
}

func (h *Hetzner) Up() (*cloud.Deployed, error) {
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

		for _, s := range h.servers {
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

	for _, s := range h.servers {
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
		server, err := hcloud.NewServer(h.ctx, s.id, s.args,
			pulumi.DependsOn(deps),
			pulumi.IgnoreChanges([]string{"sshKeys", "userData"}),
		)
		if err != nil {
			return nil, fmt.Errorf("error creating server: %w", err)
		}

		servers = append(servers, server)
	}

	return &cloud.Deployed{
		Servers: h.Servers(),
		Deps:    servers,
	}, nil
}

func newServer(ctx *pulumi.Context, cluster *cluster.Cluster, machine *cluster.Machine, sshKey *hcloud.SshKey) (*Server, error) {
	if machine.Hcloud == nil {
		return nil, fmt.Errorf("machine %q is missing hcloud configuration", machine.ID)
	}

	if machine.Hcloud.ServerType == "" {
		return nil, fmt.Errorf("machine %q must define hcloud.serverType", machine.ID)
	}

	if machine.PrivateIP == "" {
		return nil, fmt.Errorf("machine %q must define privateIP", machine.ID)
	}

	datacenter := machine.Hcloud.Datacenter
	if datacenter == "" {
		datacenter = defaultDatacenter
	}

	talosVersion := machine.TalosInitialVersion
	if talosVersion == "" {
		talosVersion = defaultTalosInitialVersion
	}

	ipv4, err := hcloud.NewPrimaryIp(ctx, fmt.Sprintf("%s-ipv4", machine.ID), &hcloud.PrimaryIpArgs{
		Name:         pulumi.Sprintf("%s-%s-ipv4", cluster.Name, machine.ID),
		Datacenter:   pulumi.String(datacenter),
		Type:         pulumi.String("ipv4"),
		AssigneeType: pulumi.String("server"),
		AutoDelete:   pulumi.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	ipv6, err := hcloud.NewPrimaryIp(ctx, fmt.Sprintf("%s-ipv6", machine.ID), &hcloud.PrimaryIpArgs{
		Name:         pulumi.Sprintf("%s-%s-ipv6", cluster.Name, machine.ID),
		Datacenter:   pulumi.String(datacenter),
		Type:         pulumi.String("ipv6"),
		AssigneeType: pulumi.String("server"),
		AutoDelete:   pulumi.Bool(true),
	})
	if err != nil {
		return nil, err
	}

	arch := architectureForServer(machine.Hcloud.ServerType)
	imageSelector := fmt.Sprintf(
		"%s,version=%s,variant=%s,arch=%s",
		testImageSelector,
		talosVersion,
		machine.Platform,
		arch,
	)

	return &Server{
		id:            machine.ID,
		privateIP:     machine.PrivateIP,
		arch:          arch,
		imageSelector: imageSelector,
		args: &hcloud.ServerArgs{
			Name: pulumi.Sprintf("%s-%s", cluster.Name, machine.ID),
			SshKeys: pulumi.StringArray{
				sshKey.ID(),
			},
			ServerType: pulumi.String(machine.Hcloud.ServerType),
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
		ip: ipv4.IpAddress,
	}, nil
}

func architectureForServer(serverType string) string {
	if strings.HasPrefix(serverType, "cax") || strings.HasPrefix(serverType, "cpx") {
		return "arm"
	}

	return "x86"
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
