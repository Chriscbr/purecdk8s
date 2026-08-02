package cdk8splus34

import (
	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

type ConnectionScheme string

const (
	// Use HTTP request for connecting to host.
	ConnectionScheme_HTTP ConnectionScheme = "HTTP"
	// Use HTTPS request for connecting to host.
	ConnectionScheme_HTTPS ConnectionScheme = "HTTPS"
)

type HttpHeader struct {
	// The HTTP Header name to be used.
	Name *string `field:"required" json:"name" yaml:"name"`
	// The HTTP header value to be set.
	Value *string `field:"required" json:"value" yaml:"value"`
}

// Probe options.
type ProbeOptions struct {
	// Minimum consecutive failures for the probe to be considered failed after having succeeded.
	//
	// Defaults to 3. Minimum value is 1. Default: 3.
	FailureThreshold *float64 `field:"optional" json:"failureThreshold" yaml:"failureThreshold"`
	// Number of seconds after the container has started before liveness probes are initiated. See: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	//
	// Default: - immediate.
	InitialDelaySeconds cdk8s.Duration `field:"optional" json:"initialDelaySeconds" yaml:"initialDelaySeconds"`
	// How often (in seconds) to perform the probe.
	//
	// Default to 10 seconds. Minimum value is 1. Default: Duration.seconds(10) Minimum value is 1.
	PeriodSeconds cdk8s.Duration `field:"optional" json:"periodSeconds" yaml:"periodSeconds"`
	// Minimum consecutive successes for the probe to be considered successful after having failed. Defaults to 1.
	//
	// Must be 1 for liveness and startup. Minimum value is 1. Default: 1 Must be 1 for liveness and startup. Minimum value is 1.
	SuccessThreshold *float64 `field:"optional" json:"successThreshold" yaml:"successThreshold"`
	// Number of seconds after which the probe times out.
	//
	// Defaults to 1 second. Minimum value is 1. See: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	//
	// Default: Duration.seconds(1)
	TimeoutSeconds cdk8s.Duration `field:"optional" json:"timeoutSeconds" yaml:"timeoutSeconds"`
}

// Options for `Probe.fromCommand()`.
type CommandProbeOptions = ProbeOptions

// Options for `Probe.fromHttpGet()`.
type HttpGetProbeOptions struct {
	// Minimum consecutive failures for the probe to be considered failed after having succeeded.
	//
	// Defaults to 3. Minimum value is 1. Default: 3.
	FailureThreshold *float64 `field:"optional" json:"failureThreshold" yaml:"failureThreshold"`
	// Number of seconds after the container has started before liveness probes are initiated. See: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	//
	// Default: - immediate.
	InitialDelaySeconds cdk8s.Duration `field:"optional" json:"initialDelaySeconds" yaml:"initialDelaySeconds"`
	// How often (in seconds) to perform the probe.
	//
	// Default to 10 seconds. Minimum value is 1. Default: Duration.seconds(10) Minimum value is 1.
	PeriodSeconds cdk8s.Duration `field:"optional" json:"periodSeconds" yaml:"periodSeconds"`
	// Minimum consecutive successes for the probe to be considered successful after having failed. Defaults to 1.
	//
	// Must be 1 for liveness and startup. Minimum value is 1. Default: 1 Must be 1 for liveness and startup. Minimum value is 1.
	SuccessThreshold *float64 `field:"optional" json:"successThreshold" yaml:"successThreshold"`
	// Number of seconds after which the probe times out.
	//
	// Defaults to 1 second. Minimum value is 1. See: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	//
	// Default: Duration.seconds(1)
	TimeoutSeconds cdk8s.Duration `field:"optional" json:"timeoutSeconds" yaml:"timeoutSeconds"`
	// The host name to connect to on the container. Default: - defaults to the pod IP.
	Host *string `field:"optional" json:"host" yaml:"host"`
	// Custom HTTP headers to set in the probe request.
	//
	// Note that HTTP allows repeated headers. Default: - no custom headers are set.
	HttpHeaders *[]*HttpHeader `field:"optional" json:"httpHeaders" yaml:"httpHeaders"`
	// The TCP port to use when sending the GET request. Default: - defaults to `container.port`.
	Port *float64 `field:"optional" json:"port" yaml:"port"`
	// Scheme to use for connecting to the host (HTTP or HTTPS). Default: ConnectionScheme.HTTP
	Scheme ConnectionScheme `field:"optional" json:"scheme" yaml:"scheme"`
}

// Options for `Probe.fromTcpSocket()`.
type TcpSocketProbeOptions struct {
	// Minimum consecutive failures for the probe to be considered failed after having succeeded.
	//
	// Defaults to 3. Minimum value is 1. Default: 3.
	FailureThreshold *float64 `field:"optional" json:"failureThreshold" yaml:"failureThreshold"`
	// Number of seconds after the container has started before liveness probes are initiated. See: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	//
	// Default: - immediate.
	InitialDelaySeconds cdk8s.Duration `field:"optional" json:"initialDelaySeconds" yaml:"initialDelaySeconds"`
	// How often (in seconds) to perform the probe.
	//
	// Default to 10 seconds. Minimum value is 1. Default: Duration.seconds(10) Minimum value is 1.
	PeriodSeconds cdk8s.Duration `field:"optional" json:"periodSeconds" yaml:"periodSeconds"`
	// Minimum consecutive successes for the probe to be considered successful after having failed. Defaults to 1.
	//
	// Must be 1 for liveness and startup. Minimum value is 1. Default: 1 Must be 1 for liveness and startup. Minimum value is 1.
	SuccessThreshold *float64 `field:"optional" json:"successThreshold" yaml:"successThreshold"`
	// Number of seconds after which the probe times out.
	//
	// Defaults to 1 second. Minimum value is 1. See: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	//
	// Default: Duration.seconds(1)
	TimeoutSeconds cdk8s.Duration `field:"optional" json:"timeoutSeconds" yaml:"timeoutSeconds"`
	// The host name to connect to on the container. Default: - defaults to the pod IP.
	Host *string `field:"optional" json:"host" yaml:"host"`
	// The TCP port to connect to on the container. Default: - defaults to `container.port`.
	Port *float64 `field:"optional" json:"port" yaml:"port"`
}

// Options for `Probe.fromGrpc()`.
type GrpcProbeOptions struct {
	// Minimum consecutive failures for the probe to be considered failed after having succeeded.
	//
	// Defaults to 3. Minimum value is 1. Default: 3.
	FailureThreshold *float64 `field:"optional" json:"failureThreshold" yaml:"failureThreshold"`
	// Number of seconds after the container has started before liveness probes are initiated. See: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	//
	// Default: - immediate.
	InitialDelaySeconds cdk8s.Duration `field:"optional" json:"initialDelaySeconds" yaml:"initialDelaySeconds"`
	// How often (in seconds) to perform the probe.
	//
	// Default to 10 seconds. Minimum value is 1. Default: Duration.seconds(10) Minimum value is 1.
	PeriodSeconds cdk8s.Duration `field:"optional" json:"periodSeconds" yaml:"periodSeconds"`
	// Minimum consecutive successes for the probe to be considered successful after having failed. Defaults to 1.
	//
	// Must be 1 for liveness and startup. Minimum value is 1. Default: 1 Must be 1 for liveness and startup. Minimum value is 1.
	SuccessThreshold *float64 `field:"optional" json:"successThreshold" yaml:"successThreshold"`
	// Number of seconds after which the probe times out.
	//
	// Defaults to 1 second. Minimum value is 1. See: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	//
	// Default: Duration.seconds(1)
	TimeoutSeconds cdk8s.Duration `field:"optional" json:"timeoutSeconds" yaml:"timeoutSeconds"`
	// The TCP port to connect to on the container. Default: - defaults to `container.port`.
	Port *float64 `field:"optional" json:"port" yaml:"port"`
	// Service is the name of the service to place in the gRPC HealthCheckRequest (see https://github.com/grpc/grpc/blob/master/doc/health-checking.md). Default: - If this is not specified, the default behavior is defined by gRPC.
	Service *string `field:"optional" json:"service" yaml:"service"`
}

// Probe describes a health check to be performed against a container to determine whether it is alive or ready to receive traffic.
type Probe interface {
	toManifest(container *containerImpl) map[string]interface{}
}

type commandProbe struct {
	command *[]*string
	options *CommandProbeOptions
}

// Defines a probe based on a command which is executed within the container.
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

// Defines a probe based on an HTTP GET request to the IP address of the container.
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

// Defines a probe based opening a connection to a TCP socket on the container.
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

// Defines a probe based on a gRPC request to the container.
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
