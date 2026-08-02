package cdk8splus34

// HandlerFromTcpSocketOptions configures a lifecycle TCP action.
type HandlerFromTcpSocketOptions struct {
	Host *string  `field:"optional" json:"host" yaml:"host"`
	Port *float64 `field:"optional" json:"port" yaml:"port"`
}

// HandlerFromHttpGetOptions configures a lifecycle HTTP action.
type HandlerFromHttpGetOptions struct {
	Port *float64 `field:"optional" json:"port" yaml:"port"`
}

// Handler describes an action performed by a container lifecycle hook.
type Handler interface {
	toManifest(container *containerImpl) map[string]interface{}
}

// Handler_FromCommand returns a handler that executes command in its container.
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

// Handler_FromHttpGet returns a handler that performs an HTTP GET request.
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
		"path": h.path,
		"port": probePort(container, port),
	}}
}

// Handler_FromTcpSocket returns a handler that opens a TCP socket.
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

// ContainerLifecycle configures handlers for container start and stop events.
type ContainerLifecycle struct {
	PostStart Handler `field:"optional" json:"postStart" yaml:"postStart"`
	PreStop   Handler `field:"optional" json:"preStop" yaml:"preStop"`
}
