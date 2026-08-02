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

// Trait marker for classes that can be depended upon.
//
// The presence of this interface indicates that an object has
// an `IDependable` implementation.
//
// This interface can be used to take an (ordering) dependency on a set of
// constructs. An ordering dependency implies that the resources represented by
// those constructs are deployed before the resources depending ON them are
// deployed.
type IDependable interface{}

// Represents a construct.
type IConstruct interface {
	IDependable
	// Applies one or more mixins to this construct.
	//
	// Mixins are applied in order. The list of constructs is captured at the
	// start of the call, so constructs added by a mixin will not be visited.
	//
	// Returns: This construct for chaining.
	With(mixins ...IMixin) IConstruct
	// The tree node.
	Node() Node
}

// Represents the building block of the construct graph.
//
// All constructs besides the root construct must be created within the scope of
// another construct.
type Construct interface {
	IConstruct
	// The tree node.
	Node() Node
	// Returns a string representation of this construct.
	ToString() *string
	// Applies one or more mixins to this construct.
	//
	// Mixins are applied in order. The list of constructs is captured at the
	// start of the call, so constructs added by a mixin will not be visited.
	// Use multiple `with()` calls if subsequent mixins should apply to added
	// constructs.
	//
	// Returns: This construct for chaining.
	With(mixins ...IMixin) IConstruct
}

// Creates a new root construct node.
//
// The root construct represents the top of the construct tree and is not contained within a parent scope itself.
// For root constructs, the id is optional.
type RootConstruct interface {
	Construct
	// The tree node.
	Node() Node
	// Returns a string representation of this construct.
	ToString() *string
	// Applies one or more mixins to this construct.
	//
	// Mixins are applied in order. The list of constructs is captured at the
	// start of the call, so constructs added by a mixin will not be visited.
	// Use multiple `with()` calls if subsequent mixins should apply to added
	// constructs.
	//
	// Returns: This construct for chaining.
	With(mixins ...IMixin) IConstruct
}

// A mixin is a reusable piece of functionality that can be applied to constructs to add behavior, properties, or modify existing functionality without inheritance.
type IMixin interface {
	// Applies the mixin functionality to the target construct.
	ApplyTo(construct IConstruct)
	// Determines whether this mixin can be applied to the given construct.
	Supports(construct IConstruct) *bool
}

// Implement this interface in order for the construct to be able to validate itself.
type IValidation interface {
	// Validate the current construct.
	//
	// This method can be implemented by derived constructs in order to perform
	// validation logic. It is called on all constructs before synthesis.
	//
	// Returns: An array of validation error messages, or an empty array if there the construct is valid.
	Validate() *[]*string
}

// In what order to return constructs.
type ConstructOrder string

const (
	// Depth-first, pre-order.
	ConstructOrder_PREORDER ConstructOrder = "PREORDER"
	// Depth-first, post-order (leaf nodes first).
	ConstructOrder_POSTORDER ConstructOrder = "POSTORDER"
)

// An entry in the construct metadata table.
type MetadataEntry struct {
	// The data.
	Data interface{} `field:"required" json:"data" yaml:"data"`
	// The metadata entry type.
	Type *string `field:"required" json:"type" yaml:"type"`
	// Stack trace at the point of adding the metadata.
	//
	// Only available if `addMetadata()` is called with `stackTrace: true`.
	// Default: - no trace information.
	//
	Trace *[]*string `field:"optional" json:"trace" yaml:"trace"`
}

// Options for `construct.addMetadata()`.
type MetadataOptions struct {
	// Include stack trace with metadata entry.
	// Default: false.
	//
	StackTrace *bool `field:"optional" json:"stackTrace" yaml:"stackTrace"`
	// The actual stack trace to be added to the metadata.
	//
	// If this
	// parameter is passed, the stackTrace parameter is ignored.
	StackTraceOverride *[]*string `field:"optional" json:"stackTraceOverride" yaml:"stackTraceOverride"`
	// A JavaScript function to begin tracing from.
	//
	// This option is ignored unless `stackTrace` is `true`.
	// Default: addMetadata().
	//
	TraceFromFunction interface{} `field:"optional" json:"traceFromFunction" yaml:"traceFromFunction"`
}

