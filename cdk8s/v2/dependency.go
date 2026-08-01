package cdk8s

import (
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/purecdk8s/purecdk8s/constructs/v10"
)

type dependencyVertexImpl struct {
	value    constructs.IConstruct
	children []DependencyVertex
	parents  []DependencyVertex
	host     DependencyVertex
}

func NewDependencyVertex(value constructs.IConstruct) DependencyVertex {
	result := &dependencyVertexImpl{value: value}
	result.host = result
	return result
}

func NewDependencyVertex_Override(vertex DependencyVertex, value constructs.IConstruct) {
	if vertex == nil {
		panic("parameter vertex is required, but nil was provided")
	}
	implementation := &dependencyVertexImpl{value: value, host: vertex}
	if target, ok := vertex.(*dependencyVertexImpl); ok {
		*target = *implementation
		return
	}
	if !setEmbeddedImplementation(vertex, implementation) {
		panic("cdk8s: DependencyVertex override must embed cdk8s.DependencyVertex")
	}
}

func (v *dependencyVertexImpl) Inbound() *[]DependencyVertex {
	result := append([]DependencyVertex(nil), v.parents...)
	return &result
}

func (v *dependencyVertexImpl) Outbound() *[]DependencyVertex {
	result := append([]DependencyVertex(nil), v.children...)
	return &result
}

func (v *dependencyVertexImpl) Value() constructs.IConstruct {
	return v.value
}

func (v *dependencyVertexImpl) AddChild(dependency DependencyVertex) {
	if dependency == nil {
		panic("parameter dep is required, but nil was provided")
	}
	self := v.self()
	cycle := dependencyRoute(dependency, self)
	if len(cycle) != 0 {
		cycle = append(cycle, dependency)
		paths := make([]string, 0, len(cycle))
		for _, vertex := range cycle {
			if value := vertex.Value(); value != nil {
				paths = append(paths, dependencyString(value.Node().Path()))
			}
		}
		panic("Dependency cycle detected: " + strings.Join(paths, " => "))
	}

	if !containsDependencyVertex(v.children, dependency) {
		v.children = append(v.children, dependency)
	}
	if concrete := nativeDependencyVertex(dependency); concrete != nil {
		if !containsDependencyVertex(concrete.parents, self) {
			concrete.parents = append(concrete.parents, self)
		}
	}
}

func nativeDependencyVertex(vertex DependencyVertex) *dependencyVertexImpl {
	if concrete, ok := vertex.(*dependencyVertexImpl); ok {
		return concrete
	}
	value := reflect.ValueOf(vertex)
	if !value.IsValid() || value.Kind() != reflect.Pointer || value.IsNil() {
		return nil
	}
	value = value.Elem()
	if value.Kind() != reflect.Struct {
		return nil
	}
	for index := 0; index < value.NumField(); index++ {
		fieldInfo := value.Type().Field(index)
		field := value.Field(index)
		if !fieldInfo.Anonymous || !field.CanInterface() || field.Kind() != reflect.Interface || field.IsNil() {
			continue
		}
		if concrete, ok := field.Interface().(*dependencyVertexImpl); ok {
			return concrete
		}
	}
	return nil
}

func (v *dependencyVertexImpl) Topology() *[]constructs.IConstruct {
	found := make(map[*dependencyVertexImpl]bool)
	foundOther := make([]DependencyVertex, 0)
	topology := make([]constructs.IConstruct, 0)

	var visit func(DependencyVertex)
	visit = func(vertex DependencyVertex) {
		for _, child := range *vertex.Outbound() {
			visit(child)
		}
		if concrete, ok := vertex.(*dependencyVertexImpl); ok {
			if found[concrete] {
				return
			}
			found[concrete] = true
		} else {
			if containsDependencyVertex(foundOther, vertex) {
				return
			}
			foundOther = append(foundOther, vertex)
		}
		if value := vertex.Value(); value != nil {
			topology = append(topology, value)
		}
	}
	visit(v.self())
	return &topology
}

func (v *dependencyVertexImpl) self() DependencyVertex {
	if v.host != nil {
		return v.host
	}
	return v
}

