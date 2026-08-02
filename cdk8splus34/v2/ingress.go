package cdk8splus34

import (
	"sort"

	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
)

type ServiceIngressBackendOptions struct {
	Port *float64 `field:"optional" json:"port" yaml:"port"`
}

type (
	IngressBackend interface{ manifest() interface{} }
	ingressBackend struct{ value interface{} }
)

func (b *ingressBackend) manifest() interface{} {
	return b.value
}

func IngressBackend_FromService(service Service, options *ServiceIngressBackendOptions) IngressBackend {
	if service == nil {
		panic("service is required")
	}
	port := service.Port()
	if options != nil && options.Port != nil {
		port = options.Port
	}
	if port == nil {
		panic("service port is required")
	}
	return &ingressBackend{value: map[string]interface{}{"service": map[string]interface{}{"name": service.Name(), "port": map[string]interface{}{"number": port}}}}
}

func IngressBackend_FromResource(resource IResource) IngressBackend {
	if resource == nil {
		panic("resource is required")
	}
	return &ingressBackend{value: map[string]interface{}{"resource": map[string]interface{}{"apiGroup": resource.ApiGroup(), "kind": resource.Kind(), "name": resource.Name()}}}
}

type IngressRule struct {
	Host     *string             `field:"optional" json:"host" yaml:"host"`
	Path     *string             `field:"optional" json:"path" yaml:"path"`
	Backend  IngressBackend      `field:"required" json:"backend" yaml:"backend"`
	PathType HttpIngressPathType `field:"optional" json:"pathType" yaml:"pathType"`
}

type IngressTls struct {
	Hosts  *[]*string `field:"optional" json:"hosts" yaml:"hosts"`
	Secret ISecret    `field:"optional" json:"secret" yaml:"secret"`
}

type IngressProps struct {
	Metadata       *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	ClassName      *string                  `field:"optional" json:"className" yaml:"className"`
	DefaultBackend IngressBackend           `field:"optional" json:"defaultBackend" yaml:"defaultBackend"`
	Rules          *[]*IngressRule          `field:"optional" json:"rules" yaml:"rules"`
	Tls            *[]*IngressTls           `field:"optional" json:"tls" yaml:"tls"`
}

type Ingress interface {
	Resource
	AddDefaultBackend(backend IngressBackend)
	AddHostDefaultBackend(host *string, backend IngressBackend)
	AddHostRule(host, path *string, backend IngressBackend, pathType HttpIngressPathType)
	AddRule(path *string, backend IngressBackend, pathType HttpIngressPathType)
	AddRules(rules ...*IngressRule)
	AddTls(tls *[]*IngressTls)
}

type ingressImpl struct {
	resourceBase
	className      *string
	defaultBackend IngressBackend
	rules          []*IngressRule
	tls            []*IngressTls
}

func NewIngress(scope constructs.Construct, id *string, props *IngressProps) Ingress {
	if props == nil {
		props = &IngressProps{}
	}
	result := &ingressImpl{className: props.ClassName, defaultBackend: props.DefaultBackend}
	manifest := map[string]interface{}{}
	result.resourceBase.initialize(result, scope, id, "networking.k8s.io/v1", "Ingress", "ingresses", props.Metadata, manifest)
	if props.Rules != nil {
		result.AddRules((*props.Rules)...)
	}
	if props.Tls != nil {
		result.tls = append(result.tls, (*props.Tls)...)
	}
	manifest["spec"] = cdk8s.Lazy_Any(lazyProducer{produce: func() interface{} { return result.toManifest() }})
	return result
}

func NewIngress_Override(ingress Ingress, scope constructs.Construct, id *string, props *IngressProps) {
	applyOverride(ingress, NewIngress(scope, id, props), "Ingress")
}

func Ingress_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func (i *ingressImpl) AddDefaultBackend(backend IngressBackend) {
	i.defaultBackend = backend
}

func (i *ingressImpl) AddHostDefaultBackend(host *string, backend IngressBackend) {
	i.AddRules(&IngressRule{Host: host, Backend: backend})
}

func (i *ingressImpl) AddHostRule(host, path *string, backend IngressBackend, pathType HttpIngressPathType) {
	i.AddRules(&IngressRule{Host: host, Path: path, Backend: backend, PathType: pathType})
}

func (i *ingressImpl) AddRule(path *string, backend IngressBackend, pathType HttpIngressPathType) {
	i.AddHostRule(nil, path, backend, pathType)
}

func (i *ingressImpl) AddRules(rules ...*IngressRule) {
	for _, rule := range rules {
		if rule == nil || rule.Backend == nil {
			panic("ingress rule backend is required")
		}
		if rule.Host == nil && rule.Path == nil {
			if i.defaultBackend != nil {
				panic("a default backend is already defined for this ingress")
			}
			i.defaultBackend = rule.Backend
			continue
		}
		i.rules = append(i.rules, rule)
	}
}

func (i *ingressImpl) AddTls(tls *[]*IngressTls) {
	if tls != nil {
		i.tls = append(i.tls, (*tls)...)
	}
}

func (i *ingressImpl) toManifest() interface{} {
	result := map[string]interface{}{}
	if i.className != nil {
		result["ingressClassName"] = i.className
	}
	if i.defaultBackend != nil {
		result["defaultBackend"] = i.defaultBackend.manifest()
	}
	if len(i.rules) > 0 {
		type hostPaths struct {
			host  *string
			paths []map[string]interface{}
		}
		byHost := map[string]*hostPaths{}
		orderedHosts := []string{}
		for _, rule := range i.rules {
			if rule == nil || rule.Backend == nil {
				panic("ingress rule backend is required")
			}
			pathType := rule.PathType
			if pathType == "" {
				pathType = HttpIngressPathType_PREFIX
			}
			path := "/"
			if rule.Path != nil {
				path = *rule.Path
			}
			hostKey := ""
			if rule.Host != nil {
				hostKey = *rule.Host
			}
			entry := byHost[hostKey]
			if entry == nil {
				entry = &hostPaths{host: rule.Host}
				byHost[hostKey] = entry
				orderedHosts = append(orderedHosts, hostKey)
			}
			entry.paths = append(entry.paths, map[string]interface{}{"path": path, "pathType": string(pathType), "backend": rule.Backend.manifest()})
		}
		rules := make([]interface{}, 0, len(orderedHosts))
		for _, host := range orderedHosts {
			entry := byHost[host]
			sort.SliceStable(entry.paths, func(a, b int) bool {
				return entry.paths[a]["path"].(string) < entry.paths[b]["path"].(string)
			})
			paths := make([]interface{}, len(entry.paths))
			for index, path := range entry.paths {
				paths[index] = path
			}
			item := map[string]interface{}{"http": map[string]interface{}{"paths": paths}}
			if entry.host != nil {
				item["host"] = entry.host
			}
			rules = append(rules, item)
		}
		result["rules"] = rules
	}
	return result
}

// ISecret is declared here because container and ingress APIs reference it.
// Its concrete resource implementation is provided with the Secret constructs.
type ISecret interface {
	IResource
	EnvValue(key *string, options *EnvValueFromSecretOptions) EnvValue
}