// Represents the construct node in the scope tree.
type Node interface {
	// An opaque, deterministic address for this construct, derived from its path.
	//
	// The address is a 42 character string: the prefix "c8" followed by 40
	// lowercase hexadecimal characters (0-9a-f). It is a SHA-1 over the ids of
	// the constructs on the path from the root of the tree down to this
	// construct.
	//
	// To enable refactoring of construct trees, constructs with the ID `Default`
	// are excluded from the calculation. Within a tree, `a/Default/b` and `a/b`
	// have the same address.
	//
	// This means the address is *not* guaranteed to identify a construct uniquely:
	//
	// - Any construct whose path is made up of the same ids has the same address.
	//   Two trees that are shaped alike therefore hand out the same addresses.
	// - As described above, a construct under a `Default` scope has the same
	//   address as its counterpart outside of that scope.
	// - The digest is of fixed width, so even distinct paths can in principle hash
	//   to the same address. SHA-1 in particular is no longer collision resistant.
	//
	// Use an address to derive stable, deterministic names from the location of a
	// construct in the tree. Do not use it as the identity of a construct:
	// instead, compare construct instances or use the `path`.
	//
	// Example:
	//   c83a2846e506bcc5f10682b564084bca2d275709ee
	//
	Addr() *string
	// All direct children of this construct.
	Children() *[]IConstruct
	// Returns the child construct that has the id `Default` or `Resource`.
	//
	// This is usually the construct that provides the bulk of the underlying functionality.
	// Useful for modifications of the underlying construct that are not available at the higher levels.
	// Override the defaultChild property.
	//
	// This should only be used in the cases where the correct
	// default child is not named 'Resource' or 'Default' as it
	// should be.
	//
	// If you set this to undefined, the default behavior of finding
	// the child named 'Resource' or 'Default' will be used.
	//
	// Returns: a construct or undefined if there is no default child.
	DefaultChild() IConstruct
	SetDefaultChild(val IConstruct)
	// Constructs that this construct depends on directly.
	//
	// This expands the sets of `Dependables` to the set of `Constructs` that they
	// represent.
	Dependencies() *[]IConstruct
	// The id of this construct within the current scope.
	//
	// This is a scope-unique id. To obtain an id that reflects the full location
	// of this construct in the tree, use `path` or `addr`.
	Id() *string
	// Returns true if this construct or the scopes in which it is defined are locked.
	Locked() *bool
	// An immutable array of metadata objects associated with this construct.
	//
	// This can be used, for example, to implement support for deprecation notices, source mapping, etc.
	Metadata() *[]*MetadataEntry
	// The full, absolute path of this construct in the tree.
	//
	// Components are separated by '/'.
	Path() *string
	// Returns the root of the construct tree.
	//
	// Returns: The root of the construct tree.
	Root() IConstruct
	// Returns the scope in which this construct is defined.
	//
	// The value is `undefined` at the root of the construct scope tree.
	Scope() IConstruct
	// All parent scopes of this construct.
	//
	// Returns: a list of parent scopes. The last element in the list will always
	// be the current construct and the first element will be the root of the
	// tree.
	Scopes() *[]IConstruct
	// Add an ordering dependency on a Dependable (a construct or set of constructs).
	//
	// It is up to the target document language to decide what an ordering
	// relationship means and how it should be rendered; for example, in the AWS
	// CDK for CloudFormation it means a dependency from every resource in scope
	// of the current construct to every resource in scope of the target set
	// of constructs, that get realized as either stack dependencies or
	// `DependsOn` relationships in the synthesized template.
	//
	// `IDependable` is a marker interface that indicates that a class has used
	// `Dependable.implement()` to implement the `IDependable` interface. It
	// can be used to make the target object represent more than one set of
	// constructs at a time. For example, a `DependencyGroup` uses this
	// interface to represent an explicit list of 0 or more constructs that should
	// be involved in the dependency relationship.
	AddDependency(deps ...IDependable)
	// Adds a metadata entry to this construct.
	//
	// Entries are arbitrary values and will also include a stack trace to allow tracing back to
	// the code location for when the entry was added. It can be used, for example, to include source
	// mapping in CloudFormation templates to improve diagnostics.
	// Note that construct metadata is not the same as CloudFormation resource metadata and is never written to the CloudFormation template.
	// The metadata entries are written to the Cloud Assembly Manifest if the `treeMetadata` property is specified in the props of the App that contains this Construct.
	AddMetadata(type_ *string, data interface{}, options *MetadataOptions)
	// Adds a validation to this construct.
	//
	// When `node.validate()` is called, the `validate()` method will be called on
	// all validations and all errors will be returned.
	AddValidation(validation IValidation)
	// Return this construct and all of its children in the given order.
	FindAll(order ConstructOrder) *[]IConstruct
	// Return a direct child by id.
	//
	// Throws an error if the child is not found.
	//
	// Returns: Child with the given id.
	FindChild(id *string) IConstruct
	// Retrieves the all context of a node from tree context.
	//
	// Context is usually initialized at the root, but can be overridden at any point in the tree.
	//
	// Returns: The context object or an empty object if there is discovered context.
	GetAllContext(defaults *map[string]interface{}) interface{}
	// Retrieves a value from tree context if present. Otherwise, would throw an error.
	//
	// Context is usually initialized at the root, but can be overridden at any point in the tree.
	//
	// Returns: The context value or throws error if there is no context value for this key.
	GetContext(key *string) interface{}
	// Locks this construct from allowing more children to be added.
	//
	// After this
	// call, no more children can be added to this construct or to any children.
	Lock()
	// Remove an ordering dependency on a Dependenable (a construct or set of constructs).
	//
	// This removes any dependency added using `node.addDependency()`. It must
	// use the exact same object that was involved in the `addDependency()` call.
	RemoveDependency(deps ...IDependable)
	// This can be used to set contextual values.
	//
	// Context must be set before any children are added, since children may consult context info during construction.
	// If the key already exists, it will be overridden.
	SetContext(key *string, value interface{})
	// Return a direct child by id, or undefined.
	//
	// Returns: the child if found, or undefined.
	TryFindChild(id *string) IConstruct
	// Retrieves a value from tree context.
	//
	// Context is usually initialized at the root, but can be overridden at any point in the tree.
	//
	// Returns: The context value or `undefined` if there is no context value for this key.
	TryGetContext(key *string) interface{}
	// Remove the child with the given name, if present.
	//
	// Returns: Whether a child with the given name was deleted.
	TryRemoveChild(childName *string) *bool
	// Validates this construct.
	//
	// Invokes the `validate()` method on all validations added through
	// `addValidation()`.
	//
	// Returns: an array of validation error messages associated with this
	// construct.
	Validate() *[]*string
	// Applies one or more mixins to this construct.
	//
	// Mixins are applied in order. The list of constructs is captured at the
	// start of the call, so constructs added by a mixin will not be visited.
	// Use multiple `with()` calls if subsequent mixins should apply to added
	// constructs.
	//
	// Returns: This construct for chaining.
	With(mixins ...IMixin) IConstruct
}

