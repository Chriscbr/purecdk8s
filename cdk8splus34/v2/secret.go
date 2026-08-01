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
