package cdk8splus34

import (
	"github.com/purecdk8s/purecdk8s/cdk8s/v2"
	"github.com/purecdk8s/purecdk8s/jsii"
)

// ConnectionScheme controls the protocol used by an HTTP probe.
type ConnectionScheme string

const (
	ConnectionScheme_HTTP  ConnectionScheme = "HTTP"
	ConnectionScheme_HTTPS ConnectionScheme = "HTTPS"
)

// HttpHeader is one HTTP header used by an HTTP probe.
type HttpHeader struct {
	Name  *string `field:"required" json:"name" yaml:"name"`
	Value *string `field:"required" json:"value" yaml:"value"`
}

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

// HttpGetProbeOptions configures an HTTP health probe.
type HttpGetProbeOptions struct {
	FailureThreshold    *float64         `field:"optional" json:"failureThreshold" yaml:"failureThreshold"`
	InitialDelaySeconds cdk8s.Duration   `field:"optional" json:"initialDelaySeconds" yaml:"initialDelaySeconds"`
	PeriodSeconds       cdk8s.Duration   `field:"optional" json:"periodSeconds" yaml:"periodSeconds"`
	SuccessThreshold    *float64         `field:"optional" json:"successThreshold" yaml:"successThreshold"`
	TimeoutSeconds      cdk8s.Duration   `field:"optional" json:"timeoutSeconds" yaml:"timeoutSeconds"`
	Host                *string          `field:"optional" json:"host" yaml:"host"`
	HttpHeaders         *[]*HttpHeader   `field:"optional" json:"httpHeaders" yaml:"httpHeaders"`
	Port                *float64         `field:"optional" json:"port" yaml:"port"`
	Scheme              ConnectionScheme `field:"optional" json:"scheme" yaml:"scheme"`
}

// TcpSocketProbeOptions configures a TCP health probe.
type TcpSocketProbeOptions struct {
	FailureThreshold    *float64       `field:"optional" json:"failureThreshold" yaml:"failureThreshold"`
	InitialDelaySeconds cdk8s.Duration `field:"optional" json:"initialDelaySeconds" yaml:"initialDelaySeconds"`
	PeriodSeconds       cdk8s.Duration `field:"optional" json:"periodSeconds" yaml:"periodSeconds"`
	SuccessThreshold    *float64       `field:"optional" json:"successThreshold" yaml:"successThreshold"`
	TimeoutSeconds      cdk8s.Duration `field:"optional" json:"timeoutSeconds" yaml:"timeoutSeconds"`
	Host                *string        `field:"optional" json:"host" yaml:"host"`
	Port                *float64       `field:"optional" json:"port" yaml:"port"`
}

// GrpcProbeOptions configures a gRPC health probe.
type GrpcProbeOptions struct {
	FailureThreshold    *float64       `field:"optional" json:"failureThreshold" yaml:"failureThreshold"`
	InitialDelaySeconds cdk8s.Duration `field:"optional" json:"initialDelaySeconds" yaml:"initialDelaySeconds"`
	PeriodSeconds       cdk8s.Duration `field:"optional" json:"periodSeconds" yaml:"periodSeconds"`
	SuccessThreshold    *float64       `field:"optional" json:"successThreshold" yaml:"successThreshold"`
	TimeoutSeconds      cdk8s.Duration `field:"optional" json:"timeoutSeconds" yaml:"timeoutSeconds"`
	Port                *float64       `field:"optional" json:"port" yaml:"port"`
	Service             *string        `field:"optional" json:"service" yaml:"service"`
}

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
	result := probeOptionsManifest(p.options)
	result["exec"] = map[string]interface{}{"command": p.command}
	return result
}

// Probe_FromHttpGet creates an HTTP GET health probe.
func Probe_FromHttpGet(path *string, options *HttpGetProbeOptions) Probe {
	if path == nil {
		panic("path is required")
	}
	return &httpGetProbe{path: path, options: options}
}

type httpGetProbe struct {
	path    *string
	options *HttpGetProbeOptions
}

