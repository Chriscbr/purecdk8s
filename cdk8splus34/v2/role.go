package cdk8splus34

import (
	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

type (
	// A reference to any Role or ClusterRole.
	IRole interface{ IResource }
	// Represents a cluster-level role.
	IClusterRole interface{ IResource }
	// Policy rule of a `Role.
	RolePolicyRule struct {
		// Verbs to allow.
		//
		// (e.g ['get', 'watch'])
		Verbs *[]*string `field:"required" json:"verbs" yaml:"verbs"`
		// Resources this rule applies to.
		Resources *[]IApiResource `field:"required" json:"resources" yaml:"resources"`
	}
)

// Policy rule of a `ClusterRole.
type ClusterRolePolicyRule struct {
	// Verbs to allow.
	//
	// (e.g ['get', 'watch'])
	Verbs *[]*string `field:"required" json:"verbs" yaml:"verbs"`
	// Endpoints this rule applies to.
	//
	// Can be either api resources or non api resources.
	Endpoints *[]IApiEndpoint `field:"required" json:"endpoints" yaml:"endpoints"`
}

// Properties for `Role`.
type RoleProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// A list of rules the role should allow. Default: [].
	Rules *[]*RolePolicyRule `field:"optional" json:"rules" yaml:"rules"`
}

// Properties for `ClusterRole`.
type ClusterRoleProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// A list of rules the role should allow. Default: [].
	Rules *[]*ClusterRolePolicyRule `field:"optional" json:"rules" yaml:"rules"`
	// Specify labels that should be used to locate ClusterRoles, whose rules will be automatically filled into this ClusterRole's rules.
	AggregationLabels *map[string]*string `field:"optional" json:"aggregationLabels" yaml:"aggregationLabels"`
}

// Role is a namespaced, logical grouping of PolicyRules that can be referenced as a unit by a RoleBinding.
type Role interface {
	Resource
	IRole
	// Rules associaated with this Role.
	//
	// Returns a copy, use `allow` to add rules.
	Rules() *[]*RolePolicyRule
	// Add permission to perform a list of HTTP verbs on a collection of resources. See: https://kubernetes.io/docs/reference/access-authn-authz/authorization/#determine-the-request-verb
	Allow(verbs *[]*string, resources ...IApiResource)
	// Add "create" permission for the resources.
	AllowCreate(resources ...IApiResource)
	// Add "get" permission for the resources.
	AllowGet(resources ...IApiResource)
	// Add "list" permission for the resources.
	AllowList(resources ...IApiResource)
	// Add "watch" permission for the resources.
	AllowWatch(resources ...IApiResource)
	// Add "update" permission for the resources.
	AllowUpdate(resources ...IApiResource)
	// Add "patch" permission for the resources.
	AllowPatch(resources ...IApiResource)
	// Add "delete" permission for the resources.
	AllowDelete(resources ...IApiResource)
	// Add "deletecollection" permission for the resources.
	AllowDeleteCollection(resources ...IApiResource)
	// Add "get", "list", and "watch" permissions for the resources.
	AllowRead(resources ...IApiResource)
	// Add "get", "list", "watch", "create", "update", "patch", "delete", and "deletecollection" permissions for the resources.
	AllowReadWrite(resources ...IApiResource)
	// Create a RoleBinding that binds the permissions in this Role to a list of subjects, that will only apply this role's namespace.
	Bind(subjects ...ISubject) RoleBinding
}

type roleImpl struct {
	resourceBase
	rules []*RolePolicyRule
}

func NewRole(scope constructs.Construct, id *string, props *RoleProps) Role {
	if props == nil {
		props = &RoleProps{}
	}
	result := &roleImpl{}
	manifest := map[string]interface{}{}
	result.resourceBase.initialize(result, scope, id, "rbac.authorization.k8s.io/v1", "Role", "roles", props.Metadata, manifest)
	if props.Rules != nil {
		for _, rule := range *props.Rules {
			if rule == nil {
				panic("role policy rule is required")
			}
			result.rules = append(result.rules, rule)
		}
	}
	manifest["rules"] = cdk8s.Lazy_Any(lazyProducer{produce: func() interface{} { return synthesizeRoleRules(result.rules) }})
	return result
}

