package cdk8splus34

import "github.com/purecdk8s/purecdk8s/jsii"

// Capability is a POSIX capability for a container process.
type Capability string

const (
	Capability_ALL                Capability = "ALL"
	Capability_AUDIT_CONTROL      Capability = "AUDIT_CONTROL"
	Capability_AUDIT_READ         Capability = "AUDIT_READ"
	Capability_AUDIT_WRITE        Capability = "AUDIT_WRITE"
	Capability_BLOCK_SUSPEND      Capability = "BLOCK_SUSPEND"
	Capability_BPF                Capability = "BPF"
	Capability_CHECKPOINT_RESTORE Capability = "CHECKPOINT_RESTORE"
	Capability_CHOWN              Capability = "CHOWN"
	Capability_DAC_OVERRIDE       Capability = "DAC_OVERRIDE"
	Capability_DAC_READ_SEARCH    Capability = "DAC_READ_SEARCH"
	Capability_FOWNER             Capability = "FOWNER"
	Capability_FSETID             Capability = "FSETID"
	Capability_IPC_LOCK           Capability = "IPC_LOCK"
	Capability_IPC_OWNER          Capability = "IPC_OWNER"
	Capability_KILL               Capability = "KILL"
	Capability_LEASE              Capability = "LEASE"
	Capability_LINUX_IMMUTABLE    Capability = "LINUX_IMMUTABLE"
	Capability_MAC_ADMIN          Capability = "MAC_ADMIN"
	Capability_MAC_OVERRIDE       Capability = "MAC_OVERRIDE"
	Capability_MKNOD              Capability = "MKNOD"
	Capability_NET_ADMIN          Capability = "NET_ADMIN"
	Capability_NET_BIND_SERVICE   Capability = "NET_BIND_SERVICE"
	Capability_NET_BROADCAST      Capability = "NET_BROADCAST"
	Capability_NET_RAW            Capability = "NET_RAW"
	Capability_PERFMON            Capability = "PERFMON"
	Capability_SETGID             Capability = "SETGID"
	Capability_SETFCAP            Capability = "SETFCAP"
	Capability_SETPCAP            Capability = "SETPCAP"
	Capability_SETUID             Capability = "SETUID"
	Capability_SYS_ADMIN          Capability = "SYS_ADMIN"
	Capability_SYS_BOOT           Capability = "SYS_BOOT"
	Capability_SYS_CHROOT         Capability = "SYS_CHROOT"
	Capability_SYS_MODULE         Capability = "SYS_MODULE"
	Capability_SYS_NICE           Capability = "SYS_NICE"
	Capability_SYS_PACCT          Capability = "SYS_PACCT"
	Capability_SYS_PTRACE         Capability = "SYS_PTRACE"
	Capability_SYS_RAWIO          Capability = "SYS_RAWIO"
	Capability_SYS_RESOURCE       Capability = "SYS_RESOURCE"
	Capability_SYS_TIME           Capability = "SYS_TIME"
	Capability_SYS_TTY_CONFIG     Capability = "SYS_TTY_CONFIG"
	Capability_SYSLOG             Capability = "SYSLOG"
	Capability_WAKE_ALARM         Capability = "WAKE_ALARM"
)

type SeccompProfileType string

const (
	SeccompProfileType_LOCALHOST       SeccompProfileType = "LOCALHOST"
	SeccompProfileType_RUNTIME_DEFAULT SeccompProfileType = "RUNTIME_DEFAULT"
	SeccompProfileType_UNCONFINED      SeccompProfileType = "UNCONFINED"
)

type SeccompProfile struct {
	Type             SeccompProfileType `field:"required" json:"type" yaml:"type"`
	LocalhostProfile *string            `field:"optional" json:"localhostProfile" yaml:"localhostProfile"`
}

type ContainerSecutiryContextCapabilities struct {
	Add  *[]Capability `field:"optional" json:"add" yaml:"add"`
	Drop *[]Capability `field:"optional" json:"drop" yaml:"drop"`
}

type ContainerSecurityContextProps struct {
	AllowPrivilegeEscalation *bool                                 `field:"optional" json:"allowPrivilegeEscalation" yaml:"allowPrivilegeEscalation"`
	Capabilities             *ContainerSecutiryContextCapabilities `field:"optional" json:"capabilities" yaml:"capabilities"`
	EnsureNonRoot            *bool                                 `field:"optional" json:"ensureNonRoot" yaml:"ensureNonRoot"`
	Group                    *float64                              `field:"optional" json:"group" yaml:"group"`
	Privileged               *bool                                 `field:"optional" json:"privileged" yaml:"privileged"`
	ReadOnlyRootFilesystem   *bool                                 `field:"optional" json:"readOnlyRootFilesystem" yaml:"readOnlyRootFilesystem"`
	SeccompProfile           *SeccompProfile                       `field:"optional" json:"seccompProfile" yaml:"seccompProfile"`
	User                     *float64                              `field:"optional" json:"user" yaml:"user"`
}

type ContainerSecurityContext interface {
	AllowPrivilegeEscalation() *bool
	Capabilities() *ContainerSecutiryContextCapabilities
	EnsureNonRoot() *bool
	Group() *float64
	Privileged() *bool
	ReadOnlyRootFilesystem() *bool
	SeccompProfile() *SeccompProfile
	User() *float64
	toManifest() map[string]interface{}
}

