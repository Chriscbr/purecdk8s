package cdk8splus34

import (
	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// Options for `LabelSelector.of`.
type LabelSelectorOptions struct {
	// Expression based label matchers.
	Expressions *[]LabelExpression `field:"optional" json:"expressions" yaml:"expressions"`
	// Strict label matchers.
	Labels *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
}

// A label selector requirement is a selector that contains values, a key, and an operator that relates the key and values.
type LabelSelectorRequirement struct {
	// The label key that the selector applies to.
	Key *string `field:"required" json:"key" yaml:"key"`
	// Represents a key's relationship to a set of values.
	Operator *string `field:"required" json:"operator" yaml:"operator"`
	// An array of string values.
	//
	// If the operator is In or NotIn, the values array must be non-empty. If the operator is Exists or DoesNotExist, the values array must be empty. This array is replaced during a strategic merge patch.
	Values *[]*string `field:"optional" json:"values" yaml:"values"`
}

// Match a resource by labels.
type LabelSelector interface {
	IsEmpty() *bool
	toManifest() map[string]interface{}
	labels() map[string]*string
}

type labelSelectorImpl struct {
	expressions []LabelExpression
	values      map[string]*string
}

func LabelSelector_Of(options *LabelSelectorOptions) LabelSelector {
	if options == nil {
		options = &LabelSelectorOptions{}
	}
	return newLabelSelector(options.Expressions, options.Labels)
}

func newLabelSelector(expressions *[]LabelExpression, labels *map[string]*string) *labelSelectorImpl {
	result := &labelSelectorImpl{values: map[string]*string{}}
	if expressions != nil {
		for _, expression := range *expressions {
			if expression == nil {
				panic("label expression is required")
			}
			result.expressions = append(result.expressions, expression)
		}
	}
	if labels != nil {
		for key, value := range *labels {
			result.values[key] = value
		}
	}
	return result
}

func labelSelectorFromLabels(labels *map[string]*string) LabelSelector {
	return newLabelSelector(nil, labels)
}

func newLabelSelectorFromRequirements(requirements []*LabelSelectorRequirement, labels *map[string]*string) LabelSelector {
	expressions := make([]LabelExpression, 0, len(requirements))
	for _, requirement := range requirements {
		if requirement == nil || requirement.Key == nil || requirement.Operator == nil {
			panic("label selector requirement is required")
		}
		expressions = append(expressions, &labelExpressionImpl{key: requirement.Key, operator: requirement.Operator, values: requirement.Values})
	}
	return newLabelSelector(&expressions, labels)
}

func (l *labelSelectorImpl) IsEmpty() *bool {
	return jsii.Bool(len(l.expressions) == 0 && len(l.values) == 0)
}

func (l *labelSelectorImpl) labels() map[string]*string {
	result := map[string]*string{}
	for key, value := range l.values {
		result[key] = value
	}
	return result
}

func labelSelectorLabels(selector LabelSelector) map[string]*string {
	return selector.labels()
}

func labelSelectorRequirements(selector LabelSelector) []*LabelSelectorRequirement {
	impl, ok := selector.(*labelSelectorImpl)
	if !ok {
		panic("unsupported label selector")
	}
	result := make([]*LabelSelectorRequirement, 0, len(impl.expressions))
	for _, expression := range impl.expressions {
		result = append(result, &LabelSelectorRequirement{Key: expression.Key(), Operator: expression.Operator(), Values: expression.Values()})
	}
	return result
}

func (l *labelSelectorImpl) toManifest() map[string]interface{} {
	result := map[string]interface{}{}
	if len(l.expressions) > 0 {
		expressions := make([]interface{}, 0, len(l.expressions))
		for _, expression := range l.expressions {
			entry := map[string]interface{}{"key": expression.Key(), "operator": expression.Operator()}
			if expression.Values() != nil {
				entry["values"] = expression.Values()
			}
			expressions = append(expressions, entry)
		}
		result["matchExpressions"] = expressions
	}
	if len(l.values) > 0 {
		result["matchLabels"] = l.values
	}
	return result
}

func labelSelectorManifest(selector LabelSelector) map[string]interface{} {
	return selector.toManifest()
}

// Represents a query that can be performed against resources with labels.
type LabelExpression interface {
	Key() *string
	Operator() *string
	Values() *[]*string
}

type labelExpressionImpl struct {
	key, operator *string
	values        *[]*string
}

func (l *labelExpressionImpl) Key() *string {
	return l.key
}

func (l *labelExpressionImpl) Operator() *string {
	return l.operator
}

func (l *labelExpressionImpl) Values() *[]*string {
	return l.values
}

func newLabelExpression(key *string, operator string, values *[]*string) LabelExpression {
	if key == nil {
		panic("key is required")
	}
	return &labelExpressionImpl{key: key, operator: jsii.String(operator), values: values}
}

// Requires value of label `key` to be one of `values`.
func LabelExpression_In(key *string, values *[]*string) LabelExpression {
	if values == nil {
		panic("values are required")
	}
	return newLabelExpression(key, "In", values)
}

// Requires value of label `key` to be none of `values`.
func LabelExpression_NotIn(key *string, values *[]*string) LabelExpression {
	if values == nil {
		panic("values are required")
	}
	return newLabelExpression(key, "NotIn", values)
}

// Requires label `key` to exist.
func LabelExpression_Exists(key *string) LabelExpression {
	return newLabelExpression(key, "Exists", nil)
}

// Requires label `key` to not exist.
func LabelExpression_DoesNotExist(key *string) LabelExpression {
	return newLabelExpression(key, "DoesNotExist", nil)
}

// Configuration for network peers.
//
// A peer can either by an ip block, or a selection of pods, not both.
type NetworkPolicyPeerConfig struct {
	// The ip block this peer represents.
	IpBlock NetworkPolicyIpBlock `field:"optional" json:"ipBlock" yaml:"ipBlock"`
	// The pod selector this peer represents.
	PodSelector *PodSelectorConfig `field:"optional" json:"podSelector" yaml:"podSelector"`
}

// Describes a peer to allow traffic to/from.
type INetworkPolicyPeer interface {
	constructs.IConstruct
	// Return the configuration of this peer.
	ToNetworkPolicyPeerConfig() *NetworkPolicyPeerConfig
	// Convert the peer into a pod selector, if possible.
	ToPodSelector() IPodSelector
}

// Represents an object that can select namespaces.
type INamespaceSelector interface {
	constructs.IConstruct
	// Return the configuration of this selector.
	ToNamespaceSelectorConfig() *NamespaceSelectorConfig
}

// Options for `Pods.all`.
type PodsAllOptions struct {
	// Namespaces the pods are allowed to be in.
	//
	// Use `Namespaces.all()` to allow all namespaces. Default: - unset, implies the namespace of the resource this selection is used in.
	Namespaces Namespaces `field:"optional" json:"namespaces" yaml:"namespaces"`
}

// Options for `Pods.select`.
type PodsSelectOptions struct {
	// Expressions the pods must satisify. Default: - no expressions requirements.
	Expressions *[]LabelExpression `field:"optional" json:"expressions" yaml:"expressions"`
	// Labels the pods must have. Default: - no strict labels requirements.
	Labels *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	// Namespaces the pods are allowed to be in.
	//
	// Use `Namespaces.all()` to allow all namespaces. Default: - unset, implies the namespace of the resource this selection is used in.
	Namespaces Namespaces `field:"optional" json:"namespaces" yaml:"namespaces"`
}

// Represents a group of pods.
type Pods interface {
	constructs.Construct
	IPodSelector
	INetworkPolicyPeer
	// See: INetworkPolicyPeer.toNetworkPolicyPeerConfig ()
	ToNetworkPolicyPeerConfig() *NetworkPolicyPeerConfig
	// See: INetworkPolicyPeer.toPodSelector ()
	ToPodSelector() IPodSelector
}

type podsImpl struct {
	node        constructs.Node
	expressions *[]LabelExpression
	labels      *map[string]*string
	namespaces  INamespaceSelector
}

func (p *podsImpl) Node() constructs.Node {
	return p.node
}

func (p *podsImpl) SetNodeInternal(node constructs.Node) {
	p.node = node
}

func (p *podsImpl) ToString() *string {
	return p.node.Path()
}

func (p *podsImpl) With(mixins ...constructs.IMixin) constructs.IConstruct {
	return p.node.With(mixins...)
}

func NewPods(scope constructs.Construct, id *string, expressions *[]LabelExpression, labels *map[string]*string, namespaces INamespaceSelector) Pods {
	if scope == nil || id == nil {
		panic("scope and id are required")
	}
	result := &podsImpl{expressions: expressions, labels: labels, namespaces: namespaces}
	constructs.NewConstruct_Override(result, scope, id)
	return result
}

func NewPods_Override(pods Pods, scope constructs.Construct, id *string, expressions *[]LabelExpression, labels *map[string]*string, namespaces INamespaceSelector) {
	applyOverride(pods, NewPods(scope, id, expressions, labels, namespaces), "Pods")
}

// Select pods in the cluster with various selectors.
func Pods_Select(scope constructs.Construct, id *string, options *PodsSelectOptions) Pods {
	if options == nil {
		panic("options is required")
	}
	return NewPods(scope, id, options.Expressions, options.Labels, options.Namespaces)
}

// Select all pods.
func Pods_All(scope constructs.Construct, id *string, options *PodsAllOptions) Pods {
	if options == nil {
		options = &PodsAllOptions{}
	}
	return NewPods(scope, id, nil, nil, options.Namespaces)
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func Pods_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func (p *podsImpl) ToPodSelectorConfig() *PodSelectorConfig {
	var namespaces *NamespaceSelectorConfig
	if p.namespaces != nil {
		namespaces = p.namespaces.ToNamespaceSelectorConfig()
	}
	return &PodSelectorConfig{LabelSelector: newLabelSelector(p.expressions, p.labels), Namespaces: namespaces}
}

func (p *podsImpl) ToNetworkPolicyPeerConfig() *NetworkPolicyPeerConfig {
	return &NetworkPolicyPeerConfig{PodSelector: p.ToPodSelectorConfig()}
}

func (p *podsImpl) ToPodSelector() IPodSelector {
	return p
}

// Options for `Namespaces.select`.
type NamespacesSelectOptions struct {
	// Namespaces must satisfy these selectors.
	//
	// The selectors query labels, just like the `labels` property, but they provide a more advanced matching mechanism. Default: - no selector requirements.
	Expressions *[]LabelExpression `field:"optional" json:"expressions" yaml:"expressions"`
	// Labels the namespaces must have.
	//
	// This is equivalent to using an 'Is' selector. Default: - no strict labels requirements.
	Labels *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	// Namespaces names must be one of these. Default: - no name requirements.
	Names *[]*string `field:"optional" json:"names" yaml:"names"`
}

// Represents a group of namespaces.
type Namespaces interface {
	constructs.Construct
	INamespaceSelector
	INetworkPolicyPeer
}

type namespacesImpl struct {
	node        constructs.Node
	expressions *[]LabelExpression
	names       *[]*string
	labels      *map[string]*string
	pods        Pods
}

func (n *namespacesImpl) Node() constructs.Node {
	return n.node
}

func (n *namespacesImpl) SetNodeInternal(node constructs.Node) {
	n.node = node
}

func (n *namespacesImpl) ToString() *string {
	return n.node.Path()
}

func (n *namespacesImpl) With(mixins ...constructs.IMixin) constructs.IConstruct {
	return n.node.With(mixins...)
}

func NewNamespaces(scope constructs.Construct, id *string, expressions *[]LabelExpression, names *[]*string, labels *map[string]*string) Namespaces {
	if scope == nil || id == nil {
		panic("scope and id are required")
	}
	result := &namespacesImpl{expressions: expressions, names: names, labels: labels}
	constructs.NewConstruct_Override(result, scope, id)
	result.pods = Pods_All(result, jsii.String("Pods"), &PodsAllOptions{Namespaces: result})
	return result
}

func NewNamespaces_Override(namespaces Namespaces, scope constructs.Construct, id *string, expressions *[]LabelExpression, names *[]*string, labels *map[string]*string) {
	applyOverride(namespaces, NewNamespaces(scope, id, expressions, names, labels), "Namespaces")
}

// Select specific namespaces.
func Namespaces_Select(scope constructs.Construct, id *string, options *NamespacesSelectOptions) Namespaces {
	if options == nil {
		panic("options is required")
	}
	return NewNamespaces(scope, id, options.Expressions, options.Names, options.Labels)
}

// Select all namespaces.
func Namespaces_All(scope constructs.Construct, id *string) Namespaces {
	return NewNamespaces(scope, id, &[]LabelExpression{}, nil, &map[string]*string{})
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func Namespaces_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func (n *namespacesImpl) ToNamespaceSelectorConfig() *NamespaceSelectorConfig {
	return &NamespaceSelectorConfig{LabelSelector: newLabelSelector(n.expressions, n.labels), Names: n.names}
}

func (n *namespacesImpl) ToNetworkPolicyPeerConfig() *NetworkPolicyPeerConfig {
	return n.pods.ToNetworkPolicyPeerConfig()
}

func (n *namespacesImpl) ToPodSelector() IPodSelector {
	return n.pods.ToPodSelector()
}

// Node is a factory for node selectors.
type (
	// Represents a node in the cluster.
	Node     interface{}
	nodeImpl struct{}
)

func NewNode() Node {
	return &nodeImpl{}
}

func NewNode_Override(_ Node) {
	_ = NewNode()
}

type (
	// A node that is matched by its name.
	NamedNode     interface{ Name() *string }
	namedNodeImpl struct{ name *string }
)

func NewNamedNode(name *string) NamedNode {
	if name == nil {
		panic("name is required")
	}
	return &namedNodeImpl{name: name}
}

func NewNamedNode_Override(node NamedNode, name *string) {
	applyOverride(node, NewNamedNode(name), "NamedNode")
}

func (n *namedNodeImpl) Name() *string {
	return n.name
}

type (
	// A node that is matched by taint selectors.
	TaintedNode     interface{ TaintSelector() *[]NodeTaintQuery }
	taintedNodeImpl struct{ selector []NodeTaintQuery }
)

func NewTaintedNode(selector *[]NodeTaintQuery) TaintedNode {
	if selector == nil {
		panic("taintSelector is required")
	}
	result := &taintedNodeImpl{}
	result.selector = append(result.selector, (*selector)...)
	return result
}

func NewTaintedNode_Override(node TaintedNode, selector *[]NodeTaintQuery) {
	applyOverride(node, NewTaintedNode(selector), "TaintedNode")
}

func (n *taintedNodeImpl) TaintSelector() *[]NodeTaintQuery {
	values := append([]NodeTaintQuery(nil), n.selector...)
	return &values
}

func NewLabeledNode(selector *[]NodeLabelQuery) LabeledNode {
	if selector == nil {
		panic("labelSelector is required")
	}
	return &labeledNode{selectors: append([]NodeLabelQuery(nil), (*selector)...)}
}

func NewLabeledNode_Override(node LabeledNode, selector *[]NodeLabelQuery) {
	applyOverride(node, NewLabeledNode(selector), "LabeledNode")
}

// Match a node by its name.
func Node_Named(name *string) NamedNode {
	return NewNamedNode(name)
}

// Match a node by its taints.
func Node_Tainted(selector ...NodeTaintQuery) TaintedNode {
	return NewTaintedNode(&selector)
}

// Taint effects.
type TaintEffect string

const (
	// This means that no pod will be able to schedule onto the node unless it has a matching toleration.
	TaintEffect_NO_SCHEDULE TaintEffect = "NO_SCHEDULE"
	// This is a "preference" or "soft" version of `NO_SCHEDULE` -- the system will try to avoid placing a pod that does not tolerate the taint on the node, but it is not required.
	TaintEffect_PREFER_NO_SCHEDULE TaintEffect = "PREFER_NO_SCHEDULE"
	// This affects pods that are already running on the node as follows:.
	//
	//   - Pods that do not tolerate the taint are evicted immediately.
	//   - Pods that tolerate the taint without specifying `duration` remain bound forever.
	//   - Pods that tolerate the taint with a specified `duration` remain bound for the specified amount of time.
	TaintEffect_NO_EXECUTE TaintEffect = "NO_EXECUTE"
)

// Options for `NodeTaintQuery`.
type NodeTaintQueryOptions struct {
	// The taint effect to match. Default: - all effects are matched.
	Effect TaintEffect `field:"optional" json:"effect" yaml:"effect"`
	// How much time should a pod that tolerates the `NO_EXECUTE` effect be bound to the node.
	//
	// Only applies for the `NO_EXECUTE` effect. Default: - bound forever.
	EvictAfter cdk8s.Duration `field:"optional" json:"evictAfter" yaml:"evictAfter"`
}

type (
	// Taint queries that can be perfomed against nodes.
	NodeTaintQuery     interface{ toManifest() map[string]interface{} }
	nodeTaintQueryImpl struct {
		operator   string
		key, value *string
		effect     TaintEffect
		evictAfter cdk8s.Duration
	}
)

// Matches a taint with a specific key and value.
func NodeTaintQuery_Is(key, value *string, options *NodeTaintQueryOptions) NodeTaintQuery {
	if key == nil || value == nil {
		panic("key and value are required")
	}
	return newNodeTaintQuery("Equal", key, value, options)
}

// Matches a tain with any value of a specific key.
func NodeTaintQuery_Exists(key *string, options *NodeTaintQueryOptions) NodeTaintQuery {
	if key == nil {
		panic("key is required")
	}
	return newNodeTaintQuery("Exists", key, nil, options)
}

// Matches any taint.
func NodeTaintQuery_Any() NodeTaintQuery {
	return &nodeTaintQueryImpl{operator: "Exists"}
}

func newNodeTaintQuery(operator string, key, value *string, options *NodeTaintQueryOptions) NodeTaintQuery {
	result := &nodeTaintQueryImpl{operator: operator, key: key, value: value}
	if options != nil {
		result.effect = options.Effect
		result.evictAfter = options.EvictAfter
		if result.evictAfter != nil && result.effect != TaintEffect_NO_EXECUTE {
			panic("Only 'NO_EXECUTE' effects can specify 'evictAfter'")
		}
	}
	return result
}

func (q *nodeTaintQueryImpl) toManifest() map[string]interface{} {
	result := map[string]interface{}{"operator": q.operator}
	if q.key != nil {
		result["key"] = q.key
	}
	if q.value != nil {
		result["value"] = q.value
	}
	if q.effect != "" {
		result["effect"] = taintEffectManifest(q.effect)
	}
	if q.evictAfter != nil {
		result["tolerationSeconds"] = q.evictAfter.ToSeconds(nil)
	}
	return result
}

func taintEffectManifest(value TaintEffect) string {
	switch value {
	case TaintEffect_NO_SCHEDULE:
		return "NoSchedule"
	case TaintEffect_PREFER_NO_SCHEDULE:
		return "PreferNoSchedule"
	case TaintEffect_NO_EXECUTE:
		return "NoExecute"
	default:
		panic("invalid taint effect")
	}
}

type (
	// Available topology domains.
	Topology     interface{ Key() *string }
	topologyImpl struct{ key *string }
)

func (t *topologyImpl) Key() *string {
	return t.key
}

// Custom key for the node label that the system uses to denote the topology domain.
func Topology_Custom(key *string) Topology {
	if key == nil {
		panic("key is required")
	}
	return &topologyImpl{key: key}
}

func Topology_HOSTNAME() Topology {
	return Topology_Custom(jsii.String("kubernetes.io/hostname"))
}

func Topology_ZONE() Topology {
	return Topology_Custom(jsii.String("topology.kubernetes.io/zone"))
}

func Topology_REGION() Topology {
	return Topology_Custom(jsii.String("topology.kubernetes.io/region"))
}
