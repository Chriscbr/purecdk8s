package cdk8splus34

import (
	"sort"

	"github.com/purecdk8s/purecdk8s/jsii"
)

// Protocol is the network protocol for container and service ports.
type Protocol string

const (
	Protocol_TCP  Protocol = "TCP"
	Protocol_UDP  Protocol = "UDP"
	Protocol_SCTP Protocol = "SCTP"
)

// ImagePullPolicy specifies how Kubernetes retrieves a container image.
type ImagePullPolicy string

const (
	ImagePullPolicy_ALWAYS         ImagePullPolicy = "Always"
	ImagePullPolicy_IF_NOT_PRESENT ImagePullPolicy = "IfNotPresent"
	ImagePullPolicy_NEVER          ImagePullPolicy = "Never"
)

// ContainerPort defines an exposed container port.
type ContainerPort struct {
	Number   *float64 `field:"required" json:"number" yaml:"number"`
	HostIp   *string  `field:"optional" json:"hostIp" yaml:"hostIp"`
	HostPort *float64 `field:"optional" json:"hostPort" yaml:"hostPort"`
	Name     *string  `field:"optional" json:"name" yaml:"name"`
	Protocol Protocol `field:"optional" json:"protocol" yaml:"protocol"`
}

// ContainerProps configures one workload container.
type ContainerProps struct {
	Image           *string              `field:"required" json:"image" yaml:"image"`
	Name            *string              `field:"optional" json:"name" yaml:"name"`
	Port            *float64             `field:"optional" json:"port" yaml:"port"`
	PortNumber      *float64             `field:"optional" json:"portNumber" yaml:"portNumber"`
	Ports           *[]*ContainerPort    `field:"optional" json:"ports" yaml:"ports"`
	Command         *[]*string           `field:"optional" json:"command" yaml:"command"`
	Args            *[]*string           `field:"optional" json:"args" yaml:"args"`
	WorkingDir      *string              `field:"optional" json:"workingDir" yaml:"workingDir"`
	EnvVariables    *map[string]EnvValue `field:"optional" json:"envVariables" yaml:"envVariables"`
	ImagePullPolicy ImagePullPolicy      `field:"optional" json:"imagePullPolicy" yaml:"imagePullPolicy"`
	Liveness        Probe                `field:"optional" json:"liveness" yaml:"liveness"`
	Readiness       Probe                `field:"optional" json:"readiness" yaml:"readiness"`
	Startup         Probe                `field:"optional" json:"startup" yaml:"startup"`
}

// ContainerOpts is the optional portion of ContainerProps.
type ContainerOpts struct {
	Name            *string
	Port            *float64
	PortNumber      *float64
	Ports           *[]*ContainerPort
	Command         *[]*string
	Args            *[]*string
	WorkingDir      *string
	EnvVariables    *map[string]EnvValue
	ImagePullPolicy ImagePullPolicy
	Liveness        Probe
	Readiness       Probe
	Startup         Probe
}

// Container represents a container attached to a pod workload.
type Container interface {
	Args() *[]*string
	Command() *[]*string
	Image() *string
	ImagePullPolicy() ImagePullPolicy
	Mounts() *[]*VolumeMount
	Name() *string
	Port() *float64
	PortNumber() *float64
	Ports() *[]*ContainerPort
	WorkingDir() *string
	Env() Env
	AddPort(port *ContainerPort)
	Mount(path *string, storage IStorage, options *MountOptions)
}

type containerImpl struct {
	image      *string
	name       *string
	portNumber *float64
	ports      []*ContainerPort
	command    *[]*string
	args       *[]*string
	workingDir *string
	env        map[string]EnvValue
	mounts     []*VolumeMount
	pullPolicy ImagePullPolicy
	liveness   Probe
	readiness  Probe
	startup    Probe
}

func NewContainer(props *ContainerProps) Container {
	if props == nil || props.Image == nil {
		panic("props.image is required")
	}
	return newContainer(props)
}

