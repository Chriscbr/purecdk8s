package cdk8splus34

import (
	"os"

	"github.com/Chriscbr/purecdk8s/jsii"
)

type EnvFieldPaths string

const (
	// The name of the pod.
	EnvFieldPaths_POD_NAME EnvFieldPaths = "POD_NAME"
	// The namespace of the pod.
	EnvFieldPaths_POD_NAMESPACE EnvFieldPaths = "POD_NAMESPACE"
	// The uid of the pod.
	EnvFieldPaths_POD_UID EnvFieldPaths = "POD_UID"
	// The labels of the pod.
	EnvFieldPaths_POD_LABEL EnvFieldPaths = "POD_LABEL"
	// The annotations of the pod.
	EnvFieldPaths_POD_ANNOTATION EnvFieldPaths = "POD_ANNOTATION"
	// The ipAddress of the pod.
	EnvFieldPaths_POD_IP EnvFieldPaths = "POD_IP"
	// The service account name of the pod.
	EnvFieldPaths_SERVICE_ACCOUNT_NAME EnvFieldPaths = "SERVICE_ACCOUNT_NAME"
	// The name of the node.
	EnvFieldPaths_NODE_NAME EnvFieldPaths = "NODE_NAME"
	// The ipAddress of the node.
	EnvFieldPaths_NODE_IP EnvFieldPaths = "NODE_IP"
	// The ipAddresess of the pod.
	EnvFieldPaths_POD_IPS EnvFieldPaths = "POD_IPS"
)

var envFieldPathValues = map[EnvFieldPaths]string{
	EnvFieldPaths_POD_NAME:             "metadata.name",
	EnvFieldPaths_POD_NAMESPACE:        "metadata.namespace",
	EnvFieldPaths_POD_UID:              "metadata.uid",
	EnvFieldPaths_POD_LABEL:            "metadata.labels",
	EnvFieldPaths_POD_ANNOTATION:       "metadata.annotations",
	EnvFieldPaths_POD_IP:               "status.podIP",
	EnvFieldPaths_SERVICE_ACCOUNT_NAME: "spec.serviceAccountName",
	EnvFieldPaths_NODE_NAME:            "spec.nodeName",
	EnvFieldPaths_NODE_IP:              "status.hostIP",
	EnvFieldPaths_POD_IPS:              "status.podIPs",
}

type ResourceFieldPaths string

const (
	// CPU limit of the container.
	ResourceFieldPaths_CPU_LIMIT ResourceFieldPaths = "CPU_LIMIT"
	// Memory limit of the container.
	ResourceFieldPaths_MEMORY_LIMIT ResourceFieldPaths = "MEMORY_LIMIT"
	// CPU request of the container.
	ResourceFieldPaths_CPU_REQUEST ResourceFieldPaths = "CPU_REQUEST"
	// Memory request of the container.
	ResourceFieldPaths_MEMORY_REQUEST ResourceFieldPaths = "MEMORY_REQUEST"
	// Ephemeral storage limit of the container.
	ResourceFieldPaths_STORAGE_LIMIT ResourceFieldPaths = "STORAGE_LIMIT"
	// Ephemeral storage request of the container.
	ResourceFieldPaths_STORAGE_REQUEST ResourceFieldPaths = "STORAGE_REQUEST"
)

var resourceFieldPathValues = map[ResourceFieldPaths]string{
	ResourceFieldPaths_CPU_LIMIT:       "limits.cpu",
	ResourceFieldPaths_MEMORY_LIMIT:    "limits.memory",
	ResourceFieldPaths_CPU_REQUEST:     "requests.cpu",
	ResourceFieldPaths_MEMORY_REQUEST:  "requests.memory",
	ResourceFieldPaths_STORAGE_LIMIT:   "limits.ephemeral-storage",
	ResourceFieldPaths_STORAGE_REQUEST: "requests.ephemeral-storage",
}

// Options to specify an environment variable value from the process environment.
type EnvValueFromProcessOptions struct {
	// Specify whether the key must exist in the environment.
	//
	// If this is set to true, and the key does not exist, an error will thrown. Default: false.
	Required *bool `field:"optional" json:"required" yaml:"required"`
}

