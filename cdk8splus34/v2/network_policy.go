package cdk8splus34

import (
	"net"
	"strings"

	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

type NetworkProtocol string

const (
	NetworkProtocol_TCP  NetworkProtocol = "TCP"
	NetworkProtocol_UDP  NetworkProtocol = "UDP"
	NetworkProtocol_SCTP NetworkProtocol = "SCTP"
)

type NetworkPolicyPortProps struct {
	EndPort  *float64        `field:"optional" json:"endPort" yaml:"endPort"`
	Port     *float64        `field:"optional" json:"port" yaml:"port"`
	Protocol NetworkProtocol `field:"optional" json:"protocol" yaml:"protocol"`
}

type NetworkPolicyPort interface{ toManifest() map[string]interface{} }

type networkPolicyPortImpl struct {
	port, endPort *float64
	protocol      NetworkProtocol
}

func (p *networkPolicyPortImpl) toManifest() map[string]interface{} {
	result := map[string]interface{}{}
	if p.port != nil {
		result["port"] = p.port
	}
	if p.endPort != nil {
		result["endPort"] = p.endPort
	}
	if p.protocol != "" {
		result["protocol"] = string(p.protocol)
	}
	return result
}

func NewNetworkPolicyPort(props *NetworkPolicyPortProps) NetworkPolicyPort {
	return NetworkPolicyPort_Of(props)
}

func NewNetworkPolicyPort_Override(port NetworkPolicyPort, props *NetworkPolicyPortProps) {
	applyOverride(port, NetworkPolicyPort_Of(props), "NetworkPolicyPort")
}

func NetworkPolicyPort_Of(props *NetworkPolicyPortProps) NetworkPolicyPort {
	if props == nil {
		panic("props is required")
	}
	return &networkPolicyPortImpl{port: props.Port, endPort: props.EndPort, protocol: props.Protocol}
}

func NetworkPolicyPort_Tcp(port *float64) NetworkPolicyPort {
	if port == nil {
		panic("port is required")
	}
	return &networkPolicyPortImpl{port: port, protocol: NetworkProtocol_TCP}
}

func NetworkPolicyPort_TcpRange(startPort, endPort *float64) NetworkPolicyPort {
	if startPort == nil || endPort == nil {
		panic("startPort and endPort are required")
	}
	return &networkPolicyPortImpl{port: startPort, endPort: endPort, protocol: NetworkProtocol_TCP}
}

func NetworkPolicyPort_AllTcp() NetworkPolicyPort {
	return &networkPolicyPortImpl{port: jsii.Number(0), endPort: jsii.Number(65535), protocol: NetworkProtocol_TCP}
}

func NetworkPolicyPort_Udp(port *float64) NetworkPolicyPort {
	if port == nil {
		panic("port is required")
	}
	return &networkPolicyPortImpl{port: port, protocol: NetworkProtocol_UDP}
}

func NetworkPolicyPort_UdpRange(startPort, endPort *float64) NetworkPolicyPort {
	if startPort == nil || endPort == nil {
		panic("startPort and endPort are required")
	}
	return &networkPolicyPortImpl{port: startPort, endPort: endPort, protocol: NetworkProtocol_UDP}
}

func NetworkPolicyPort_AllUdp() NetworkPolicyPort {
	return &networkPolicyPortImpl{port: jsii.Number(0), endPort: jsii.Number(65535), protocol: NetworkProtocol_UDP}
}

type NetworkPolicyIpBlock interface {
	constructs.Construct
	INetworkPolicyPeer
	Cidr() *string
	Except() *[]*string
	toManifest() map[string]interface{}
}

type networkPolicyIpBlockImpl struct {
	node   constructs.Node
	cidr   *string
	except []*string
}

func (n *networkPolicyIpBlockImpl) Node() constructs.Node {
	return n.node
}

func (n *networkPolicyIpBlockImpl) SetNodeInternal(node constructs.Node) {
	n.node = node
}

func (n *networkPolicyIpBlockImpl) ToString() *string {
	return n.node.Path()
}

func (n *networkPolicyIpBlockImpl) With(mixins ...constructs.IMixin) constructs.IConstruct {
	return n.node.With(mixins...)
}

func (n *networkPolicyIpBlockImpl) Cidr() *string {
	return n.cidr
}

func (n *networkPolicyIpBlockImpl) Except() *[]*string {
	result := append([]*string(nil), n.except...)
	return &result
}

func (n *networkPolicyIpBlockImpl) ToNetworkPolicyPeerConfig() *NetworkPolicyPeerConfig {
	return &NetworkPolicyPeerConfig{IpBlock: n}
}

func (n *networkPolicyIpBlockImpl) ToPodSelector() IPodSelector {
	return nil
}

func (n *networkPolicyIpBlockImpl) toManifest() map[string]interface{} {
	result := map[string]interface{}{"cidr": n.cidr}
	if len(n.except) > 0 {
		result["except"] = n.except
	}
	return result
}

func newNetworkPolicyIpBlock(scope constructs.Construct, id, cidr *string, except *[]*string, family int) NetworkPolicyIpBlock {
	if scope == nil || id == nil || cidr == nil {
		panic("scope, id and cidrIp are required")
	}
	if !strings.Contains(*cidr, "/") {
		if family == 4 {
			panic("CIDR mask is missing in IPv4: \"" + *cidr + "\". Did you mean \"" + *cidr + "/32\"?")
		}
		panic("CIDR mask is missing in IPv6: \"" + *cidr + "\". Did you mean \"" + *cidr + "/128\"?")
	}
	ip, _, err := net.ParseCIDR(*cidr)
	if err != nil || (family == 4 && ip.To4() == nil) || (family == 6 && ip.To4() != nil) {
		if family == 4 {
			panic("Invalid IPv4 CIDR: \"" + *cidr + "\"")
		}
		panic("Invalid IPv6 CIDR: \"" + *cidr + "\"")
	}
	result := &networkPolicyIpBlockImpl{cidr: cidr}
	if except != nil {
		result.except = append(result.except, (*except)...)
	}
	constructs.NewConstruct_Override(result, scope, id)
	return result
}

func NetworkPolicyIpBlock_Ipv4(scope constructs.Construct, id, cidr *string, except *[]*string) NetworkPolicyIpBlock {
	return newNetworkPolicyIpBlock(scope, id, cidr, except, 4)
}

func NetworkPolicyIpBlock_Ipv6(scope constructs.Construct, id, cidr *string, except *[]*string) NetworkPolicyIpBlock {
	return newNetworkPolicyIpBlock(scope, id, cidr, except, 6)
}

func NetworkPolicyIpBlock_AnyIpv4(scope constructs.Construct, id *string) NetworkPolicyIpBlock {
	return NetworkPolicyIpBlock_Ipv4(scope, id, jsii.String("0.0.0.0/0"), nil)
}

func NetworkPolicyIpBlock_AnyIpv6(scope constructs.Construct, id *string) NetworkPolicyIpBlock {
	return NetworkPolicyIpBlock_Ipv6(scope, id, jsii.String("::/0"), nil)
}

func NetworkPolicyIpBlock_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

type NetworkPolicyTrafficDefault string

const (
	NetworkPolicyTrafficDefault_DENY  NetworkPolicyTrafficDefault = "DENY"
	NetworkPolicyTrafficDefault_ALLOW NetworkPolicyTrafficDefault = "ALLOW"
)

type NetworkPolicyRule struct {
	Peer  INetworkPolicyPeer   `field:"required" json:"peer" yaml:"peer"`
	Ports *[]NetworkPolicyPort `field:"optional" json:"ports" yaml:"ports"`
}

type NetworkPolicyTraffic struct {
	Default NetworkPolicyTrafficDefault `field:"optional" json:"default" yaml:"default"`
	Rules   *[]*NetworkPolicyRule       `field:"optional" json:"rules" yaml:"rules"`
}

type NetworkPolicyAddEgressRuleOptions struct {
	Ports *[]NetworkPolicyPort `field:"optional" json:"ports" yaml:"ports"`
}

type NetworkPolicyProps struct {
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	Egress   *NetworkPolicyTraffic    `field:"optional" json:"egress" yaml:"egress"`
	Ingress  *NetworkPolicyTraffic    `field:"optional" json:"ingress" yaml:"ingress"`
	Selector IPodSelector             `field:"optional" json:"selector" yaml:"selector"`
}

type NetworkPolicy interface {
	Resource
	AddEgressRule(peer INetworkPolicyPeer, ports *[]NetworkPolicyPort)
	AddIngressRule(peer INetworkPolicyPeer, ports *[]NetworkPolicyPort)
}

type networkPolicyImpl struct {
	resourceBase
	selector     *PodSelectorConfig
	egressRules  []interface{}
	ingressRules []interface{}
	policyTypes  map[string]bool
}

func NewNetworkPolicy(scope constructs.Construct, id *string, props *NetworkPolicyProps) NetworkPolicy {
	if props == nil {
		props = &NetworkPolicyProps{}
	}
	selectorConfig := (*PodSelectorConfig)(nil)
	if props.Selector != nil {
		selectorConfig = props.Selector.ToPodSelectorConfig()
	}
	metadata := networkPolicyMetadata(props.Metadata, selectorConfig)
	result := &networkPolicyImpl{policyTypes: map[string]bool{}}
	manifest := map[string]interface{}{}
	result.resourceBase.initialize(result, scope, id, "networking.k8s.io/v1", "NetworkPolicy", "networkpolicies", metadata, manifest)
	if props.Selector == nil {
		result.selector = Pods_All(result, jsii.String("AllPods"), nil).ToPodSelectorConfig()
	} else {
		result.selector = selectorConfig
	}
	if result.selector == nil || result.selector.LabelSelector == nil {
		panic("network policy selector is required")
	}
	manifest["spec"] = cdk8s.Lazy_Any(lazyProducer{produce: func() interface{} { return result.toManifest() }})
	result.configureTraffic("Egress", props.Egress)
	result.configureTraffic("Ingress", props.Ingress)
	return result
}

func networkPolicyMetadata(metadata *cdk8s.ApiObjectMetadata, selector *PodSelectorConfig) *cdk8s.ApiObjectMetadata {
	if metadata != nil && metadata.Namespace != nil {
		return metadata
	}
	if selector == nil || selector.Namespaces == nil {
		return metadata
	}
	if selector.Namespaces.LabelSelector != nil && !*selector.Namespaces.LabelSelector.IsEmpty() {
		panic("Unable to create a network policy for a selector that selects pods in namespaces based on labels")
	}
	if selector.Namespaces.Names != nil && len(*selector.Namespaces.Names) > 1 {
		panic("Unable to create a network policy for a selector that selects pods in multiple namespaces")
	}
	if selector.Namespaces.Names == nil || len(*selector.Namespaces.Names) == 0 {
		return metadata
	}
	result := &cdk8s.ApiObjectMetadata{Namespace: (*selector.Namespaces.Names)[0]}
	if metadata != nil {
		copy := *metadata
		copy.Namespace = result.Namespace
		result = &copy
	}
	return result
}

func NewNetworkPolicy_Override(policy NetworkPolicy, scope constructs.Construct, id *string, props *NetworkPolicyProps) {
	applyOverride(policy, NewNetworkPolicy(scope, id, props), "NetworkPolicy")
}

func NetworkPolicy_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func (n *networkPolicyImpl) configureTraffic(direction string, traffic *NetworkPolicyTraffic) {
	if traffic == nil {
		return
	}
	if traffic.Default == NetworkPolicyTrafficDefault_DENY {
		n.policyTypes[direction] = true
	}
	if traffic.Default == NetworkPolicyTrafficDefault_ALLOW {
		n.policyTypes[direction] = true
		if direction == "Egress" {
			n.egressRules = append(n.egressRules, map[string]interface{}{})
		} else {
			n.ingressRules = append(n.ingressRules, map[string]interface{}{})
		}
	}
	if traffic.Rules != nil {
		for _, rule := range *traffic.Rules {
			if rule == nil || rule.Peer == nil {
				panic("network policy rule peer is required")
			}
			if direction == "Egress" {
				n.AddEgressRule(rule.Peer, rule.Ports)
			} else {
				n.AddIngressRule(rule.Peer, rule.Ports)
			}
		}
	}
}

func (n *networkPolicyImpl) AddEgressRule(peer INetworkPolicyPeer, ports *[]NetworkPolicyPort) {
	n.policyTypes["Egress"] = true
	n.egressRules = append(n.egressRules, map[string]interface{}{"ports": networkPolicyPortManifests(ports), "to": networkPolicyPeers(peer)})
}

func (n *networkPolicyImpl) AddIngressRule(peer INetworkPolicyPeer, ports *[]NetworkPolicyPort) {
	n.policyTypes["Ingress"] = true
	n.ingressRules = append(n.ingressRules, map[string]interface{}{"ports": networkPolicyPortManifests(ports), "from": networkPolicyPeers(peer)})
}

func networkPolicyPortManifests(ports *[]NetworkPolicyPort) []interface{} {
	result := []interface{}{}
	if ports != nil {
		for _, port := range *ports {
			if port == nil {
				panic("network policy port is required")
			}
			result = append(result, port.toManifest())
		}
	}
	return result
}

func networkPolicyPeers(peer INetworkPolicyPeer) []interface{} {
	if peer == nil {
		panic("peer is required")
	}
	config := peer.ToNetworkPolicyPeerConfig()
	if config == nil || (config.IpBlock == nil && config.PodSelector == nil) || (config.IpBlock != nil && config.PodSelector != nil) {
		panic("Invalid peer: either ipBlock or podSelector must be defined")
	}
	if config.IpBlock != nil {
		return []interface{}{map[string]interface{}{"ipBlock": config.IpBlock.toManifest()}}
	}
	selector := config.PodSelector
	podSelector := labelSelectorManifest(selector.LabelSelector)
	if selector.Namespaces == nil || selector.Namespaces.Names == nil {
		entry := map[string]interface{}{"podSelector": podSelector}
		if selector.Namespaces != nil && selector.Namespaces.LabelSelector != nil {
			entry["namespaceSelector"] = labelSelectorManifest(selector.Namespaces.LabelSelector)
		}
		return []interface{}{entry}
	}
	result := []interface{}{}
	namespaceSelector := map[string]interface{}{}
	if selector.Namespaces.LabelSelector != nil {
		namespaceSelector = labelSelectorManifest(selector.Namespaces.LabelSelector)
	}
	for _, name := range *selector.Namespaces.Names {
		labels := map[string]interface{}{}
		if current, ok := namespaceSelector["matchLabels"].(map[string]*string); ok {
			for key, value := range current {
				labels[key] = value
			}
		}
		labels["kubernetes.io/metadata.name"] = name
		entry := map[string]interface{}{"podSelector": podSelector, "namespaceSelector": map[string]interface{}{"matchExpressions": namespaceSelector["matchExpressions"], "matchLabels": labels}}
		result = append(result, entry)
	}
	return result
}

func (n *networkPolicyImpl) toManifest() map[string]interface{} {
	result := map[string]interface{}{"podSelector": labelSelectorManifest(n.selector.LabelSelector)}
	if len(n.egressRules) > 0 {
		result["egress"] = n.egressRules
	}
	if len(n.ingressRules) > 0 {
		result["ingress"] = n.ingressRules
	}
	if len(n.policyTypes) > 0 {
		values := []interface{}{}
		if n.policyTypes["Egress"] {
			values = append(values, "Egress")
		}
		if n.policyTypes["Ingress"] {
			values = append(values, "Ingress")
		}
		result["policyTypes"] = values
	}
	return result
}
