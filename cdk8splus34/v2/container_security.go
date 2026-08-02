package cdk8splus34

import "github.com/Chriscbr/purecdk8s/jsii"

// Capability - complete list of POSIX capabilities.
type Capability string

const (
	// ALL.
	Capability_ALL Capability = "ALL"
	// CAP_AUDIT_CONTROL.
	Capability_AUDIT_CONTROL Capability = "AUDIT_CONTROL"
	// CAP_AUDIT_READ.
	Capability_AUDIT_READ Capability = "AUDIT_READ"
	// CAP_AUDIT_WRITE.
	Capability_AUDIT_WRITE Capability = "AUDIT_WRITE"
	// CAP_BLOCK_SUSPEND.
	Capability_BLOCK_SUSPEND Capability = "BLOCK_SUSPEND"
	// CAP_BPF.
	Capability_BPF Capability = "BPF"
	// CAP_CHECKPOINT_RESTORE.
	Capability_CHECKPOINT_RESTORE Capability = "CHECKPOINT_RESTORE"
	// CAP_CHOWN.
	Capability_CHOWN Capability = "CHOWN"
	// CAP_DAC_OVERRIDE.
	Capability_DAC_OVERRIDE Capability = "DAC_OVERRIDE"
	// CAP_DAC_READ_SEARCH.
	Capability_DAC_READ_SEARCH Capability = "DAC_READ_SEARCH"
	// CAP_FOWNER.
	Capability_FOWNER Capability = "FOWNER"
	// CAP_FSETID.
	Capability_FSETID Capability = "FSETID"
	// CAP_IPC_LOCK.
	Capability_IPC_LOCK Capability = "IPC_LOCK"
	// CAP_IPC_OWNER.
	Capability_IPC_OWNER Capability = "IPC_OWNER"
	// CAP_KILL.
	Capability_KILL Capability = "KILL"
	// CAP_LEASE.
	Capability_LEASE Capability = "LEASE"
	// CAP_LINUX_IMMUTABLE.
	Capability_LINUX_IMMUTABLE Capability = "LINUX_IMMUTABLE"
	// CAP_MAC_ADMIN.
	Capability_MAC_ADMIN Capability = "MAC_ADMIN"
	// CAP_MAC_OVERRIDE.
	Capability_MAC_OVERRIDE Capability = "MAC_OVERRIDE"
	// CAP_MKNOD.
	Capability_MKNOD Capability = "MKNOD"
	// CAP_NET_ADMIN.
	Capability_NET_ADMIN Capability = "NET_ADMIN"
	// CAP_NET_BIND_SERVICE.
	Capability_NET_BIND_SERVICE Capability = "NET_BIND_SERVICE"
	// CAP_NET_BROADCAST.
	Capability_NET_BROADCAST Capability = "NET_BROADCAST"
	// CAP_NET_RAW.
	Capability_NET_RAW Capability = "NET_RAW"
	// CAP_PERFMON.
	Capability_PERFMON Capability = "PERFMON"
	// CAP_SETGID.
	Capability_SETGID Capability = "SETGID"
	// CAP_SETFCAP.
	Capability_SETFCAP Capability = "SETFCAP"
	// CAP_SETPCAP.
	Capability_SETPCAP Capability = "SETPCAP"
	// CAP_SETUID.
	Capability_SETUID Capability = "SETUID"
	// CAP_SYS_ADMIN.
	Capability_SYS_ADMIN Capability = "SYS_ADMIN"
	// CAP_SYS_BOOT.
	Capability_SYS_BOOT Capability = "SYS_BOOT"
	// CAP_SYS_CHROOT.
	Capability_SYS_CHROOT Capability = "SYS_CHROOT"
	// CAP_SYS_MODULE.
	Capability_SYS_MODULE Capability = "SYS_MODULE"
	// CAP_SYS_NICE.
	Capability_SYS_NICE Capability = "SYS_NICE"
	// CAP_SYS_PACCT.
	Capability_SYS_PACCT Capability = "SYS_PACCT"
	// CAP_SYS_PTRACE.
	Capability_SYS_PTRACE Capability = "SYS_PTRACE"
	// CAP_SYS_RAWIO.
	Capability_SYS_RAWIO Capability = "SYS_RAWIO"
	// CAP_SYS_RESOURCE.
	Capability_SYS_RESOURCE Capability = "SYS_RESOURCE"
	// CAP_SYS_TIME.
	Capability_SYS_TIME Capability = "SYS_TIME"
	// CAP_SYS_TTY_CONFIG.
	Capability_SYS_TTY_CONFIG Capability = "SYS_TTY_CONFIG"
	// CAP_SYSLOG.
	Capability_SYSLOG Capability = "SYSLOG"
	// CAP_WAKE_ALARM.
	Capability_WAKE_ALARM Capability = "WAKE_ALARM"
)

type SeccompProfileType string

const (
	// A profile defined in a file on the node should be used.
	SeccompProfileType_LOCALHOST SeccompProfileType = "LOCALHOST"
	// The container runtime default profile should be used.
	SeccompProfileType_RUNTIME_DEFAULT SeccompProfileType = "RUNTIME_DEFAULT"
	// No profile should be applied.
	SeccompProfileType_UNCONFINED SeccompProfileType = "UNCONFINED"
)

