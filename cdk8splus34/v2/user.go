package cdk8splus34

import (
	"github.com/purecdk8s/purecdk8s/constructs/v10"
	"github.com/purecdk8s/purecdk8s/jsii"
)

type User interface {
	constructs.Construct
	ISubject
	ApiGroup() *string
	Kind() *string
	Name() *string
}

type userImpl struct {
	node constructs.Node
	name *string
}

func (u *userImpl) Node() constructs.Node {
	return u.node
}

func (u *userImpl) SetNodeInternal(node constructs.Node) {
	u.node = node
}

func (u *userImpl) ToString() *string {
	return u.node.Path()
}

func (u *userImpl) With(mixins ...constructs.IMixin) constructs.IConstruct {
	return u.node.With(mixins...)
}

func (u *userImpl) ApiGroup() *string {
	return jsii.String("rbac.authorization.k8s.io")
}

func (u *userImpl) Kind() *string {
	return jsii.String("User")
}

func (u *userImpl) Name() *string {
	return u.name
}

func (u *userImpl) ToSubjectConfiguration() *SubjectConfiguration {
	return &SubjectConfiguration{ApiGroup: u.ApiGroup(), Kind: u.Kind(), Name: u.Name()}
}

func User_FromName(scope constructs.Construct, id, name *string) User {
	if scope == nil || id == nil || name == nil {
		panic("scope, id and name are required")
	}
	result := &userImpl{name: name}
	constructs.NewConstruct_Override(result, scope, id)
	return result
}

func User_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}
