package cdk8splus34

import (
	"github.com/purecdk8s/purecdk8s/cdk8s/v2"
	"github.com/purecdk8s/purecdk8s/constructs/v10"
)

// NamespaceProps configures a Kubernetes Namespace.
type NamespaceProps struct {
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
}

// Namespace is a native Kubernetes Namespace construct.
type Namespace interface {
	Resource
	ToNamespaceSelectorConfig() *NamespaceSelectorConfig
}

type NamespaceSelectorConfig struct {
	Namespaces *[]*string `field:"optional" json:"namespaces" yaml:"namespaces"`
}

type namespaceImpl struct{ resourceBase }

func NewNamespace(scope constructs.Construct, id *string, props *NamespaceProps) Namespace {
	if props == nil {
		props = &NamespaceProps{}
	}
	result := &namespaceImpl{}
	result.resourceBase.initialize(result, scope, id, "v1", "Namespace", "namespaces", props.Metadata, map[string]interface{}{})
	return result
}

func NewNamespace_Override(namespace Namespace, scope constructs.Construct, id *string, props *NamespaceProps) {
	panic("native cdk8splus34 overrides are not implemented")
}

func Namespace_IsConstruct(x interface{}) *bool { return constructs.Construct_IsConstruct(x) }

func (n *namespaceImpl) ToNamespaceSelectorConfig() *NamespaceSelectorConfig {
	name := n.Name()
	return &NamespaceSelectorConfig{Namespaces: &[]*string{name}}
}
