package cdk8splus34

import (
	"reflect"

	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

const podAddressLabel = "cdk8s.io/metadata.addr"

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

type lazyProducer struct{ produce func() interface{} }

func (p lazyProducer) Produce() interface{} {
	return p.produce()
}

// Represents a resource or collection of resources.
type IApiResource interface {
	// The group portion of the API version (e.g. `authorization.k8s.io`).
	ApiGroup() *string
	// The unique, namespace-global, name of an object inside the Kubernetes cluster.
	//
	// If this is omitted, the ApiResource should represent all objects of the given type.
	ResourceName() *string
	// The name of a resource type as it appears in the relevant API endpoint.
	//
	// Example:
	//   - "pods" or "pods/log"
	//
	// See: https://kubernetes.io/docs/reference/access-authn-authz/rbac/#referring-to-resources
	ResourceType() *string
}

// An API Endpoint can either be a resource descriptor (e.g /pods) or a non resource url (e.g /healthz). It must be one or the other, and not both.
type IApiEndpoint interface {
	// Return the IApiResource this object represents.
	AsApiResource() IApiResource
	// Return the non resource url this object represents.
	AsNonApiResource() *string
}

// Represents a resource.
type IResource interface {
	constructs.IConstruct
	IApiResource
	// The object's API version (e.g. "authorization.k8s.io/v1").
	ApiVersion() *string
	// The object kind (e.g. "Deployment").
	Kind() *string
	// The Kubernetes name of this resource.
	Name() *string
}

// Base class for all Kubernetes objects in stdk8s.
//
// Represents a single resource.
type Resource interface {
	constructs.Construct
	IApiEndpoint
	IApiResource
	IResource
	// The underlying cdk8s API object.
	ApiObject() cdk8s.ApiObject
	Metadata() cdk8s.ApiObjectMetadataDefinition
	Permissions() ResourcePermissions
}

// Initialization properties for resources.
type ResourceProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
}

// Controls permissions for operations on resources.
type ResourcePermissions interface {
	Instance() Resource
	// Grants the list of subjects permissions to read this resource.
	GrantRead(subjects ...ISubject) RoleBinding
	// Grants the list of subjects permissions to read and write this resource.
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

func (r *resourceBase) Node() constructs.Node {
	return r.node
}

// SetNodeInternal lets constructs initialize native subclasses without a JSII
// subclass proxy.
func (r *resourceBase) SetNodeInternal(node constructs.Node) {
	r.node = node
}

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
		r.permissions = NewResourcePermissions(resource)
	}
	r.apiObject = cdk8s.NewApiObjectWithManifest(host, jsii.String("Resource"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String(apiVersion),
		Kind:       jsii.String(kind),
		Metadata:   metadata,
	}, manifest)
}

func (r *resourceBase) ApiObject() cdk8s.ApiObject {
	return r.apiObject
}

func (r *resourceBase) ApiVersion() *string {
	return r.apiObject.ApiVersion()
}

func (r *resourceBase) ApiGroup() *string {
	return r.apiObject.ApiGroup()
}

func (r *resourceBase) Kind() *string {
	return r.apiObject.Kind()
}

func (r *resourceBase) Metadata() cdk8s.ApiObjectMetadataDefinition {
	return r.apiObject.Metadata()
}

func (r *resourceBase) Name() *string {
	return r.apiObject.Name()
}

func (r *resourceBase) ResourceName() *string {
	return r.apiObject.Name()
}

func (r *resourceBase) ResourceType() *string {
	return jsii.String(r.resourceType)
}

func (r *resourceBase) AsApiResource() IApiResource {
	return r.resource
}

func (r *resourceBase) AsNonApiResource() *string {
	return nil
}

func (r *resourceBase) Permissions() ResourcePermissions {
	return r.permissions
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func Resource_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func NewResource_Override(_ Resource, _ constructs.Construct, _ *string) {
	panic("Resource is an abstract base; use a concrete resource constructor")
}

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

// applyOverride provides the idiomatic Go equivalent of a JSII `_Override`
// constructor. A caller can embed one of the public interfaces in its own
// struct; this replaces that embedded implementation with the native value.
// Concrete resources are initialized through their ordinary constructor first,
// so they retain the same manifest behavior as the non-override path.
func applyOverride(target, implementation interface{}, name string) {
	if target == nil {
		panic(name + " override target is required")
	}
	targetValue := reflect.ValueOf(target)
	implementationValue := reflect.ValueOf(implementation)
	if targetValue.Kind() != reflect.Ptr || targetValue.IsNil() {
		panic(name + " override target must be a non-nil pointer")
	}
	if targetValue.Type() == implementationValue.Type() {
		targetValue.Elem().Set(implementationValue.Elem())
		return
	}
	value := targetValue.Elem()
	if value.Kind() != reflect.Struct {
		panic(name + " override target must embed the public interface")
	}
	for index := 0; index < value.NumField(); index++ {
		field := value.Field(index)
		if field.CanSet() && implementationValue.Type().AssignableTo(field.Type()) {
			field.Set(implementationValue)
			return
		}
	}
	panic(name + " override target must embed the public interface")
}