func NewContainer_Override(container Container, props *ContainerProps) {
	panic("native cdk8splus34 does not support overriding value objects")
}

func newContainer(props *ContainerProps) *containerImpl {
	name := props.Name
	if name == nil {
		name = jsii.String("main")
	}
	result := &containerImpl{
		image:      props.Image,
		name:       name,
		command:    props.Command,
		args:       props.Args,
		workingDir: props.WorkingDir,
		pullPolicy: props.ImagePullPolicy,
		liveness:   props.Liveness,
		readiness:  props.Readiness,
		startup:    props.Startup,
		env:        map[string]EnvValue{},
	}
	if result.pullPolicy == "" {
		result.pullPolicy = ImagePullPolicy_ALWAYS
	}
	if props.PortNumber != nil {
		result.portNumber = props.PortNumber
	} else {
		result.portNumber = props.Port
	}
	if result.portNumber != nil {
		result.ports = append(result.ports, &ContainerPort{Number: result.portNumber})
	}
	if props.Ports != nil {
		for _, port := range *props.Ports {
			if port != nil {
				result.AddPort(port)
			}
		}
	}
	if props.EnvVariables != nil {
		for key, value := range *props.EnvVariables {
			result.env[key] = value
		}
	}
	return result
}

func (c *containerImpl) Args() *[]*string                 { return c.args }
func (c *containerImpl) Command() *[]*string              { return c.command }
func (c *containerImpl) Image() *string                   { return c.image }
func (c *containerImpl) ImagePullPolicy() ImagePullPolicy { return c.pullPolicy }
func (c *containerImpl) Name() *string                    { return c.name }
func (c *containerImpl) Port() *float64                   { return c.portNumber }
func (c *containerImpl) PortNumber() *float64             { return c.portNumber }
func (c *containerImpl) WorkingDir() *string              { return c.workingDir }
func (c *containerImpl) Env() Env                         { return &envImpl{variables: c.env} }
func (c *containerImpl) Ports() *[]*ContainerPort {
	values := append([]*ContainerPort(nil), c.ports...)
	return &values
}
func (c *containerImpl) Mounts() *[]*VolumeMount {
	values := append([]*VolumeMount(nil), c.mounts...)
	return &values
}
func (c *containerImpl) AddPort(port *ContainerPort) {
	if port == nil || port.Number == nil {
		panic("container port number is required")
	}
	c.ports = append(c.ports, port)
}
func (c *containerImpl) Mount(path *string, storage IStorage, options *MountOptions) {
	if path == nil || storage == nil {
		panic("path and storage are required")
	}
	volume := storage.AsVolume()
	c.mounts = append(c.mounts, &VolumeMount{Path: path, Volume: volume, MountPath: path, ReadOnly: optionsReadOnly(options)})
}

func optionsReadOnly(options *MountOptions) *bool {
	if options == nil {
		return nil
	}
	return options.ReadOnly
}

