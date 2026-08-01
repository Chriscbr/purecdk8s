// Package cdk8splus36 forward-ports the native cdk8s+ high-level API to the
// Kubernetes 1.36 schema. The high-level API is unchanged from 1.34; users
// can use this package together with its version-specific k8s subpackage.
package cdk8splus36

import (
	plus34 "github.com/purecdk8s/purecdk8s/cdk8splus34/v2"
	"github.com/purecdk8s/purecdk8s/constructs/v10"
)

type IApiResource = plus34.IApiResource
type IApiEndpoint = plus34.IApiEndpoint
type IResource = plus34.IResource
type Resource = plus34.Resource
type ResourceProps = plus34.ResourceProps
type ResourcePermissions = plus34.ResourcePermissions
type Protocol = plus34.Protocol

const (
	Protocol_TCP  = plus34.Protocol_TCP
	Protocol_UDP  = plus34.Protocol_UDP
	Protocol_SCTP = plus34.Protocol_SCTP
)

type ImagePullPolicy = plus34.ImagePullPolicy

const (
	ImagePullPolicy_ALWAYS         = plus34.ImagePullPolicy_ALWAYS
	ImagePullPolicy_IF_NOT_PRESENT = plus34.ImagePullPolicy_IF_NOT_PRESENT
	ImagePullPolicy_NEVER          = plus34.ImagePullPolicy_NEVER
)

type ContainerPort = plus34.ContainerPort
type ContainerProps = plus34.ContainerProps
type ContainerOpts = plus34.ContainerOpts
type Container = plus34.Container
type EnvValue = plus34.EnvValue
type EnvValueFromConfigMapOptions = plus34.EnvValueFromConfigMapOptions
type EnvValueFromSecretOptions = plus34.EnvValueFromSecretOptions
type Env = plus34.Env
type ProbeOptions = plus34.ProbeOptions
type CommandProbeOptions = plus34.CommandProbeOptions
type Probe = plus34.Probe
type MountOptions = plus34.MountOptions
type VolumeMount = plus34.VolumeMount
type IStorage = plus34.IStorage
type Volume = plus34.Volume
type ConfigMapVolumeOptions = plus34.ConfigMapVolumeOptions
type EmptyDirMedium = plus34.EmptyDirMedium

const (
	EmptyDirMedium_DEFAULT = plus34.EmptyDirMedium_DEFAULT
	EmptyDirMedium_MEMORY  = plus34.EmptyDirMedium_MEMORY
)

type EmptyDirVolumeOptions = plus34.EmptyDirVolumeOptions
type IConfigMap = plus34.IConfigMap
type ConfigMap = plus34.ConfigMap
type ConfigMapProps = plus34.ConfigMapProps
type AddDirectoryOptions = plus34.AddDirectoryOptions
type DeploymentProps = plus34.DeploymentProps
type DeploymentExposeViaServiceOptions = plus34.DeploymentExposeViaServiceOptions
type ExposeDeploymentViaIngressOptions = plus34.ExposeDeploymentViaIngressOptions
type Deployment = plus34.Deployment
type StatefulSetProps = plus34.StatefulSetProps
type StatefulSet = plus34.StatefulSet
type NodeLabelQuery = plus34.NodeLabelQuery
type LabeledNode = plus34.LabeledNode
type PodSchedulingAttractOptions = plus34.PodSchedulingAttractOptions
type PodSchedulingColocateOptions = plus34.PodSchedulingColocateOptions
type PodScheduling = plus34.PodScheduling
type WorkloadScheduling = plus34.WorkloadScheduling
type PodConnections = plus34.PodConnections
type PodConnectionsAllowToOptions = plus34.PodConnectionsAllowToOptions
type IScalable = plus34.IScalable
type ScalingTarget = plus34.ScalingTarget
type HorizontalPodAutoscalerProps = plus34.HorizontalPodAutoscalerProps
type HorizontalPodAutoscaler = plus34.HorizontalPodAutoscaler
type Metric = plus34.Metric
type MetricTarget = plus34.MetricTarget
type PodSelectorConfig = plus34.PodSelectorConfig
type LabelSelector = plus34.LabelSelector
type ServiceType = plus34.ServiceType

