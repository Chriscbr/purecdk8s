package cdk8splus34

import (
	"encoding/json"

	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// Options for `Secret`.
type SecretProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// If set to true, ensures that data stored in the Secret cannot be updated (only object metadata can be modified).
	//
	// If not set to true, the field can be modified at any time. Default: false.
	Immutable *bool `field:"optional" json:"immutable" yaml:"immutable"`
	// stringData allows specifying non-binary secret data in string form.
	//
	// It is provided as a write-only convenience method. All keys and values are merged into the data field on write, overwriting any existing values. It is never output when reading from the API.
	StringData *map[string]*string `field:"optional" json:"stringData" yaml:"stringData"`
	// Optional type associated with the secret.
	//
	// Used to facilitate programmatic handling of secret data by various controllers. Default: undefined - Don't set a type.
	Type *string `field:"optional" json:"type" yaml:"type"`
}

// Common properties for `Secret`.
type CommonSecretProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// If set to true, ensures that data stored in the Secret cannot be updated (only object metadata can be modified).
	//
	// If not set to true, the field can be modified at any time. Default: false.
	Immutable *bool `field:"optional" json:"immutable" yaml:"immutable"`
}

// Options for `BasicAuthSecret`.
type BasicAuthSecretProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// If set to true, ensures that data stored in the Secret cannot be updated (only object metadata can be modified).
	//
	// If not set to true, the field can be modified at any time. Default: false.
	Immutable *bool `field:"optional" json:"immutable" yaml:"immutable"`
	// The password or token for authentication.
	Password *string `field:"required" json:"password" yaml:"password"`
	// The user name for authentication.
	Username *string `field:"required" json:"username" yaml:"username"`
}

// Options for `SshAuthSecret`.
type SshAuthSecretProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// If set to true, ensures that data stored in the Secret cannot be updated (only object metadata can be modified).
	//
	// If not set to true, the field can be modified at any time. Default: false.
	Immutable *bool `field:"optional" json:"immutable" yaml:"immutable"`
	// The SSH private key to use.
	SshPrivateKey *string `field:"required" json:"sshPrivateKey" yaml:"sshPrivateKey"`
}

// Options for `TlsSecret`.
type TlsSecretProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// If set to true, ensures that data stored in the Secret cannot be updated (only object metadata can be modified).
	//
	// If not set to true, the field can be modified at any time. Default: false.
	Immutable *bool `field:"optional" json:"immutable" yaml:"immutable"`
	// The TLS cert.
	TlsCert *string `field:"required" json:"tlsCert" yaml:"tlsCert"`
	// The TLS key.
	TlsKey *string `field:"required" json:"tlsKey" yaml:"tlsKey"`
}

// Options for `DockerConfigSecret`.
type DockerConfigSecretProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// If set to true, ensures that data stored in the Secret cannot be updated (only object metadata can be modified).
	//
	// If not set to true, the field can be modified at any time. Default: false.
	Immutable *bool `field:"optional" json:"immutable" yaml:"immutable"`
	// JSON content to provide for the `~/.docker/config.json` file. This will be stringified and inserted as stringData. See: https://docs.docker.com/engine/reference/commandline/cli/#sample-configuration-file
	Data *map[string]interface{} `field:"required" json:"data" yaml:"data"`
}

// Options for `ServiceAccountTokenSecret`.
type ServiceAccountTokenSecretProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// If set to true, ensures that data stored in the Secret cannot be updated (only object metadata can be modified).
	//
	// If not set to true, the field can be modified at any time. Default: false.
	Immutable *bool `field:"optional" json:"immutable" yaml:"immutable"`
	// The service account to store a secret for.
	ServiceAccount IServiceAccount `field:"required" json:"serviceAccount" yaml:"serviceAccount"`
}

// Options to specify an environment variable value from a Secret.
type EnvValueFromSecretOptions struct {
	// Specify whether the Secret or its key must be defined. Default: false.
	Optional *bool `field:"optional" json:"optional" yaml:"optional"`
}

// Represents a specific value in JSON secret.
type SecretValue struct {
	// The secret.
	Secret ISecret `field:"required" json:"secret" yaml:"secret"`
	// The JSON key.
	Key *string `field:"required" json:"key" yaml:"key"`
}

// Kubernetes Secrets let you store and manage sensitive information, such as passwords, OAuth tokens, and ssh keys.
//
// Storing confidential information in a Secret is safer and more flexible than putting it verbatim in a Pod definition or in a container image. See: https://kubernetes.io/docs/concepts/configuration/secret
type Secret interface {
	Resource
	ISecret
	// Whether or not the secret is immutable.
	Immutable() *bool
	// Adds a string data field to the secret.
	AddStringData(key, value *string)
	// Gets a string data by key or undefined.
	GetStringData(key *string) *string
	// Returns EnvValue object from a secret's key.
	EnvValue(key *string, options *EnvValueFromSecretOptions) EnvValue
}

// Create a secret for basic authentication. See: https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret
type BasicAuthSecret interface{ Secret }

