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
	cloud "github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cloud/go"
	cluster "github.com/spigell/pulumi-talos-cluster/integration-tests/pkg/cluster/go"
	"golang.org/x/crypto/ssh"
)

const (
	defaultDatacenter = "nbg1-dc3"
	// It should be renamed; this value is not related to Talos.
	defaultTalosInitialVersion = "v1.10.3"
	testImageSelector          = "os=talos"
	defaultServerType          = "cx11"
)

type Hetzner struct {
	ctx               *pulumi.Context
	privateNetwork    string
	privateSubnetwork string
	clusterName       string

	servers []*hetznerServer
}

type hetznerServer struct {
	args          *hcloud.ServerArgs
	privateIP     string
	arch          string
	imageSelector string

	id string
	ip pulumi.StringOutput
}

func NewWithIPS(ctx *pulumi.Context, clu *cluster.Cluster) (*Hetzner, error) {
	servers := make([]*hetznerServer, 0, len(clu.Machines))

	pubKey, err := generateECDSAPubKey()
	if err != nil {
		return nil, fmt.Errorf("generate ECDSA public key: %w", err)
	}

	sshKey, err := hcloud.NewSshKey(ctx, "sshKey", &hcloud.SshKeyArgs{
		PublicKey: pulumi.String(pubKey),
	}, pulumi.IgnoreChanges([]string{"publicKey"}))
	if err != nil {
		return nil, fmt.Errorf("create SSH key: %w", err)
	}

	for _, machine := range clu.Machines {
		server, err := newServer(ctx, clu, machine, sshKey)
		if err != nil {
			return nil, err
		}

		servers = append(servers, server)
	}

	return &Hetzner{
		ctx:               ctx,
		servers:           servers,
		privateNetwork:    clu.PrivateNetwork,
		privateSubnetwork: clu.PrivateSubnetwork,
		clusterName:       clu.Name,
	}, nil
}

func (h *Hetzner) Servers() []cloud.Server {
	result := make([]cloud.Server, len(h.servers))
	for i, s := range h.servers {
		result[i] = s
	}
	return result
}

func (s *hetznerServer) Args() *hcloud.ServerArgs {
	return s.args
}

func (s *hetznerServer) ID() string {
	return s.id
}

func (s *hetznerServer) IP() pulumi.StringOutput {
	return s.ip
}

func (s *hetznerServer) WithUserdata(userdata pulumi.StringOutput) cloud.Server {
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
			return nil, fmt.Errorf("find image with selector %q: %w", s.imageSelector, err)
		}
		s.args.Image = pulumi.Sprintf("%d", image.Id)
		// Define the server
		server, err := hcloud.NewServer(h.ctx, s.id, s.args,
			pulumi.DependsOn(deps),
			pulumi.DeleteBeforeReplace(true),
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

func newServer(ctx *pulumi.Context, clu *cluster.Cluster, machine *cluster.Machine, sshKey *hcloud.SshKey) (*hetznerServer, error) {
	if machine.Hcloud == nil {
		machine.Hcloud = &cluster.HcloudMachine{}
	}

	if machine.Hcloud.ServerType == "" {
		machine.Hcloud.ServerType = defaultServerType
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
		Name:         pulumi.Sprintf("%s-%s-ipv4", clu.Name, machine.ID),
		Datacenter:   pulumi.String(datacenter),
		Type:         pulumi.String("ipv4"),
		AssigneeType: pulumi.String("server"),
		AutoDelete:   pulumi.Bool(false),
	})
	if err != nil {
		return nil, fmt.Errorf("allocate ipv4 for machine %q: %w", machine.ID, err)
	}

	ipv6, err := hcloud.NewPrimaryIp(ctx, fmt.Sprintf("%s-ipv6", machine.ID), &hcloud.PrimaryIpArgs{
		Name:         pulumi.Sprintf("%s-%s-ipv6", clu.Name, machine.ID),
		Datacenter:   pulumi.String(datacenter),
		Type:         pulumi.String("ipv6"),
		AssigneeType: pulumi.String("server"),
		AutoDelete:   pulumi.Bool(false),
	})
	if err != nil {
		return nil, fmt.Errorf("allocate ipv6 for machine %q: %w", machine.ID, err)
	}

	arch := architectureForServer(machine.Hcloud.ServerType)
	imageSelector := fmt.Sprintf(
		"%s,version=%s,variant=%s,arch=%s",
		testImageSelector,
		talosVersion,
		machine.Variant,
		arch,
	)

	return &hetznerServer{
		id:            machine.ID,
		privateIP:     machine.PrivateIP,
		arch:          arch,
		imageSelector: imageSelector,
		args: &hcloud.ServerArgs{
			Name: pulumi.Sprintf("%s-%s", clu.Name, machine.ID),
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
		return "", fmt.Errorf("generate private key: %w", err)
	}

	publicRsaKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", fmt.Errorf("build public key: %w", err)
	}

	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)

	return string(pubKeyBytes), nil
}
