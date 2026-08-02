package cdk8splus34

import "github.com/Chriscbr/purecdk8s/jsii"

// Pod DNS policies.
type DnsPolicy string

const (
	// Any DNS query that does not match the configured cluster domain suffix, such as "www.kubernetes.io", is forwarded to the upstream nameserver inherited from the node. Cluster administrators may have extra stub-domain and upstream DNS servers configured.
	DnsPolicy_CLUSTER_FIRST DnsPolicy = "CLUSTER_FIRST"
	// For Pods running with hostNetwork, you should explicitly set its DNS policy "ClusterFirstWithHostNet".
	DnsPolicy_CLUSTER_FIRST_WITH_HOST_NET DnsPolicy = "CLUSTER_FIRST_WITH_HOST_NET"
	// The Pod inherits the name resolution configuration from the node that the pods run on.
	DnsPolicy_DEFAULT DnsPolicy = "DEFAULT"
	// It allows a Pod to ignore DNS settings from the Kubernetes environment.
	//
	// All DNS settings are supposed to be provided using the dnsConfig field in the Pod Spec.
	DnsPolicy_NONE DnsPolicy = "NONE"
)

func dnsPolicyManifestValue(value DnsPolicy) string {
	switch value {
	case DnsPolicy_CLUSTER_FIRST:
		return "ClusterFirst"
	case DnsPolicy_CLUSTER_FIRST_WITH_HOST_NET:
		return "ClusterFirstWithHostNet"
	case DnsPolicy_DEFAULT:
		return "Default"
	case DnsPolicy_NONE:
		return "None"
	default:
		return string(value)
	}
}

// Custom DNS option.
type DnsOption struct {
	// Option name.
	Name *string `field:"required" json:"name" yaml:"name"`
	// Option value. Default: - No value.
	Value *string `field:"optional" json:"value" yaml:"value"`
}

// HostAlias holds the mapping between IP and hostnames that will be injected as an entry in the pod's /etc/hosts file.
type HostAlias struct {
	// Hostnames for the chosen IP address.
	Hostnames *[]*string `field:"required" json:"hostnames" yaml:"hostnames"`
	// IP address of the host file entry.
	Ip *string `field:"required" json:"ip" yaml:"ip"`
}

// Properties for `PodDns`.
type PodDnsProps struct {
	// Specifies the hostname of the Pod. Default: - Set to a system-defined value.
	Hostname *string `field:"optional" json:"hostname" yaml:"hostname"`
	// If true the pod's hostname will be configured as the pod's FQDN, rather than the leaf name (the default).
	//
	// In Linux containers, this means setting the FQDN in the hostname field of the kernel (the nodename field of struct utsname). In Windows containers, this means setting the registry value of hostname for the registry key HKEY_LOCAL_MACHINE\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters to FQDN. If a pod does not have FQDN, this has no effect. Default: false.
	HostnameAsFQDN *bool `field:"optional" json:"hostnameAsFQDN" yaml:"hostnameAsFQDN"`
	// A list of IP addresses that will be used as DNS servers for the Pod.
	//
	// There can be at most 3 IP addresses specified. When the policy is set to "NONE", the list must contain at least one IP address, otherwise this property is optional. The servers listed will be combined to the base nameservers generated from the specified DNS policy with duplicate addresses removed.
	Nameservers *[]*string `field:"optional" json:"nameservers" yaml:"nameservers"`
	// List of objects where each object may have a name property (required) and a value property (optional).
	//
	// The contents in this property will be merged to the options generated from the specified DNS policy. Duplicate entries are removed.
	Options *[]*DnsOption `field:"optional" json:"options" yaml:"options"`
	// Set DNS policy for the pod.
	//
	// If policy is set to `None`, other configuration must be supplied. Default: DnsPolicy.CLUSTER_FIRST
	Policy DnsPolicy `field:"optional" json:"policy" yaml:"policy"`
	// A list of DNS search domains for hostname lookup in the Pod.
	//
	// When specified, the provided list will be merged into the base search domain names generated from the chosen DNS policy. Duplicate domain names are removed.
	//
	// Kubernetes allows for at most 6 search domains.
	Searches *[]*string `field:"optional" json:"searches" yaml:"searches"`
	// If specified, the fully qualified Pod hostname will be "<hostname>.<subdomain>.<pod namespace>.svc.<cluster domain>". Default: - No subdomain.
	Subdomain *string `field:"optional" json:"subdomain" yaml:"subdomain"`
}

