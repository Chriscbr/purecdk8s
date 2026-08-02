package cdk8splus34

import (
	"os"

	"github.com/purecdk8s/purecdk8s/jsii"
)

// EnvFieldPaths identifies a Pod field available to an environment variable.
type EnvFieldPaths string

const (
	EnvFieldPaths_POD_NAME             EnvFieldPaths = "POD_NAME"
	EnvFieldPaths_POD_NAMESPACE        EnvFieldPaths = "POD_NAMESPACE"
	EnvFieldPaths_POD_UID              EnvFieldPaths = "POD_UID"
	EnvFieldPaths_POD_LABEL            EnvFieldPaths = "POD_LABEL"
	EnvFieldPaths_POD_ANNOTATION       EnvFieldPaths = "POD_ANNOTATION"
	EnvFieldPaths_POD_IP               EnvFieldPaths = "POD_IP"
	EnvFieldPaths_SERVICE_ACCOUNT_NAME EnvFieldPaths = "SERVICE_ACCOUNT_NAME"
	EnvFieldPaths_NODE_NAME            EnvFieldPaths = "NODE_NAME"
	EnvFieldPaths_NODE_IP              EnvFieldPaths = "NODE_IP"
	EnvFieldPaths_POD_IPS              EnvFieldPaths = "POD_IPS"
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

// ResourceFieldPaths identifies a container resource available to an environment variable.
type ResourceFieldPaths string

const (
	ResourceFieldPaths_CPU_LIMIT       ResourceFieldPaths = "CPU_LIMIT"
	ResourceFieldPaths_MEMORY_LIMIT    ResourceFieldPaths = "MEMORY_LIMIT"
	ResourceFieldPaths_CPU_REQUEST     ResourceFieldPaths = "CPU_REQUEST"
	ResourceFieldPaths_MEMORY_REQUEST  ResourceFieldPaths = "MEMORY_REQUEST"
	ResourceFieldPaths_STORAGE_LIMIT   ResourceFieldPaths = "STORAGE_LIMIT"
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

type EnvValueFromProcessOptions struct {
	Required *bool `field:"optional" json:"required" yaml:"required"`
}

type EnvValueFromFieldRefOptions struct {
	ApiVersion *string `field:"optional" json:"apiVersion" yaml:"apiVersion"`
	Key        *string `field:"optional" json:"key" yaml:"key"`
}

type EnvValueFromResourceOptions struct {
	Container Container `field:"optional" json:"container" yaml:"container"`
	Divisor   *string   `field:"optional" json:"divisor" yaml:"divisor"`
}

// EnvFrom is an environment-variable source copied from a ConfigMap or Secret.
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

func Env_FromConfigMap(configMap IConfigMap, prefix *string) EnvFrom {
	if configMap == nil {
		panic("configMap is required")
	}
	return NewEnvFrom(configMap, prefix, nil)
}

func Env_FromSecret(secret ISecret) EnvFrom {
	if secret == nil {
		panic("secret is required")
	}
	return NewEnvFrom(nil, nil, secret)
}

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
