package cdk8splus34

import (
	"github.com/purecdk8s/purecdk8s/cdk8s/v2"
	"github.com/purecdk8s/purecdk8s/constructs/v10"
	"github.com/purecdk8s/purecdk8s/jsii"
)

const podAddressLabel = "cdk8s.io/metadata.addr"

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

type lazyProducer struct{ produce func() interface{} }

func (p lazyProducer) Produce() interface{} { return p.produce() }

// IApiResource represents a Kubernetes API resource or collection.
type IApiResource interface {
	ApiGroup() *string
	ResourceName() *string
	ResourceType() *string
}

// IApiEndpoint is either an API resource or a non-resource URL.
type IApiEndpoint interface {
	AsApiResource() IApiResource
	AsNonApiResource() *string
}

// IResource is the common contract of cdk8s+ resources.
type IResource interface {
	constructs.IConstruct
	IApiResource
	ApiVersion() *string
	Kind() *string
	Name() *string
}

// Resource is the base interface of all high-level Kubernetes resources.
type Resource interface {
	constructs.Construct
	IApiEndpoint
	IApiResource
	IResource
	ApiObject() cdk8s.ApiObject
	Metadata() cdk8s.ApiObjectMetadataDefinition
	Permissions() ResourcePermissions
}

// ResourceProps contains common persisted-object properties.
type ResourceProps struct {
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
}

// ResourcePermissions controls RBAC grants for a resource.
type ResourcePermissions interface {
	GrantRead(subjects ...ISubject) RoleBinding
	GrantReadWrite(subjects ...ISubject) RoleBinding
}

type resourceBase struct {
	node         constructs.Node
	apiObject    cdk8s.ApiObject
	resourceType string
	resource     IResource
	manifest     map[string]interface{}
	permissions  ResourcePermissions
}

func (r *resourceBase) Node() constructs.Node { return r.node }

// SetNodeInternal lets constructs initialize native subclasses without a JSII
// subclass proxy.
func (r *resourceBase) SetNodeInternal(node constructs.Node) { r.node = node }

func (r *resourceBase) ToString() *string {
	if r.node == nil {
		return jsii.String("<root>")
	}
	return r.node.Path()
}

func (r *resourceBase) With(mixins ...constructs.IMixin) constructs.IConstruct {
	return r.node.With(mixins...)
}

func (r *resourceBase) initialize(host constructs.Construct, scope constructs.Construct, id *string, apiVersion, kind, resourceType string, metadata *cdk8s.ApiObjectMetadata, manifest map[string]interface{}) {
	if scope == nil || id == nil {
		panic("scope and id are required")
	}
	constructs.NewConstruct_Override(host, scope, id)
	r.initializeApiObject(host, apiVersion, kind, resourceType, metadata, manifest)
}

// initializeApiObject attaches the Kubernetes API object after a resource has
// already been initialized as a construct. StatefulSet uses this to create its
// required headless Service before its own API object.
func (r *resourceBase) initializeApiObject(host constructs.Construct, apiVersion, kind, resourceType string, metadata *cdk8s.ApiObjectMetadata, manifest map[string]interface{}) {
	r.resourceType = resourceType
	r.manifest = manifest
	r.resource, _ = host.(IResource)
	if resource, ok := host.(Resource); ok {
		r.permissions = &resourcePermissions{instance: resource}
	}
	r.apiObject = cdk8s.NewApiObjectWithManifest(host, jsii.String("Resource"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String(apiVersion),
		Kind:       jsii.String(kind),
		Metadata:   metadata,
	}, manifest)
}

func (r *resourceBase) ApiObject() cdk8s.ApiObject                  { return r.apiObject }
func (r *resourceBase) ApiVersion() *string                         { return r.apiObject.ApiVersion() }
func (r *resourceBase) ApiGroup() *string                           { return r.apiObject.ApiGroup() }
func (r *resourceBase) Kind() *string                               { return r.apiObject.Kind() }
func (r *resourceBase) Metadata() cdk8s.ApiObjectMetadataDefinition { return r.apiObject.Metadata() }
func (r *resourceBase) Name() *string                               { return r.apiObject.Name() }
func (r *resourceBase) ResourceName() *string                       { return r.apiObject.Name() }
func (r *resourceBase) ResourceType() *string                       { return jsii.String(r.resourceType) }
func (r *resourceBase) AsApiResource() IApiResource                 { return r.resource }
func (r *resourceBase) AsNonApiResource() *string                   { return nil }
func (r *resourceBase) Permissions() ResourcePermissions            { return r.permissions }

func metadataMap(metadata *cdk8s.ApiObjectMetadata) map[string]interface{} {
	if metadata == nil {
		return map[string]interface{}{}
	}
	result := map[string]interface{}{}
	if metadata.Name != nil {
		result["name"] = metadata.Name
	}
	if metadata.Namespace != nil {
		result["namespace"] = metadata.Namespace
	}
	if metadata.Labels != nil {
		result["labels"] = *metadata.Labels
	}
	if metadata.Annotations != nil {
		result["annotations"] = *metadata.Annotations
	}
	if metadata.Finalizers != nil {
		result["finalizers"] = *metadata.Finalizers
	}
	if metadata.OwnerReferences != nil {
		result["ownerReferences"] = *metadata.OwnerReferences
	}
	return result
}

func copyLabels(values *map[string]*string) map[string]interface{} {
	result := map[string]interface{}{}
	if values == nil {
		return result
	}
	for key, value := range *values {
		result[key] = value
	}
	return result
}
