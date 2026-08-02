package cdk8splus34

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
)

// Options for setting up backends for ingress rules.
type ServiceIngressBackendOptions struct {
	// The port to use to access the service.
	//
	//   - This option will fail if the service does not expose any ports.
	//   - If the service exposes multiple ports, this option must be specified.
	//   - If the service exposes a single port, this option is optional and if specified, it must be the same port exposed by the service.
	//
	// Default: - if the service exposes a single port, this port will be used.
	Port *float64 `field:"optional" json:"port" yaml:"port"`
}

type (
	// The backend for an ingress path.
	IngressBackend interface{ manifest() interface{} }
	ingressBackend struct{ value interface{} }
)

func (b *ingressBackend) manifest() interface{} {
	return b.value
}

// A Kubernetes `Service` to use as the backend for this path.
func IngressBackend_FromService(service Service, options *ServiceIngressBackendOptions) IngressBackend {
	if service == nil {
		panic("service is required")
	}
	ports := *service.Ports()
	if len(ports) == 0 {
		panic("service does not expose any ports")
	}
	var port *float64
	if options == nil || options.Port == nil {
		if len(ports) > 1 {
			panic("unable to determine service port since service exposes multiple ports")
		}
		port = ports[0].Port
	} else {
		port = options.Port
		for _, exposed := range ports {
			if exposed != nil && exposed.Port != nil && *exposed.Port == *port {
				return &ingressBackend{value: map[string]interface{}{"service": map[string]interface{}{"name": service.Name(), "port": map[string]interface{}{"number": port}}}}
			}
		}
		if len(ports) == 1 {
			panic(fmt.Sprintf("backend defines port %s but service exposes port %s", numberString(*port), numberString(*ports[0].Port)))
		}
		values := make([]string, 0, len(ports))
		for _, exposed := range ports {
			if exposed != nil && exposed.Port != nil {
				values = append(values, numberString(*exposed.Port))
			}
		}
		panic(fmt.Sprintf("service exposes ports %s but backend is defined to use port %s", strings.Join(values, ","), numberString(*port)))
	}
	return &ingressBackend{value: map[string]interface{}{"service": map[string]interface{}{"name": service.Name(), "port": map[string]interface{}{"number": port}}}}
}

// A Resource backend is an ObjectRef to another Kubernetes resource within the same namespace as the Ingress object.
//
// A common usage for a Resource backend is to ingress data to an object storage backend with static assets.
func IngressBackend_FromResource(resource IResource) IngressBackend {
	if resource == nil {
		panic("resource is required")
	}
	return &ingressBackend{value: map[string]interface{}{"resource": map[string]interface{}{"apiGroup": resource.ApiGroup(), "kind": resource.Kind(), "name": resource.Name()}}}
}

// Represents the rules mapping the paths under a specified host to the related backend services.
//
// Incoming requests are first evaluated for a host match, then routed to the backend associated with the matching path.
type IngressRule struct {
	// Host is the fully qualified domain name of a network host, as defined by RFC 3986.
	//
	// Note the following deviations from the "host" part of the URI as defined in the RFC: 1. IPs are not allowed. Currently an IngressRuleValue can only apply to the IP in the Spec of the parent Ingress. 2. The `:` delimiter is not respected because ports are not allowed. Currently the port of an Ingress is implicitly :80 for http and :443 for https. Both these may change in the future. Incoming requests are matched against the host before the IngressRuleValue. Default: - If the host is unspecified, the Ingress routes all traffic based on the specified IngressRuleValue.
	Host *string `field:"optional" json:"host" yaml:"host"`
	// Path is an extended POSIX regex as defined by IEEE Std 1003.1, (i.e this follows the egrep/unix syntax, not the perl syntax) matched against the path of an incoming request. Currently it can contain characters disallowed from the conventional "path" part of a URL as defined by RFC 3986. Paths must begin with a '/'. Default: - If unspecified, the path defaults to a catch all sending traffic to the backend.
	Path *string `field:"optional" json:"path" yaml:"path"`
	// Backend defines the referenced service endpoint to which the traffic will be forwarded to.
	Backend IngressBackend `field:"required" json:"backend" yaml:"backend"`
	// Specify how the path is matched against request paths.
	//
	// By default, path types will be matched by prefix. See: https://kubernetes.io/docs/concepts/services-networking/ingress/#path-types
	PathType HttpIngressPathType `field:"optional" json:"pathType" yaml:"pathType"`
}

// Represents the TLS configuration mapping that is passed to the ingress controller for SSL termination.
type IngressTls struct {
	// Hosts are a list of hosts included in the TLS certificate.
	//
	// The values in this list must match the name/s used in the TLS Secret. Default: - If unspecified, it defaults to the wildcard host setting for the loadbalancer controller fulfilling this Ingress.
	Hosts *[]*string `field:"optional" json:"hosts" yaml:"hosts"`
	// Secret is the secret that contains the certificate and key used to terminate SSL traffic on 443.
	//
	// If the SNI host in a listener conflicts with the "Host" header field used by an IngressRule, the SNI host is used for termination and value of the Host header is used for routing. Default: - If unspecified, it allows SSL routing based on SNI hostname.
	Secret ISecret `field:"optional" json:"secret" yaml:"secret"`
}