func (p *httpGetProbe) toManifest(container *containerImpl) map[string]interface{} {
	result := probeOptionsManifest(httpProbeOptions(p.options))
	options := p.options
	port := (*float64)(nil)
	scheme := ConnectionScheme_HTTP
	if options != nil {
		port = options.Port
		if options.Scheme != "" {
			scheme = options.Scheme
		}
	}
	httpGet := map[string]interface{}{
		"path":   p.path,
		"port":   probePort(container, port),
		"scheme": string(scheme),
	}
	if options != nil {
		if options.Host != nil {
			httpGet["host"] = options.Host
		}
		if options.HttpHeaders != nil {
			headers := make([]interface{}, 0, len(*options.HttpHeaders))
			for _, header := range *options.HttpHeaders {
				if header == nil || header.Name == nil || header.Value == nil {
					panic("HTTP header name and value are required")
				}
				headers = append(headers, map[string]interface{}{"name": header.Name, "value": header.Value})
			}
			httpGet["httpHeaders"] = headers
		}
	}
	result["httpGet"] = httpGet
	return result
}

// Probe_FromTcpSocket creates a TCP health probe.
func Probe_FromTcpSocket(options *TcpSocketProbeOptions) Probe {
	return &tcpSocketProbe{options: options}
}

type tcpSocketProbe struct{ options *TcpSocketProbeOptions }

func (p *tcpSocketProbe) toManifest(container *containerImpl) map[string]interface{} {
	result := probeOptionsManifest(tcpProbeOptions(p.options))
	port := (*float64)(nil)
	tcpSocket := map[string]interface{}{}
	if p.options != nil {
		port = p.options.Port
		if p.options.Host != nil {
			tcpSocket["host"] = p.options.Host
		}
	}
	tcpSocket["port"] = probePort(container, port)
	result["tcpSocket"] = tcpSocket
	return result
}

// Probe_FromGrpc creates a gRPC health probe.
func Probe_FromGrpc(options *GrpcProbeOptions) Probe {
	return &grpcProbe{options: options}
}

type grpcProbe struct{ options *GrpcProbeOptions }

func (p *grpcProbe) toManifest(container *containerImpl) map[string]interface{} {
	result := probeOptionsManifest(grpcOptions(p.options))
	port := (*float64)(nil)
	grpc := map[string]interface{}{}
	if p.options != nil {
		port = p.options.Port
		if p.options.Service != nil {
			grpc["service"] = p.options.Service
		}
	}
	grpc["port"] = probePort(container, port)
	result["grpc"] = grpc
	return result
}

func probePort(container *containerImpl, requested *float64) *float64 {
	if requested != nil {
		return requested
	}
	if container.PortNumber() != nil {
		return container.PortNumber()
	}
	return jsii.Number(80)
}

func probeOptionsManifest(options *ProbeOptions) map[string]interface{} {
	result := map[string]interface{}{"failureThreshold": 3}
	if options == nil {
		return result
	}
	if options.FailureThreshold != nil {
		result["failureThreshold"] = options.FailureThreshold
	}
	if options.InitialDelaySeconds != nil {
		result["initialDelaySeconds"] = options.InitialDelaySeconds.ToSeconds(nil)
	}
	if options.PeriodSeconds != nil {
		result["periodSeconds"] = options.PeriodSeconds.ToSeconds(nil)
	}
	if options.SuccessThreshold != nil {
		result["successThreshold"] = options.SuccessThreshold
	}
	if options.TimeoutSeconds != nil {
		result["timeoutSeconds"] = options.TimeoutSeconds.ToSeconds(nil)
	}
	return result
}

func httpProbeOptions(options *HttpGetProbeOptions) *ProbeOptions {
	if options == nil {
		return nil
	}
	return &ProbeOptions{FailureThreshold: options.FailureThreshold, InitialDelaySeconds: options.InitialDelaySeconds, PeriodSeconds: options.PeriodSeconds, SuccessThreshold: options.SuccessThreshold, TimeoutSeconds: options.TimeoutSeconds}
}

func tcpProbeOptions(options *TcpSocketProbeOptions) *ProbeOptions {
	if options == nil {
		return nil
	}
	return &ProbeOptions{FailureThreshold: options.FailureThreshold, InitialDelaySeconds: options.InitialDelaySeconds, PeriodSeconds: options.PeriodSeconds, SuccessThreshold: options.SuccessThreshold, TimeoutSeconds: options.TimeoutSeconds}
}

func grpcOptions(options *GrpcProbeOptions) *ProbeOptions {
	if options == nil {
		return nil
	}
	return &ProbeOptions{FailureThreshold: options.FailureThreshold, InitialDelaySeconds: options.InitialDelaySeconds, PeriodSeconds: options.PeriodSeconds, SuccessThreshold: options.SuccessThreshold, TimeoutSeconds: options.TimeoutSeconds}
}
