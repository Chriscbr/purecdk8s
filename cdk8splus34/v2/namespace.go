package cdk8splus34

import (
	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// Properties for `Namespace`.
type NamespaceProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
}

// In Kubernetes, namespaces provides a mechanism for isolating groups of resources within a single cluster.
//
// Names of resources need to be unique within a namespace, but not across namespaces. Namespace-based scoping is applicable only for namespaced objects (e.g. Deployments, Services, etc) and not for cluster-wide objects (e.g. StorageClass, Nodes, PersistentVolumes, etc).
type Namespace interface {
	Resource
	INamespaceSelector
	INetworkPolicyPeer
	// Return the configuration of this selector. See: INamespaceSelector.toNamespaceSelectorConfig ()
	ToNamespaceSelectorConfig() *NamespaceSelectorConfig
}

// Configuration for selecting namespaces.
type NamespaceSelectorConfig struct {
	// A selector to select namespaces by labels.
	LabelSelector LabelSelector `field:"optional" json:"labelSelector" yaml:"labelSelector"`
	// A list of names to select namespaces by names.
	Names *[]*string `field:"optional" json:"names" yaml:"names"`
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
	result.resourceBase.initialize(result, scope, id, "v1", "Namespace", "namespaces", props.Metadata, map[string]interface{}{"spec": map[string]interface{}{}})
	result.pods = Pods_All(result, jsii.String("Pods"), &PodsAllOptions{Namespaces: Namespaces_Select(result, jsii.String("Namespaces"), &NamespacesSelectOptions{Names: &[]*string{result.Name()}})})
	return result
}

func NewNamespace_Override(namespace Namespace, scope constructs.Construct, id *string, props *NamespaceProps) {
	applyOverride(namespace, NewNamespace(scope, id, props), "Namespace")
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func Namespace_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

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
