package cdk8splus34

import (
	"github.com/Chriscbr/purecdk8s/constructs/v10"
)

// Represents a user.
type User interface {
	constructs.Construct
	ISubject
	ApiGroup() *string
	Kind() *string
	Name() *string
}

// Reference a user in the cluster by name.
func User_FromName(scope constructs.Construct, id, name *string) User {
	return newNamedSubject(scope, id, name, "User")
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
