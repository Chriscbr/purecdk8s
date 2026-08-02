package cdk8splus34

import "github.com/purecdk8s/purecdk8s/cdk8s/v2"

// WorkloadProps configures the common pod-template portion of a workload.
type WorkloadProps struct {
	Metadata                     *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	AutomountServiceAccountToken *bool                    `field:"optional" json:"automountServiceAccountToken" yaml:"automountServiceAccountToken"`
	Containers                   *[]*ContainerProps       `field:"optional" json:"containers" yaml:"containers"`
	Dns                          *PodDnsProps             `field:"optional" json:"dns" yaml:"dns"`
	DockerRegistryAuth           ISecret                  `field:"optional" json:"dockerRegistryAuth" yaml:"dockerRegistryAuth"`
	EnableServiceLinks           *bool                    `field:"optional" json:"enableServiceLinks" yaml:"enableServiceLinks"`
	HostAliases                  *[]*HostAlias            `field:"optional" json:"hostAliases" yaml:"hostAliases"`
	HostNetwork                  *bool                    `field:"optional" json:"hostNetwork" yaml:"hostNetwork"`
	InitContainers               *[]*ContainerProps       `field:"optional" json:"initContainers" yaml:"initContainers"`
	Isolate                      *bool                    `field:"optional" json:"isolate" yaml:"isolate"`
	RestartPolicy                RestartPolicy            `field:"optional" json:"restartPolicy" yaml:"restartPolicy"`
	SecurityContext              *PodSecurityContextProps `field:"optional" json:"securityContext" yaml:"securityContext"`
	ServiceAccount               IServiceAccount          `field:"optional" json:"serviceAccount" yaml:"serviceAccount"`
	ShareProcessNamespace        *bool                    `field:"optional" json:"shareProcessNamespace" yaml:"shareProcessNamespace"`
	TerminationGracePeriod       cdk8s.Duration           `field:"optional" json:"terminationGracePeriod" yaml:"terminationGracePeriod"`
	Volumes                      *[]Volume                `field:"optional" json:"volumes" yaml:"volumes"`
	PodMetadata                  *cdk8s.ApiObjectMetadata `field:"optional" json:"podMetadata" yaml:"podMetadata"`
	Select                       *bool                    `field:"optional" json:"select" yaml:"select"`
	Spread                       *bool                    `field:"optional" json:"spread" yaml:"spread"`
}

func (p *WorkloadProps) podProps() *PodProps {
	if p == nil {
		return &PodProps{}
	}
	return &PodProps{
		Metadata: p.Metadata, AutomountServiceAccountToken: p.AutomountServiceAccountToken,
		Containers: p.Containers, Dns: p.Dns, DockerRegistryAuth: p.DockerRegistryAuth,
		EnableServiceLinks: p.EnableServiceLinks, HostAliases: p.HostAliases, HostNetwork: p.HostNetwork,
		InitContainers: p.InitContainers, Isolate: p.Isolate, RestartPolicy: p.RestartPolicy,
		SecurityContext: p.SecurityContext, ServiceAccount: p.ServiceAccount,
		ShareProcessNamespace: p.ShareProcessNamespace, TerminationGracePeriod: p.TerminationGracePeriod,
		Volumes: p.Volumes,
	}
}
