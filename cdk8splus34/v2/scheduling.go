package cdk8splus34

import (
	"fmt"

	"github.com/Chriscbr/purecdk8s/constructs/v10"
)

// NodeLabelQuery describes a label requirement for nodes.
type (
	NodeLabelQuery interface{ toManifest() map[string]interface{} }
	nodeLabelQuery struct {
		key      *string
		operator string
		values   *[]*string
	}
)

func (q *nodeLabelQuery) toManifest() map[string]interface{} {
	result := map[string]interface{}{"key": q.key, "operator": q.operator}
	if q.values != nil {
		result["values"] = q.values
	}
	return result
}

func nodeLabelQueryWithValues(key *string, operator string, values *[]*string) NodeLabelQuery {
	if key == nil || values == nil {
		panic("key and values are required")
	}
	return &nodeLabelQuery{key: key, operator: operator, values: values}
}

func NodeLabelQuery_Is(key, value *string) NodeLabelQuery {
	if value == nil {
		panic("key and value are required")
	}
	return nodeLabelQueryWithValues(key, "In", &[]*string{value})
}

func NodeLabelQuery_In(key *string, values *[]*string) NodeLabelQuery {
	return nodeLabelQueryWithValues(key, "In", values)
}

func NodeLabelQuery_NotIn(key *string, values *[]*string) NodeLabelQuery {
	return nodeLabelQueryWithValues(key, "NotIn", values)
}

func NodeLabelQuery_Gt(key *string, values *[]*string) NodeLabelQuery {
	return nodeLabelQueryWithValues(key, "Gt", values)
}

func NodeLabelQuery_Lt(key *string, values *[]*string) NodeLabelQuery {
	return nodeLabelQueryWithValues(key, "Lt", values)
}

func NodeLabelQuery_Exists(key *string) NodeLabelQuery {
	if key == nil {
		panic("key is required")
	}
	return &nodeLabelQuery{key: key, operator: "Exists"}
}

func NodeLabelQuery_DoesNotExist(key *string) NodeLabelQuery {
	if key == nil {
		panic("key is required")
	}
	return &nodeLabelQuery{key: key, operator: "DoesNotExist"}
}

// LabeledNode selects nodes whose labels match all supplied requirements.
type (
	LabeledNode interface{ LabelSelector() *[]NodeLabelQuery }
	labeledNode struct{ selectors []NodeLabelQuery }
)

func (n *labeledNode) LabelSelector() *[]NodeLabelQuery {
	values := append([]NodeLabelQuery(nil), n.selectors...)
	return &values
}

func Node_Labeled(selectors ...NodeLabelQuery) LabeledNode {
	return &labeledNode{selectors: append([]NodeLabelQuery(nil), selectors...)}
}

// PodSchedulingAttractOptions controls a node-affinity rule.
type PodSchedulingAttractOptions struct {
	Weight *float64 `field:"optional" json:"weight" yaml:"weight"`
}

// PodSchedulingColocateOptions controls a pod-affinity rule.
type PodSchedulingColocateOptions struct {
	Topology Topology `field:"optional" json:"topology" yaml:"topology"`
	Weight   *float64 `field:"optional" json:"weight" yaml:"weight"`
}

// PodSchedulingSeparateOptions controls a pod anti-affinity rule.
type PodSchedulingSeparateOptions struct {
	Topology Topology `field:"optional" json:"topology" yaml:"topology"`
	Weight   *float64 `field:"optional" json:"weight" yaml:"weight"`
}

// PodScheduling configures node selection, affinity, and tolerations.
type PodScheduling interface {
	Instance() AbstractPod
	Assign(NamedNode)
	Attract(LabeledNode, *PodSchedulingAttractOptions)
	Colocate(IPodSelector, *PodSchedulingColocateOptions)
	Separate(IPodSelector, *PodSchedulingSeparateOptions)
	Tolerate(TaintedNode)
	toManifest() map[string]interface{}
}

type podSchedulingImpl struct {
	instance              AbstractPod
	nodeAffinityPreferred []interface{}
	nodeAffinityRequired  []interface{}
	podAffinityPreferred  []interface{}
	podAffinityRequired   []interface{}
	podAntiPreferred      []interface{}
	podAntiRequired       []interface{}
	tolerations           []interface{}
	nodeName              *string
}