type constructImpl struct {
	node Node
}

func (c *constructImpl) Node() Node {
	return c.node
}

func (c *constructImpl) setNode(node Node) {
	c.node = node
}

func (c *constructImpl) SetNodeInternal(node Node) {
	c.node = node
}

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

// Creates a new construct node.
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

// Creates a new construct node.
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

// Creates a new root construct node.
func NewRootConstruct(id *string) RootConstruct {
	result := &constructImpl{}
	NewRootConstruct_Override(result, id)
	return result
}

// Creates a new root construct node.
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
	identifier := sanitizeID(value(id))
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

// Returns the node associated with a construct.
// Deprecated: use `construct.node` instead
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

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct`
// instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on
// disk are seen as independent, completely different libraries. As a
// consequence, the class `Construct` in each copy of the `constructs` library
// is seen as a different class, and an instance of one class will not test as
// `instanceof` the other class. `npm install` will not create installations
// like this, but users may manually symlink construct libraries together or
// use a monorepo tool: in those cases, multiple copies of the `constructs`
// library can be accidentally installed, and `instanceof` will behave
// unpredictably. It is safest to avoid using `instanceof`, and using
// this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func Construct_IsConstruct(x interface{}) *bool {
	if x == nil {
		result := false
		return &result
	}
	_, ok := x.(IConstruct)
	return &ok
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct`
// instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on
// disk are seen as independent, completely different libraries. As a
// consequence, the class `Construct` in each copy of the `constructs` library
// is seen as a different class, and an instance of one class will not test as
// `instanceof` the other class. `npm install` will not create installations
// like this, but users may manually symlink construct libraries together or
// use a monorepo tool: in those cases, multiple copies of the `constructs`
// library can be accidentally installed, and `instanceof` will behave
// unpredictably. It is safest to avoid using `instanceof`, and using
// this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func RootConstruct_IsConstruct(x interface{}) *bool {
	return Construct_IsConstruct(x)
}

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

