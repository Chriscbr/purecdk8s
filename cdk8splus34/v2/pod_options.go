package cdk8splus34

import "github.com/purecdk8s/purecdk8s/jsii"

// DnsPolicy controls how a pod resolves DNS names.
type DnsPolicy string

const (
	DnsPolicy_CLUSTER_FIRST               DnsPolicy = "ClusterFirst"
	DnsPolicy_CLUSTER_FIRST_WITH_HOST_NET DnsPolicy = "ClusterFirstWithHostNet"
	DnsPolicy_DEFAULT                     DnsPolicy = "Default"
	DnsPolicy_NONE                        DnsPolicy = "None"
)

// DnsOption is a custom DNS resolver option.
type DnsOption struct {
	Name  *string `field:"required" json:"name" yaml:"name"`
	Value *string `field:"optional" json:"value" yaml:"value"`
}

// HostAlias adds host names to a pod's /etc/hosts file.
type HostAlias struct {
	Hostnames *[]*string `field:"required" json:"hostnames" yaml:"hostnames"`
	Ip        *string    `field:"required" json:"ip" yaml:"ip"`
}

// PodDnsProps configures Pod DNS settings.
type PodDnsProps struct {
	Hostname       *string       `field:"optional" json:"hostname" yaml:"hostname"`
	HostnameAsFQDN *bool         `field:"optional" json:"hostnameAsFQDN" yaml:"hostnameAsFQDN"`
	Nameservers    *[]*string    `field:"optional" json:"nameservers" yaml:"nameservers"`
	Options        *[]*DnsOption `field:"optional" json:"options" yaml:"options"`
	Policy         DnsPolicy     `field:"optional" json:"policy" yaml:"policy"`
	Searches       *[]*string    `field:"optional" json:"searches" yaml:"searches"`
	Subdomain      *string       `field:"optional" json:"subdomain" yaml:"subdomain"`
}

// PodDns holds mutable pod DNS settings.
type PodDns interface {
	Hostname() *string
	HostnameAsFQDN() *bool
	Nameservers() *[]*string
	Options() *[]*DnsOption
	Policy() DnsPolicy
	Searches() *[]*string
	Subdomain() *string
	AddNameserver(nameservers ...*string)
	AddOption(options ...*DnsOption)
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
	result := map[string]interface{}{"dnsPolicy": string(p.policy), "setHostnameAsFQDN": p.hostnameAsFQDN}
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

// FsGroupChangePolicy controls when Kubernetes changes volume ownership.
type FsGroupChangePolicy string

const (
	FsGroupChangePolicy_ON_ROOT_MISMATCH FsGroupChangePolicy = "OnRootMismatch"
	FsGroupChangePolicy_ALWAYS           FsGroupChangePolicy = "Always"
)

// Sysctl defines a kernel parameter to be set for a pod.
type Sysctl struct {
	Name  *string `field:"required" json:"name" yaml:"name"`
	Value *string `field:"required" json:"value" yaml:"value"`
}

// PodSecurityContextProps configures pod-level security settings.
type PodSecurityContextProps struct {
	EnsureNonRoot       *bool               `field:"optional" json:"ensureNonRoot" yaml:"ensureNonRoot"`
	FsGroup             *float64            `field:"optional" json:"fsGroup" yaml:"fsGroup"`
	FsGroupChangePolicy FsGroupChangePolicy `field:"optional" json:"fsGroupChangePolicy" yaml:"fsGroupChangePolicy"`
	Group               *float64            `field:"optional" json:"group" yaml:"group"`
	Sysctls             *[]*Sysctl          `field:"optional" json:"sysctls" yaml:"sysctls"`
	User                *float64            `field:"optional" json:"user" yaml:"user"`
}

// PodSecurityContext holds pod-level security attributes.
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
	result := map[string]interface{}{"runAsNonRoot": p.ensureNonRoot, "fsGroupChangePolicy": string(p.fsGroupChangePolicy)}
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
