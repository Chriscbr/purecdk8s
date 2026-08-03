package cdk8s_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	constructs "github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

func dependencyPaths(values []constructs.IConstruct) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, dependencyStringValue(value.Node().Path()))
	}
	return result
}

func dependencyStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func dependencyRequirePanic(t *testing.T, want string, callback func()) {
	t.Helper()
	defer func() {
		value := recover()
		if value == nil {
			t.Fatalf("expected panic %q", want)
		}
		if got := fmt.Sprint(value); got != want {
			t.Fatalf("panic = %q, want %q", got, want)
		}
	}()
	callback()
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/dependency.test.ts#L4
func TestDependencyTopologyReturnsCorrectOrder(t *testing.T) {
	root := constructs.NewRootConstruct(jsii.String("App"))
	group := constructs.NewConstruct(root, jsii.String("chart1"))
	obj1 := constructs.NewConstruct(group, jsii.String("obj1"))
	obj2 := constructs.NewConstruct(group, jsii.String("obj2"))
	obj3 := constructs.NewConstruct(group, jsii.String("obj3"))
	obj1.Node().AddDependency(obj2)
	obj2.Node().AddDependency(obj3)

	got := *cdk8s.NewDependencyGraph(group.Node()).Topology()
	want := []constructs.IConstruct{group, obj3, obj2, obj1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("topology = %#v (%#v), want %#v (%#v)", got, dependencyPaths(got), want, dependencyPaths(want))
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/dependency.test.ts#L21
func TestDependencyCycleDetection(t *testing.T) {
	root := constructs.NewRootConstruct(jsii.String("App"))
	group := constructs.NewConstruct(root, jsii.String("chart1"))
	obj1 := constructs.NewConstruct(group, jsii.String("obj1"))
	obj2 := constructs.NewConstruct(group, jsii.String("obj2"))
	obj3 := constructs.NewConstruct(group, jsii.String("obj3"))
	obj1.Node().AddDependency(obj2)
	obj2.Node().AddDependency(obj3)
	obj3.Node().AddDependency(obj1)

	paths := []string{
		dependencyStringValue(obj1.Node().Path()),
		dependencyStringValue(obj2.Node().Path()),
		dependencyStringValue(obj3.Node().Path()),
		dependencyStringValue(obj1.Node().Path()),
	}
	dependencyRequirePanic(t, "Dependency cycle detected: "+strings.Join(paths, " => "), func() {
		cdk8s.NewDependencyGraph(group.Node())
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/dependency.test.ts#L40
func TestDependencyValueOfRootIsNil(t *testing.T) {
	root := constructs.NewRootConstruct(jsii.String("App"))
	group := constructs.NewConstruct(root, jsii.String("chart1"))
	obj1 := constructs.NewConstruct(group, jsii.String("obj1"))
	obj2 := constructs.NewConstruct(group, jsii.String("obj2"))
	obj3 := constructs.NewConstruct(group, jsii.String("obj3"))
	obj1.Node().AddDependency(obj2)
	obj2.Node().AddDependency(obj3)

	if got := cdk8s.NewDependencyGraph(group.Node()).Root().Value(); got != nil {
		t.Fatalf("dummy root value = %#v, want nil", got)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/dependency.test.ts#L56
func TestDependencyChildrenOfRootContainAllOrphans(t *testing.T) {
	root := constructs.NewRootConstruct(jsii.String("App"))
	group := constructs.NewConstruct(root, jsii.String("chart1"))
	obj1 := constructs.NewConstruct(group, jsii.String("obj1"))
	obj2 := constructs.NewConstruct(group, jsii.String("obj2"))
	obj1.Node().AddDependency(obj2)

	got := make(map[constructs.IConstruct]bool)
	for _, child := range *cdk8s.NewDependencyGraph(group.Node()).Root().Outbound() {
		got[child.Value()] = true
	}
	want := map[constructs.IConstruct]bool{group: true, obj1: true}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("dummy root children = %#v, want %#v", got, want)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/dependency.test.ts#L76
func TestDependencyIgnoresCrossScopeNodes(t *testing.T) {
	root := constructs.NewRootConstruct(jsii.String("App"))
	group1 := constructs.NewConstruct(root, jsii.String("group1"))
	group2 := constructs.NewConstruct(root, jsii.String("group2"))
	obj1 := constructs.NewConstruct(group1, jsii.String("obj1"))
	obj2 := constructs.NewConstruct(group1, jsii.String("obj2"))
	obj3 := constructs.NewConstruct(group2, jsii.String("obj3"))
	obj1.Node().AddDependency(obj2)
	obj2.Node().AddDependency(obj3)

	got := *cdk8s.NewDependencyGraph(group1.Node()).Topology()
	want := []constructs.IConstruct{group1, obj2, obj1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("topology = %#v (%#v), want %#v (%#v)", got, dependencyPaths(got), want, dependencyPaths(want))
	}
}
