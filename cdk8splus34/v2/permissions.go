package cdk8splus34

import (
	"sort"

	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// Represents an object that can be used as a role binding subject.
type ISubject interface {
	constructs.IConstruct
	// Return the subject configuration.
	ToSubjectConfiguration() *SubjectConfiguration
}

// Subject contains a reference to the object or user identities a role binding applies to.
//
// This can either hold a direct API object reference, or a value for non-objects such as user and group names.
type SubjectConfiguration struct {
	// Kind of object being referenced.
	//
	// Values defined by this API group are "User", "Group", and "ServiceAccount". If the Authorizer does not recognized the kind value, the Authorizer should report an error.
	Kind *string `field:"required" json:"kind" yaml:"kind"`
	// Name of the object being referenced.
	Name *string `field:"required" json:"name" yaml:"name"`
	// APIGroup holds the API group of the referenced subject.
	//
	// Defaults to "" for ServiceAccount subjects. Defaults to "rbac.authorization.k8s.io" for User and Group subjects.
	ApiGroup *string `field:"optional" json:"apiGroup" yaml:"apiGroup"`
	// Namespace of the referenced object.
	//
	// If the object kind is non-namespace, such as "User" or "Group", and this value is not empty the Authorizer should report an error.
	Namespace *string `field:"optional" json:"namespace" yaml:"namespace"`
}

func subjectManifest(subject ISubject) map[string]interface{} {
	if subject == nil {
		panic("subject is required")
	}
	config := subject.ToSubjectConfiguration()
	if config == nil || config.Kind == nil || config.Name == nil {
		panic("subject configuration is required")
	}
	result := map[string]interface{}{"kind": config.Kind, "name": config.Name}
	if config.ApiGroup != nil {
		result["apiGroup"] = apiGroupForRBAC(config.ApiGroup)
	}
	if config.Namespace != nil {
		result["namespace"] = config.Namespace
	}
	return result
}

// Represents a group.
type Group interface {
	constructs.Construct
	ISubject
	ApiGroup() *string
	Kind() *string
	Name() *string
}

type groupImpl struct {
	node constructs.Node
	name *string
}

func (g *groupImpl) Node() constructs.Node {
	return g.node
}

func (g *groupImpl) SetNodeInternal(node constructs.Node) {
	g.node = node
}

func (g *groupImpl) ToString() *string {
	return g.node.Path()
}

func (g *groupImpl) With(mixins ...constructs.IMixin) constructs.IConstruct {
	return g.node.With(mixins...)
}

func (g *groupImpl) ApiGroup() *string {
	return jsii.String("rbac.authorization.k8s.io")
}

func (g *groupImpl) Kind() *string {
	return jsii.String("Group")
}

func (g *groupImpl) Name() *string {
	return g.name
}

func (g *groupImpl) ToSubjectConfiguration() *SubjectConfiguration {
	return &SubjectConfiguration{ApiGroup: g.ApiGroup(), Kind: g.Kind(), Name: g.Name()}
}

// Reference a group by name.
func Group_FromName(scope constructs.Construct, id, name *string) Group {
	if scope == nil || id == nil || name == nil {
		panic("scope, id and name are required")
	}
	result := &groupImpl{name: name}
	constructs.NewConstruct_Override(result, scope, id)
	return result
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func Group_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

// A RoleBinding grants permissions within a specific namespace to a user or set of users.
type RoleBinding interface {
	Resource
	Role() IRole
	Subjects() *[]ISubject
	// Adds a subject to the role.
	AddSubjects(subjects ...ISubject)
}

type roleBindingImpl struct {
	resourceBase
	role     IRole
	subjects []ISubject
}

// Properties for `RoleBinding`.
type RoleBindingProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// The role to bind to.
	//
	// A RoleBinding can reference a Role or a ClusterRole.
	Role IRole `field:"required" json:"role" yaml:"role"`
}

// Properties for `ClusterRoleBinding`.
type ClusterRoleBindingProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// The role to bind to.
	Role IClusterRole `field:"required" json:"role" yaml:"role"`
}

func NewRoleBinding(scope constructs.Construct, id *string, props *RoleBindingProps) RoleBinding {
	if props == nil || props.Role == nil {
		panic("role is required")
	}
	result := &roleBindingImpl{role: props.Role}
	manifest := map[string]interface{}{}
	result.resourceBase.initialize(result, scope, id, "rbac.authorization.k8s.io/v1", "RoleBinding", "rolebindings", props.Metadata, manifest)
	manifest["subjects"] = cdk8s.Lazy_Any(lazyProducer{produce: func() interface{} { return result.subjectManifests() }})
	manifest["roleRef"] = cdk8s.Lazy_Any(lazyProducer{produce: func() interface{} { return roleReference(result.role) }})
	return result
}

