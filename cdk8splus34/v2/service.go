package cdk8splus34

import (
	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// Configuration for selecting pods, optionally in particular namespaces.
type PodSelectorConfig struct {
	// A selector to select pods by labels.
	LabelSelector LabelSelector `field:"required" json:"labelSelector" yaml:"labelSelector"`
	// Configuration for selecting which namepsaces are the pods allowed to be in.
	Namespaces *NamespaceSelectorConfig `field:"optional" json:"namespaces" yaml:"namespaces"`
}

// For some parts of your application (for example, frontends) you may want to expose a Service onto an external IP address, that's outside of your cluster.
//
// Kubernetes ServiceTypes allow you to specify what kind of Service you want. The default is ClusterIP.
type ServiceType string

const (
	// Exposes the Service on a cluster-internal IP.
	//
	// Choosing this value makes the Service only reachable from within the cluster. This is the default ServiceType.
	ServiceType_CLUSTER_IP ServiceType = "CLUSTER_IP"
	// Exposes the Service on each Node's IP at a static port (the NodePort).
	//
	// A ClusterIP Service, to which the NodePort Service routes, is automatically created. You'll be able to contact the NodePort Service, from outside the cluster, by requesting <NodeIP>:<NodePort>.
	ServiceType_NODE_PORT ServiceType = "NODE_PORT"
	// Exposes the Service externally using a cloud provider's load balancer.
	//
	// NodePort and ClusterIP Services, to which the external load balancer routes, are automatically created.
	ServiceType_LOAD_BALANCER ServiceType = "LOAD_BALANCER"
	// Maps the Service to the contents of the externalName field (e.g. foo.bar.example.com), by returning a CNAME record with its value. No proxying of any kind is set up.
	//
	// > Note: You need either kube-dns version 1.7 or CoreDNS version 0.0.8 or higher to use the ExternalName type.
	ServiceType_EXTERNAL_NAME ServiceType = "EXTERNAL_NAME"
)

func serviceTypeManifestValue(value ServiceType) string {
	switch value {
	case ServiceType_CLUSTER_IP:
		return "ClusterIP"
	case ServiceType_NODE_PORT:
		return "NodePort"
	case ServiceType_LOAD_BALANCER:
		return "LoadBalancer"
	case ServiceType_EXTERNAL_NAME:
		return "ExternalName"
	default:
		return string(value)
	}
}

// Options for `Service.bind`.
type ServiceBindOptions struct {
	// The name of this port within the service.
	//
	// This must be a DNS_LABEL. All ports within a ServiceSpec must have unique names. This maps to the 'Name' field in EndpointPort objects. Optional if only one ServicePort is defined on this service.
	Name *string `field:"optional" json:"name" yaml:"name"`
	// The port on each node on which this service is exposed when type=NodePort or LoadBalancer.
	//
	// Usually assigned by the system. If specified, it will be allocated to the service if unused or else creation of the service will fail. Default is to auto-allocate a port if the ServiceType of this Service requires one. See: https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport
	//
	// Default: - auto-allocate a port if the ServiceType of this Service requires one.
	NodePort *float64 `field:"optional" json:"nodePort" yaml:"nodePort"`
	// The IP protocol for this port.
	//
	// Supports "TCP", "UDP", and "SCTP". Default is TCP. Default: Protocol.TCP
	Protocol Protocol `field:"optional" json:"protocol" yaml:"protocol"`
	// The port number the service will redirect to. Default: - The value of `port` will be used.
	TargetPort *float64 `field:"optional" json:"targetPort" yaml:"targetPort"`
}

// Options to add a deployment to a service.
type AddDeploymentOptions struct {
	// The name of this port within the service.
	//
	// This must be a DNS_LABEL. All ports within a ServiceSpec must have unique names. This maps to the 'Name' field in EndpointPort objects. Optional if only one ServicePort is defined on this service.
	Name *string `field:"optional" json:"name" yaml:"name"`
	// The port on each node on which this service is exposed when type=NodePort or LoadBalancer.
	//
	// Usually assigned by the system. If specified, it will be allocated to the service if unused or else creation of the service will fail. Default is to auto-allocate a port if the ServiceType of this Service requires one. See: https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport
	//
	// Default: - auto-allocate a port if the ServiceType of this Service requires one.
	NodePort *float64 `field:"optional" json:"nodePort" yaml:"nodePort"`
	// The port number the service will bind to. Default: - Copied from the first container of the deployment.
	Port *float64 `field:"optional" json:"port" yaml:"port"`
	// The IP protocol for this port.
	//
	// Supports "TCP", "UDP", and "SCTP". Default is TCP. Default: Protocol.TCP
	Protocol Protocol `field:"optional" json:"protocol" yaml:"protocol"`
	// The port number the service will redirect to. Default: - The value of `port` will be used.
	TargetPort *float64 `field:"optional" json:"targetPort" yaml:"targetPort"`
}

// Definition of a service port.
type ServicePort struct {
	// The port number the service will bind to.
	Port *float64 `field:"required" json:"port" yaml:"port"`
	// The name of this port within the service.
	//
	// This must be a DNS_LABEL. All ports within a ServiceSpec must have unique names. This maps to the 'Name' field in EndpointPort objects. Optional if only one ServicePort is defined on this service.
	Name *string `field:"optional" json:"name" yaml:"name"`
	// The port on each node on which this service is exposed when type=NodePort or LoadBalancer.
	//
	// Usually assigned by the system. If specified, it will be allocated to the service if unused or else creation of the service will fail. Default is to auto-allocate a port if the ServiceType of this Service requires one. See: https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport
	//
	// Default: - auto-allocate a port if the ServiceType of this Service requires one.
	NodePort *float64 `field:"optional" json:"nodePort" yaml:"nodePort"`
	// The IP protocol for this port.
	//
	// Supports "TCP", "UDP", and "SCTP". Default is TCP. Default: Protocol.TCP
	Protocol Protocol `field:"optional" json:"protocol" yaml:"protocol"`
	// The port number the service will redirect to. Default: - The value of `port` will be used.
	TargetPort *float64 `field:"optional" json:"targetPort" yaml:"targetPort"`
}

// Properties for `Service`.
type ServiceProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// Which pods should the service select and route to.
	//
	// You can pass one of the following:
	//
	// - An instance of `Pod` or any workload resource (e.g `Deployment`, `StatefulSet`, ...) - Pods selected by the `Pods.select` function. Note that in this case only labels can be specified.
	//
	// Example:
	//
	//	// select the pods of a specific deployment
	//	const backend = new kplus.Deployment(this, 'Backend', ...);
	//	new kplus.Service(this, 'Service', { selector: backend });
	//
	//	// select all pods labeled with the `tier=backend` label
	//	const backend = kplus.Pod.labeled({ tier: 'backend' });
	//	new kplus.Service(this, 'Service', { selector: backend });
	//
	// Default: - unset, the service is assumed to have an external process managing its endpoints, which Kubernetes will not modify.
	Selector IPodSelector `field:"optional" json:"selector" yaml:"selector"`
	// The IP address of the service and is usually assigned randomly by the master.
	//
	// If an address is specified manually and is not in use by others, it will be allocated to the service; otherwise, creation of the service will fail. This field can not be changed through updates. Valid values are "None", empty string (""), or a valid IP address. "None" can be specified for headless services when proxying is not required. Only applies to types ClusterIP, NodePort, and LoadBalancer. Ignored if type is ExternalName. See: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies
	//
	// Default: - Automatically assigned.
	ClusterIP *string `field:"optional" json:"clusterIP" yaml:"clusterIP"`
	// A list of IP addresses for which nodes in the cluster will also accept traffic for this service.
	//
	// These IPs are not managed by Kubernetes. The user is responsible for ensuring that traffic arrives at a node with this IP. A common example is external load-balancers that are not part of the Kubernetes system. Default: - No external IPs.
	ExternalIPs *[]*string `field:"optional" json:"externalIPs" yaml:"externalIPs"`
	// Determines how the Service is exposed.
	//
	// More info: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types Default: ServiceType.ClusterIP
	Type ServiceType `field:"optional" json:"type" yaml:"type"`
	// The ports this service binds to.
	//
	// If the selector of the service is a managed pod / workload, its ports will are automatically extracted and used as the default value. Otherwise, no ports are bound. Default: - either the selector ports, or none.
	Ports *[]*ServicePort `field:"optional" json:"ports" yaml:"ports"`
	// The externalName to be used when ServiceType.EXTERNAL_NAME is set. Default: - No external name.
	ExternalName *string `field:"optional" json:"externalName" yaml:"externalName"`
	// A list of CIDR IP addresses, if specified and supported by the platform, will restrict traffic through the cloud-provider load-balancer to the specified client IPs.
	//
	// More info: https://kubernetes.io/docs/tasks/access-application-cluster/configure-cloud-provider-firewall/
	LoadBalancerSourceRanges *[]*string `field:"optional" json:"loadBalancerSourceRanges" yaml:"loadBalancerSourceRanges"`
	// The publishNotReadyAddresses indicates that any agent which deals with endpoints for this Service should disregard any indications of ready/not-ready.
	//
	// More info: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/#servicespec-v1-core Default: - false.
	PublishNotReadyAddresses *bool `field:"optional" json:"publishNotReadyAddresses" yaml:"publishNotReadyAddresses"`
}