// Create a secret for ssh authentication. See: https://kubernetes.io/docs/concepts/configuration/secret/#ssh-authentication-secrets
type SshAuthSecret interface{ Secret }

// Create a secret for storing a TLS certificate and its associated key. See: https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets
type TlsSecret interface{ Secret }

// Create a secret for storing credentials for accessing a container image registry. See: https://kubernetes.io/docs/concepts/configuration/secret/#docker-config-secrets
type DockerConfigSecret interface{ Secret }

// Create a secret for a service account token. See: https://kubernetes.io/docs/concepts/configuration/secret/#service-account-token-secrets
type ServiceAccountTokenSecret interface{ Secret }

type secretImpl struct {
	resourceBase
	immutable  *bool
	stringData map[string]*string
}

// importedSecret is a construct-only reference to a Secret that already
// exists in the cluster. It deliberately does not synthesize a manifest.
type importedSecret struct {
	node constructs.Node
	name *string
}

func (s *importedSecret) Node() constructs.Node {
	return s.node
}

func (s *importedSecret) SetNodeInternal(node constructs.Node) {
	s.node = node
}

func (s *importedSecret) ToString() *string {
	return s.node.Path()
}

func (s *importedSecret) With(mixins ...constructs.IMixin) constructs.IConstruct {
	return s.node.With(mixins...)
}

func (s *importedSecret) ApiVersion() *string {
	return jsii.String("v1")
}

func (s *importedSecret) ApiGroup() *string {
	return jsii.String("")
}

func (s *importedSecret) Kind() *string {
	return jsii.String("Secret")
}

func (s *importedSecret) Name() *string {
	return s.name
}

func (s *importedSecret) ResourceName() *string {
	return s.name
}

func (s *importedSecret) ResourceType() *string {
	return jsii.String("secrets")
}

func (s *importedSecret) AsApiResource() IApiResource {
	return s
}

func (s *importedSecret) AsNonApiResource() *string {
	return nil
}

func (s *importedSecret) EnvValue(key *string, options *EnvValueFromSecretOptions) EnvValue {
	return EnvValue_FromSecretValue(&SecretValue{Secret: s, Key: key}, options)
}

// Imports a secret from the cluster as a reference.
func Secret_FromSecretName(scope constructs.Construct, id, name *string) ISecret {
	if scope == nil || id == nil || name == nil {
		panic("scope, id and name are required")
	}
	result := &importedSecret{name: name}
	constructs.NewConstruct_Override(result, scope, id)
	return result
}

// Imports a secret from the cluster as a reference.
func BasicAuthSecret_FromSecretName(scope constructs.Construct, id, name *string) ISecret {
	return Secret_FromSecretName(scope, id, name)
}

// Imports a secret from the cluster as a reference.
func SshAuthSecret_FromSecretName(scope constructs.Construct, id, name *string) ISecret {
	return Secret_FromSecretName(scope, id, name)
}

// Imports a secret from the cluster as a reference.
func TlsSecret_FromSecretName(scope constructs.Construct, id, name *string) ISecret {
	return Secret_FromSecretName(scope, id, name)
}

// Imports a secret from the cluster as a reference.
func DockerConfigSecret_FromSecretName(scope constructs.Construct, id, name *string) ISecret {
	return Secret_FromSecretName(scope, id, name)
}

// Imports a secret from the cluster as a reference.
func ServiceAccountTokenSecret_FromSecretName(scope constructs.Construct, id, name *string) ISecret {
	return Secret_FromSecretName(scope, id, name)
}

func NewSecret(scope constructs.Construct, id *string, props *SecretProps) Secret {
	if props == nil {
		props = &SecretProps{}
	}
	immutable := props.Immutable
	if immutable == nil {
		immutable = jsii.Bool(false)
	}
	result := &secretImpl{immutable: immutable, stringData: map[string]*string{}}
	if props.StringData != nil {
		for key, value := range *props.StringData {
			result.stringData[key] = value
		}
	}
	manifest := map[string]interface{}{"immutable": immutable}
	if props.Type != nil {
		manifest["type"] = props.Type
	}
	result.resourceBase.initialize(result, scope, id, "v1", "Secret", "secrets", props.Metadata, manifest)
	manifest["stringData"] = cdk8s.Lazy_Any(lazyProducer{produce: func() interface{} {
		if len(result.stringData) == 0 {
			return map[string]interface{}{}
		}
		value := make(map[string]interface{}, len(result.stringData))
		for key, data := range result.stringData {
			value[key] = data
		}
		return value
	}})
	return result
}

func NewBasicAuthSecret(scope constructs.Construct, id *string, props *BasicAuthSecretProps) BasicAuthSecret {
	if props == nil || props.Username == nil || props.Password == nil {
		panic("username and password are required")
	}
	return NewSecret(scope, id, &SecretProps{
		Metadata: props.Metadata, Immutable: props.Immutable, Type: jsii.String("kubernetes.io/basic-auth"),
		StringData: &map[string]*string{"username": props.Username, "password": props.Password},
	})
}

