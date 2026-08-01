package cdk8splus34

import (
	"github.com/purecdk8s/purecdk8s/constructs/v10"
	"github.com/purecdk8s/purecdk8s/jsii"
)

// PodConnections configures NetworkPolicy connections for a workload.
type PodConnections interface {
	AllowTo(peer IPodSelector, options *PodConnectionsAllowToOptions)
	Isolate()
}

// PodConnectionsAllowToOptions configures an allowed connection. The native
// implementation currently supports the default behavior used by cdk8s+.
type PodConnectionsAllowToOptions struct{}

type connectableWorkload interface {
	Resource
	IPodSelector
	Containers() *[]Container
}

type podConnectionsImpl struct{ workload connectableWorkload }

func (c *podConnectionsImpl) Isolate() {
	newNetworkPolicy(c.workload, "DefaultDenyAll", c.workload, map[string]interface{}{
		"policyTypes": []interface{}{"Egress", "Ingress"},
	})
}

func (c *podConnectionsImpl) AllowTo(peer IPodSelector, options *PodConnectionsAllowToOptions) {
	if peer == nil {
		panic("peer is required")
	}
	if options != nil {
		panic("pod connection options are not implemented")
	}
	ports := networkPolicyPorts(peer)
	peerAddress := stringValue(peer.Node().Addr())
	newNetworkPolicy(c.workload, "AllowEgress"+peerAddress, c.workload, map[string]interface{}{
		"egress": []interface{}{map[string]interface{}{
			"ports": ports,
			"to":    []interface{}{map[string]interface{}{"podSelector": podSelectorManifest(peer)}},
		}},
		"policyTypes": []interface{}{"Egress"},
	})
	newNetworkPolicy(c.workload, "AllowIngressundefined"+peerAddress, peer, map[string]interface{}{
		"ingress": []interface{}{map[string]interface{}{
			"from":  []interface{}{map[string]interface{}{"podSelector": podSelectorManifest(c.workload)}},
			"ports": ports,
		}},
		"policyTypes": []interface{}{"Ingress"},
	})
}

func networkPolicyPorts(peer IPodSelector) []interface{} {
	workload, ok := peer.(interface{ Containers() *[]Container })
	if !ok {
		return []interface{}{}
	}
	ports := []interface{}{}
	for _, container := range *workload.Containers() {
		for _, port := range *container.Ports() {
			entry := map[string]interface{}{"port": port.Number}
			if port.Protocol != "" {
				entry["protocol"] = string(port.Protocol)
			} else {
				entry["protocol"] = "TCP"
			}
			ports = append(ports, entry)
		}
	}
	return ports
}

func newNetworkPolicy(scope constructs.Construct, id string, selector IPodSelector, spec map[string]interface{}) {
	policy := &networkPolicyImpl{}
	policySpec := map[string]interface{}{}
	for key, value := range spec {
		policySpec[key] = value
	}
	policySpec["podSelector"] = podSelectorManifest(selector)
	manifest := map[string]interface{}{"spec": policySpec}
	policy.resourceBase.initialize(policy, scope, jsii.String(id), "networking.k8s.io/v1", "NetworkPolicy", "networkpolicies", nil, manifest)
}

type networkPolicyImpl struct{ resourceBase }

func podSelectorManifest(selector IPodSelector) map[string]interface{} {
	config := selector.ToPodSelectorConfig()
	labels := map[string]interface{}{}
	if config != nil && config.LabelSelector != nil && config.LabelSelector.Labels != nil {
		for key, value := range *config.LabelSelector.Labels {
			labels[key] = value
		}
	}
	return map[string]interface{}{"matchLabels": labels}
}
