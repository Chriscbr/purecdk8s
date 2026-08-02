package cdk8splus34

import (
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// Represents a user.
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

// Reference a user in the cluster by name.
func User_FromName(scope constructs.Construct, id, name *string) User {
	if scope == nil || id == nil || name == nil {
		panic("scope, id and name are required")
	}
	result := &userImpl{name: name}
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
func User_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}