func NewSshAuthSecret(scope constructs.Construct, id *string, props *SshAuthSecretProps) SshAuthSecret {
	if props == nil || props.SshPrivateKey == nil {
		panic("sshPrivateKey is required")
	}
	return NewSecret(scope, id, &SecretProps{
		Metadata: props.Metadata, Immutable: props.Immutable, Type: jsii.String("kubernetes.io/ssh-auth"),
		StringData: &map[string]*string{"ssh-privatekey": props.SshPrivateKey},
	})
}

func NewTlsSecret(scope constructs.Construct, id *string, props *TlsSecretProps) TlsSecret {
	if props == nil || props.TlsCert == nil || props.TlsKey == nil {
		panic("tlsCert and tlsKey are required")
	}
	return NewSecret(scope, id, &SecretProps{
		Metadata: props.Metadata, Immutable: props.Immutable, Type: jsii.String("kubernetes.io/tls"),
		StringData: &map[string]*string{"tls.crt": props.TlsCert, "tls.key": props.TlsKey},
	})
}

func NewDockerConfigSecret(scope constructs.Construct, id *string, props *DockerConfigSecretProps) DockerConfigSecret {
	if props == nil || props.Data == nil {
		panic("data is required")
	}
	data, err := json.Marshal(*props.Data)
	if err != nil {
		panic(err)
	}
	return NewSecret(scope, id, &SecretProps{
		Metadata: props.Metadata, Immutable: props.Immutable, Type: jsii.String("kubernetes.io/dockerconfigjson"),
		StringData: &map[string]*string{".dockerconfigjson": jsii.String(string(data))},
	})
}

func NewServiceAccountTokenSecret(scope constructs.Construct, id *string, props *ServiceAccountTokenSecretProps) ServiceAccountTokenSecret {
	if props == nil || props.ServiceAccount == nil {
		panic("serviceAccount is required")
	}
	secret := NewSecret(scope, id, &SecretProps{
		Metadata: props.Metadata, Immutable: props.Immutable, Type: jsii.String("kubernetes.io/service-account-token"),
	})
	secret.Metadata().AddAnnotation(jsii.String("kubernetes.io/service-account.name"), props.ServiceAccount.Name())
	return secret
}

func NewSecret_Override(secret Secret, scope constructs.Construct, id *string, props *SecretProps) {
	applyOverride(secret, NewSecret(scope, id, props), "Secret")
}

func NewBasicAuthSecret_Override(secret BasicAuthSecret, scope constructs.Construct, id *string, props *BasicAuthSecretProps) {
	applyOverride(secret, NewBasicAuthSecret(scope, id, props), "BasicAuthSecret")
}

func NewSshAuthSecret_Override(secret SshAuthSecret, scope constructs.Construct, id *string, props *SshAuthSecretProps) {
	applyOverride(secret, NewSshAuthSecret(scope, id, props), "SshAuthSecret")
}

func NewTlsSecret_Override(secret TlsSecret, scope constructs.Construct, id *string, props *TlsSecretProps) {
	applyOverride(secret, NewTlsSecret(scope, id, props), "TlsSecret")
}

func NewDockerConfigSecret_Override(secret DockerConfigSecret, scope constructs.Construct, id *string, props *DockerConfigSecretProps) {
	applyOverride(secret, NewDockerConfigSecret(scope, id, props), "DockerConfigSecret")
}

func NewServiceAccountTokenSecret_Override(secret ServiceAccountTokenSecret, scope constructs.Construct, id *string, props *ServiceAccountTokenSecretProps) {
	applyOverride(secret, NewServiceAccountTokenSecret(scope, id, props), "ServiceAccountTokenSecret")
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func Secret_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func BasicAuthSecret_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func SshAuthSecret_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func TlsSecret_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func DockerConfigSecret_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func ServiceAccountTokenSecret_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func (s *secretImpl) Immutable() *bool {
	return s.immutable
}

func (s *secretImpl) AddStringData(key, value *string) {
	if key == nil || *key == "" {
		panic("secret key is required")
	}
	s.stringData[*key] = value
}

func (s *secretImpl) GetStringData(key *string) *string {
	if key == nil {
		return nil
	}
	return s.stringData[*key]
}

func (s *secretImpl) EnvValue(key *string, options *EnvValueFromSecretOptions) EnvValue {
	if key == nil || *key == "" {
		panic("secret key is required")
	}
	secretKeyRef := map[string]interface{}{"name": s.Name(), "key": key}
	if options != nil && options.Optional != nil {
		secretKeyRef["optional"] = options.Optional
	}
	return &envValue{valueFrom: map[string]interface{}{"secretKeyRef": secretKeyRef}}
}

// Defines an environment value from a secret JSON value.
func EnvValue_FromSecretValue(secretValue *SecretValue, options *EnvValueFromSecretOptions) EnvValue {
	if secretValue == nil || secretValue.Secret == nil || secretValue.Key == nil || *secretValue.Key == "" {
		panic("secret and key are required")
	}
	ref := map[string]interface{}{"name": secretValue.Secret.Name(), "key": secretValue.Key}
	if options != nil && options.Optional != nil {
		ref["optional"] = options.Optional
	}
	return &envValue{valueFrom: map[string]interface{}{"secretKeyRef": ref}}
}
