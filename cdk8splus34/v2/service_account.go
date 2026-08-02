package cdk8splus34

import (
	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// Properties for initialization of `ServiceAccount`.
type ServiceAccountProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// Indicates whether pods running as this service account should have an API token automatically mounted.
	//
	// Can be overridden at the pod level. See: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#use-the-default-service-account-to-access-the-api-server
	//
	// Default: false.
	AutomountToken *bool `field:"optional" json:"automountToken" yaml:"automountToken"`
	// List of secrets allowed to be used by pods running using this ServiceAccount. See: https://kubernetes.io/docs/concepts/configuration/secret
	Secrets *[]ISecret `field:"optional" json:"secrets" yaml:"secrets"`
}

// A service account provides an identity for processes that run in a Pod.
//
// When you (a human) access the cluster (for example, using kubectl), you are authenticated by the apiserver as a particular User Account (currently this is usually admin, unless your cluster administrator has customized your cluster). Processes in containers inside pods can also contact the apiserver. When they do, they are authenticated as a particular Service Account (for example, default). See: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account
type ServiceAccount interface {
	Resource
	IServiceAccount
	ISubject
	// Whether or not a token is automatically mounted for this service account.
	AutomountToken() *bool
	// List of secrets allowed to be used by pods running using this service account.
	//
	// Returns a copy. To add a secret, use `addSecret()`.
	Secrets() *[]ISecret
	// Allow a secret to be accessed by pods using this service account.
	AddSecret(secret ISecret)
}

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
			return nil
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

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func ServiceAccount_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

// Imports a service account from the cluster as a reference.
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
	// The name of the namespace the service account belongs to. Default: "default".
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
