package cdk8splus34

import "github.com/purecdk8s/purecdk8s/cdk8s/v2"

// ProbeOptions configures Kubernetes health probes.
type ProbeOptions struct {
	FailureThreshold    *float64       `field:"optional" json:"failureThreshold" yaml:"failureThreshold"`
	InitialDelaySeconds cdk8s.Duration `field:"optional" json:"initialDelaySeconds" yaml:"initialDelaySeconds"`
	PeriodSeconds       cdk8s.Duration `field:"optional" json:"periodSeconds" yaml:"periodSeconds"`
	SuccessThreshold    *float64       `field:"optional" json:"successThreshold" yaml:"successThreshold"`
	TimeoutSeconds      cdk8s.Duration `field:"optional" json:"timeoutSeconds" yaml:"timeoutSeconds"`
}

// CommandProbeOptions configures a command probe.
type CommandProbeOptions = ProbeOptions

// Probe describes a health check for a container.
type Probe interface {
	toManifest(container *containerImpl) map[string]interface{}
}

type commandProbe struct {
	command *[]*string
	options *CommandProbeOptions
}

// Probe_FromCommand creates a probe that executes command in its container.
func Probe_FromCommand(command *[]*string, options *CommandProbeOptions) Probe {
	if command == nil {
		panic("command is required")
	}
	return &commandProbe{command: command, options: options}
}

func (p *commandProbe) toManifest(_ *containerImpl) map[string]interface{} {
	result := map[string]interface{}{
		"failureThreshold": 3,
		"exec":             map[string]interface{}{"command": p.command},
	}
	if p.options == nil {
		return result
	}
	if p.options.FailureThreshold != nil {
		result["failureThreshold"] = p.options.FailureThreshold
	}
	if p.options.InitialDelaySeconds != nil {
		result["initialDelaySeconds"] = p.options.InitialDelaySeconds.ToSeconds(nil)
	}
	if p.options.PeriodSeconds != nil {
		result["periodSeconds"] = p.options.PeriodSeconds.ToSeconds(nil)
	}
	if p.options.SuccessThreshold != nil {
		result["successThreshold"] = p.options.SuccessThreshold
	}
	if p.options.TimeoutSeconds != nil {
		result["timeoutSeconds"] = p.options.TimeoutSeconds.ToSeconds(nil)
	}
	return result
}