// Holds dns settings of the pod.
type PodDns interface {
	// The configured hostname of the pod.
	//
	// Undefined means its set to a system-defined value.
	Hostname() *string
	// Whether or not the pods hostname is set to its FQDN.
	HostnameAsFQDN() *bool
	// Nameservers defined for this pod.
	Nameservers() *[]*string
	// Custom dns options defined for this pod.
	Options() *[]*DnsOption
	// The DNS policy of this pod.
	Policy() DnsPolicy
	// Search domains defined for this pod.
	Searches() *[]*string
	// The configured subdomain of the pod.
	Subdomain() *string
	// Add a nameserver.
	AddNameserver(nameservers ...*string)
	// Add a custom option.
	AddOption(options ...*DnsOption)
	// Add a search domain.
	AddSearch(searches ...*string)
	toManifest() map[string]interface{}
}

type podDnsImpl struct {
	hostname, subdomain *string
	hostnameAsFQDN      bool
	policy              DnsPolicy
	nameservers         []*string
	searches            []*string
	options             []*DnsOption
}

func NewPodDns(props *PodDnsProps) PodDns {
	if props == nil {
		props = &PodDnsProps{}
	}
	policy := props.Policy
	if policy == "" {
		policy = DnsPolicy_CLUSTER_FIRST
	}
	result := &podDnsImpl{hostname: props.Hostname, subdomain: props.Subdomain, policy: policy}
	if props.HostnameAsFQDN != nil {
		result.hostnameAsFQDN = *props.HostnameAsFQDN
	}
	if props.Nameservers != nil {
		result.nameservers = append(result.nameservers, (*props.Nameservers)...)
	}
	if props.Searches != nil {
		result.searches = append(result.searches, (*props.Searches)...)
	}
	if props.Options != nil {
		result.options = append(result.options, (*props.Options)...)
	}
	return result
}

func NewPodDns_Override(dns PodDns, props *PodDnsProps) {
	applyOverride(dns, NewPodDns(props), "PodDns")
}

func (p *podDnsImpl) Hostname() *string {
	return p.hostname
}

func (p *podDnsImpl) HostnameAsFQDN() *bool {
	return jsii.Bool(p.hostnameAsFQDN)
}

func (p *podDnsImpl) Policy() DnsPolicy {
	return p.policy
}

func (p *podDnsImpl) Subdomain() *string {
	return p.subdomain
}

func (p *podDnsImpl) Nameservers() *[]*string {
	result := append([]*string(nil), p.nameservers...)
	return &result
}

func (p *podDnsImpl) Searches() *[]*string {
	result := append([]*string(nil), p.searches...)
	return &result
}

func (p *podDnsImpl) Options() *[]*DnsOption {
	result := append([]*DnsOption(nil), p.options...)
	return &result
}

func (p *podDnsImpl) AddNameserver(nameservers ...*string) {
	p.nameservers = append(p.nameservers, nameservers...)
}

func (p *podDnsImpl) AddSearch(searches ...*string) {
	p.searches = append(p.searches, searches...)
}

func (p *podDnsImpl) AddOption(options ...*DnsOption) {
	p.options = append(p.options, options...)
}

func (p *podDnsImpl) toManifest() map[string]interface{} {
	if p.policy == DnsPolicy_NONE && len(p.nameservers) == 0 {
		panic("When dns policy is set to NONE, at least one nameserver is required")
	}
	if len(p.nameservers) > 3 {
		panic("There can be at most 3 nameservers specified")
	}
	if len(p.searches) > 6 {
		panic("There can be at most 6 search domains specified")
	}
	result := map[string]interface{}{"dnsPolicy": dnsPolicyManifestValue(p.policy), "setHostnameAsFQDN": p.hostnameAsFQDN}
	if p.hostname != nil {
		result["hostname"] = p.hostname
	}
	if p.subdomain != nil {
		result["subdomain"] = p.subdomain
	}
	config := map[string]interface{}{}
	if len(p.nameservers) > 0 {
		config["nameservers"] = p.nameservers
	}
	if len(p.searches) > 0 {
		config["searches"] = p.searches
	}
	if len(p.options) > 0 {
		config["options"] = p.options
	}
	if len(config) > 0 {
		result["dnsConfig"] = config
	}
	return result
}

type FsGroupChangePolicy string

const (
	// Only change permissions and ownership if permission and ownership of root directory does not match with expected permissions of the volume.
	//
	// This could help shorten the time it takes to change ownership and permission of a volume.
	FsGroupChangePolicy_ON_ROOT_MISMATCH FsGroupChangePolicy = "ON_ROOT_MISMATCH"
	// Always change permission and ownership of the volume when volume is mounted.
	FsGroupChangePolicy_ALWAYS FsGroupChangePolicy = "ALWAYS"
)

func fsGroupChangePolicyManifestValue(value FsGroupChangePolicy) string {
	switch value {
	case FsGroupChangePolicy_ON_ROOT_MISMATCH:
		return "OnRootMismatch"
	case FsGroupChangePolicy_ALWAYS:
		return "Always"
	default:
		return string(value)
	}
}

