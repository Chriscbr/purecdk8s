package cdk8splus34

import (
	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

type (
	IRole          interface{ IResource }
	IClusterRole   interface{ IResource }
	RolePolicyRule struct {
		Verbs     *[]*string      `field:"required" json:"verbs" yaml:"verbs"`
		Resources *[]IApiResource `field:"required" json:"resources" yaml:"resources"`
	}
)

type ClusterRolePolicyRule struct {
	Verbs     *[]*string      `field:"required" json:"verbs" yaml:"verbs"`
	Endpoints *[]IApiEndpoint `field:"required" json:"endpoints" yaml:"endpoints"`
}

type RoleProps struct {
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	Rules    *[]*RolePolicyRule       `field:"optional" json:"rules" yaml:"rules"`
}

type ClusterRoleProps struct {
	Metadata          *cdk8s.ApiObjectMetadata  `field:"optional" json:"metadata" yaml:"metadata"`
	Rules             *[]*ClusterRolePolicyRule `field:"optional" json:"rules" yaml:"rules"`
	AggregationLabels *map[string]*string       `field:"optional" json:"aggregationLabels" yaml:"aggregationLabels"`
}

type Role interface {
	Resource
	IRole
	Rules() *[]*RolePolicyRule
	Allow(verbs *[]*string, resources ...IApiResource)
	AllowCreate(resources ...IApiResource)
	AllowGet(resources ...IApiResource)
	AllowList(resources ...IApiResource)
	AllowWatch(resources ...IApiResource)
	AllowUpdate(resources ...IApiResource)
	AllowPatch(resources ...IApiResource)
	AllowDelete(resources ...IApiResource)
	AllowDeleteCollection(resources ...IApiResource)
	AllowRead(resources ...IApiResource)
	AllowReadWrite(resources ...IApiResource)
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

func Role_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

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

type ClusterRole interface {
	Resource
	IClusterRole
	IRole
	Rules() *[]*ClusterRolePolicyRule
	Allow(verbs *[]*string, endpoints ...IApiEndpoint)
	AllowCreate(endpoints ...IApiEndpoint)
	AllowGet(endpoints ...IApiEndpoint)
	AllowList(endpoints ...IApiEndpoint)
	AllowWatch(endpoints ...IApiEndpoint)
	AllowUpdate(endpoints ...IApiEndpoint)
	AllowPatch(endpoints ...IApiEndpoint)
	AllowDelete(endpoints ...IApiEndpoint)
	AllowDeleteCollection(endpoints ...IApiEndpoint)
	AllowRead(endpoints ...IApiEndpoint)
	AllowReadWrite(endpoints ...IApiEndpoint)
	Aggregate(key, value *string)
	Combine(role ClusterRole)
	Bind(subjects ...ISubject) ClusterRoleBinding
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

func ClusterRole_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

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
