package cdk8splus34

import (
	"sort"

	"github.com/purecdk8s/purecdk8s/cdk8s/v2"
	"github.com/purecdk8s/purecdk8s/constructs/v10"
	"github.com/purecdk8s/purecdk8s/jsii"
)

// ISubject can be bound to a Kubernetes RBAC role.
type ISubject interface {
	constructs.IConstruct
	toSubjectManifest() map[string]interface{}
}

// Group is an RBAC group subject.
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

func (g *groupImpl) Node() constructs.Node                { return g.node }
func (g *groupImpl) SetNodeInternal(node constructs.Node) { g.node = node }
func (g *groupImpl) ToString() *string                    { return g.node.Path() }
func (g *groupImpl) With(mixins ...constructs.IMixin) constructs.IConstruct {
	return g.node.With(mixins...)
}
func (g *groupImpl) ApiGroup() *string { return jsii.String("rbac.authorization.k8s.io") }
func (g *groupImpl) Kind() *string     { return jsii.String("Group") }
func (g *groupImpl) Name() *string     { return g.name }
func (g *groupImpl) toSubjectManifest() map[string]interface{} {
	return map[string]interface{}{"apiGroup": "rbac.authorization.k8s.io", "kind": "Group", "name": g.name}
}

func Group_FromName(scope constructs.Construct, id, name *string) Group {
	if scope == nil || id == nil || name == nil {
		panic("scope, id and name are required")
	}
	result := &groupImpl{name: name}
	constructs.NewConstruct_Override(result, scope, id)
	return result
}

func Group_IsConstruct(x interface{}) *bool { return constructs.Construct_IsConstruct(x) }

// RoleBinding is a namespaced role binding resource.
type RoleBinding interface{ Resource }

type roleBindingImpl struct{ resourceBase }

// resourcePermissions creates native Role and RoleBinding resources for grants.
type resourcePermissions struct{ instance Resource }

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
	role := &nativeRole{}
	roleManifest := map[string]interface{}{"rules": []interface{}{map[string]interface{}{
		"apiGroups":     []interface{}{apiGroupForRBAC(p.instance.ApiGroup())},
		"resourceNames": []interface{}{p.instance.ResourceName()},
		"resources":     []interface{}{p.instance.ResourceType()},
		"verbs":         stringSliceInterfaces(verbs),
	}}}
	role.resourceBase.initialize(role, p.instance, jsii.String("Role"+address), "rbac.authorization.k8s.io/v1", "Role", "roles", &cdk8s.ApiObjectMetadata{Namespace: p.instance.Metadata().Namespace()}, roleManifest)

	binding := &roleBindingImpl{}
	subjectValues := make([]interface{}, 0, len(subjects))
	for _, subject := range subjects {
		if subject == nil {
			panic("subject is required")
		}
		subjectValues = append(subjectValues, subject.toSubjectManifest())
	}
	bindingManifest := map[string]interface{}{
		"roleRef":  map[string]interface{}{"apiGroup": "rbac.authorization.k8s.io", "kind": "Role", "name": role.Name()},
		"subjects": subjectValues,
	}
	binding.resourceBase.initialize(binding, role, jsii.String("RoleBinding"+address), "rbac.authorization.k8s.io/v1", "RoleBinding", "rolebindings", &cdk8s.ApiObjectMetadata{Namespace: p.instance.Metadata().Namespace()}, bindingManifest)
	return binding
}

type nativeRole struct{ resourceBase }

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