func NewRoleBinding_Override(binding RoleBinding, scope constructs.Construct, id *string, props *RoleBindingProps) {
	applyOverride(binding, NewRoleBinding(scope, id, props), "RoleBinding")
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func RoleBinding_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func (r *roleBindingImpl) Role() IRole {
	return r.role
}

func (r *roleBindingImpl) Subjects() *[]ISubject {
	values := append([]ISubject(nil), r.subjects...)
	return &values
}

func (r *roleBindingImpl) AddSubjects(subjects ...ISubject) {
	for _, subject := range subjects {
		if subject == nil {
			panic("subject is required")
		}
		r.subjects = append(r.subjects, subject)
	}
}

func (r *roleBindingImpl) subjectManifests() interface{} {
	result := make([]interface{}, 0, len(r.subjects))
	for _, subject := range r.subjects {
		result = append(result, subjectManifest(subject))
	}
	return result
}

// A ClusterRoleBinding grants permissions cluster-wide to a user or set of users.
type ClusterRoleBinding interface {
	Resource
	Role() IClusterRole
	Subjects() *[]ISubject
	// Adds a subject to the role.
	AddSubjects(subjects ...ISubject)
}

type clusterRoleBindingImpl struct {
	resourceBase
	role     IClusterRole
	subjects []ISubject
}

func NewClusterRoleBinding(scope constructs.Construct, id *string, props *ClusterRoleBindingProps) ClusterRoleBinding {
	if props == nil || props.Role == nil {
		panic("role is required")
	}
	result := &clusterRoleBindingImpl{role: props.Role}
	manifest := map[string]interface{}{}
	result.resourceBase.initialize(result, scope, id, "rbac.authorization.k8s.io/v1", "ClusterRoleBinding", "clusterrolebindings", props.Metadata, manifest)
	manifest["subjects"] = cdk8s.Lazy_Any(lazyProducer{produce: func() interface{} { return result.subjectManifests() }})
	manifest["roleRef"] = cdk8s.Lazy_Any(lazyProducer{produce: func() interface{} { return roleReference(result.role) }})
	return result
}

func NewClusterRoleBinding_Override(binding ClusterRoleBinding, scope constructs.Construct, id *string, props *ClusterRoleBindingProps) {
	applyOverride(binding, NewClusterRoleBinding(scope, id, props), "ClusterRoleBinding")
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func ClusterRoleBinding_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func (r *clusterRoleBindingImpl) Role() IClusterRole {
	return r.role
}

func (r *clusterRoleBindingImpl) Subjects() *[]ISubject {
	values := append([]ISubject(nil), r.subjects...)
	return &values
}

func (r *clusterRoleBindingImpl) AddSubjects(subjects ...ISubject) {
	for _, subject := range subjects {
		if subject == nil {
			panic("subject is required")
		}
		r.subjects = append(r.subjects, subject)
	}
}

func (r *clusterRoleBindingImpl) subjectManifests() interface{} {
	result := make([]interface{}, 0, len(r.subjects))
	for _, subject := range r.subjects {
		result = append(result, subjectManifest(subject))
	}
	return result
}

func roleReference(role IRole) map[string]interface{} {
	return map[string]interface{}{"apiGroup": apiGroupForRBAC(role.ApiGroup()), "kind": role.Kind(), "name": role.Name()}
}

// resourcePermissions creates native Role and RoleBinding resources for grants.
type resourcePermissions struct{ instance Resource }

func NewResourcePermissions(instance Resource) ResourcePermissions {
	if instance == nil {
		panic("instance is required")
	}
	return &resourcePermissions{instance: instance}
}

func NewResourcePermissions_Override(permissions ResourcePermissions, instance Resource) {
	applyOverride(permissions, NewResourcePermissions(instance), "ResourcePermissions")
}

func (p *resourcePermissions) Instance() Resource {
	return p.instance
}

func (p *resourcePermissions) GrantRead(subjects ...ISubject) RoleBinding {
	return p.grant([]string{"get", "list", "watch"}, subjects...)
}

func (p *resourcePermissions) GrantReadWrite(subjects ...ISubject) RoleBinding {
	return p.grant([]string{"get", "list", "watch", "create", "update", "patch", "delete", "deletecollection"}, subjects...)
}

func (p *resourcePermissions) grant(verbs []string, subjects ...ISubject) RoleBinding {
	if len(subjects) == 0 {
		panic("at least one subject is required")
	}
	address := constructAddress(subjects)
	verbsCopy := make([]*string, len(verbs))
	for index, verb := range verbs {
		verbsCopy[index] = jsii.String(verb)
	}
	role := NewRole(p.instance, jsii.String("Role"+address), &RoleProps{Metadata: &cdk8s.ApiObjectMetadata{Namespace: p.instance.Metadata().Namespace()}})
	role.Allow(&verbsCopy, p.instance.AsApiResource())
	binding := NewRoleBinding(role, jsii.String("RoleBinding"+address), &RoleBindingProps{Metadata: &cdk8s.ApiObjectMetadata{Namespace: p.instance.Metadata().Namespace()}, Role: role})
	binding.AddSubjects(subjects...)
	return binding
}

func constructAddress(subjects []ISubject) string {
	values := make([]string, 0, len(subjects))
	for _, subject := range subjects {
		if subject == nil {
			panic("subject is required")
		}
		values = append(values, stringValue(subject.Node().Addr()))
	}
	sort.Strings(values)
	return joinStrings(values)
}

func joinStrings(values []string) string {
	result := ""
	for _, value := range values {
		result += value
	}
	return result
}

func apiGroupForRBAC(group *string) interface{} {
	if group == nil || *group == "" || *group == "core" {
		return ""
	}
	return group
}

func stringSliceInterfaces(values []string) []interface{} {
	result := make([]interface{}, len(values))
	for index, value := range values {
		result[index] = value
	}
	return result
}