const (
	ServiceType_CLUSTER_IP    = plus34.ServiceType_CLUSTER_IP
	ServiceType_NODE_PORT     = plus34.ServiceType_NODE_PORT
	ServiceType_LOAD_BALANCER = plus34.ServiceType_LOAD_BALANCER
	ServiceType_EXTERNAL_NAME = plus34.ServiceType_EXTERNAL_NAME
)

type ServiceBindOptions = plus34.ServiceBindOptions
type ServicePort = plus34.ServicePort
type ServiceProps = plus34.ServiceProps
type IPodSelector = plus34.IPodSelector
type Service = plus34.Service
type HttpIngressPathType = plus34.HttpIngressPathType

const (
	HttpIngressPathType_EXACT                   = plus34.HttpIngressPathType_EXACT
	HttpIngressPathType_PREFIX                  = plus34.HttpIngressPathType_PREFIX
	HttpIngressPathType_IMPLEMENTATION_SPECIFIC = plus34.HttpIngressPathType_IMPLEMENTATION_SPECIFIC
)

type ExposeServiceViaIngressOptions = plus34.ExposeServiceViaIngressOptions
type ServiceIngressBackendOptions = plus34.ServiceIngressBackendOptions
type IngressBackend = plus34.IngressBackend
type IngressRule = plus34.IngressRule
type IngressTls = plus34.IngressTls
type IngressProps = plus34.IngressProps
type Ingress = plus34.Ingress
type ISecret = plus34.ISecret
type Secret = plus34.Secret
type SecretProps = plus34.SecretProps
type SecretValue = plus34.SecretValue
type ISubject = plus34.ISubject
type Group = plus34.Group
type RoleBinding = plus34.RoleBinding
type Namespace = plus34.Namespace
type NamespaceProps = plus34.NamespaceProps
type NamespaceSelectorConfig = plus34.NamespaceSelectorConfig
type IServiceAccount = plus34.IServiceAccount
type ServiceAccount = plus34.ServiceAccount
type ServiceAccountProps = plus34.ServiceAccountProps
type RestartPolicy = plus34.RestartPolicy
type PodProps = plus34.PodProps
type Pod = plus34.Pod
type Job = plus34.Job
type JobProps = plus34.JobProps
type CronJob = plus34.CronJob
type CronJobProps = plus34.CronJobProps

