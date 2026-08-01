package cdk8splus34

import (
	"github.com/purecdk8s/purecdk8s/cdk8s/v2"
	"github.com/purecdk8s/purecdk8s/constructs/v10"
	"github.com/purecdk8s/purecdk8s/jsii"
)

// SecretProps configures a Kubernetes Secret.
type SecretProps struct {
	Metadata   *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	Immutable  *bool                    `field:"optional" json:"immutable" yaml:"immutable"`
	StringData *map[string]*string      `field:"optional" json:"stringData" yaml:"stringData"`
	Type       *string                  `field:"optional" json:"type" yaml:"type"`
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

func (s *importedSecret) Node() constructs.Node                { return s.node }
func (s *importedSecret) SetNodeInternal(node constructs.Node) { s.node = node }
func (s *importedSecret) ToString() *string                    { return s.node.Path() }
func (s *importedSecret) With(mixins ...constructs.IMixin) constructs.IConstruct {
	return s.node.With(mixins...)
}
func (s *importedSecret) ApiVersion() *string   { return jsii.String("v1") }
func (s *importedSecret) ApiGroup() *string     { return jsii.String("") }
func (s *importedSecret) Kind() *string         { return jsii.String("Secret") }
func (s *importedSecret) Name() *string         { return s.name }
func (s *importedSecret) ResourceName() *string { return s.name }
func (s *importedSecret) ResourceType() *string { return jsii.String("secrets") }
func (s *importedSecret) AsApiResource() IApiResource {
	return s
}
func (s *importedSecret) AsNonApiResource() *string { return nil }
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

func NewSecret_Override(secret Secret, scope constructs.Construct, id *string, props *SecretProps) {
	panic("native cdk8splus34 overrides are not implemented")
}

func Secret_IsConstruct(x interface{}) *bool { return constructs.Construct_IsConstruct(x) }

func (s *secretImpl) Immutable() *bool { return s.immutable }

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