func NewRole_Override(role Role, scope constructs.Construct, id *string, props *RoleProps) {
	applyOverride(role, NewRole(scope, id, props), "Role")
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func Role_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

// Imports a role from the cluster as a reference.
func Role_FromRoleName(scope constructs.Construct, id, name *string) IRole {
	return newImportedRole(scope, id, name, "Role", "roles")
}

func (r *roleImpl) Rules() *[]*RolePolicyRule {
	values := append([]*RolePolicyRule(nil), r.rules...)
	return &values
}

func (r *roleImpl) Allow(verbs *[]*string, resources ...IApiResource) {
	if verbs == nil {
		panic("verbs are required")
	}
	for _, resource := range resources {
		if resource == nil {
			panic("resource is required")
		}
	}
	values := append([]IApiResource(nil), resources...)
	r.rules = append(r.rules, &RolePolicyRule{Verbs: verbs, Resources: &values})
}

func (r *roleImpl) AllowCreate(resources ...IApiResource) {
	r.Allow(&[]*string{jsii.String("create")}, resources...)
}

func (r *roleImpl) AllowGet(resources ...IApiResource) {
	r.Allow(&[]*string{jsii.String("get")}, resources...)
}

func (r *roleImpl) AllowList(resources ...IApiResource) {
	r.Allow(&[]*string{jsii.String("list")}, resources...)
}

func (r *roleImpl) AllowWatch(resources ...IApiResource) {
	r.Allow(&[]*string{jsii.String("watch")}, resources...)
}

func (r *roleImpl) AllowUpdate(resources ...IApiResource) {
	r.Allow(&[]*string{jsii.String("update")}, resources...)
}

func (r *roleImpl) AllowPatch(resources ...IApiResource) {
	r.Allow(&[]*string{jsii.String("patch")}, resources...)
}

func (r *roleImpl) AllowDelete(resources ...IApiResource) {
	r.Allow(&[]*string{jsii.String("delete")}, resources...)
}

func (r *roleImpl) AllowDeleteCollection(resources ...IApiResource) {
	r.Allow(&[]*string{jsii.String("deletecollection")}, resources...)
}

func (r *roleImpl) AllowRead(resources ...IApiResource) {
	r.Allow(&[]*string{jsii.String("get"), jsii.String("list"), jsii.String("watch")}, resources...)
}

func (r *roleImpl) AllowReadWrite(resources ...IApiResource) {
	r.Allow(&[]*string{jsii.String("get"), jsii.String("list"), jsii.String("watch"), jsii.String("create"), jsii.String("update"), jsii.String("patch"), jsii.String("delete"), jsii.String("deletecollection")}, resources...)
}

func (r *roleImpl) Bind(subjects ...ISubject) RoleBinding {
	binding := NewRoleBinding(r, jsii.String("RoleBinding"+constructAddress(subjects)), &RoleBindingProps{Metadata: &cdk8s.ApiObjectMetadata{Namespace: r.Metadata().Namespace()}, Role: r})
	binding.AddSubjects(subjects...)
	return binding
}

// ClusterRole is a cluster level, logical grouping of PolicyRules that can be referenced as a unit by a RoleBinding or ClusterRoleBinding.
type ClusterRole interface {
	Resource
	IClusterRole
	IRole
	// Rules associaated with this Role.
	//
	// Returns a copy, use `allow` to add rules.
	Rules() *[]*ClusterRolePolicyRule
	// Add permission to perform a list of HTTP verbs on a collection of resources. See: https://kubernetes.io/docs/reference/access-authn-authz/authorization/#determine-the-request-verb
	Allow(verbs *[]*string, endpoints ...IApiEndpoint)
	// Add "create" permission for the resources.
	AllowCreate(endpoints ...IApiEndpoint)
	// Add "get" permission for the resources.
	AllowGet(endpoints ...IApiEndpoint)
	// Add "list" permission for the resources.
	AllowList(endpoints ...IApiEndpoint)
	// Add "watch" permission for the resources.
	AllowWatch(endpoints ...IApiEndpoint)
	// Add "update" permission for the resources.
	AllowUpdate(endpoints ...IApiEndpoint)
	// Add "patch" permission for the resources.
	AllowPatch(endpoints ...IApiEndpoint)
	// Add "delete" permission for the resources.
	AllowDelete(endpoints ...IApiEndpoint)
	// Add "deletecollection" permission for the resources.
	AllowDeleteCollection(endpoints ...IApiEndpoint)
	// Add "get", "list", and "watch" permissions for the resources.
	AllowRead(endpoints ...IApiEndpoint)
	// Add "get", "list", "watch", "create", "update", "patch", "delete", and "deletecollection" permissions for the resources.
	AllowReadWrite(endpoints ...IApiEndpoint)
	// Aggregate rules from roles matching this label selector.
	Aggregate(key, value *string)
	// Combines the rules of the argument ClusterRole into this ClusterRole using aggregation labels.
	Combine(role ClusterRole)
	// Create a ClusterRoleBinding that binds the permissions in this ClusterRole to a list of subjects, without namespace restrictions.
	Bind(subjects ...ISubject) ClusterRoleBinding
	// Create a RoleBinding that binds the permissions in this ClusterRole to a list of subjects, that will only apply to the given namespace.
	BindInNamespace(namespace *string, subjects ...ISubject) RoleBinding
}

type clusterRoleImpl struct {
	resourceBase
	rules  []*ClusterRolePolicyRule
	labels map[string]*string
}

func NewClusterRole(scope constructs.Construct, id *string, props *ClusterRoleProps) ClusterRole {
	if props == nil {
		props = &ClusterRoleProps{}
	}
	result := &clusterRoleImpl{labels: map[string]*string{}}
	manifest := map[string]interface{}{}
	result.resourceBase.initialize(result, scope, id, "rbac.authorization.k8s.io/v1", "ClusterRole", "clusterroles", props.Metadata, manifest)
	if props.Rules != nil {
		for _, rule := range *props.Rules {
			if rule == nil {
				panic("cluster role policy rule is required")
			}
			result.rules = append(result.rules, rule)
		}
	}
	if props.AggregationLabels != nil {
		for key, value := range *props.AggregationLabels {
			result.labels[key] = value
		}
	}
	manifest["rules"] = cdk8s.Lazy_Any(lazyProducer{produce: func() interface{} { return synthesizeClusterRoleRules(result.rules) }})
	manifest["aggregationRule"] = cdk8s.Lazy_Any(lazyProducer{produce: func() interface{} { return result.aggregationRule() }})
	return result
}

func NewClusterRole_Override(role ClusterRole, scope constructs.Construct, id *string, props *ClusterRoleProps) {
	applyOverride(role, NewClusterRole(scope, id, props), "ClusterRole")
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func ClusterRole_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

// Imports a role from the cluster as a reference.
func ClusterRole_FromClusterRoleName(scope constructs.Construct, id, name *string) IClusterRole {
	return newImportedRole(scope, id, name, "ClusterRole", "clusterroles")
}

func (r *clusterRoleImpl) Rules() *[]*ClusterRolePolicyRule {
	values := append([]*ClusterRolePolicyRule(nil), r.rules...)
	return &values
}

func (r *clusterRoleImpl) Allow(verbs *[]*string, endpoints ...IApiEndpoint) {
	if verbs == nil {
		panic("verbs are required")
	}
	for _, endpoint := range endpoints {
		if endpoint == nil {
			panic("endpoint is required")
		}
	}
	values := append([]IApiEndpoint(nil), endpoints...)
	r.rules = append(r.rules, &ClusterRolePolicyRule{Verbs: verbs, Endpoints: &values})
}

func (r *clusterRoleImpl) AllowCreate(values ...IApiEndpoint) {
	r.Allow(&[]*string{jsii.String("create")}, values...)
}

func (r *clusterRoleImpl) AllowGet(values ...IApiEndpoint) {
	r.Allow(&[]*string{jsii.String("get")}, values...)
}

func (r *clusterRoleImpl) AllowList(values ...IApiEndpoint) {
	r.Allow(&[]*string{jsii.String("list")}, values...)
}

func (r *clusterRoleImpl) AllowWatch(values ...IApiEndpoint) {
	r.Allow(&[]*string{jsii.String("watch")}, values...)
}

func (r *clusterRoleImpl) AllowUpdate(values ...IApiEndpoint) {
	r.Allow(&[]*string{jsii.String("update")}, values...)
}

func (r *clusterRoleImpl) AllowPatch(values ...IApiEndpoint) {
	r.Allow(&[]*string{jsii.String("patch")}, values...)
}

func (r *clusterRoleImpl) AllowDelete(values ...IApiEndpoint) {
	r.Allow(&[]*string{jsii.String("delete")}, values...)
}

func (r *clusterRoleImpl) AllowDeleteCollection(values ...IApiEndpoint) {
	r.Allow(&[]*string{jsii.String("deletecollection")}, values...)
}

func (r *clusterRoleImpl) AllowRead(values ...IApiEndpoint) {
	r.Allow(&[]*string{jsii.String("get"), jsii.String("list"), jsii.String("watch")}, values...)
}

func (r *clusterRoleImpl) AllowReadWrite(values ...IApiEndpoint) {
	r.Allow(&[]*string{jsii.String("get"), jsii.String("list"), jsii.String("watch"), jsii.String("create"), jsii.String("update"), jsii.String("patch"), jsii.String("delete"), jsii.String("deletecollection")}, values...)
}

func (r *clusterRoleImpl) Aggregate(key, value *string) {
	if key == nil || value == nil {
		panic("key and value are required")
	}
	r.labels[*key] = value
}

func (r *clusterRoleImpl) Combine(role ClusterRole) {
	if role == nil {
		panic("role is required")
	}
	key := jsii.String("cdk8s.cluster-role/aggregate-to-" + stringValue(cdk8s.Names_ToLabelValue(r, nil)))
	role.Metadata().AddLabel(key, jsii.String("true"))
	r.Aggregate(key, jsii.String("true"))
}

func (r *clusterRoleImpl) Bind(subjects ...ISubject) ClusterRoleBinding {
	binding := NewClusterRoleBinding(r, jsii.String("ClusterRoleBinding"+constructAddress(subjects)), &ClusterRoleBindingProps{Role: r})
	binding.AddSubjects(subjects...)
	return binding
}

func (r *clusterRoleImpl) BindInNamespace(namespace *string, subjects ...ISubject) RoleBinding {
	if namespace == nil {
		panic("namespace is required")
	}
	binding := NewRoleBinding(r, jsii.String("RoleBinding-"+*namespace), &RoleBindingProps{
		Metadata: &cdk8s.ApiObjectMetadata{Namespace: namespace},
		Role:     r,
	})
	binding.AddSubjects(subjects...)
	return binding
}

func (r *clusterRoleImpl) aggregationRule() interface{} {
	if len(r.labels) == 0 {
		return nil
	}
	return map[string]interface{}{"clusterRoleSelectors": []interface{}{map[string]interface{}{"matchLabels": r.labels}}}
}

func synthesizeRoleRules(rules []*RolePolicyRule) []interface{} {
	result := []interface{}{}
	for _, rule := range rules {
		if rule == nil || rule.Verbs == nil || rule.Resources == nil {
			panic("role policy rule verbs and resources are required")
		}
		for _, resource := range *rule.Resources {
			if resource == nil {
				panic("resource is required")
			}
			entry := map[string]interface{}{"verbs": *rule.Verbs, "apiGroups": []interface{}{apiGroupForRBAC(resource.ApiGroup())}}
			if resource.ResourceType() != nil {
				entry["resources"] = []interface{}{resource.ResourceType()}
			}
			if resource.ResourceName() != nil {
				entry["resourceNames"] = []interface{}{resource.ResourceName()}
			}
			result = append(result, entry)
		}
	}
	return result
}

func synthesizeClusterRoleRules(rules []*ClusterRolePolicyRule) []interface{} {
	result := []interface{}{}
	for _, rule := range rules {
		if rule == nil || rule.Verbs == nil || rule.Endpoints == nil {
			panic("cluster role policy rule verbs and endpoints are required")
		}
		for _, endpoint := range *rule.Endpoints {
			if endpoint == nil {
				panic("endpoint is required")
			}
			entry := map[string]interface{}{"verbs": *rule.Verbs}
			if endpoint.AsApiResource() != nil {
				resource := endpoint.AsApiResource()
				entry["apiGroups"] = []interface{}{apiGroupForRBAC(resource.ApiGroup())}
				if resource.ResourceType() != nil {
					entry["resources"] = []interface{}{resource.ResourceType()}
				}
				if resource.ResourceName() != nil {
					entry["resourceNames"] = []interface{}{resource.ResourceName()}
				}
			} else if endpoint.AsNonApiResource() != nil {
				entry["nonResourceURLs"] = []interface{}{endpoint.AsNonApiResource()}
			}
			result = append(result, entry)
		}
	}
	return result
}

type importedRole struct {
	node               constructs.Node
	name               *string
	kind, resourceType string
}

func (i *importedRole) Node() constructs.Node {
	return i.node
}

func (i *importedRole) SetNodeInternal(node constructs.Node) {
	i.node = node
}

func (i *importedRole) ToString() *string {
	return i.node.Path()
}

func (i *importedRole) With(m ...constructs.IMixin) constructs.IConstruct {
	return i.node.With(m...)
}

func (i *importedRole) ApiVersion() *string {
	return jsii.String("rbac.authorization.k8s.io/v1")
}

func (i *importedRole) ApiGroup() *string {
	return jsii.String("rbac.authorization.k8s.io")
}

func (i *importedRole) Kind() *string {
	return jsii.String(i.kind)
}

func (i *importedRole) Name() *string {
	return i.name
}

func (i *importedRole) ResourceName() *string {
	return i.name
}

func (i *importedRole) ResourceType() *string {
	return jsii.String(i.resourceType)
}

func newImportedRole(scope constructs.Construct, id, name *string, kind, resourceType string) *importedRole {
	if scope == nil || id == nil || name == nil {
		panic("scope, id and name are required")
	}
	result := &importedRole{name: name, kind: kind, resourceType: resourceType}
	constructs.NewConstruct_Override(result, scope, id)
	return result
}
