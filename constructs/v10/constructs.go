// Package constructs implements the constructs v10 graph in pure Go.
package constructs

import (
	"crypto/sha1"
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type IDependable interface{}

type IConstruct interface {
	IDependable
	With(mixins ...IMixin) IConstruct
	Node() Node
}

type Construct interface {
	IConstruct
	Node() Node
	ToString() *string
	With(mixins ...IMixin) IConstruct
}

type RootConstruct interface {
	Construct
	Node() Node
	ToString() *string
	With(mixins ...IMixin) IConstruct
}

type IMixin interface {
	ApplyTo(construct IConstruct)
	Supports(construct IConstruct) *bool
}

type IValidation interface {
	Validate() *[]*string
}

type ConstructOrder string

const (
	ConstructOrder_PREORDER  ConstructOrder = "PREORDER"
	ConstructOrder_POSTORDER ConstructOrder = "POSTORDER"
)

type MetadataEntry struct {
	Data  interface{} `field:"required" json:"data" yaml:"data"`
	Type  *string     `field:"required" json:"type" yaml:"type"`
	Trace *[]*string  `field:"optional" json:"trace" yaml:"trace"`
}

type MetadataOptions struct {
	StackTrace         *bool       `field:"optional" json:"stackTrace" yaml:"stackTrace"`
	StackTraceOverride *[]*string  `field:"optional" json:"stackTraceOverride" yaml:"stackTraceOverride"`
	TraceFromFunction  interface{} `field:"optional" json:"traceFromFunction" yaml:"traceFromFunction"`
}

type Node interface {
	Addr() *string
	Children() *[]IConstruct
	DefaultChild() IConstruct
	SetDefaultChild(val IConstruct)
	Dependencies() *[]IConstruct
	Id() *string
	Locked() *bool
	Metadata() *[]*MetadataEntry
	Path() *string
	Root() IConstruct
	Scope() IConstruct
	Scopes() *[]IConstruct
	AddDependency(deps ...IDependable)
	AddMetadata(type_ *string, data interface{}, options *MetadataOptions)
	AddValidation(validation IValidation)
	FindAll(order ConstructOrder) *[]IConstruct
	FindChild(id *string) IConstruct
	GetAllContext(defaults *map[string]interface{}) interface{}
	GetContext(key *string) interface{}
	Lock()
	SetContext(key *string, value interface{})
	TryFindChild(id *string) IConstruct
	TryGetContext(key *string) interface{}
	TryRemoveChild(childName *string) *bool
	Validate() *[]*string
	With(mixins ...IMixin) IConstruct
}

type constructImpl struct {
	node Node
}

func (c *constructImpl) Node() Node { return c.node }

func (c *constructImpl) setNode(node Node) { c.node = node }

func (c *constructImpl) SetNodeInternal(node Node) { c.node = node }

func (c *constructImpl) ToString() *string {
	if c.node == nil || value(c.node.Path()) == "" {
		result := "<root>"
		return &result
	}
	return c.node.Path()
}

func (c *constructImpl) With(mixins ...IMixin) IConstruct {
	return c.node.With(mixins...)
}

type nodeSetter interface {
	SetNodeInternal(Node)
}

type nodeImpl struct {
	host         Construct
	scope        IConstruct
	id           string
	children     map[string]IConstruct
	childOrder   []string
	defaultChild IConstruct
	hasDefault   bool
	dependencies []IDependable
	context      map[string]interface{}
	metadata     []*MetadataEntry
	validations  []IValidation
	locked       bool
	addr         string
}

func NewConstruct(scope Construct, id *string) Construct {
	if scope == nil {
		panic("parameter scope is required, but nil was provided")
	}
	if id == nil {
		panic("parameter id is required, but nil was provided")
	}
	result := &constructImpl{}
	NewConstruct_Override(result, scope, id)
	return result
}

func NewConstruct_Override(c Construct, scope Construct, id *string) {
	if c == nil {
		panic("parameter c is required, but nil was provided")
	}
	if scope == nil {
		panic("parameter scope is required, but nil was provided")
	}
	if id == nil {
		panic("parameter id is required, but nil was provided")
	}
	initializeConstruct(c, scope, id)
}

func NewRootConstruct(id *string) RootConstruct {
	result := &constructImpl{}
	NewRootConstruct_Override(result, id)
	return result
}

func NewRootConstruct_Override(r RootConstruct, id *string) {
	if r == nil {
		panic("parameter r is required, but nil was provided")
	}
	initializeConstruct(r, nil, id)
}

func initializeConstruct(host Construct, scope IConstruct, id *string) {
	setter, ok := host.(nodeSetter)
	if ok {
		node := newNode(host, scope, id)
		setter.SetNodeInternal(node)
		Dependable_Implement(host, &singleDependable{root: host})
		return
	}

	// JSII subclassing in Go embeds constructs.Construct in the user's struct.
	// Fill that embedded interface with a native implementation so the same
	// source pattern remains usable without a runtime proxy.
	base := &constructImpl{}
	base.node = newNode(host, scope, id)
	if !setEmbeddedImplementation(host, base) {
		panic("construct override must embed constructs.Construct or implement SetNodeInternal(constructs.Node)")
	}
	Dependable_Implement(host, &singleDependable{root: host})
}

func NewNode(host Construct, scope IConstruct, id *string) Node {
	if host == nil {
		panic("parameter host is required, but nil was provided")
	}
	if id == nil {
		panic("parameter id is required, but nil was provided")
	}
	return newNode(host, scope, id)
}

func NewNode_Override(n Node, host Construct, scope IConstruct, id *string) {
	if n == nil {
		panic("parameter n is required, but nil was provided")
	}
	if host == nil {
		panic("parameter host is required, but nil was provided")
	}
	if id == nil {
		panic("parameter id is required, but nil was provided")
	}
	created := newNode(host, scope, id)
	if target, ok := n.(*nodeImpl); ok {
		*target = *created
		return
	}
	if !setEmbeddedImplementation(n, created) {
		panic("constructs: Node override must embed constructs.Node")
	}
}

func newNode(host Construct, scope IConstruct, id *string) *nodeImpl {
	identifier := value(id)
	identifier = strings.ReplaceAll(identifier, "/", "--")
	if scope != nil && identifier == "" {
		panic("Only root constructs may have an empty ID")
	}
	result := &nodeImpl{
		host:     host,
		scope:    scope,
		id:       identifier,
		children: make(map[string]IConstruct),
		context:  make(map[string]interface{}),
	}
	if scope != nil {
		parent, ok := scope.Node().(*nodeImpl)
		if !ok {
			panic("construct scope was not created by purecdk8s constructs")
		}
		parent.addChild(host, identifier)
	}
	return result
}

func Node_Of(construct IConstruct) Node {
	if construct == nil {
		panic("parameter construct is required, but nil was provided")
	}
	return construct.Node()
}

func Node_PATH_SEP() *string {
	result := "/"
	return &result
}

func Construct_IsConstruct(x interface{}) *bool {
	if x == nil {
		panic("parameter x is required, but nil was provided")
	}
	_, ok := x.(IConstruct)
	return &ok
}

func RootConstruct_IsConstruct(x interface{}) *bool { return Construct_IsConstruct(x) }

func (n *nodeImpl) Addr() *string {
	if n.addr == "" {
		hash := sha1.New()
		for _, scope := range *n.Scopes() {
			id := value(scope.Node().Id())
			if id == "Default" {
				continue
			}
			_, _ = hash.Write([]byte(id))
			_, _ = hash.Write([]byte{'\n'})
		}
		n.addr = fmt.Sprintf("c8%x", hash.Sum(nil))
	}
	return &n.addr
}

func (n *nodeImpl) Children() *[]IConstruct {
	result := make([]IConstruct, 0, len(n.childOrder))
	for _, id := range n.childIDsInObjectOrder() {
		if child, ok := n.children[id]; ok {
			result = append(result, child)
		}
	}
	return &result
}

func (n *nodeImpl) DefaultChild() IConstruct {
	if n.hasDefault {
		return n.defaultChild
	}
	resource := n.TryFindChild(stringPointer("Resource"))
	defaultValue := n.TryFindChild(stringPointer("Default"))
	if resource != nil && defaultValue != nil {
		panic(fmt.Sprintf(`Cannot determine default child for %s. There is both a child with id "Resource" and id "Default"`, value(n.Path())))
	}
	if defaultValue != nil {
		return defaultValue
	}
	return resource
}

func (n *nodeImpl) SetDefaultChild(val IConstruct) {
	n.defaultChild = val
	n.hasDefault = val != nil
}

func (n *nodeImpl) Dependencies() *[]IConstruct {
	result := make([]IConstruct, 0)
	for _, dependency := range n.dependencies {
		roots := Dependable_Of(dependency).DependencyRoots()
		if roots == nil {
			continue
		}
		for _, root := range *roots {
			result = append(result, root)
		}
	}
	return &result
}

func (n *nodeImpl) Id() *string { return &n.id }

func (n *nodeImpl) Locked() *bool {
	locked := n.locked
	if !locked && n.scope != nil {
		locked = valueBool(n.scope.Node().Locked())
	}
	return &locked
}

func (n *nodeImpl) Metadata() *[]*MetadataEntry {
	result := append([]*MetadataEntry(nil), n.metadata...)
	return &result
}

func (n *nodeImpl) Path() *string {
	parts := make([]string, 0)
	for _, scope := range *n.Scopes() {
		id := value(scope.Node().Id())
		if id != "" {
			parts = append(parts, id)
		}
	}
	result := strings.Join(parts, "/")
	return &result
}

func (n *nodeImpl) Root() IConstruct {
	scopes := n.Scopes()
	if len(*scopes) == 0 {
		return nil
	}
	return (*scopes)[0]
}

func (n *nodeImpl) Scope() IConstruct { return n.scope }

func (n *nodeImpl) Scopes() *[]IConstruct {
	result := make([]IConstruct, 0)
	current := IConstruct(n.host)
	for current != nil {
		result = append([]IConstruct{current}, result...)
		current = current.Node().Scope()
	}
	return &result
}

func (n *nodeImpl) AddDependency(deps ...IDependable) {
	for _, dependency := range deps {
		found := false
		for _, existing := range n.dependencies {
			if interfaceEqual(existing, dependency) {
				found = true
				break
			}
		}
		if !found {
			n.dependencies = append(n.dependencies, dependency)
		}
	}
}

func (n *nodeImpl) AddMetadata(type_ *string, data interface{}, options *MetadataOptions) {
	if type_ == nil {
		panic("parameter type_ is required, but nil was provided")
	}
	if data == nil {
		panic("parameter data is required, but nil was provided")
	}
	entry := &MetadataEntry{Type: type_, Data: data}
	if options != nil && options.StackTraceOverride != nil && len(*options.StackTraceOverride) > 0 {
		entry.Trace = options.StackTraceOverride
	} else if options != nil && valueBool(options.StackTrace) {
		trace := make([]*string, 0)
		pcs := make([]uintptr, 32)
		count := runtime.Callers(2, pcs)
		frames := runtime.CallersFrames(pcs[:count])
		for {
			frame, more := frames.Next()
			line := fmt.Sprintf("%s (%s:%d)", frame.Function, frame.File, frame.Line)
			trace = append(trace, &line)
			if !more {
				break
			}
		}
		entry.Trace = &trace
	}
	n.metadata = append(n.metadata, entry)
}

func (n *nodeImpl) AddValidation(validation IValidation) {
	if validation == nil {
		panic("parameter validation is required, but nil was provided")
	}
	n.validations = append(n.validations, validation)
}

func (n *nodeImpl) FindAll(order ConstructOrder) *[]IConstruct {
	if order == "" {
		order = ConstructOrder_PREORDER
	}
	result := make([]IConstruct, 0)
	var visit func(IConstruct)
	visit = func(current IConstruct) {
		if order == ConstructOrder_PREORDER {
			result = append(result, current)
		}
		for _, child := range *current.Node().Children() {
			visit(child)
		}
		if order == ConstructOrder_POSTORDER {
			result = append(result, current)
		}
	}
	visit(n.host)
	return &result
}

func (n *nodeImpl) FindChild(id *string) IConstruct {
	if id == nil {
		panic("parameter id is required, but nil was provided")
	}
	result := n.TryFindChild(id)
	if result == nil {
		panic(fmt.Sprintf("No child with id: '%s'", value(id)))
	}
	return result
}

func (n *nodeImpl) GetAllContext(defaults *map[string]interface{}) interface{} {
	result := make(map[string]interface{})
	for _, scope := range *n.Scopes() {
		if impl, ok := scope.Node().(*nodeImpl); ok {
			for key, item := range impl.context {
				result[key] = item
			}
		}
	}
	// Despite the parameter name, defaults are explicit overrides in the
	// constructs v10 API and therefore have the highest precedence.
	if defaults != nil {
		for key, item := range *defaults {
			result[key] = item
		}
	}
	return result
}

func (n *nodeImpl) GetContext(key *string) interface{} {
	if key == nil {
		panic("parameter key is required, but nil was provided")
	}
	result := n.TryGetContext(key)
	if result == nil {
		panic(fmt.Sprintf("No context value present for %s key", value(key)))
	}
	return result
}

func (n *nodeImpl) Lock() { n.locked = true }

func (n *nodeImpl) SetContext(key *string, item interface{}) {
	if key == nil {
		panic("parameter key is required, but nil was provided")
	}
	if item == nil {
		panic("parameter value is required, but nil was provided")
	}
	if len(n.children) > 0 {
		ids := n.childIDsInObjectOrder()
		panic("Cannot set context after children have been added: " + strings.Join(ids, ","))
	}
	n.context[value(key)] = item
}

func (n *nodeImpl) TryFindChild(id *string) IConstruct {
	if id == nil {
		panic("parameter id is required, but nil was provided")
	}
	return n.children[strings.ReplaceAll(value(id), "/", "--")]
}

func (n *nodeImpl) TryGetContext(key *string) interface{} {
	if key == nil {
		panic("parameter key is required, but nil was provided")
	}
	if result, ok := n.context[value(key)]; ok {
		return result
	}
	if n.scope != nil {
		return n.scope.Node().TryGetContext(key)
	}
	return nil
}

func (n *nodeImpl) TryRemoveChild(childName *string) *bool {
	if childName == nil {
		panic("parameter childName is required, but nil was provided")
	}
	// Unlike lookup, removal does not sanitize the supplied id upstream. A
	// child created as "a/b" is stored as "a--b" and must be removed by that
	// stored id.
	id := value(childName)
	_, ok := n.children[id]
	if ok {
		delete(n.children, id)
		for index, childID := range n.childOrder {
			if childID == id {
				n.childOrder = append(n.childOrder[:index], n.childOrder[index+1:]...)
				break
			}
		}
	}
	return &ok
}

func (n *nodeImpl) Validate() *[]*string {
	result := make([]*string, 0)
	for _, validation := range n.validations {
		if values := validation.Validate(); values != nil {
			result = append(result, (*values)...)
		}
	}
	return &result
}

func (n *nodeImpl) With(mixins ...IMixin) IConstruct {
	all := append([]IConstruct(nil), (*n.FindAll(ConstructOrder_PREORDER))...)
	for _, mixin := range mixins {
		if mixin == nil {
			panic("mixin is required, but nil was provided")
		}
		for _, construct := range all {
			if valueBool(mixin.Supports(construct)) {
				mixin.ApplyTo(construct)
			}
		}
	}
	return n.host
}

func (n *nodeImpl) addChild(child IConstruct, id string) {
	if valueBool(n.Locked()) {
		if value(n.Path()) == "" {
			panic("Cannot add children during synthesis")
		}
		panic(fmt.Sprintf(`Cannot add children to "%s" during synthesis`, value(n.Path())))
	}
	if _, exists := n.children[id]; exists {
		name := n.id
		suffix := ""
		if name != "" {
			suffix = " [" + name + "]"
		}
		panic(fmt.Sprintf("There is already a Construct with name '%s' in Construct%s", id, suffix))
	}
	n.children[id] = child
	n.childOrder = append(n.childOrder, id)
}

// JavaScript Object.values() enumerates canonical array-index keys first in
// numeric order, followed by all other string keys in insertion order. Node's
// upstream child table is a plain object, so numeric construct ids follow that
// ordering as well.
func (n *nodeImpl) childIDsInObjectOrder() []string {
	type numericID struct {
		id    string
		index uint64
	}
	numeric := make([]numericID, 0)
	ordinary := make([]string, 0, len(n.childOrder))
	for _, id := range n.childOrder {
		if _, exists := n.children[id]; !exists {
			continue
		}
		if index, ok := javascriptArrayIndex(id); ok {
			numeric = append(numeric, numericID{id: id, index: index})
		} else {
			ordinary = append(ordinary, id)
		}
	}
	sort.Slice(numeric, func(i, j int) bool { return numeric[i].index < numeric[j].index })
	result := make([]string, 0, len(numeric)+len(ordinary))
	for _, item := range numeric {
		result = append(result, item.id)
	}
	return append(result, ordinary...)
}

func javascriptArrayIndex(value string) (uint64, bool) {
	if value == "" || (len(value) > 1 && value[0] == '0') {
		return 0, false
	}
	index, err := strconv.ParseUint(value, 10, 32)
	if err != nil || index == (1<<32)-1 || strconv.FormatUint(index, 10) != value {
		return 0, false
	}
	return index, true
}

type Dependable interface {
	DependencyRoots() *[]IConstruct
}

var dependableTraits sync.Map

func Dependable_Implement(instance IDependable, trait Dependable) {
	if instance == nil {
		panic("parameter instance is required, but nil was provided")
	}
	if trait == nil {
		panic("parameter trait is required, but nil was provided")
	}
	if !isComparable(instance) {
		panic("Dependable instance must be comparable")
	}
	dependableTraits.Store(instance, trait)
}

func NewDependable_Override(d Dependable) {}

func Dependable_Of(instance IDependable) Dependable {
	if instance == nil {
		panic("parameter instance is required, but nil was provided")
	}
	if trait, ok := dependableTraits.Load(instance); ok {
		return trait.(Dependable)
	}
	// Generated Go resources wrap an underlying construct interface. They are
	// still constructs (and therefore dependables), but the wrapper itself was
	// not present when the underlying node was initialized. Resolve such
	// forwarding wrappers to the node's actual host.
	if construct, ok := instance.(IConstruct); ok {
		root := construct
		if node, native := construct.Node().(*nodeImpl); native && node.host != nil {
			root = node.host
		}
		return &singleDependable{root: root}
	}
	panic(fmt.Sprintf(`%v does not implement IDependable. Use "Dependable_Implement()" to implement`, instance))
}

func Dependable_Get(instance IDependable) Dependable { return Dependable_Of(instance) }

type singleDependable struct{ root IConstruct }

func (d *singleDependable) DependencyRoots() *[]IConstruct {
	result := []IConstruct{d.root}
	return &result
}

type DependencyGroup interface {
	IDependable
	Add(scopes ...IDependable)
}

type dependencyGroup struct{ dependencies []IDependable }

func NewDependencyGroup(deps ...IDependable) DependencyGroup {
	result := &dependencyGroup{}
	Dependable_Implement(result, result)
	result.Add(deps...)
	return result
}

func NewDependencyGroup_Override(d DependencyGroup, deps ...IDependable) {
	if _, native := d.(*dependencyGroup); !native {
		base := NewDependencyGroup()
		if !setEmbeddedImplementation(d, base) {
			panic("constructs: DependencyGroup override must embed constructs.DependencyGroup")
		}
	}
	d.Add(deps...)
	Dependable_Implement(d, &groupDependable{group: d})
}

func (d *dependencyGroup) Add(scopes ...IDependable) {
	d.dependencies = append(d.dependencies, scopes...)
}

func (d *dependencyGroup) DependencyRoots() *[]IConstruct {
	result := make([]IConstruct, 0)
	for _, dependency := range d.dependencies {
		if roots := Dependable_Of(dependency).DependencyRoots(); roots != nil {
			result = append(result, (*roots)...)
		}
	}
	return &result
}

type groupDependable struct{ group DependencyGroup }

func (d *groupDependable) DependencyRoots() *[]IConstruct {
	if value, ok := d.group.(Dependable); ok {
		return value.DependencyRoots()
	}
	result := make([]IConstruct, 0)
	return &result
}

func value(pointer *string) string {
	if pointer == nil {
		return ""
	}
	return *pointer
}

func valueBool(pointer *bool) bool {
	return pointer != nil && *pointer
}

func stringPointer(value string) *string { return &value }

func interfaceEqual(left, right interface{}) bool {
	if !isComparable(left) || !isComparable(right) {
		return false
	}
	return left == right
}

func isComparable(value interface{}) bool {
	if value == nil {
		return true
	}
	defer func() { _ = recover() }()
	switch fmt.Sprintf("%T", value) {
	default:
		// Interface comparison below is the definitive test.
	}
	_ = value == value
	return true
}

func setEmbeddedImplementation(target interface{}, implementation interface{}) bool {
	value := reflect.ValueOf(target)
	if !value.IsValid() || value.Kind() != reflect.Pointer || value.IsNil() {
		return false
	}
	value = value.Elem()
	if value.Kind() != reflect.Struct {
		return false
	}
	implementationValue := reflect.ValueOf(implementation)
	for index := 0; index < value.NumField(); index++ {
		fieldInfo := value.Type().Field(index)
		field := value.Field(index)
		if !fieldInfo.Anonymous || !field.CanSet() || field.Kind() != reflect.Interface {
			continue
		}
		if implementationValue.Type().Implements(field.Type()) {
			field.Set(implementationValue)
			return true
		}
	}
	return false
}

// SortedChildren is an optional native helper; the upstream API preserves
// insertion order, while callers that need deterministic ID order can use it.
func SortedChildren(node Node) []IConstruct {
	children := append([]IConstruct(nil), (*node.Children())...)
	sort.SliceStable(children, func(i, j int) bool {
		return value(children[i].Node().Id()) < value(children[j].Node().Id())
	})
	return children
}
