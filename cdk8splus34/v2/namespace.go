package cdk8splus34

import (
	"github.com/purecdk8s/purecdk8s/cdk8s/v2"
	"github.com/purecdk8s/purecdk8s/constructs/v10"
	"github.com/purecdk8s/purecdk8s/jsii"
)

// NamespaceProps configures a Kubernetes Namespace.
type NamespaceProps struct {
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
}

// Namespace is a native Kubernetes Namespace construct.
type Namespace interface {
	Resource
	INamespaceSelector
	INetworkPolicyPeer
	ToNamespaceSelectorConfig() *NamespaceSelectorConfig
}

type NamespaceSelectorConfig struct {
	LabelSelector LabelSelector `field:"optional" json:"labelSelector" yaml:"labelSelector"`
	Names         *[]*string    `field:"optional" json:"names" yaml:"names"`
}

type namespaceImpl struct {
	resourceBase
	pods Pods
}

func NewNamespace(scope constructs.Construct, id *string, props *NamespaceProps) Namespace {
	if props == nil {
		props = &NamespaceProps{}
	}
	result := &namespaceImpl{}
	result.resourceBase.initialize(result, scope, id, "v1", "Namespace", "namespaces", props.Metadata, map[string]interface{}{})
	result.pods = Pods_All(result, jsii.String("Pods"), &PodsAllOptions{Namespaces: Namespaces_Select(result, jsii.String("Namespaces"), &NamespacesSelectOptions{Names: &[]*string{result.Name()}})})
	return result
}

func NewNamespace_Override(namespace Namespace, scope constructs.Construct, id *string, props *NamespaceProps) {
	applyOverride(namespace, NewNamespace(scope, id, props), "Namespace")
}

func Namespace_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

// Namespace_NAME_LABEL is the Kubernetes label used to select a namespace by
// its name.
func Namespace_NAME_LABEL() *string {
	return jsii.String("kubernetes.io/metadata.name")
}

func (n *namespaceImpl) ToNamespaceSelectorConfig() *NamespaceSelectorConfig {
	name := n.Name()
	return &NamespaceSelectorConfig{Names: &[]*string{name}}
}

func (n *namespaceImpl) ToNetworkPolicyPeerConfig() *NetworkPolicyPeerConfig {
	return n.pods.ToNetworkPolicyPeerConfig()
}

func (n *namespaceImpl) ToPodSelector() IPodSelector {
	return n.pods.ToPodSelector()
}
