package cdk8splus34

// Options for `Handler.fromTcpSocket`.
type HandlerFromTcpSocketOptions struct {
	// The host name to connect to on the container. Default: - defaults to the pod IP.
	Host *string `field:"optional" json:"host" yaml:"host"`
	// The TCP port to connect to on the container. Default: - defaults to `container.port`.
	Port *float64 `field:"optional" json:"port" yaml:"port"`
}

// Options for `Handler.fromHttpGet`.
type HandlerFromHttpGetOptions struct {
	// The TCP port to use when sending the GET request. Default: - defaults to `container.port`.
	Port *float64 `field:"optional" json:"port" yaml:"port"`
}

// Defines a specific action that should be taken.
type Handler interface {
	toManifest(container *containerImpl) map[string]interface{}
}

// Defines a handler based on a command which is executed within the container.
func Handler_FromCommand(command *[]*string) Handler {
	if command == nil {
		panic("command is required")
	}
	return &commandHandler{command: command}
}

type commandHandler struct{ command *[]*string }

func (h *commandHandler) toManifest(_ *containerImpl) map[string]interface{} {
	return map[string]interface{}{"exec": map[string]interface{}{"command": h.command}}
}

// Defines a handler based on an HTTP GET request to the IP address of the container.
func Handler_FromHttpGet(path *string, options *HandlerFromHttpGetOptions) Handler {
	if path == nil {
		panic("path is required")
	}
	return &httpGetHandler{path: path, options: options}
}

type httpGetHandler struct {
	path    *string
	options *HandlerFromHttpGetOptions
}

func (h *httpGetHandler) toManifest(container *containerImpl) map[string]interface{} {
	port := (*float64)(nil)
	if h.options != nil {
		port = h.options.Port
	}
	return map[string]interface{}{"httpGet": map[string]interface{}{
		"path":   h.path,
		"port":   probePort(container, port),
		"scheme": string(ConnectionScheme_HTTP),
	}}
}

// Defines a handler based opening a connection to a TCP socket on the container.
func Handler_FromTcpSocket(options *HandlerFromTcpSocketOptions) Handler {
	return &tcpSocketHandler{options: options}
}

type tcpSocketHandler struct{ options *HandlerFromTcpSocketOptions }

func (h *tcpSocketHandler) toManifest(container *containerImpl) map[string]interface{} {
	port := (*float64)(nil)
	tcpSocket := map[string]interface{}{}
	if h.options != nil {
		port = h.options.Port
		if h.options.Host != nil {
			tcpSocket["host"] = h.options.Host
		}
	}
	tcpSocket["port"] = probePort(container, port)
	return map[string]interface{}{"tcpSocket": tcpSocket}
}

// Container lifecycle properties.
type ContainerLifecycle struct {
	// This hook is executed immediately after a container is created.
	//
	// However, there is no guarantee that the hook will execute before the container ENTRYPOINT. Default: - No post start handler.
	PostStart Handler `field:"optional" json:"postStart" yaml:"postStart"`
	// This hook is called immediately before a container is terminated due to an API request or management event such as a liveness/startup probe failure, preemption, resource contention and others.
	//
	// A call to the PreStop hook fails if the container is already in a terminated or completed state and the hook must complete before the TERM signal to stop the container can be sent. The Pod's termination grace period countdown begins before the PreStop hook is executed, so regardless of the outcome of the handler, the container will eventually terminate within the Pod's termination grace period. No parameters are passed to the handler. See: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination
	//
	// Default: - No pre stop handler.
	PreStop Handler `field:"optional" json:"preStop" yaml:"preStop"`
}