func (c *containerImpl) toManifest() map[string]interface{} {
	ports := make([]interface{}, 0, len(c.ports))
	for _, port := range c.ports {
		entry := map[string]interface{}{"containerPort": port.Number}
		if port.Name != nil {
			entry["name"] = port.Name
		}
		if port.HostIp != nil {
			entry["hostIP"] = port.HostIp
		}
		if port.HostPort != nil {
			entry["hostPort"] = port.HostPort
		}
		if port.Protocol != "" {
			entry["protocol"] = string(port.Protocol)
		}
		ports = append(ports, entry)
	}
	mounts := make([]interface{}, 0, len(c.mounts))
	for _, mount := range c.mounts {
		entry := map[string]interface{}{"name": mount.Volume.Name(), "mountPath": mount.Path}
		if mount.ReadOnly != nil {
			entry["readOnly"] = mount.ReadOnly
		}
		mounts = append(mounts, entry)
	}
	env := make([]interface{}, 0, len(c.env))
	keys := make([]string, 0, len(c.env))
	for key := range c.env {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := c.env[key]
		entry := map[string]interface{}{"name": key}
		if value.Value() != nil {
			entry["value"] = value.Value()
		}
		if value.ValueFrom() != nil {
			entry["valueFrom"] = value.ValueFrom()
		}
		env = append(env, entry)
	}
	result := map[string]interface{}{
		"image":           c.image,
		"name":            c.name,
		"imagePullPolicy": string(c.pullPolicy),
		"resources": map[string]interface{}{
			"limits": map[string]interface{}{
				"cpu":    "1500m",
				"memory": "2048Mi",
			},
			"requests": map[string]interface{}{
				"cpu":    "1000m",
				"memory": "512Mi",
			},
		},
		"securityContext": map[string]interface{}{
			"allowPrivilegeEscalation": false,
			"privileged":               false,
			"readOnlyRootFilesystem":   true,
			"runAsNonRoot":             true,
		},
	}
	if c.command != nil {
		result["command"] = c.command
	}
	if c.args != nil {
		result["args"] = c.args
	}
	if c.workingDir != nil {
		result["workingDir"] = c.workingDir
	}
	if len(ports) > 0 {
		result["ports"] = ports
	}
	if len(mounts) > 0 {
		result["volumeMounts"] = mounts
	}
	if len(env) > 0 {
		result["env"] = env
	}
	if c.portNumber != nil {
		result["startupProbe"] = map[string]interface{}{
			"failureThreshold": 3,
			"tcpSocket":        map[string]interface{}{"port": c.portNumber},
		}
	}
	if c.liveness != nil {
		result["livenessProbe"] = c.liveness.toManifest(c)
	}
	if c.readiness != nil {
		result["readinessProbe"] = c.readiness.toManifest(c)
	}
	if c.startup != nil {
		result["startupProbe"] = c.startup.toManifest(c)
	}
	return result
}

// EnvValue represents either a literal environment variable value or a
// Kubernetes EnvVarSource.
type EnvValue interface {
	Value() interface{}
	ValueFrom() interface{}
}

type envValue struct{ value, valueFrom interface{} }

func (v *envValue) Value() interface{}     { return v.value }
func (v *envValue) ValueFrom() interface{} { return v.valueFrom }

func EnvValue_FromValue(value *string) EnvValue {
	if value == nil {
		panic("value is required")
	}
	return &envValue{value: value}
}

// Env is a mutable container environment.
type Env interface {
	Variables() *map[string]EnvValue
	AddVariable(name *string, value EnvValue)
}

type envImpl struct{ variables map[string]EnvValue }

func NewEnv(_ interface{}, variables *map[string]EnvValue) Env {
	values := map[string]EnvValue{}
	if variables != nil {
		for name, value := range *variables {
			values[name] = value
		}
	}
	return &envImpl{variables: values}
}
func (e *envImpl) Variables() *map[string]EnvValue {
	values := make(map[string]EnvValue, len(e.variables))
	for name, value := range e.variables {
		values[name] = value
	}
	return &values
}
func (e *envImpl) AddVariable(name *string, value EnvValue) {
	if name == nil || *name == "" || value == nil {
		panic("environment variable name and value are required")
	}
	e.variables[*name] = value
}

type EnvValueFromConfigMapOptions struct {
	Optional *bool `field:"optional" json:"optional" yaml:"optional"`
}

func EnvValue_FromConfigMap(configMap IConfigMap, key *string, options *EnvValueFromConfigMapOptions) EnvValue {
	if configMap == nil || key == nil {
		panic("configMap and key are required")
	}
	ref := map[string]interface{}{"name": configMap.Name(), "key": key}
	if options != nil && options.Optional != nil {
		ref["optional"] = options.Optional
	}
	return &envValue{valueFrom: map[string]interface{}{"configMapKeyRef": ref}}
}