// Represents an object that can select pods.
type IPodSelector interface {
	constructs.IConstruct
	// Return the configuration of this selector.
	ToPodSelectorConfig() *PodSelectorConfig
}

// An abstract way to expose an application running on a set of Pods as a network service.
//
// With Kubernetes you don't need to modify your application to use an unfamiliar service discovery mechanism. Kubernetes gives Pods their own IP addresses and a single DNS name for a set of Pods, and can load-balance across them.
//
// For example, consider a stateless image-processing backend which is running with 3 replicas. Those replicas are fungible—frontends do not care which backend they use. While the actual Pods that compose the backend set may change, the frontend clients should not need to be aware of that, nor should they need to keep track of the set of backends themselves. The Service abstraction enables this decoupling.
//
// If you're able to use Kubernetes APIs for service discovery in your application, you can query the API server for Endpoints, that get updated whenever the set of Pods in a Service changes. For non-native applications, Kubernetes offers ways to place a network port or load balancer in between your application and the backend Pods.
type Service interface {
	Resource
	// The IP address of the service and is usually assigned randomly by the master.
	ClusterIP() *string
	// Determines how the Service is exposed.
	Type() ServiceType
	// The externalName to be used for EXTERNAL_NAME types.
	ExternalName() *string
	// Ports for this service.
	//
	// Use `bind()` to bind additional service ports.
	Ports() *[]*ServicePort
	// Return the first port of the service.
	Port() *float64
	// Configure a port the service will bind to.
	//
	// This method can be called multiple times.
	Bind(port *float64, options *ServiceBindOptions)
	// Require this service to select pods matching the selector.
	//
	// Note that invoking this method multiple times acts as an AND operator on the resulting labels.
	Select(selector IPodSelector)
	// Require this service to select pods with this label.
	//
	// Note that invoking this method multiple times acts as an AND operator on the resulting labels.
	SelectLabel(key, value *string)
	// Expose a service via an ingress using the specified path.
	//
	// Returns: The `Ingress` resource that was used.
	ExposeViaIngress(path *string, options *ExposeServiceViaIngressOptions) Ingress
}

