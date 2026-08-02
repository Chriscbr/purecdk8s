package cdk8splus34

import (
	"encoding/json"

	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// SecretProps configures a Kubernetes Secret.
type SecretProps struct {
	Metadata   *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	Immutable  *bool                    `field:"optional" json:"immutable" yaml:"immutable"`
	StringData *map[string]*string      `field:"optional" json:"stringData" yaml:"stringData"`
	Type       *string                  `field:"optional" json:"type" yaml:"type"`
}

// CommonSecretProps contains fields shared by the specialized Secret types.
type CommonSecretProps struct {
	Metadata  *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	Immutable *bool                    `field:"optional" json:"immutable" yaml:"immutable"`
}

type BasicAuthSecretProps struct {
	Metadata  *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	Immutable *bool                    `field:"optional" json:"immutable" yaml:"immutable"`
	Password  *string                  `field:"required" json:"password" yaml:"password"`
	Username  *string                  `field:"required" json:"username" yaml:"username"`
}

type SshAuthSecretProps struct {
	Metadata      *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	Immutable     *bool                    `field:"optional" json:"immutable" yaml:"immutable"`
	SshPrivateKey *string                  `field:"required" json:"sshPrivateKey" yaml:"sshPrivateKey"`
}

type TlsSecretProps struct {
	Metadata  *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	Immutable *bool                    `field:"optional" json:"immutable" yaml:"immutable"`
	TlsCert   *string                  `field:"required" json:"tlsCert" yaml:"tlsCert"`
	TlsKey    *string                  `field:"required" json:"tlsKey" yaml:"tlsKey"`
}

type DockerConfigSecretProps struct {
	Metadata  *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	Immutable *bool                    `field:"optional" json:"immutable" yaml:"immutable"`
	Data      *map[string]interface{}  `field:"required" json:"data" yaml:"data"`
}

type ServiceAccountTokenSecretProps struct {
	Metadata       *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	Immutable      *bool                    `field:"optional" json:"immutable" yaml:"immutable"`
	ServiceAccount IServiceAccount          `field:"required" json:"serviceAccount" yaml:"serviceAccount"`
}

// EnvValueFromSecretOptions controls a secret-backed environment value.
type EnvValueFromSecretOptions struct {
	Optional *bool `field:"optional" json:"optional" yaml:"optional"`
}

// SecretValue identifies one key in a Secret.
type SecretValue struct {
	Secret ISecret `field:"required" json:"secret" yaml:"secret"`
	Key    *string `field:"required" json:"key" yaml:"key"`
}

// Secret is a native Kubernetes Secret construct.
type Secret interface {
	Resource
	ISecret
	Immutable() *bool
	AddStringData(key, value *string)
	GetStringData(key *string) *string
	EnvValue(key *string, options *EnvValueFromSecretOptions) EnvValue
}

// BasicAuthSecret stores username and password credentials.
type BasicAuthSecret interface{ Secret }

// SshAuthSecret stores an SSH private key.
type SshAuthSecret interface{ Secret }

// TlsSecret stores a TLS certificate and private key.
type TlsSecret interface{ Secret }

// DockerConfigSecret stores Docker registry configuration.
type DockerConfigSecret interface{ Secret }

// ServiceAccountTokenSecret requests a token for a ServiceAccount.
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

// Secret_FromSecretName imports an existing Secret by name.
func Secret_FromSecretName(scope constructs.Construct, id, name *string) ISecret {
	if scope == nil || id == nil || name == nil {
		panic("scope, id and name are required")
	}
	result := &importedSecret{name: name}
	constructs.NewConstruct_Override(result, scope, id)
	return result
}

func BasicAuthSecret_FromSecretName(scope constructs.Construct, id, name *string) ISecret {
	return Secret_FromSecretName(scope, id, name)
}

func SshAuthSecret_FromSecretName(scope constructs.Construct, id, name *string) ISecret {
	return Secret_FromSecretName(scope, id, name)
}

func TlsSecret_FromSecretName(scope constructs.Construct, id, name *string) ISecret {
	return Secret_FromSecretName(scope, id, name)
}

func DockerConfigSecret_FromSecretName(scope constructs.Construct, id, name *string) ISecret {
	return Secret_FromSecretName(scope, id, name)
}

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

func Secret_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func BasicAuthSecret_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func SshAuthSecret_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func TlsSecret_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func DockerConfigSecret_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

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

// EnvValue_FromSecretValue creates an environment value backed by a Secret
// key, including references to imported Secrets.
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
