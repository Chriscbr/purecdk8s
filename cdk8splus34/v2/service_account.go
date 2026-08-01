package cdk8splus34

import (
	"github.com/purecdk8s/purecdk8s/cdk8s/v2"
	"github.com/purecdk8s/purecdk8s/constructs/v10"
	"github.com/purecdk8s/purecdk8s/jsii"
)

// ServiceAccountProps configures a Kubernetes ServiceAccount.
type ServiceAccountProps struct {
	Metadata       *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	AutomountToken *bool                    `field:"optional" json:"automountToken" yaml:"automountToken"`
	Secrets        *[]ISecret               `field:"optional" json:"secrets" yaml:"secrets"`
}

// ServiceAccount is a native Kubernetes ServiceAccount construct.
type ServiceAccount interface {
	Resource
	IServiceAccount
	AutomountToken() *bool
	Secrets() *[]ISecret
	AddSecret(secret ISecret)
}

// IServiceAccount identifies a service account resource.
type IServiceAccount interface{ IResource }

type serviceAccountImpl struct {
	resourceBase
	automountToken *bool
	secrets        []ISecret
}

func NewServiceAccount(scope constructs.Construct, id *string, props *ServiceAccountProps) ServiceAccount {
	if props == nil {
		props = &ServiceAccountProps{}
	}
	automountToken := props.AutomountToken
	if automountToken == nil {
		automountToken = jsii.Bool(false)
	}
	result := &serviceAccountImpl{automountToken: automountToken}
	if props.Secrets != nil {
		result.secrets = append(result.secrets, (*props.Secrets)...)
	}
	manifest := map[string]interface{}{"automountServiceAccountToken": automountToken}
	result.resourceBase.initialize(result, scope, id, "v1", "ServiceAccount", "serviceaccounts", props.Metadata, manifest)
	manifest["secrets"] = cdk8s.Lazy_Any(lazyProducer{produce: func() interface{} {
		if len(result.secrets) == 0 {
			return []interface{}{}
		}
		values := make([]interface{}, 0, len(result.secrets))
		for _, secret := range result.secrets {
			values = append(values, map[string]interface{}{"name": secret.ResourceName()})
		}
		return values
	}})
	return result
}

func NewServiceAccount_Override(account ServiceAccount, scope constructs.Construct, id *string, props *ServiceAccountProps) {
	panic("native cdk8splus34 overrides are not implemented")
}

func ServiceAccount_IsConstruct(x interface{}) *bool { return constructs.Construct_IsConstruct(x) }

func (s *serviceAccountImpl) AutomountToken() *bool { return s.automountToken }

func (s *serviceAccountImpl) Secrets() *[]ISecret {
	values := append([]ISecret(nil), s.secrets...)
	return &values
}

func (s *serviceAccountImpl) AddSecret(secret ISecret) {
	if secret == nil {
		panic("secret is required")
	}
	s.secrets = append(s.secrets, secret)
}