type serviceImpl struct {
	resourceBase
	clusterIP                *string
	type_                    ServiceType
	externalName             *string
	ports                    []*ServicePort
	selector                 map[string]*string
	externalIPs              *[]*string
	loadBalancerSourceRanges *[]*string
	publishNotReadyAddresses *bool
}

func NewService(scope constructs.Construct, id *string, props *ServiceProps) Service {
	if props == nil {
		props = &ServiceProps{}
	}
	result := &serviceImpl{clusterIP: props.ClusterIP, externalName: props.ExternalName, selector: map[string]*string{}, externalIPs: props.ExternalIPs, loadBalancerSourceRanges: props.LoadBalancerSourceRanges, publishNotReadyAddresses: props.PublishNotReadyAddresses}
	result.type_ = props.Type
	if result.externalName != nil {
		result.type_ = ServiceType_EXTERNAL_NAME
	} else if result.type_ == "" {
		result.type_ = ServiceType_CLUSTER_IP
	}
	manifest := map[string]interface{}{}
	result.resourceBase.initialize(result, scope, id, "v1", "Service", "services", props.Metadata, manifest)
	if props.Selector != nil {
		result.Select(props.Selector)
	}
	if props.Ports != nil {
		for _, port := range *props.Ports {
			if port != nil {
				result.ports = append(result.ports, port)
			}
		}
	}
	manifest["spec"] = cdk8s.Lazy_Any(lazyProducer{produce: func() interface{} { return result.toManifest() }})
	return result
}

