package cdk8splus34

import (
	"github.com/purecdk8s/purecdk8s/cdk8s/v2"
	"github.com/purecdk8s/purecdk8s/constructs/v10"
	"github.com/purecdk8s/purecdk8s/jsii"
)

type LabelSelectorOptions struct {
	Expressions *[]LabelExpression  `field:"optional" json:"expressions" yaml:"expressions"`
	Labels      *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
}

type LabelSelectorRequirement struct {
	Key      *string    `field:"required" json:"key" yaml:"key"`
	Operator *string    `field:"required" json:"operator" yaml:"operator"`
	Values   *[]*string `field:"optional" json:"values" yaml:"values"`
}

// LabelSelector matches Kubernetes resources by labels and expressions.
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

// LabelExpression is an expression-based label matcher.
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

func LabelExpression_In(key *string, values *[]*string) LabelExpression {
	if values == nil {
		panic("values are required")
	}
	return newLabelExpression(key, "In", values)
}

func LabelExpression_NotIn(key *string, values *[]*string) LabelExpression {
	if values == nil {
		panic("values are required")
	}
	return newLabelExpression(key, "NotIn", values)
}

func LabelExpression_Exists(key *string) LabelExpression {
	return newLabelExpression(key, "Exists", nil)
}

func LabelExpression_DoesNotExist(key *string) LabelExpression {
	return newLabelExpression(key, "DoesNotExist", nil)
}

// NetworkPolicyPeerConfig describes an IP block or a pod selection.
type NetworkPolicyPeerConfig struct {
	IpBlock     NetworkPolicyIpBlock `field:"optional" json:"ipBlock" yaml:"ipBlock"`
	PodSelector *PodSelectorConfig   `field:"optional" json:"podSelector" yaml:"podSelector"`
}

type INetworkPolicyPeer interface {
	constructs.IConstruct
	ToNetworkPolicyPeerConfig() *NetworkPolicyPeerConfig
	ToPodSelector() IPodSelector
}

type INamespaceSelector interface {
	constructs.IConstruct
	ToNamespaceSelectorConfig() *NamespaceSelectorConfig
}

type PodsAllOptions struct {
	Namespaces Namespaces `field:"optional" json:"namespaces" yaml:"namespaces"`
}

type PodsSelectOptions struct {
	Expressions *[]LabelExpression  `field:"optional" json:"expressions" yaml:"expressions"`
	Labels      *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	Namespaces  Namespaces          `field:"optional" json:"namespaces" yaml:"namespaces"`
}

type Pods interface {
	constructs.Construct
	IPodSelector
	INetworkPolicyPeer
	ToNetworkPolicyPeerConfig() *NetworkPolicyPeerConfig
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

func Pods_Select(scope constructs.Construct, id *string, options *PodsSelectOptions) Pods {
	if options == nil {
		panic("options is required")
	}
	return NewPods(scope, id, options.Expressions, options.Labels, options.Namespaces)
}

func Pods_All(scope constructs.Construct, id *string, options *PodsAllOptions) Pods {
	if options == nil {
		options = &PodsAllOptions{}
	}
	return NewPods(scope, id, nil, nil, options.Namespaces)
}

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

type NamespacesSelectOptions struct {
	Expressions *[]LabelExpression  `field:"optional" json:"expressions" yaml:"expressions"`
	Labels      *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	Names       *[]*string          `field:"optional" json:"names" yaml:"names"`
}

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

func Namespaces_Select(scope constructs.Construct, id *string, options *NamespacesSelectOptions) Namespaces {
	if options == nil {
		panic("options is required")
	}
	return NewNamespaces(scope, id, options.Expressions, options.Names, options.Labels)
}

func Namespaces_All(scope constructs.Construct, id *string) Namespaces {
	return NewNamespaces(scope, id, &[]LabelExpression{}, nil, &map[string]*string{})
}

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

func Node_Named(name *string) NamedNode {
	return NewNamedNode(name)
}

func Node_Tainted(selector ...NodeTaintQuery) TaintedNode {
	return NewTaintedNode(&selector)
}

type TaintEffect string

const (
	TaintEffect_NO_SCHEDULE        TaintEffect = "NO_SCHEDULE"
	TaintEffect_PREFER_NO_SCHEDULE TaintEffect = "PREFER_NO_SCHEDULE"
	TaintEffect_NO_EXECUTE         TaintEffect = "NO_EXECUTE"
)

type NodeTaintQueryOptions struct {
	Effect     TaintEffect    `field:"optional" json:"effect" yaml:"effect"`
	EvictAfter cdk8s.Duration `field:"optional" json:"evictAfter" yaml:"evictAfter"`
}

type (
	NodeTaintQuery     interface{ toManifest() map[string]interface{} }
	nodeTaintQueryImpl struct {
		operator   string
		key, value *string
		effect     TaintEffect
		evictAfter cdk8s.Duration
	}
)

func NodeTaintQuery_Is(key, value *string, options *NodeTaintQueryOptions) NodeTaintQuery {
	if key == nil || value == nil {
		panic("key and value are required")
	}
	return newNodeTaintQuery("Equal", key, value, options)
}

func NodeTaintQuery_Exists(key *string, options *NodeTaintQueryOptions) NodeTaintQuery {
	if key == nil {
		panic("key is required")
	}
	return newNodeTaintQuery("Exists", key, nil, options)
}

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
	Topology     interface{ Key() *string }
	topologyImpl struct{ key *string }
)

func (t *topologyImpl) Key() *string {
	return t.key
}

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