type SeccompProfile struct {
	// Indicates which kind of seccomp profile will be applied.
	Type SeccompProfileType `field:"required" json:"type" yaml:"type"`
	// localhostProfile indicates a profile defined in a file on the node should be used.
	//
	// The profile must be preconfigured on the node to work. Must be a descending path, relative to the kubelet's configured seccomp profile location. Must only be set if type is "Localhost". Default: - empty string.
	LocalhostProfile *string `field:"optional" json:"localhostProfile" yaml:"localhostProfile"`
}

type ContainerSecutiryContextCapabilities struct {
	// Added capabilities.
	Add *[]Capability `field:"optional" json:"add" yaml:"add"`
	// Removed capabilities.
	Drop *[]Capability `field:"optional" json:"drop" yaml:"drop"`
}

// Properties for `ContainerSecurityContext`.
type ContainerSecurityContextProps struct {
	// Whether a process can gain more privileges than its parent process. Default: false.
	AllowPrivilegeEscalation *bool `field:"optional" json:"allowPrivilegeEscalation" yaml:"allowPrivilegeEscalation"`
	// POSIX capabilities for running containers. Default: none.
	Capabilities *ContainerSecutiryContextCapabilities `field:"optional" json:"capabilities" yaml:"capabilities"`
	// Indicates that the container must run as a non-root user.
	//
	// If true, the Kubelet will validate the image at runtime to ensure that it does not run as UID 0 (root) and fail to start the container if it does. Default: true.
	EnsureNonRoot *bool `field:"optional" json:"ensureNonRoot" yaml:"ensureNonRoot"`
	// The GID to run the entrypoint of the container process. Default: - 26000. An arbitrary number bigger than 9999 is selected here. This is so that the container is blocked to access host files even if somehow it manages to get access to host file system.
	Group *float64 `field:"optional" json:"group" yaml:"group"`
	// Run container in privileged mode.
	//
	// Processes in privileged containers are essentially equivalent to root on the host. Default: false.
	Privileged *bool `field:"optional" json:"privileged" yaml:"privileged"`
	// Whether this container has a read-only root filesystem. Default: true.
	ReadOnlyRootFilesystem *bool `field:"optional" json:"readOnlyRootFilesystem" yaml:"readOnlyRootFilesystem"`
	// Container's seccomp profile settings.
	//
	// Only one profile source may be set. Default: none.
	SeccompProfile *SeccompProfile `field:"optional" json:"seccompProfile" yaml:"seccompProfile"`
	// The UID to run the entrypoint of the container process. Default: - 25000. An arbitrary number bigger than 9999 is selected here. This is so that the container is blocked to access host files even if somehow it manages to get access to host file system.
	User *float64 `field:"optional" json:"user" yaml:"user"`
}

// Container security attributes and settings.
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

// RestartPolicy defines the restart behavior of individual containers in a pod.
//
// This field may only be set for init containers, and the only allowed value is "Always". For non-init containers or when this field is not specified, the restart behavior is defined by the Pod's restart policy and the container type. Setting the RestartPolicy as "Always" for the init container will have the following effect: this init container will be continually restarted on exit until all regular containers have terminated. Once all regular containers have completed, all init containers with restartPolicy "Always" will be shut down. This lifecycle differs from normal init containers and is often referred to as a "sidecar" container. See: https://kubernetes.io/docs/concepts/workloads/pods/sidecar-containers/
type ContainerRestartPolicy string

// If an init container is created with its restartPolicy set to Always, it will start and remain running during the entire life of the Pod.
//
// For regular containers, this is ignored by Kubernetes.
const ContainerRestartPolicy_ALWAYS ContainerRestartPolicy = "ALWAYS"

func containerRestartPolicyManifest(value ContainerRestartPolicy) string {
	if value != ContainerRestartPolicy_ALWAYS {
		panic("invalid container restart policy")
	}
	return "Always"
}

type MountPropagation string

const (
	// This volume mount will not receive any subsequent mounts that are mounted to this volume or any of its subdirectories by the host.
	//
	// In similar fashion, no mounts created by the Container will be visible on the host.
	//
	// This is the default mode.
	//
	// This mode is equal to `private` mount propagation as described in the Linux kernel documentation.
	MountPropagation_NONE MountPropagation = "NONE"
	// This volume mount will receive all subsequent mounts that are mounted to this volume or any of its subdirectories.
	//
	// In other words, if the host mounts anything inside the volume mount, the Container will see it mounted there.
	//
	// Similarly, if any Pod with Bidirectional mount propagation to the same volume mounts anything there, the Container with HostToContainer mount propagation will see it.
	//
	// This mode is equal to `rslave` mount propagation as described in the Linux kernel documentation.
	MountPropagation_HOST_TO_CONTAINER MountPropagation = "HOST_TO_CONTAINER"
	// This volume mount behaves the same the HostToContainer mount.
	//
	// In addition, all volume mounts created by the Container will be propagated back
	// to the host and to all Containers of all Pods that use the same volume
	//
	// A typical use case for this mode is a Pod with a FlexVolume or CSI driver or a
	// Pod that needs to mount something on the host using a hostPath volume.
	//
	// This mode is equal to `rshared` mount propagation as described in the Linux
	// kernel documentation
	//
	// Caution: Bidirectional mount propagation can be dangerous. It can damage the
	// host operating system and therefore it is allowed only in privileged Containers.
	// Familiarity with Linux kernel behavior is strongly recommended. In addition,
	// any volume mounts created by Containers in Pods must be destroyed (unmounted) by
	// the Containers on termination.
	MountPropagation_BIDIRECTIONAL MountPropagation = "BIDIRECTIONAL"
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