func dependencyRoute(source DependencyVertex, destination DependencyVertex) []DependencyVertex {
	route := make([]DependencyVertex, 0)
	var visit func(DependencyVertex) bool
	visit = func(vertex DependencyVertex) bool {
		route = append(route, vertex)
		found := false
		for _, child := range *vertex.Outbound() {
			if sameDependencyVertex(child, destination) {
				route = append(route, child)
				return true
			}
			found = visit(child)
		}
		if !found {
			route = route[:len(route)-1]
		}
		return found
	}
	visit(source)
	return route
}

func containsDependencyVertex(vertices []DependencyVertex, target DependencyVertex) bool {
	for _, vertex := range vertices {
		if sameDependencyVertex(vertex, target) {
			return true
		}
	}
	return false
}

func sameDependencyVertex(left, right DependencyVertex) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	leftValue, rightValue := reflect.ValueOf(left), reflect.ValueOf(right)
	if leftValue.Type() != rightValue.Type() || !leftValue.Type().Comparable() {
		return false
	}
	return leftValue.Interface() == rightValue.Interface()
}

type dependencyGraphImpl struct {
	root DependencyVertex
}

func NewDependencyGraph(node constructs.Node) DependencyGraph {
	if node == nil {
		panic("parameter node is required, but nil was provided")
	}
	return newDependencyGraph(node)
}

func NewDependencyGraph_Override(graph DependencyGraph, node constructs.Node) {
	if graph == nil {
		panic("parameter graph is required, but nil was provided")
	}
	if node == nil {
		panic("parameter node is required, but nil was provided")
	}
	implementation := newDependencyGraph(node)
	if target, ok := graph.(*dependencyGraphImpl); ok {
		*target = *implementation
		return
	}
	if !setEmbeddedImplementation(graph, implementation) {
		panic("cdk8s: DependencyGraph override must embed cdk8s.DependencyGraph")
	}
}

func newDependencyGraph(node constructs.Node) *dependencyGraphImpl {
	fosterParent := NewDependencyVertex(nil)
	all := append([]constructs.IConstruct(nil), (*node.FindAll(constructs.ConstructOrder_PREORDER))...)
	vertices := make(map[string]DependencyVertex, len(all))
	vertexKeys := make([]string, 0, len(all))
	for _, construct := range all {
		key := dependencyString(construct.Node().Path())
		vertex := NewDependencyVertex(construct)
		vertices[key] = vertex
		vertexKeys = append(vertexKeys, key)
	}

	for _, child := range all {
		source := vertices[dependencyString(child.Node().Path())]
		for _, dependency := range *child.Node().Dependencies() {
			target := vertices[dependencyString(dependency.Node().Path())]
			if target == nil {
				continue
			}
			source.AddChild(target)
		}
	}

	for _, key := range dependencyObjectKeyOrder(vertexKeys) {
		vertex := vertices[key]
		if len(*vertex.Inbound()) == 0 {
			fosterParent.AddChild(vertex)
		}
	}
	return &dependencyGraphImpl{root: fosterParent}
}

func (g *dependencyGraphImpl) Root() DependencyVertex {
	return g.root
}

func (g *dependencyGraphImpl) Topology() *[]constructs.IConstruct {
	return g.root.Topology()
}

func dependencyString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// DependencyGraph stores vertices in a JavaScript object upstream, so integer
// index keys are enumerated numerically before ordinary string keys.
func dependencyObjectKeyOrder(keys []string) []string {
	type indexedKey struct {
		key   string
		index uint64
	}
	indexed := make([]indexedKey, 0)
	ordinary := make([]string, 0)
	for _, key := range keys {
		index, err := strconv.ParseUint(key, 10, 32)
		if err == nil && index != (1<<32)-1 && strconv.FormatUint(index, 10) == key {
			indexed = append(indexed, indexedKey{key: key, index: index})
		} else {
			ordinary = append(ordinary, key)
		}
	}
	sort.Slice(indexed, func(left, right int) bool {
		return indexed[left].index < indexed[right].index
	})
	result := make([]string, 0, len(keys))
	for _, key := range indexed {
		result = append(result, key.key)
	}
	return append(result, ordinary...)
}