// Options to specify an environment variable value from a field reference.
type EnvValueFromFieldRefOptions struct {
	// Version of the schema the FieldPath is written in terms of.
	ApiVersion *string `field:"optional" json:"apiVersion" yaml:"apiVersion"`
	// The key to select the pod label or annotation.
	Key *string `field:"optional" json:"key" yaml:"key"`
}

// Options to specify an environment variable value from a resource.
type EnvValueFromResourceOptions struct {
	// The container to select the value from.
	Container Container `field:"optional" json:"container" yaml:"container"`
	// The output format of the exposed resource.
	Divisor *string `field:"optional" json:"divisor" yaml:"divisor"`
}

// A collection of env variables defined in other resources.
type EnvFrom interface{ toManifest() map[string]interface{} }

type envFromImpl struct {
	configMap IConfigMap
	prefix    *string
	secret    ISecret
}

func (e *envFromImpl) toManifest() map[string]interface{} {
	result := map[string]interface{}{}
	if e.configMap != nil {
		result["configMapRef"] = map[string]interface{}{"name": e.configMap.Name()}
	}
	if e.secret != nil {
		result["secretRef"] = map[string]interface{}{"name": e.secret.Name()}
	}
	if e.prefix != nil {
		result["prefix"] = e.prefix
	}
	return result
}

func NewEnvFrom(configMap IConfigMap, prefix *string, secret ISecret) EnvFrom {
	if configMap == nil && secret == nil {
		panic("configMap or secret is required")
	}
	return &envFromImpl{configMap: configMap, prefix: prefix, secret: secret}
}

func NewEnvFrom_Override(env EnvFrom, configMap IConfigMap, prefix *string, secret ISecret) {
	applyOverride(env, NewEnvFrom(configMap, prefix, secret), "EnvFrom")
}

// Selects a ConfigMap to populate the environment variables with.
//
// The contents of the target ConfigMap's Data field will represent the key-value pairs as environment variables.
func Env_FromConfigMap(configMap IConfigMap, prefix *string) EnvFrom {
	if configMap == nil {
		panic("configMap is required")
	}
	return NewEnvFrom(configMap, prefix, nil)
}

// Selects a Secret to populate the environment variables with.
//
// The contents of the target Secret's Data field will represent the key-value pairs as environment variables.
func Env_FromSecret(secret ISecret) EnvFrom {
	if secret == nil {
		panic("secret is required")
	}
	return NewEnvFrom(nil, nil, secret)
}

// Create a value from a field reference.
func EnvValue_FromFieldRef(fieldPath EnvFieldPaths, options *EnvValueFromFieldRefOptions) EnvValue {
	path, ok := envFieldPathValues[fieldPath]
	if !ok {
		panic("fieldPath is required")
	}
	needsKey := fieldPath == EnvFieldPaths_POD_LABEL || fieldPath == EnvFieldPaths_POD_ANNOTATION
	if needsKey && (options == nil || options.Key == nil) {
		panic(path + " requires a key")
	}
	if needsKey {
		path += "['" + *options.Key + "']"
	}
	fieldRef := map[string]interface{}{"fieldPath": path}
	if options != nil && options.ApiVersion != nil {
		fieldRef["apiVersion"] = options.ApiVersion
	}
	return &envValue{valueFrom: map[string]interface{}{"fieldRef": fieldRef}}
}

// Create a value from a resource.
func EnvValue_FromResource(resource ResourceFieldPaths, options *EnvValueFromResourceOptions) EnvValue {
	path, ok := resourceFieldPathValues[resource]
	if !ok {
		panic("resource is required")
	}
	resourceRef := map[string]interface{}{"resource": path}
	if options != nil {
		if options.Divisor != nil {
			resourceRef["divisor"] = options.Divisor
		}
		if options.Container != nil {
			resourceRef["containerName"] = options.Container.Name()
		}
	}
	return &envValue{valueFrom: map[string]interface{}{"resourceFieldRef": resourceRef}}
}

// Create a value from a key in the current process environment.
func EnvValue_FromProcess(key *string, options *EnvValueFromProcessOptions) EnvValue {
	if key == nil {
		panic("key is required")
	}
	if value, found := os.LookupEnv(*key); found {
		return EnvValue_FromValue(jsii.String(value))
	}
	if options != nil && options.Required != nil && *options.Required {
		panic("Missing " + *key + " env variable")
	}
	return &envValue{}
}
