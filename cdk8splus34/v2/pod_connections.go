package cdk8splus34

import (
	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// PodConnectionsIsolation chooses which endpoint receives a NetworkPolicy.
type PodConnectionsIsolation string

const (
	PodConnectionsIsolation_POD  PodConnectionsIsolation = "POD"
	PodConnectionsIsolation_PEER PodConnectionsIsolation = "PEER"
)

type PodConnectionsAllowToOptions struct {
	Isolation PodConnectionsIsolation `field:"optional" json:"isolation" yaml:"isolation"`
	Ports     *[]NetworkPolicyPort    `field:"optional" json:"ports" yaml:"ports"`
}

type PodConnectionsAllowFromOptions struct {
	Isolation PodConnectionsIsolation `field:"optional" json:"isolation" yaml:"isolation"`
	Ports     *[]NetworkPolicyPort    `field:"optional" json:"ports" yaml:"ports"`
}

// PodConnections controls NetworkPolicy isolation rules for one pod or
// workload pod template.
type PodConnections interface {
	Instance() AbstractPod
	AllowFrom(INetworkPolicyPeer, *PodConnectionsAllowFromOptions)
	AllowTo(INetworkPolicyPeer, *PodConnectionsAllowToOptions)
	Isolate()
}

type podConnectionsImpl struct{ instance AbstractPod }

func NewPodConnections(instance AbstractPod) PodConnections {
	if instance == nil {
		panic("instance is required")
	}
	return &podConnectionsImpl{instance: instance}
}

func NewPodConnections_Override(connections PodConnections, instance AbstractPod) {
	applyOverride(connections, NewPodConnections(instance), "PodConnections")
}

func (p *podConnectionsImpl) Instance() AbstractPod {
	return p.instance
}

func (p *podConnectionsImpl) AllowTo(peer INetworkPolicyPeer, options *PodConnectionsAllowToOptions) {
	if peer == nil {
		panic("peer is required")
	}
	ports := extractedNetworkPolicyPorts(peer)
	isolation := PodConnectionsIsolation("")
	if options != nil {
		isolation = options.Isolation
		if options.Ports != nil {
			ports = options.Ports
		}
	}
	p.allow("Egress", peer, isolation, ports)
}

func (p *podConnectionsImpl) AllowFrom(peer INetworkPolicyPeer, options *PodConnectionsAllowFromOptions) {
	if peer == nil {
		panic("peer is required")
	}
	ports := extractedNetworkPolicyPorts(p.instance)
	isolation := PodConnectionsIsolation("")
	if options != nil {
		isolation = options.Isolation
		if options.Ports != nil {
			ports = options.Ports
		}
	}
	p.allow("Ingress", peer, isolation, ports)
}

func (p *podConnectionsImpl) Isolate() {
	NewNetworkPolicy(p.instance, jsii.String("DefaultDenyAll"), &NetworkPolicyProps{
		Metadata: &cdk8s.ApiObjectMetadata{Namespace: p.instance.Metadata().Namespace()}, Selector: p.instance,
		Egress:  &NetworkPolicyTraffic{Default: NetworkPolicyTrafficDefault_DENY},
		Ingress: &NetworkPolicyTraffic{Default: NetworkPolicyTrafficDefault_DENY},
	})
}

func extractedNetworkPolicyPorts(peer INetworkPolicyPeer) *[]NetworkPolicyPort {
	containers, ok := peer.(interface{ Containers() *[]Container })
	if !ok {
		return &[]NetworkPolicyPort{}
	}
	result := make([]NetworkPolicyPort, 0)
	for _, container := range *containers.Containers() {
		for _, port := range *container.Ports() {
			if port != nil && port.Number != nil {
				result = append(result, NetworkPolicyPort_Tcp(port.Number))
			}
		}
	}
	return &result
}

func peerAddress(peer INetworkPolicyPeer) string {
	if peer == nil || peer.Node() == nil {
		return "Peer"
	}
	return stringValue(peer.Node().Addr())
}

func (p *podConnectionsImpl) allow(direction string, peer INetworkPolicyPeer, isolation PodConnectionsIsolation, ports *[]NetworkPolicyPort) {
	config := peer.ToNetworkPolicyPeerConfig()
	if config == nil || (config.IpBlock == nil && config.PodSelector == nil) {
		panic("network policy peer configuration is required")
	}
	address := peerAddress(peer)
	if isolation == "" || isolation == PodConnectionsIsolation_POD {
		policy := NewNetworkPolicy(p.instance, jsii.String("Allow"+direction+address), &NetworkPolicyProps{
			Metadata: &cdk8s.ApiObjectMetadata{Namespace: p.instance.Metadata().Namespace()}, Selector: p.instance,
		})
		if direction == "Egress" {
			policy.AddEgressRule(peer, ports)
		} else {
			policy.AddIngressRule(peer, ports)
		}
	}
	if isolation != "" && isolation != PodConnectionsIsolation_PEER {
		return
	}
	if config.IpBlock != nil {
		return
	}
	selector := peer.ToPodSelector()
	if selector == nil {
		panic("Unable to create opposite network policy because peer is not a pod selector")
	}
	selectorConfig := selector.ToPodSelectorConfig()
	if selectorConfig == nil {
		panic("pod selector configuration is required")
	}
	namespaces := []*string{p.instance.Metadata().Namespace()}
	if selectorConfig.Namespaces != nil {
		if selectorConfig.Namespaces.LabelSelector != nil && !*selectorConfig.Namespaces.LabelSelector.IsEmpty() {
			panic("Unable to create opposite network policy for peer with namespace label selector")
		}
		if selectorConfig.Namespaces.Names == nil {
			panic("Unable to create opposite network policy for peer without namespace names")
		}
		namespaces = *selectorConfig.Namespaces.Names
	}
	for _, namespace := range namespaces {
		// TypeScript interpolates an omitted namespace as "undefined" in the
		// construct id, while leaving the resource metadata namespace unset. Do
		// the same here: an unset namespace means the chart default namespace,
		// not that the opposite policy should be skipped.
		namespaceAddress := "undefined"
		if namespace != nil {
			namespaceAddress = *namespace
		}
		if direction == "Egress" {
			policy := NewNetworkPolicy(p.instance, jsii.String("AllowIngress"+namespaceAddress+address), &NetworkPolicyProps{Metadata: &cdk8s.ApiObjectMetadata{Namespace: namespace}, Selector: selector})
			policy.AddIngressRule(p.instance, ports)
		} else {
			policy := NewNetworkPolicy(p.instance, jsii.String("AllowEgress"+namespaceAddress+address), &NetworkPolicyProps{Metadata: &cdk8s.ApiObjectMetadata{Namespace: namespace}, Selector: selector})
			policy.AddEgressRule(p.instance, ports)
		}
	}
}
