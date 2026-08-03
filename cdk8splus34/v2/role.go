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

type roleRules[T, R any] struct {
	rules     []R
	valueName string
	newRule   func(*[]*string, *[]T) R
}

func newRoleRules[T, R any](valueName string, newRule func(*[]*string, *[]T) R) roleRules[T, R] {
	return roleRules[T, R]{valueName: valueName, newRule: newRule}
}

func (r *roleRules[T, R]) Rules() *[]R {
	values := append([]R(nil), r.rules...)
	return &values
}

func (r *roleRules[T, R]) Allow(verbs *[]*string, values ...T) {
	if verbs == nil {
		panic("verbs are required")
	}
	for _, value := range values {
		if any(value) == nil {
			panic(r.valueName + " is required")
		}
	}
	values = append([]T(nil), values...)
	r.rules = append(r.rules, r.newRule(verbs, &values))
}

func (r *roleRules[T, R]) allowVerbs(values []T, verbs ...string) {
	verbPointers := make([]*string, len(verbs))
	for index, verb := range verbs {
		verbPointers[index] = jsii.String(verb)
	}
	r.Allow(&verbPointers, values...)
}

func (r *roleRules[T, R]) AllowCreate(values ...T) { r.allowVerbs(values, "create") }
func (r *roleRules[T, R]) AllowGet(values ...T)    { r.allowVerbs(values, "get") }
func (r *roleRules[T, R]) AllowList(values ...T)   { r.allowVerbs(values, "list") }
func (r *roleRules[T, R]) AllowWatch(values ...T)  { r.allowVerbs(values, "watch") }
func (r *roleRules[T, R]) AllowUpdate(values ...T) { r.allowVerbs(values, "update") }
func (r *roleRules[T, R]) AllowPatch(values ...T)  { r.allowVerbs(values, "patch") }
func (r *roleRules[T, R]) AllowDelete(values ...T) { r.allowVerbs(values, "delete") }
func (r *roleRules[T, R]) AllowDeleteCollection(values ...T) {
	r.allowVerbs(values, "deletecollection")
}

func (r *roleRules[T, R]) AllowRead(values ...T) {
	r.allowVerbs(values, "get", "list", "watch")
}

func (r *roleRules[T, R]) AllowReadWrite(values ...T) {
	r.allowVerbs(values, "get", "list", "watch", "create", "update", "patch", "delete", "deletecollection")
}

type roleImpl struct {
	resourceBase
	roleRules[IApiResource, *RolePolicyRule]
}

func NewRole(scope constructs.Construct, id *string, props *RoleProps) Role {
	if props == nil {
		props = &RoleProps{}
	}
	result := &roleImpl{roleRules: newRoleRules("resource", func(verbs *[]*string, resources *[]IApiResource) *RolePolicyRule {
		return &RolePolicyRule{Verbs: verbs, Resources: resources}
	})}
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
	roleRules[IApiEndpoint, *ClusterRolePolicyRule]
	labels map[string]*string
}

func NewClusterRole(scope constructs.Construct, id *string, props *ClusterRoleProps) ClusterRole {
	if props == nil {
		props = &ClusterRoleProps{}
	}
	result := &clusterRoleImpl{
		roleRules: newRoleRules("endpoint", func(verbs *[]*string, endpoints *[]IApiEndpoint) *ClusterRolePolicyRule {
			return &ClusterRolePolicyRule{Verbs: verbs, Endpoints: endpoints}
		}),
		labels: map[string]*string{},
	}
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
				entry["resourceNames"] = []interface{}{}
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
	importedResourceBase
}

func newImportedRole(scope constructs.Construct, id, name *string, kind, resourceType string) *importedRole {
	if scope == nil || id == nil || name == nil {
		panic("scope, id and name are required")
	}
	result := &importedRole{importedResourceBase: newImportedResourceBase(name, "rbac.authorization.k8s.io/v1", "rbac.authorization.k8s.io", kind, resourceType)}
	constructs.NewConstruct_Override(result, scope, id)
	return result
}