type containerSecurityContextImpl struct {
	allowPrivilegeEscalation *bool
	capabilities             *ContainerSecutiryContextCapabilities
	ensureNonRoot            *bool
	group                    *float64
	privileged               *bool
	readOnlyRootFilesystem   *bool
	seccompProfile           *SeccompProfile
	user                     *float64
}

func NewContainerSecurityContext(props *ContainerSecurityContextProps) ContainerSecurityContext {
	if props == nil {
		props = &ContainerSecurityContextProps{}
	}
	if props.SeccompProfile != nil && props.SeccompProfile.LocalhostProfile != nil && props.SeccompProfile.Type != SeccompProfileType_LOCALHOST {
		panic("localhostProfile must only be set if type is \"Localhost\"")
	}
	ensureNonRoot, privileged, readOnly, allowPrivilegeEscalation := jsii.Bool(true), jsii.Bool(false), jsii.Bool(true), jsii.Bool(false)
	if props.EnsureNonRoot != nil {
		ensureNonRoot = props.EnsureNonRoot
	}
	if props.Privileged != nil {
		privileged = props.Privileged
	}
	if props.ReadOnlyRootFilesystem != nil {
		readOnly = props.ReadOnlyRootFilesystem
	}
	if props.AllowPrivilegeEscalation != nil {
		allowPrivilegeEscalation = props.AllowPrivilegeEscalation
	}
	return &containerSecurityContextImpl{allowPrivilegeEscalation: allowPrivilegeEscalation, capabilities: props.Capabilities, ensureNonRoot: ensureNonRoot, group: props.Group, privileged: privileged, readOnlyRootFilesystem: readOnly, seccompProfile: props.SeccompProfile, user: props.User}
}

func NewContainerSecurityContext_Override(context ContainerSecurityContext, props *ContainerSecurityContextProps) {
	applyOverride(context, NewContainerSecurityContext(props), "ContainerSecurityContext")
}

func (c *containerSecurityContextImpl) AllowPrivilegeEscalation() *bool {
	return c.allowPrivilegeEscalation
}

func (c *containerSecurityContextImpl) Capabilities() *ContainerSecutiryContextCapabilities {
	return c.capabilities
}

func (c *containerSecurityContextImpl) EnsureNonRoot() *bool {
	return c.ensureNonRoot
}

func (c *containerSecurityContextImpl) Group() *float64 {
	return c.group
}

func (c *containerSecurityContextImpl) Privileged() *bool {
	return c.privileged
}

func (c *containerSecurityContextImpl) ReadOnlyRootFilesystem() *bool {
	return c.readOnlyRootFilesystem
}

func (c *containerSecurityContextImpl) SeccompProfile() *SeccompProfile {
	return c.seccompProfile
}

func (c *containerSecurityContextImpl) User() *float64 {
	return c.user
}

func (c *containerSecurityContextImpl) toManifest() map[string]interface{} {
	result := map[string]interface{}{"runAsNonRoot": c.ensureNonRoot, "privileged": c.privileged, "readOnlyRootFilesystem": c.readOnlyRootFilesystem, "allowPrivilegeEscalation": c.allowPrivilegeEscalation}
	if c.group != nil {
		result["runAsGroup"] = c.group
	}
	if c.user != nil {
		result["runAsUser"] = c.user
	}
	if c.capabilities != nil {
		caps := map[string]interface{}{}
		if c.capabilities.Add != nil {
			caps["add"] = capabilitiesManifest(*c.capabilities.Add)
		}
		if c.capabilities.Drop != nil {
			caps["drop"] = capabilitiesManifest(*c.capabilities.Drop)
		}
		result["capabilities"] = caps
	}
	if c.seccompProfile != nil {
		profile := map[string]interface{}{"type": seccompProfileTypeManifest(c.seccompProfile.Type)}
		if c.seccompProfile.LocalhostProfile != nil {
			profile["localhostProfile"] = c.seccompProfile.LocalhostProfile
		}
		result["seccompProfile"] = profile
	}
	return result
}

func capabilitiesManifest(capabilities []Capability) []interface{} {
	result := make([]interface{}, 0, len(capabilities))
	for _, capability := range capabilities {
		result = append(result, string(capability))
	}
	return result
}

func seccompProfileTypeManifest(value SeccompProfileType) string {
	switch value {
	case SeccompProfileType_LOCALHOST:
		return "Localhost"
	case SeccompProfileType_RUNTIME_DEFAULT:
		return "RuntimeDefault"
	case SeccompProfileType_UNCONFINED:
		return "Unconfined"
	default:
		panic("invalid seccomp profile type")
	}
}

type ContainerRestartPolicy string

const ContainerRestartPolicy_ALWAYS ContainerRestartPolicy = "ALWAYS"

func containerRestartPolicyManifest(value ContainerRestartPolicy) string {
	if value != ContainerRestartPolicy_ALWAYS {
		panic("invalid container restart policy")
	}
	return "Always"
}

type MountPropagation string

const (
	MountPropagation_NONE              MountPropagation = "NONE"
	MountPropagation_HOST_TO_CONTAINER MountPropagation = "HOST_TO_CONTAINER"
	MountPropagation_BIDIRECTIONAL     MountPropagation = "BIDIRECTIONAL"
)

func mountPropagationManifest(value MountPropagation) string {
	switch value {
	case MountPropagation_NONE:
		return "None"
	case MountPropagation_HOST_TO_CONTAINER:
		return "HostToContainer"
	case MountPropagation_BIDIRECTIONAL:
		return "Bidirectional"
	default:
		panic("invalid mount propagation")
	}
}