func (n *nodeImpl) Id() *string {
	return &n.id
}

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

func (n *nodeImpl) Scope() IConstruct {
	return n.scope
}

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

func (n *nodeImpl) RemoveDependency(deps ...IDependable) {
	for _, dependency := range deps {
		for index, existing := range n.dependencies {
			if interfaceEqual(existing, dependency) {
				n.dependencies = append(n.dependencies[:index], n.dependencies[index+1:]...)
				break
			}
		}
	}
}

func (n *nodeImpl) AddMetadata(type_ *string, data interface{}, options *MetadataOptions) {
	if type_ == nil {
		panic("parameter type_ is required, but nil was provided")
	}
	if data == nil {
		return
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
	if defaults != nil {
		for key, item := range *defaults {
			result[key] = item
		}
	}
	for _, scope := range *n.Scopes() {
		if impl, ok := scope.Node().(*nodeImpl); ok {
			for key, item := range impl.context {
				result[key] = item
			}
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

func (n *nodeImpl) Lock() {
	n.locked = true
}

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
	return n.children[sanitizeID(value(id))]
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

// Trait for IDependable.
//
// Traits are interfaces that are privately implemented by objects. Instead of
// showing up in the public interface of a class, they need to be queried
// explicitly. This is used to implement certain framework features that are
// not intended to be used by Construct consumers, and so should be hidden
// from accidental use.
//
// Example:
//
//	// Usage
//	const roots = Dependable.of(construct).dependencyRoots;
//
//	// Definition
//	Dependable.implement(construct, {
//	      dependencyRoots: [construct],
//	});
type Dependable interface {
	// The set of constructs that form the root of this dependable.
	//
	// All resources under all returned constructs are included in the ordering
	// dependency.
	DependencyRoots() *[]IConstruct
}

var dependableTraits sync.Map

// Turn any object into an IDependable.
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

func NewDependable_Override(d Dependable) {
}

// Return the matching Dependable for the given class instance.
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

// Return the matching Dependable for the given class instance.
// Deprecated: use `of`.
func Dependable_Get(instance IDependable) Dependable {
	return Dependable_Of(instance)
}

type singleDependable struct{ root IConstruct }

func (d *singleDependable) DependencyRoots() *[]IConstruct {
	result := []IConstruct{d.root}
	return &result
}

// A set of constructs to be used as a dependable.
//
// This class can be used when a set of constructs which are disjoint in the
// construct tree needs to be combined to be used as a single dependable.
type DependencyGroup interface {
	IDependable
	// Add a construct to the dependency roots.
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

func stringPointer(value string) *string {
	return &value
}

func sanitizeID(id string) string {
	return strings.NewReplacer("/", "--", "\n", "--").Replace(id)
}

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