func NewContainer(props *ContainerProps) Container { return plus34.NewContainer(props) }
func NewContainer_Override(container Container, props *ContainerProps) {
	plus34.NewContainer_Override(container, props)
}
func EnvValue_FromValue(value *string) EnvValue { return plus34.EnvValue_FromValue(value) }
func EnvValue_FromConfigMap(configMap IConfigMap, key *string, options *EnvValueFromConfigMapOptions) EnvValue {
	return plus34.EnvValue_FromConfigMap(configMap, key, options)
}
func EnvValue_FromSecretValue(value *SecretValue, options *EnvValueFromSecretOptions) EnvValue {
	return plus34.EnvValue_FromSecretValue(value, options)
}
func Probe_FromCommand(command *[]*string, options *CommandProbeOptions) Probe {
	return plus34.Probe_FromCommand(command, options)
}
func Volume_FromConfigMap(scope constructs.Construct, id *string, configMap IConfigMap, options *ConfigMapVolumeOptions) Volume {
	return plus34.Volume_FromConfigMap(scope, id, configMap, options)
}
func Volume_FromEmptyDir(scope constructs.Construct, id, name *string, options *EmptyDirVolumeOptions) Volume {
	return plus34.Volume_FromEmptyDir(scope, id, name, options)
}
func Volume_FromName(scope constructs.Construct, id, name *string) Volume {
	return plus34.Volume_FromName(scope, id, name)
}
func Volume_IsConstruct(x interface{}) *bool { return plus34.Volume_IsConstruct(x) }
func NewConfigMap(scope constructs.Construct, id *string, props *ConfigMapProps) ConfigMap {
	return plus34.NewConfigMap(scope, id, props)
}
func NewConfigMap_Override(configMap ConfigMap, scope constructs.Construct, id *string, props *ConfigMapProps) {
	plus34.NewConfigMap_Override(configMap, scope, id, props)
}
func ConfigMap_IsConstruct(x interface{}) *bool { return plus34.ConfigMap_IsConstruct(x) }
func NewDeployment(scope constructs.Construct, id *string, props *DeploymentProps) Deployment {
	return plus34.NewDeployment(scope, id, props)
}
func NewDeployment_Override(deployment Deployment, scope constructs.Construct, id *string, props *DeploymentProps) {
	plus34.NewDeployment_Override(deployment, scope, id, props)
}
func Deployment_IsConstruct(x interface{}) *bool { return plus34.Deployment_IsConstruct(x) }
func NewStatefulSet(scope constructs.Construct, id *string, props *StatefulSetProps) StatefulSet {
	return plus34.NewStatefulSet(scope, id, props)
}
func NewStatefulSet_Override(statefulSet StatefulSet, scope constructs.Construct, id *string, props *StatefulSetProps) {
	plus34.NewStatefulSet_Override(statefulSet, scope, id, props)
}
func StatefulSet_IsConstruct(x interface{}) *bool { return plus34.StatefulSet_IsConstruct(x) }
func NodeLabelQuery_Is(key, value *string) NodeLabelQuery {
	return plus34.NodeLabelQuery_Is(key, value)
}
func NodeLabelQuery_In(key *string, values *[]*string) NodeLabelQuery {
	return plus34.NodeLabelQuery_In(key, values)
}
func NodeLabelQuery_NotIn(key *string, values *[]*string) NodeLabelQuery {
	return plus34.NodeLabelQuery_NotIn(key, values)
}
func NodeLabelQuery_Exists(key *string) NodeLabelQuery { return plus34.NodeLabelQuery_Exists(key) }
func NodeLabelQuery_DoesNotExist(key *string) NodeLabelQuery {
	return plus34.NodeLabelQuery_DoesNotExist(key)
}
func NodeLabelQuery_Gt(key *string, values *[]*string) NodeLabelQuery {
	return plus34.NodeLabelQuery_Gt(key, values)
}
func NodeLabelQuery_Lt(key *string, values *[]*string) NodeLabelQuery {
	return plus34.NodeLabelQuery_Lt(key, values)
}
func Node_Labeled(selectors ...NodeLabelQuery) LabeledNode { return plus34.Node_Labeled(selectors...) }
func NewHorizontalPodAutoscaler(scope constructs.Construct, id *string, props *HorizontalPodAutoscalerProps) HorizontalPodAutoscaler {
	return plus34.NewHorizontalPodAutoscaler(scope, id, props)
}
func NewHorizontalPodAutoscaler_Override(autoscaler HorizontalPodAutoscaler, scope constructs.Construct, id *string, props *HorizontalPodAutoscalerProps) {
	plus34.NewHorizontalPodAutoscaler_Override(autoscaler, scope, id, props)
}
func HorizontalPodAutoscaler_IsConstruct(x interface{}) *bool {
	return plus34.HorizontalPodAutoscaler_IsConstruct(x)
}
func MetricTarget_AverageUtilization(value *float64) MetricTarget {
	return plus34.MetricTarget_AverageUtilization(value)
}
func MetricTarget_AverageValue(value *float64) MetricTarget {
	return plus34.MetricTarget_AverageValue(value)
}
func MetricTarget_Value(value *float64) MetricTarget    { return plus34.MetricTarget_Value(value) }
func Metric_ResourceCpu(target MetricTarget) Metric     { return plus34.Metric_ResourceCpu(target) }
func Metric_ResourceMemory(target MetricTarget) Metric  { return plus34.Metric_ResourceMemory(target) }
func Metric_ResourceStorage(target MetricTarget) Metric { return plus34.Metric_ResourceStorage(target) }
func Metric_ResourceEphemeralStorage(target MetricTarget) Metric {
	return plus34.Metric_ResourceEphemeralStorage(target)
}
func NewService(scope constructs.Construct, id *string, props *ServiceProps) Service {
	return plus34.NewService(scope, id, props)
}
func NewService_Override(service Service, scope constructs.Construct, id *string, props *ServiceProps) {
	plus34.NewService_Override(service, scope, id, props)
}
func Service_IsConstruct(x interface{}) *bool { return plus34.Service_IsConstruct(x) }
func IngressBackend_FromService(service Service, options *ServiceIngressBackendOptions) IngressBackend {
	return plus34.IngressBackend_FromService(service, options)
}
func IngressBackend_FromResource(resource IResource) IngressBackend {
	return plus34.IngressBackend_FromResource(resource)
}
func NewIngress(scope constructs.Construct, id *string, props *IngressProps) Ingress {
	return plus34.NewIngress(scope, id, props)
}
func NewIngress_Override(ingress Ingress, scope constructs.Construct, id *string, props *IngressProps) {
	plus34.NewIngress_Override(ingress, scope, id, props)
}
func Ingress_IsConstruct(x interface{}) *bool { return plus34.Ingress_IsConstruct(x) }
func NewSecret(scope constructs.Construct, id *string, props *SecretProps) Secret {
	return plus34.NewSecret(scope, id, props)
}
func NewSecret_Override(secret Secret, scope constructs.Construct, id *string, props *SecretProps) {
	plus34.NewSecret_Override(secret, scope, id, props)
}
func Secret_IsConstruct(x interface{}) *bool { return plus34.Secret_IsConstruct(x) }
func Secret_FromSecretName(scope constructs.Construct, id, secretName *string) ISecret {
	return plus34.Secret_FromSecretName(scope, id, secretName)
}
func Group_FromName(scope constructs.Construct, id, name *string) Group {
	return plus34.Group_FromName(scope, id, name)
}
func Group_IsConstruct(x interface{}) *bool { return plus34.Group_IsConstruct(x) }
func NewNamespace(scope constructs.Construct, id *string, props *NamespaceProps) Namespace {
	return plus34.NewNamespace(scope, id, props)
}
func NewNamespace_Override(namespace Namespace, scope constructs.Construct, id *string, props *NamespaceProps) {
	plus34.NewNamespace_Override(namespace, scope, id, props)
}
func Namespace_IsConstruct(x interface{}) *bool { return plus34.Namespace_IsConstruct(x) }
func NewServiceAccount(scope constructs.Construct, id *string, props *ServiceAccountProps) ServiceAccount {
	return plus34.NewServiceAccount(scope, id, props)
}
func NewServiceAccount_Override(account ServiceAccount, scope constructs.Construct, id *string, props *ServiceAccountProps) {
	plus34.NewServiceAccount_Override(account, scope, id, props)
}
func ServiceAccount_IsConstruct(x interface{}) *bool { return plus34.ServiceAccount_IsConstruct(x) }
func NewPod(scope constructs.Construct, id *string, props *PodProps) Pod {
	return plus34.NewPod(scope, id, props)
}
func NewPod_Override(pod Pod, scope constructs.Construct, id *string, props *PodProps) {
	plus34.NewPod_Override(pod, scope, id, props)
}
func Pod_IsConstruct(x interface{}) *bool { return plus34.Pod_IsConstruct(x) }
func NewJob(scope constructs.Construct, id *string, props *JobProps) Job {
	return plus34.NewJob(scope, id, props)
}
func NewJob_Override(job Job, scope constructs.Construct, id *string, props *JobProps) {
	plus34.NewJob_Override(job, scope, id, props)
}
func Job_IsConstruct(x interface{}) *bool { return plus34.Job_IsConstruct(x) }
func NewCronJob(scope constructs.Construct, id *string, props *CronJobProps) CronJob {
	return plus34.NewCronJob(scope, id, props)
}
func NewCronJob_Override(job CronJob, scope constructs.Construct, id *string, props *CronJobProps) {
	plus34.NewCronJob_Override(job, scope, id, props)
}
func CronJob_IsConstruct(x interface{}) *bool { return plus34.CronJob_IsConstruct(x) }

const (
	RestartPolicy_ALWAYS     = plus34.RestartPolicy_ALWAYS
	RestartPolicy_NEVER      = plus34.RestartPolicy_NEVER
	RestartPolicy_ON_FAILURE = plus34.RestartPolicy_ON_FAILURE
)
