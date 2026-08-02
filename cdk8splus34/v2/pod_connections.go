package cdk8splus34

import (
	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// Isolation determines which policies are created when allowing connections from a a pod / workload to peers.
type PodConnectionsIsolation string

const (
	// Only creates network policies that select the pod.
	PodConnectionsIsolation_POD PodConnectionsIsolation = "POD"
	// Only creates network policies that select the peer.
	PodConnectionsIsolation_PEER PodConnectionsIsolation = "PEER"
)

// Options for `PodConnections.allowTo`.
type PodConnectionsAllowToOptions struct {
	// Which isolation should be applied to establish the connection. Default: - unset, isolates both the pod and the peer.
	Isolation PodConnectionsIsolation `field:"optional" json:"isolation" yaml:"isolation"`
	// Ports to allow outgoing traffic to. Default: - If the peer is a managed pod, take its ports. Otherwise, all ports are allowed.
	Ports *[]NetworkPolicyPort `field:"optional" json:"ports" yaml:"ports"`
}

// Options for `PodConnections.allowFrom`.
type PodConnectionsAllowFromOptions struct {
	// Which isolation should be applied to establish the connection. Default: - unset, isolates both the pod and the peer.
	Isolation PodConnectionsIsolation `field:"optional" json:"isolation" yaml:"isolation"`
	// Ports to allow incoming traffic to. Default: - The pod ports.
	Ports *[]NetworkPolicyPort `field:"optional" json:"ports" yaml:"ports"`
}

// Controls network isolation rules for inter-pod communication.
type PodConnections interface {
	Instance() AbstractPod
	// Allow network traffic from the peer to this pod.
	//
	// By default, this will create an ingress network policy for this pod, and an egress network policy for the peer. This is required if both sides are already isolated. Use `options.isolation` to control this behavior.
	//
	// Example:
	//
	//	// create only an egress policy that selects the 'web' pod to allow outgoing traffic
	//	// to the 'redis' pod. this requires the 'redis' pod to not be isolated for ingress.
	//	redis.connections.allowFrom(web, { isolation: Isolation.PEER })
	//
	//	// create only an ingress policy that selects the 'redis' peer to allow incoming traffic
	//	// from the 'web' pod. this requires the 'web' pod to not be isolated for egress.
	//	redis.connections.allowFrom(web, { isolation: Isolation.POD })
	AllowFrom(INetworkPolicyPeer, *PodConnectionsAllowFromOptions)
	// Allow network traffic from this pod to the peer.
	//
	// By default, this will create an egress network policy for this pod, and an ingress network policy for the peer. This is required if both sides are already isolated. Use `options.isolation` to control this behavior.
	//
	// Example:
	//
	//	// create only an egress policy that selects the 'web' pod to allow outgoing traffic
	//	// to the 'redis' pod. this requires the 'redis' pod to not be isolated for ingress.
	//	web.connections.allowTo(redis, { isolation: Isolation.POD })
	//
	//	// create only an ingress policy that selects the 'redis' peer to allow incoming traffic
	//	// from the 'web' pod. this requires the 'web' pod to not be isolated for egress.
	//	web.connections.allowTo(redis, { isolation: Isolation.PEER })
	AllowTo(INetworkPolicyPeer, *PodConnectionsAllowToOptions)
	// Sets the default network policy for Pod/Workload to have all egress and ingress connections as disabled.
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