func NewService_Override(service Service, scope constructs.Construct, id *string, props *ServiceProps) {
	applyOverride(service, NewService(scope, id, props), "Service")
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func Service_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func (s *serviceImpl) ClusterIP() *string {
	return s.clusterIP
}

func (s *serviceImpl) Type() ServiceType {
	return s.type_
}

func (s *serviceImpl) ExternalName() *string {
	return s.externalName
}

func (s *serviceImpl) Ports() *[]*ServicePort {
	values := append([]*ServicePort(nil), s.ports...)
	return &values
}

func (s *serviceImpl) Port() *float64 {
	if len(s.ports) == 0 {
		return nil
	}
	return s.ports[0].Port
}

func (s *serviceImpl) Bind(port *float64, options *ServiceBindOptions) {
	if port == nil {
		panic("port is required")
	}
	result := &ServicePort{Port: port}
	if options != nil {
		result.Name = options.Name
		result.NodePort = options.NodePort
		result.Protocol = options.Protocol
		result.TargetPort = options.TargetPort
	}
	s.ports = append(s.ports, result)
}

func (s *serviceImpl) Select(selector IPodSelector) {
	if selector == nil {
		panic("selector is required")
	}
	config := selector.ToPodSelectorConfig()
	if config == nil || config.LabelSelector == nil {
		return
	}
	for key, value := range labelSelectorLabels(config.LabelSelector) {
		s.selector[key] = value
	}
}

func (s *serviceImpl) SelectLabel(key, value *string) {
	if key == nil || value == nil {
		panic("key and value are required")
	}
	s.selector[*key] = value
}

func (s *serviceImpl) toManifest() interface{} {
	if s.type_ == ServiceType_EXTERNAL_NAME {
		if s.externalName == nil {
			panic("A service with type EXTERNAL_NAME requires an externalName prop")
		}
		return map[string]interface{}{"type": serviceTypeManifestValue(s.type_), "externalName": s.externalName}
	}
	if len(s.ports) == 0 {
		panic("A service must be configured with a port")
	}
	ports := make([]interface{}, 0, len(s.ports))
	for _, port := range s.ports {
		entry := map[string]interface{}{"port": port.Port}
		if port.Name != nil {
			entry["name"] = port.Name
		}
		if port.TargetPort != nil {
			entry["targetPort"] = port.TargetPort
		}
		if port.NodePort != nil {
			entry["nodePort"] = port.NodePort
		}
		if port.Protocol != "" {
			entry["protocol"] = string(port.Protocol)
		}
		ports = append(ports, entry)
	}
	externalIPs := []interface{}{}
	if s.externalIPs != nil {
		for _, value := range *s.externalIPs {
			externalIPs = append(externalIPs, value)
		}
	}
	result := map[string]interface{}{"type": serviceTypeManifestValue(s.type_), "ports": ports, "selector": s.selector, "externalIPs": externalIPs}
	if s.loadBalancerSourceRanges != nil {
		result["loadBalancerSourceRanges"] = s.loadBalancerSourceRanges
	}
	if s.publishNotReadyAddresses != nil {
		result["publishNotReadyAddresses"] = s.publishNotReadyAddresses
	}
	if s.clusterIP != nil {
		result["clusterIP"] = s.clusterIP
	}
	if s.externalName != nil {
		result["externalName"] = s.externalName
	}
	return result
}

// Specify how the path is matched against request paths. See: https://kubernetes.io/docs/concepts/services-networking/ingress/#path-types
type HttpIngressPathType string

const (
	// Matches based on a URL path prefix split by '/'.
	HttpIngressPathType_EXACT HttpIngressPathType = "EXACT"
	// Matches the URL path exactly.
	HttpIngressPathType_PREFIX HttpIngressPathType = "PREFIX"
	// Matching is specified by the underlying IngressClass.
	HttpIngressPathType_IMPLEMENTATION_SPECIFIC HttpIngressPathType = "IMPLEMENTATION_SPECIFIC"
)

func httpIngressPathTypeManifestValue(value HttpIngressPathType) string {
	switch value {
	case HttpIngressPathType_EXACT:
		return "Exact"
	case HttpIngressPathType_PREFIX:
		return "Prefix"
	case HttpIngressPathType_IMPLEMENTATION_SPECIFIC:
		return "ImplementationSpecific"
	default:
		return string(value)
	}
}

// Options for exposing a service using an ingress.
type ExposeServiceViaIngressOptions struct {
	// The type of the path. Default: HttpIngressPathType.PREFIX
	PathType HttpIngressPathType `field:"optional" json:"pathType" yaml:"pathType"`
	// The ingress to add rules to. Default: - An ingress will be automatically created.
	Ingress Ingress `field:"optional" json:"ingress" yaml:"ingress"`
}

func (s *serviceImpl) ExposeViaIngress(path *string, options *ExposeServiceViaIngressOptions) Ingress {
	if path == nil {
		panic("path is required")
	}
	var ingress Ingress
	pathType := HttpIngressPathType_PREFIX
	if options != nil {
		ingress = options.Ingress
		if options.PathType != "" {
			pathType = options.PathType
		}
	}
	if ingress == nil {
		ingress = NewIngress(s, jsii.String("Ingress"), nil)
	}
	ingress.AddRule(path, IngressBackend_FromService(s, nil), pathType)
	return ingress
}
