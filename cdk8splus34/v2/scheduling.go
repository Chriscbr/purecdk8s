package cdk8splus34

import (
	"fmt"

	"github.com/Chriscbr/purecdk8s/constructs/v10"
)

// NodeLabelQuery describes a label requirement for nodes.
type (
	// Represents a query that can be performed against nodes with labels.
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

// Requires value of label `key` to equal `value`.
func NodeLabelQuery_Is(key, value *string) NodeLabelQuery {
	if value == nil {
		panic("key and value are required")
	}
	return nodeLabelQueryWithValues(key, "In", &[]*string{value})
}

// Requires value of label `key` to be one of `values`.
func NodeLabelQuery_In(key *string, values *[]*string) NodeLabelQuery {
	return nodeLabelQueryWithValues(key, "In", values)
}

// Requires value of label `key` to be none of `values`.
func NodeLabelQuery_NotIn(key *string, values *[]*string) NodeLabelQuery {
	return nodeLabelQueryWithValues(key, "NotIn", values)
}

// Requires value of label `key` to greater than all elements in `values`.
func NodeLabelQuery_Gt(key *string, values *[]*string) NodeLabelQuery {
	return nodeLabelQueryWithValues(key, "Gt", values)
}

// Requires value of label `key` to less than all elements in `values`.
func NodeLabelQuery_Lt(key *string, values *[]*string) NodeLabelQuery {
	return nodeLabelQueryWithValues(key, "Lt", values)
}

// Requires label `key` to exist.
func NodeLabelQuery_Exists(key *string) NodeLabelQuery {
	if key == nil {
		panic("key is required")
	}
	return &nodeLabelQuery{key: key, operator: "Exists"}
}

// Requires label `key` to not exist.
func NodeLabelQuery_DoesNotExist(key *string) NodeLabelQuery {
	if key == nil {
		panic("key is required")
	}
	return &nodeLabelQuery{key: key, operator: "DoesNotExist"}
}

// LabeledNode selects nodes whose labels match all supplied requirements.
type (
	// A node that is matched by label selectors.
	LabeledNode interface{ LabelSelector() *[]NodeLabelQuery }
	labeledNode struct{ selectors []NodeLabelQuery }
)

func (n *labeledNode) LabelSelector() *[]NodeLabelQuery {
	values := append([]NodeLabelQuery(nil), n.selectors...)
	return &values
}

// Match a node by its labels.
func Node_Labeled(selectors ...NodeLabelQuery) LabeledNode {
	return &labeledNode{selectors: append([]NodeLabelQuery(nil), selectors...)}
}

// Options for `PodScheduling.attract`.
type PodSchedulingAttractOptions struct {
	// Indicates the attraction is optional (soft), with this weight score. Default: - no weight. assignment is assumed to be required (hard).
	Weight *float64 `field:"optional" json:"weight" yaml:"weight"`
}

// Options for `PodScheduling.colocate`.
type PodSchedulingColocateOptions struct {
	// Which topology to coloate on. Default: - Topology.HOSTNAME
	Topology Topology `field:"optional" json:"topology" yaml:"topology"`
	// Indicates the co-location is optional (soft), with this weight score. Default: - no weight. co-location is assumed to be required (hard).
	Weight *float64 `field:"optional" json:"weight" yaml:"weight"`
}

// Options for `PodScheduling.separate`.
type PodSchedulingSeparateOptions struct {
	// Which topology to separate on. Default: - Topology.HOSTNAME
	Topology Topology `field:"optional" json:"topology" yaml:"topology"`
	// Indicates the separation is optional (soft), with this weight score. Default: - no weight. separation is assumed to be required (hard).
	Weight *float64 `field:"optional" json:"weight" yaml:"weight"`
}

// Controls the pod scheduling strategy.
type PodScheduling interface {
	Instance() AbstractPod
	// Assign this pod a specific node by name.
	//
	// The scheduler ignores the Pod, and the kubelet on the named node tries to place the Pod on that node. Overrules any affinity rules of the pod.
	//
	// Some limitations of static assignment are:
	//
	//   - If the named node does not exist, the Pod will not run, and in some cases may be automatically deleted.
	//   - If the named node does not have the resources to accommodate the Pod, the Pod will fail and its reason will indicate why, for example OutOfmemory or OutOfcpu.
	//   - Node names in cloud environments are not always predictable or stable.
	//
	// Will throw is the pod is already assigned to named node.
	//
	// Under the hood, this method utilizes the `nodeName` property.
	Assign(NamedNode)
	// Attract this pod to a node matched by selectors. You can select a node by using `Node.labeled()`.
	//
	// Attracting to multiple nodes (i.e invoking this method multiple times) acts as an OR condition, meaning the pod will be assigned to either one of the nodes.
	//
	// Under the hood, this method utilizes the `nodeAffinity` property. See: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#node-affinity
	Attract(LabeledNode, *PodSchedulingAttractOptions)
	// Co-locate this pod with a scheduling selection.
	//
	// A selection can be one of:
	//
	// - An instance of a `Pod`. - An instance of a `Workload` (e.g `Deployment`, `StatefulSet`). - An un-managed pod that can be selected via `Pods.select()`.
	//
	// Co-locating with multiple selections ((i.e invoking this method multiple times)) acts as an AND condition. meaning the pod will be assigned to a node that satisfies all selections (i.e runs at least one pod that satisifies each selection).
	//
	// Under the hood, this method utilizes the `podAffinity` property. See: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#inter-pod-affinity-and-anti-affinity
	Colocate(IPodSelector, *PodSchedulingColocateOptions)
	// Seperate this pod from a scheduling selection.
	//
	// A selection can be one of:
	//
	// - An instance of a `Pod`. - An instance of a `Workload` (e.g `Deployment`, `StatefulSet`). - An un-managed pod that can be selected via `Pods.select()`.
	//
	// Seperating from multiple selections acts as an AND condition. meaning the pod will not be assigned to a node that satisfies all selections (i.e runs at least one pod that satisifies each selection).
	//
	// Under the hood, this method utilizes the `podAntiAffinity` property. See: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#inter-pod-affinity-and-anti-affinity
	Separate(IPodSelector, *PodSchedulingSeparateOptions)
	// Allow this pod to tolerate taints matching these tolerations.
	//
	// You can put multiple taints on the same node and multiple tolerations on the same pod. The way Kubernetes processes multiple taints and tolerations is like a filter: start with all of a node's taints, then ignore the ones for which the pod has a matching toleration; the remaining un-ignored taints have the indicated effects on the pod. In particular:
	//
	//   - if there is at least one un-ignored taint with effect NoSchedule then Kubernetes will not schedule the pod onto that node
	//   - if there is no un-ignored taint with effect NoSchedule but there is at least one un-ignored taint with effect PreferNoSchedule then Kubernetes will try to not schedule the pod onto the node
	//   - if there is at least one un-ignored taint with effect NoExecute then the pod will be evicted from the node (if it is already running on the node), and will not be scheduled onto the node (if it is not yet running on the node).
	//
	// Under the hood, this method utilizes the `tolerations` property. See: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
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
		if config.Namespaces.LabelSelector != nil {
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

// Options for `WorkloadScheduling.spread`.
type WorkloadSchedulingSpreadOptions struct {
	// Which topology to spread on. Default: - Topology.HOSTNAME
	Topology Topology `field:"optional" json:"topology" yaml:"topology"`
	// Indicates the spread is optional, with this weight score. Default: - no weight. spread is assumed to be required.
	Weight *float64 `field:"optional" json:"weight" yaml:"weight"`
}

// A workload is an application running on Kubernetes.
//
// Whether your workload is a single component or several that work together, on Kubernetes you run it inside a set of pods. In Kubernetes, a Pod represents a set of running containers on your cluster.
type Workload interface {
	AbstractPod
	Connections() PodConnections
	// The expression matchers this workload will use in order to select pods.
	//
	// Returns a a copy. Use `select()` to add expression matchers.
	MatchExpressions() *[]*LabelSelectorRequirement
	// The label matchers this workload will use in order to select pods.
	//
	// Returns a a copy. Use `select()` to add label matchers.
	MatchLabels() *map[string]*string
	Scheduling() WorkloadScheduling
	// Configure selectors for this workload.
	Select(...LabelSelector)
}

// Controls the pod scheduling strategy of this workload.
//
// It offers some additional API's on top of the core pod scheduling.
type WorkloadScheduling interface {
	PodScheduling
	// Spread the pods in this workload by the topology key.
	//
	// A spread is a separation of the pod from itself and is used to balance out pod replicas across a given topology.
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

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func Workload_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func NewWorkload_Override(_ Workload, _ constructs.Construct, _ *string, _ *WorkloadProps) {
	panic("Workload is an abstract base; use a concrete workload constructor")
}