// Sysctl defines a kernel parameter to be set.
type Sysctl struct {
	// Name of a property to set.
	Name *string `field:"required" json:"name" yaml:"name"`
	// Value of a property to set.
	Value *string `field:"required" json:"value" yaml:"value"`
}

// Properties for `PodSecurityContext`.
type PodSecurityContextProps struct {
	// Indicates that the container must run as a non-root user.
	//
	// If true, the Kubelet will validate the image at runtime to ensure that it does not run as UID 0 (root) and fail to start the container if it does. Default: true.
	EnsureNonRoot *bool `field:"optional" json:"ensureNonRoot" yaml:"ensureNonRoot"`
	// Modify the ownership and permissions of pod volumes to this GID. Default: - Volume ownership is not changed.
	FsGroup *float64 `field:"optional" json:"fsGroup" yaml:"fsGroup"`
	// Defines behavior of changing ownership and permission of the volume before being exposed inside Pod.
	//
	// This field will only apply to volume types which support fsGroup based ownership(and permissions). It will have no effect on ephemeral volume types such as: secret, configmaps and emptydir. Default: FsGroupChangePolicy.ALWAYS
	FsGroupChangePolicy FsGroupChangePolicy `field:"optional" json:"fsGroupChangePolicy" yaml:"fsGroupChangePolicy"`
	// The GID to run the entrypoint of the container process. Default: - Group configured by container runtime.
	Group *float64 `field:"optional" json:"group" yaml:"group"`
	// Sysctls hold a list of namespaced sysctls used for the pod.
	//
	// Pods with unsupported sysctls (by the container runtime) might fail to launch. Default: - No sysctls.
	Sysctls *[]*Sysctl `field:"optional" json:"sysctls" yaml:"sysctls"`
	// The UID to run the entrypoint of the container process. Default: - User specified in image metadata.
	User *float64 `field:"optional" json:"user" yaml:"user"`
}

// Holds pod-level security attributes and common container settings.
type PodSecurityContext interface {
	EnsureNonRoot() *bool
	FsGroup() *float64
	FsGroupChangePolicy() FsGroupChangePolicy
	Group() *float64
	Sysctls() *[]*Sysctl
	User() *float64
	toManifest() map[string]interface{}
}

type podSecurityContextImpl struct {
	ensureNonRoot       bool
	fsGroup             *float64
	fsGroupChangePolicy FsGroupChangePolicy
	group               *float64
	sysctls             []*Sysctl
	user                *float64
}

func NewPodSecurityContext(props *PodSecurityContextProps) PodSecurityContext {
	if props == nil {
		props = &PodSecurityContextProps{}
	}
	result := &podSecurityContextImpl{ensureNonRoot: true, fsGroup: props.FsGroup, group: props.Group, user: props.User, fsGroupChangePolicy: props.FsGroupChangePolicy}
	if props.EnsureNonRoot != nil {
		result.ensureNonRoot = *props.EnsureNonRoot
	}
	if result.fsGroupChangePolicy == "" {
		result.fsGroupChangePolicy = FsGroupChangePolicy_ALWAYS
	}
	if props.Sysctls != nil {
		result.sysctls = append(result.sysctls, (*props.Sysctls)...)
	}
	return result
}

func NewPodSecurityContext_Override(context PodSecurityContext, props *PodSecurityContextProps) {
	applyOverride(context, NewPodSecurityContext(props), "PodSecurityContext")
}

func (p *podSecurityContextImpl) EnsureNonRoot() *bool {
	return jsii.Bool(p.ensureNonRoot)
}

func (p *podSecurityContextImpl) FsGroup() *float64 {
	return p.fsGroup
}

func (p *podSecurityContextImpl) FsGroupChangePolicy() FsGroupChangePolicy {
	return p.fsGroupChangePolicy
}

func (p *podSecurityContextImpl) Group() *float64 {
	return p.group
}

func (p *podSecurityContextImpl) User() *float64 {
	return p.user
}

func (p *podSecurityContextImpl) Sysctls() *[]*Sysctl {
	result := append([]*Sysctl(nil), p.sysctls...)
	return &result
}

func (p *podSecurityContextImpl) toManifest() map[string]interface{} {
	result := map[string]interface{}{"runAsNonRoot": p.ensureNonRoot, "fsGroupChangePolicy": fsGroupChangePolicyManifestValue(p.fsGroupChangePolicy)}
	if p.group != nil {
		result["runAsGroup"] = p.group
	}
	if p.user != nil {
		result["runAsUser"] = p.user
	}
	if p.fsGroup != nil {
		result["fsGroup"] = p.fsGroup
	}
	if len(p.sysctls) > 0 {
		result["sysctls"] = p.sysctls
	}
	return result
}