func NewPodScheduling(instance AbstractPod) PodScheduling {
	if instance == nil {
		panic("instance is required")
	}
	return &podSchedulingImpl{instance: instance}
}

func NewPodScheduling_Override(scheduling PodScheduling, instance AbstractPod) {
	applyOverride(scheduling, NewPodScheduling(instance), "PodScheduling")
}

func (s *podSchedulingImpl) Instance() AbstractPod {
	return s.instance
}

func (s *podSchedulingImpl) Assign(node NamedNode) {
	if node == nil {
		panic("node is required")
	}
	if s.nodeName != nil {
		panic("Cannot assign pod to node " + stringValue(node.Name()) + ". It is already assigned to node " + stringValue(s.nodeName))
	}
	s.nodeName = node.Name()
}

func (s *podSchedulingImpl) Tolerate(node TaintedNode) {
	if node == nil {
		panic("node is required")
	}
	for _, query := range *node.TaintSelector() {
		if query == nil {
			panic("taint selector is required")
		}
		s.tolerations = append(s.tolerations, query.toManifest())
	}
}

func validateAffinityWeight(weight *float64) {
	if weight != nil && (*weight < 1 || *weight > 100) {
		panic("Invalid affinity weight: " + formatNumber(*weight) + ". Must be in range 1-100")
	}
}

func formatNumber(value float64) string {
	return fmt.Sprintf("%g", value)
}

func (s *podSchedulingImpl) Attract(node LabeledNode, options *PodSchedulingAttractOptions) {
	if node == nil {
		panic("node is required")
	}
	queries := node.LabelSelector()
	expressions := make([]interface{}, 0, len(*queries))
	for _, query := range *queries {
		if query == nil {
			panic("node label query is required")
		}
		expressions = append(expressions, query.toManifest())
	}
	term := map[string]interface{}{"matchExpressions": expressions}
	if options != nil && options.Weight != nil {
		validateAffinityWeight(options.Weight)
		s.nodeAffinityPreferred = append(s.nodeAffinityPreferred, map[string]interface{}{"weight": options.Weight, "preference": term})
		return
	}
	s.nodeAffinityRequired = append(s.nodeAffinityRequired, term)
}

func topologyKey(topology Topology) *string {
	if topology == nil {
		return Topology_HOSTNAME().Key()
	}
	return topology.Key()
}

func podAffinityTerm(selector IPodSelector, topology Topology) map[string]interface{} {
	if selector == nil {
		panic("selector is required")
	}
	config := selector.ToPodSelectorConfig()
	if config == nil || config.LabelSelector == nil {
		panic("selector configuration is required")
	}
	term := map[string]interface{}{"topologyKey": topologyKey(topology), "labelSelector": labelSelectorManifest(config.LabelSelector)}
	if config.Namespaces != nil {
		if config.Namespaces.LabelSelector != nil && !*config.Namespaces.LabelSelector.IsEmpty() {
			term["namespaceSelector"] = labelSelectorManifest(config.Namespaces.LabelSelector)
		}
		if config.Namespaces.Names != nil && len(*config.Namespaces.Names) > 0 {
			term["namespaces"] = config.Namespaces.Names
		}
	}
	return term
}

func (s *podSchedulingImpl) Colocate(selector IPodSelector, options *PodSchedulingColocateOptions) {
	var topology Topology
	if options != nil {
		topology = options.Topology
	}
	term := podAffinityTerm(selector, topology)
	if options != nil && options.Weight != nil {
		validateAffinityWeight(options.Weight)
		s.podAffinityPreferred = append(s.podAffinityPreferred, map[string]interface{}{"weight": options.Weight, "podAffinityTerm": term})
		return
	}
	s.podAffinityRequired = append(s.podAffinityRequired, term)
}

