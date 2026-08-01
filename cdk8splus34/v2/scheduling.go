package cdk8splus34

// NodeLabelQuery describes one node-label selector requirement.
type NodeLabelQuery interface{ toManifest() map[string]interface{} }

type nodeLabelQuery struct {
	key      *string
	operator string
	values   *[]*string
}

func (q *nodeLabelQuery) toManifest() map[string]interface{} {
	result := map[string]interface{}{"key": q.key, "operator": q.operator}
	if q.values != nil {
		result["values"] = q.values
	}
	return result
}

func NodeLabelQuery_Is(key, value *string) NodeLabelQuery {
	if key == nil || value == nil {
		panic("key and value are required")
	}
	return &nodeLabelQuery{key: key, operator: "In", values: &[]*string{value}}
}
func NodeLabelQuery_In(key *string, values *[]*string) NodeLabelQuery {
	if key == nil || values == nil {
		panic("key and values are required")
	}
	return &nodeLabelQuery{key: key, operator: "In", values: values}
}
func NodeLabelQuery_NotIn(key *string, values *[]*string) NodeLabelQuery {
	if key == nil || values == nil {
		panic("key and values are required")
	}
	return &nodeLabelQuery{key: key, operator: "NotIn", values: values}
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
func NodeLabelQuery_Gt(key *string, values *[]*string) NodeLabelQuery {
	if key == nil || values == nil {
		panic("key and values are required")
	}
	return &nodeLabelQuery{key: key, operator: "Gt", values: values}
}
func NodeLabelQuery_Lt(key *string, values *[]*string) NodeLabelQuery {
	if key == nil || values == nil {
		panic("key and values are required")
	}
	return &nodeLabelQuery{key: key, operator: "Lt", values: values}
}

// LabeledNode is a node selector based on labels.
type LabeledNode interface{ LabelSelector() *[]NodeLabelQuery }
type labeledNode struct{ selectors []NodeLabelQuery }

func (n *labeledNode) LabelSelector() *[]NodeLabelQuery {
	selectors := append([]NodeLabelQuery(nil), n.selectors...)
	return &selectors
}

// Node_Labeled selects nodes matching all supplied label queries.
func Node_Labeled(selectors ...NodeLabelQuery) LabeledNode {
	return &labeledNode{selectors: append([]NodeLabelQuery(nil), selectors...)}
}

// PodSchedulingAttractOptions controls a node attraction rule.
type PodSchedulingAttractOptions struct {
	Weight *float64 `field:"optional" json:"weight" yaml:"weight"`
}

// PodSchedulingColocateOptions controls a pod-affinity rule.
type PodSchedulingColocateOptions struct {
	Topology *string  `field:"optional" json:"topology" yaml:"topology"`
	Weight   *float64 `field:"optional" json:"weight" yaml:"weight"`
}

// PodScheduling is the scheduling configuration for a workload.
type PodScheduling interface {
	Attract(node LabeledNode, options *PodSchedulingAttractOptions)
	Colocate(selector IPodSelector, options *PodSchedulingColocateOptions)
}

// WorkloadScheduling is the scheduling API exposed by workloads.
type WorkloadScheduling = PodScheduling

type podSchedulingImpl struct {
	nodes    []LabeledNode
	colocate []IPodSelector
}

func (s *podSchedulingImpl) Attract(node LabeledNode, options *PodSchedulingAttractOptions) {
	if node == nil {
		panic("node is required")
	}
	if options != nil && options.Weight != nil {
		panic("weighted node affinity is not implemented")
	}
	s.nodes = append(s.nodes, node)
}

func (s *podSchedulingImpl) Colocate(selector IPodSelector, options *PodSchedulingColocateOptions) {
	if selector == nil {
		panic("selector is required")
	}
	if options != nil && (options.Weight != nil || options.Topology != nil) {
		panic("custom pod-affinity options are not implemented")
	}
	s.colocate = append(s.colocate, selector)
}

func (s *podSchedulingImpl) toManifest(self IPodSelector, spread bool) interface{} {
	nodeTerms := make([]interface{}, 0, len(s.nodes))
	for _, node := range s.nodes {
		queries := node.LabelSelector()
		expressions := make([]interface{}, 0, len(*queries))
		for _, query := range *queries {
			if query == nil {
				panic("node label query is required")
			}
			expressions = append(expressions, query.toManifest())
		}
		nodeTerms = append(nodeTerms, map[string]interface{}{"matchExpressions": expressions})
	}
	affinity := map[string]interface{}{}
	if len(nodeTerms) > 0 {
		affinity["nodeAffinity"] = map[string]interface{}{
			"requiredDuringSchedulingIgnoredDuringExecution": map[string]interface{}{"nodeSelectorTerms": nodeTerms},
		}
	}
	if len(s.colocate) > 0 {
		terms := make([]interface{}, 0, len(s.colocate))
		for _, selector := range s.colocate {
			terms = append(terms, podAffinityTerm(selector))
		}
		affinity["podAffinity"] = map[string]interface{}{"requiredDuringSchedulingIgnoredDuringExecution": terms}
	}
	if spread {
		terms := []interface{}{
			podAffinityTermWithTopology(self, "kubernetes.io/hostname"),
			podAffinityTermWithTopology(self, "topology.kubernetes.io/zone"),
		}
		affinity["podAntiAffinity"] = map[string]interface{}{"requiredDuringSchedulingIgnoredDuringExecution": terms}
	}
	if len(affinity) == 0 {
		return nil
	}
	return affinity
}

func podAffinityTerm(selector IPodSelector) map[string]interface{} {
	return podAffinityTermWithTopology(selector, "kubernetes.io/hostname")
}

func podAffinityTermWithTopology(selector IPodSelector, topology string) map[string]interface{} {
	config := selector.ToPodSelectorConfig()
	labels := map[string]interface{}{}
	if config != nil && config.LabelSelector != nil && config.LabelSelector.Labels != nil {
		for key, value := range *config.LabelSelector.Labels {
			labels[key] = value
		}
	}
	return map[string]interface{}{
		"labelSelector": map[string]interface{}{"matchLabels": labels},
		"topologyKey":   topology,
	}
}
