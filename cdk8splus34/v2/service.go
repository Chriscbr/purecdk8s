package cdk8splus34

import (
	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// PodSelectorConfig describes a pod selector and its namespaces.
type PodSelectorConfig struct {
	LabelSelector LabelSelector            `field:"required" json:"labelSelector" yaml:"labelSelector"`
	Namespaces    *NamespaceSelectorConfig `field:"optional" json:"namespaces" yaml:"namespaces"`
}

// ServiceType controls Service exposure.
type ServiceType string

const (
	ServiceType_CLUSTER_IP    ServiceType = "CLUSTER_IP"
	ServiceType_NODE_PORT     ServiceType = "NODE_PORT"
	ServiceType_LOAD_BALANCER ServiceType = "LOAD_BALANCER"
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

type ServiceBindOptions struct {
	Name       *string  `field:"optional" json:"name" yaml:"name"`
	NodePort   *float64 `field:"optional" json:"nodePort" yaml:"nodePort"`
	Protocol   Protocol `field:"optional" json:"protocol" yaml:"protocol"`
	TargetPort *float64 `field:"optional" json:"targetPort" yaml:"targetPort"`
}

// AddDeploymentOptions is retained for compatibility with the cdk8s+ service
// API. It extends ServiceBindOptions with the service port to bind.
type AddDeploymentOptions struct {
	Name       *string  `field:"optional" json:"name" yaml:"name"`
	NodePort   *float64 `field:"optional" json:"nodePort" yaml:"nodePort"`
	Port       *float64 `field:"optional" json:"port" yaml:"port"`
	Protocol   Protocol `field:"optional" json:"protocol" yaml:"protocol"`
	TargetPort *float64 `field:"optional" json:"targetPort" yaml:"targetPort"`
}

type ServicePort struct {
	Port       *float64 `field:"required" json:"port" yaml:"port"`
	Name       *string  `field:"optional" json:"name" yaml:"name"`
	NodePort   *float64 `field:"optional" json:"nodePort" yaml:"nodePort"`
	Protocol   Protocol `field:"optional" json:"protocol" yaml:"protocol"`
	TargetPort *float64 `field:"optional" json:"targetPort" yaml:"targetPort"`
}

type ServiceProps struct {
	Metadata                 *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	Selector                 IPodSelector             `field:"optional" json:"selector" yaml:"selector"`
	ClusterIP                *string                  `field:"optional" json:"clusterIP" yaml:"clusterIP"`
	ExternalIPs              *[]*string               `field:"optional" json:"externalIPs" yaml:"externalIPs"`
	Type                     ServiceType              `field:"optional" json:"type" yaml:"type"`
	Ports                    *[]*ServicePort          `field:"optional" json:"ports" yaml:"ports"`
	ExternalName             *string                  `field:"optional" json:"externalName" yaml:"externalName"`
	LoadBalancerSourceRanges *[]*string               `field:"optional" json:"loadBalancerSourceRanges" yaml:"loadBalancerSourceRanges"`
	PublishNotReadyAddresses *bool                    `field:"optional" json:"publishNotReadyAddresses" yaml:"publishNotReadyAddresses"`
}

type IPodSelector interface {
	constructs.IConstruct
	ToPodSelectorConfig() *PodSelectorConfig
}

type Service interface {
	Resource
	ClusterIP() *string
	Type() ServiceType
	ExternalName() *string
	Ports() *[]*ServicePort
	Port() *float64
	Bind(port *float64, options *ServiceBindOptions)
	Select(selector IPodSelector)
	SelectLabel(key, value *string)
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

// HttpIngressPathType controls matching for an ingress HTTP path.
type HttpIngressPathType string

const (
	HttpIngressPathType_EXACT                   HttpIngressPathType = "EXACT"
	HttpIngressPathType_PREFIX                  HttpIngressPathType = "PREFIX"
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

type ExposeServiceViaIngressOptions struct {
	PathType HttpIngressPathType `field:"optional" json:"pathType" yaml:"pathType"`
	Ingress  Ingress             `field:"optional" json:"ingress" yaml:"ingress"`
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
