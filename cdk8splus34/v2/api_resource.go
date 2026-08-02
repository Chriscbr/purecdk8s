package cdk8splus34

import "github.com/Chriscbr/purecdk8s/jsii"

// ApiResourceOptions identifies a Kubernetes API resource type.
type ApiResourceOptions struct {
	ApiGroup     *string `field:"required" json:"apiGroup" yaml:"apiGroup"`
	ResourceType *string `field:"required" json:"resourceType" yaml:"resourceType"`
}

// ApiResource is an API resource endpoint such as pods or deployments.
type ApiResource interface {
	IApiEndpoint
	IApiResource
	ApiGroup() *string
	ResourceType() *string
}

type apiResourceImpl struct {
	apiGroup     *string
	resourceType *string
}

func newApiResource(apiGroup, resourceType string) ApiResource {
	return &apiResourceImpl{apiGroup: jsii.String(apiGroup), resourceType: jsii.String(resourceType)}
}

func (a *apiResourceImpl) ApiGroup() *string {
	return a.apiGroup
}

func (a *apiResourceImpl) ResourceType() *string {
	return a.resourceType
}

func (a *apiResourceImpl) ResourceName() *string {
	return nil
}

func (a *apiResourceImpl) AsApiResource() IApiResource {
	return a
}

func (a *apiResourceImpl) AsNonApiResource() *string {
	return nil
}

// ApiResource_Custom returns a descriptor for a custom Kubernetes API resource.
func ApiResource_Custom(options *ApiResourceOptions) ApiResource {
	if options == nil || options.ApiGroup == nil || options.ResourceType == nil {
		panic("apiGroup and resourceType are required")
	}
	return &apiResourceImpl{apiGroup: options.ApiGroup, resourceType: options.ResourceType}
}

func ApiResource_API_SERVICES() ApiResource {
	return newApiResource("apiregistration.k8s.io", "apiservices")
}

func ApiResource_BINDINGS() ApiResource {
	return newApiResource("", "bindings")
}

func ApiResource_CERTIFICATE_SIGNING_REQUESTS() ApiResource {
	return newApiResource("certificates.k8s.io", "certificatesigningrequests")
}

func ApiResource_CLUSTER_ROLE_BINDINGS() ApiResource {
	return newApiResource("rbac.authorization.k8s.io", "clusterrolebindings")
}

func ApiResource_CLUSTER_ROLES() ApiResource {
	return newApiResource("rbac.authorization.k8s.io", "clusterroles")
}

func ApiResource_COMPONENT_STATUSES() ApiResource {
	return newApiResource("", "componentstatuses")
}

func ApiResource_CONFIG_MAPS() ApiResource {
	return newApiResource("", "configmaps")
}

func ApiResource_CONTROLLER_REVISIONS() ApiResource {
	return newApiResource("apps", "controllerrevisions")
}

func ApiResource_CRON_JOBS() ApiResource {
	return newApiResource("batch", "cronjobs")
}

func ApiResource_CSI_DRIVERS() ApiResource {
	return newApiResource("storage.k8s.io", "csidrivers")
}

func ApiResource_CSI_NODES() ApiResource {
	return newApiResource("storage.k8s.io", "csinodes")
}

func ApiResource_CSI_STORAGE_CAPACITIES() ApiResource {
	return newApiResource("storage.k8s.io", "csistoragecapacities")
}

func ApiResource_CUSTOM_RESOURCE_DEFINITIONS() ApiResource {
	return newApiResource("apiextensions.k8s.io", "customresourcedefinitions")
}

func ApiResource_DAEMON_SETS() ApiResource {
	return newApiResource("apps", "daemonsets")
}

func ApiResource_DEPLOYMENTS() ApiResource {
	return newApiResource("apps", "deployments")
}

func ApiResource_ENDPOINT_SLICES() ApiResource {
	return newApiResource("discovery.k8s.io", "endpointslices")
}

func ApiResource_ENDPOINTS() ApiResource {
	return newApiResource("", "endpoints")
}

func ApiResource_EVENTS() ApiResource {
	return newApiResource("", "events")
}

func ApiResource_FLOW_SCHEMAS() ApiResource {
	return newApiResource("flowcontrol.apiserver.k8s.io", "flowschemas")
}

func ApiResource_HORIZONTAL_POD_AUTOSCALERS() ApiResource {
	return newApiResource("autoscaling", "horizontalpodautoscalers")
}

func ApiResource_INGRESS_CLASSES() ApiResource {
	return newApiResource("networking.k8s.io", "ingressclasses")
}

func ApiResource_INGRESSES() ApiResource {
	return newApiResource("networking.k8s.io", "ingresses")
}

func ApiResource_JOBS() ApiResource {
	return newApiResource("batch", "jobs")
}

