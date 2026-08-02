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
	ISubject
	AutomountToken() *bool
	Secrets() *[]ISecret
	AddSecret(secret ISecret)
}

// IServiceAccount identifies a service account resource and is a valid RBAC
// role-binding subject.
type IServiceAccount interface {
	IResource
	ISubject
}

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
	applyOverride(account, NewServiceAccount(scope, id, props), "ServiceAccount")
}

func ServiceAccount_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

// ServiceAccount_FromServiceAccountName imports an existing ServiceAccount by name.
func ServiceAccount_FromServiceAccountName(scope constructs.Construct, id, name *string, options *FromServiceAccountNameOptions) IServiceAccount {
	if scope == nil || id == nil || name == nil {
		panic("scope, id and name are required")
	}
	result := &importedServiceAccount{name: name}
	if options != nil {
		result.namespace = options.NamespaceName
	}
	constructs.NewConstruct_Override(result, scope, id)
	return result
}

type FromServiceAccountNameOptions struct {
	NamespaceName *string `field:"optional" json:"namespaceName" yaml:"namespaceName"`
}

type importedServiceAccount struct {
	node      constructs.Node
	name      *string
	namespace *string
}

func (s *importedServiceAccount) Node() constructs.Node {
	return s.node
}

func (s *importedServiceAccount) SetNodeInternal(node constructs.Node) {
	s.node = node
}

func (s *importedServiceAccount) ToString() *string {
	return s.node.Path()
}

func (s *importedServiceAccount) With(mixins ...constructs.IMixin) constructs.IConstruct {
	return s.node.With(mixins...)
}

func (s *importedServiceAccount) ApiVersion() *string {
	return jsii.String("v1")
}

func (s *importedServiceAccount) ApiGroup() *string {
	return jsii.String("")
}

func (s *importedServiceAccount) Kind() *string {
	return jsii.String("ServiceAccount")
}

func (s *importedServiceAccount) Name() *string {
	return s.name
}

func (s *importedServiceAccount) ResourceName() *string {
	return s.name
}

func (s *importedServiceAccount) ResourceType() *string {
	return jsii.String("serviceaccounts")
}

func (s *importedServiceAccount) ToSubjectConfiguration() *SubjectConfiguration {
	return &SubjectConfiguration{ApiGroup: jsii.String(""), Kind: jsii.String("ServiceAccount"), Name: s.Name(), Namespace: s.namespace}
}

func (s *serviceAccountImpl) AutomountToken() *bool {
	return s.automountToken
}

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

func (s *serviceAccountImpl) ToSubjectConfiguration() *SubjectConfiguration {
	return &SubjectConfiguration{ApiGroup: jsii.String(""), Kind: jsii.String("ServiceAccount"), Name: s.Name(), Namespace: s.Metadata().Namespace()}
}