func (s *podSchedulingImpl) Separate(selector IPodSelector, options *PodSchedulingSeparateOptions) {
	var topology Topology
	if options != nil {
		topology = options.Topology
	}
	term := podAffinityTerm(selector, topology)
	if options != nil && options.Weight != nil {
		validateAffinityWeight(options.Weight)
		s.podAntiPreferred = append(s.podAntiPreferred, map[string]interface{}{"weight": options.Weight, "podAffinityTerm": term})
		return
	}
	s.podAntiRequired = append(s.podAntiRequired, term)
}

func (s *podSchedulingImpl) toManifest() map[string]interface{} {
	result := map[string]interface{}{}
	affinity := map[string]interface{}{}
	if len(s.nodeAffinityPreferred) > 0 || len(s.nodeAffinityRequired) > 0 {
		node := map[string]interface{}{}
		if len(s.nodeAffinityPreferred) > 0 {
			node["preferredDuringSchedulingIgnoredDuringExecution"] = s.nodeAffinityPreferred
		}
		if len(s.nodeAffinityRequired) > 0 {
			node["requiredDuringSchedulingIgnoredDuringExecution"] = map[string]interface{}{"nodeSelectorTerms": s.nodeAffinityRequired}
		}
		affinity["nodeAffinity"] = node
	}
	if len(s.podAffinityPreferred) > 0 || len(s.podAffinityRequired) > 0 {
		pod := map[string]interface{}{}
		if len(s.podAffinityPreferred) > 0 {
			pod["preferredDuringSchedulingIgnoredDuringExecution"] = s.podAffinityPreferred
		}
		if len(s.podAffinityRequired) > 0 {
			pod["requiredDuringSchedulingIgnoredDuringExecution"] = s.podAffinityRequired
		}
		affinity["podAffinity"] = pod
	}
	if len(s.podAntiPreferred) > 0 || len(s.podAntiRequired) > 0 {
		pod := map[string]interface{}{}
		if len(s.podAntiPreferred) > 0 {
			pod["preferredDuringSchedulingIgnoredDuringExecution"] = s.podAntiPreferred
		}
		if len(s.podAntiRequired) > 0 {
			pod["requiredDuringSchedulingIgnoredDuringExecution"] = s.podAntiRequired
		}
		affinity["podAntiAffinity"] = pod
	}
	if len(affinity) > 0 {
		result["affinity"] = affinity
	}
	if s.nodeName != nil {
		result["nodeName"] = s.nodeName
	}
	if len(s.tolerations) > 0 {
		result["tolerations"] = s.tolerations
	}
	return result
}

// WorkloadSchedulingSpreadOptions controls a workload's self anti-affinity.
type WorkloadSchedulingSpreadOptions struct {
	Topology Topology `field:"optional" json:"topology" yaml:"topology"`
	Weight   *float64 `field:"optional" json:"weight" yaml:"weight"`
}

// Workload is an AbstractPod with a pod selector and workload facilities.
type Workload interface {
	AbstractPod
	Connections() PodConnections
	MatchExpressions() *[]*LabelSelectorRequirement
	MatchLabels() *map[string]*string
	Scheduling() WorkloadScheduling
	Select(...LabelSelector)
}

// WorkloadScheduling adds spreading to the core PodScheduling API.
type WorkloadScheduling interface {
	PodScheduling
	Spread(*WorkloadSchedulingSpreadOptions)
}

type workloadSchedulingImpl struct{ *podSchedulingImpl }

func NewWorkloadScheduling(instance AbstractPod) WorkloadScheduling {
	return &workloadSchedulingImpl{podSchedulingImpl: NewPodScheduling(instance).(*podSchedulingImpl)}
}

func NewWorkloadScheduling_Override(scheduling WorkloadScheduling, instance AbstractPod) {
	applyOverride(scheduling, NewWorkloadScheduling(instance), "WorkloadScheduling")
}

func (s *workloadSchedulingImpl) Spread(options *WorkloadSchedulingSpreadOptions) {
	var topology Topology
	var weight *float64
	if options != nil {
		topology, weight = options.Topology, options.Weight
	}
	s.Separate(s.instance, &PodSchedulingSeparateOptions{Topology: topology, Weight: weight})
}

func Workload_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func NewWorkload_Override(_ Workload, _ constructs.Construct, _ *string, _ *WorkloadProps) {
	panic("Workload is an abstract base; use a concrete workload constructor")
}