func ApiResource_LEASES() ApiResource {
	return newApiResource("coordination.k8s.io", "leases")
}

func ApiResource_LIMIT_RANGES() ApiResource {
	return newApiResource("", "limitranges")
}

func ApiResource_LOCAL_SUBJECT_ACCESS_REVIEWS() ApiResource {
	return newApiResource("authorization.k8s.io", "localsubjectaccessreviews")
}

func ApiResource_MUTATING_WEBHOOK_CONFIGURATIONS() ApiResource {
	return newApiResource("admissionregistration.k8s.io", "mutatingwebhookconfigurations")
}

func ApiResource_NAMESPACES() ApiResource {
	return newApiResource("", "namespaces")
}

func ApiResource_NETWORK_POLICIES() ApiResource {
	return newApiResource("networking.k8s.io", "networkpolicies")
}

func ApiResource_NODES() ApiResource {
	return newApiResource("", "nodes")
}

func ApiResource_PERSISTENT_VOLUME_CLAIMS() ApiResource {
	return newApiResource("", "persistentvolumeclaims")
}

func ApiResource_PERSISTENT_VOLUMES() ApiResource {
	return newApiResource("", "persistentvolumes")
}

func ApiResource_POD_DISRUPTION_BUDGETS() ApiResource {
	return newApiResource("policy", "poddisruptionbudgets")
}

func ApiResource_POD_TEMPLATES() ApiResource {
	return newApiResource("", "podtemplates")
}

func ApiResource_PODS() ApiResource {
	return newApiResource("", "pods")
}

func ApiResource_PRIORITY_CLASSES() ApiResource {
	return newApiResource("scheduling.k8s.io", "priorityclasses")
}

func ApiResource_PRIORITY_LEVEL_CONFIGURATIONS() ApiResource {
	return newApiResource("flowcontrol.apiserver.k8s.io", "prioritylevelconfigurations")
}

func ApiResource_REPLICA_SETS() ApiResource {
	return newApiResource("apps", "replicasets")
}

func ApiResource_REPLICATION_CONTROLLERS() ApiResource {
	return newApiResource("", "replicationcontrollers")
}

func ApiResource_RESOURCE_QUOTAS() ApiResource {
	return newApiResource("", "resourcequotas")
}

func ApiResource_ROLE_BINDINGS() ApiResource {
	return newApiResource("rbac.authorization.k8s.io", "rolebindings")
}

func ApiResource_ROLES() ApiResource {
	return newApiResource("rbac.authorization.k8s.io", "roles")
}

func ApiResource_RUNTIME_CLASSES() ApiResource {
	return newApiResource("node.k8s.io", "runtimeclasses")
}

func ApiResource_SECRETS() ApiResource {
	return newApiResource("", "secrets")
}

func ApiResource_SELF_SUBJECT_ACCESS_REVIEWS() ApiResource {
	return newApiResource("authorization.k8s.io", "selfsubjectaccessreviews")
}

func ApiResource_SELF_SUBJECT_RULES_REVIEWS() ApiResource {
	return newApiResource("authorization.k8s.io", "selfsubjectrulesreviews")
}

func ApiResource_SERVICE_ACCOUNTS() ApiResource {
	return newApiResource("", "serviceaccounts")
}

func ApiResource_SERVICES() ApiResource {
	return newApiResource("", "services")
}

func ApiResource_STATEFUL_SETS() ApiResource {
	return newApiResource("apps", "statefulsets")
}

func ApiResource_STORAGE_CLASSES() ApiResource {
	return newApiResource("storage.k8s.io", "storageclasses")
}

func ApiResource_SUBJECT_ACCESS_REVIEWS() ApiResource {
	return newApiResource("authorization.k8s.io", "subjectaccessreviews")
}

func ApiResource_TOKEN_REVIEWS() ApiResource {
	return newApiResource("authentication.k8s.io", "tokenreviews")
}

func ApiResource_VALIDATING_WEBHOOK_CONFIGURATIONS() ApiResource {
	return newApiResource("admissionregistration.k8s.io", "validatingwebhookconfigurations")
}

func ApiResource_VOLUME_ATTACHMENTS() ApiResource {
	return newApiResource("storage.k8s.io", "volumeattachments")
}

// NonApiResource describes a non-resource Kubernetes API endpoint.
type NonApiResource interface {
	IApiEndpoint
}

type nonApiResourceImpl struct{ url *string }

// NonApiResource_Of returns a descriptor for url, such as /healthz.
func NonApiResource_Of(url *string) NonApiResource {
	if url == nil {
		panic("url is required")
	}
	return &nonApiResourceImpl{url: url}
}

func (n *nonApiResourceImpl) AsApiResource() IApiResource {
	return nil
}

func (n *nonApiResourceImpl) AsNonApiResource() *string {
	return n.url
}