// Properties for `Ingress`.
type IngressProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// Class Name for this ingress.
	//
	// This field is a reference to an IngressClass resource that contains additional Ingress configuration, including the name of the Ingress controller.
	ClassName *string `field:"optional" json:"className" yaml:"className"`
	// The default backend services requests that do not match any rule.
	//
	// Using this option or the `addDefaultBackend()` method is equivalent to adding a rule with both `path` and `host` undefined.
	DefaultBackend IngressBackend `field:"optional" json:"defaultBackend" yaml:"defaultBackend"`
	// Routing rules for this ingress.
	//
	// Each rule must define an `IngressBackend` that will receive the requests that match this rule. If both `host` and `path` are not specifiec, this backend will be used as the default backend of the ingress.
	//
	// You can also add rules later using `addRule()`, `addHostRule()`, `addDefaultBackend()` and `addHostDefaultBackend()`.
	Rules *[]*IngressRule `field:"optional" json:"rules" yaml:"rules"`
	// TLS settings for this ingress.
	//
	// Using this option tells the ingress controller to expose a TLS endpoint. Currently the Ingress only supports a single TLS port, 443. If multiple members of this list specify different hosts, they will be multiplexed on the same port according to the hostname specified through the SNI TLS extension, if the ingress controller fulfilling the ingress supports SNI.
	Tls *[]*IngressTls `field:"optional" json:"tls" yaml:"tls"`
}

// Ingress is a collection of rules that allow inbound connections to reach the endpoints defined by a backend.
//
// An Ingress can be configured to give services externally-reachable urls, load balance traffic, terminate SSL, offer name based virtual hosting etc.
type Ingress interface {
	Resource
	// Defines the default backend for this ingress.
	//
	// A default backend capable of servicing requests that don't match any rule.
	AddDefaultBackend(backend IngressBackend)
	// Specify a default backend for a specific host name.
	//
	// This backend will be used as a catch-all for requests targeted to this host name (the `Host` header matches this value).
	AddHostDefaultBackend(host *string, backend IngressBackend)
	// Adds an ingress rule applied to requests to a specific host and a specific HTTP path (the `Host` header matches this value).
	AddHostRule(host, path *string, backend IngressBackend, pathType HttpIngressPathType)
	// Adds an ingress rule applied to requests sent to a specific HTTP path.
	AddRule(path *string, backend IngressBackend, pathType HttpIngressPathType)
	// Adds rules to this ingress.
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

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func Ingress_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func (i *ingressImpl) AddDefaultBackend(backend IngressBackend) {
	if backend == nil {
		panic("ingress backend is required")
	}
	if i.defaultBackend != nil {
		panic("a default backend is already defined for this ingress")
	}
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
			i.AddDefaultBackend(rule.Backend)
			continue
		}
		path := "/"
		if rule.Path != nil {
			path = *rule.Path
		}
		if !strings.HasPrefix(path, "/") {
			panic(fmt.Sprintf("ingress paths must begin with a \"/\": %s", path))
		}
		host := ""
		if rule.Host != nil {
			host = *rule.Host
		}
		for _, existing := range i.rules {
			existingHost := ""
			if existing.Host != nil {
				existingHost = *existing.Host
			}
			existingPath := "/"
			if existing.Path != nil {
				existingPath = *existing.Path
			}
			if existingHost == host && existingPath == path {
				panic(fmt.Sprintf("there is already an ingress rule for %s%s", host, path))
			}
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
	if i.defaultBackend == nil && len(i.rules) == 0 {
		panic("ingress with no rules or default backend")
	}
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
			entry.paths = append(entry.paths, map[string]interface{}{"path": path, "pathType": httpIngressPathTypeManifestValue(pathType), "backend": rule.Backend.manifest()})
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
	if len(i.tls) > 0 {
		tlsEntries := make([]interface{}, 0, len(i.tls))
		for _, tls := range i.tls {
			if tls == nil {
				panic("ingress TLS entry is required")
			}
			entry := map[string]interface{}{}
			if tls.Hosts != nil {
				entry["hosts"] = tls.Hosts
			}
			if tls.Secret != nil {
				entry["secretName"] = tls.Secret.Name()
			}
			tlsEntries = append(tlsEntries, entry)
		}
		result["tls"] = tlsEntries
	}
	return result
}

type ISecret interface {
	IResource
	// Returns EnvValue object from a secret's key.
	EnvValue(key *string, options *EnvValueFromSecretOptions) EnvValue
}
