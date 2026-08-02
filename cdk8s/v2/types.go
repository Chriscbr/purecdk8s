package cdk8s

import "github.com/Chriscbr/purecdk8s/constructs/v10"

type ApiObjectMetadata struct {
	Annotations     *map[string]*string `field:"optional" json:"annotations" yaml:"annotations"`
	Finalizers      *[]*string          `field:"optional" json:"finalizers" yaml:"finalizers"`
	Labels          *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	Name            *string             `field:"optional" json:"name" yaml:"name"`
	Namespace       *string             `field:"optional" json:"namespace" yaml:"namespace"`
	OwnerReferences *[]*OwnerReference  `field:"optional" json:"ownerReferences" yaml:"ownerReferences"`
}

type ApiObjectMetadataDefinitionOptions struct {
	Annotations     *map[string]*string `field:"optional" json:"annotations" yaml:"annotations"`
	Finalizers      *[]*string          `field:"optional" json:"finalizers" yaml:"finalizers"`
	Labels          *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	Name            *string             `field:"optional" json:"name" yaml:"name"`
	Namespace       *string             `field:"optional" json:"namespace" yaml:"namespace"`
	OwnerReferences *[]*OwnerReference  `field:"optional" json:"ownerReferences" yaml:"ownerReferences"`
	ApiObject       ApiObject           `field:"required" json:"apiObject" yaml:"apiObject"`
}

type ApiObjectProps struct {
	ApiVersion *string            `field:"required" json:"apiVersion" yaml:"apiVersion"`
	Kind       *string            `field:"required" json:"kind" yaml:"kind"`
	Metadata   *ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
}

type AppProps struct {
	Outdir                  *string        `field:"optional" json:"outdir" yaml:"outdir"`
	OutputFileExtension     *string        `field:"optional" json:"outputFileExtension" yaml:"outputFileExtension"`
	RecordConstructMetadata *bool          `field:"optional" json:"recordConstructMetadata" yaml:"recordConstructMetadata"`
	Resolvers               *[]IResolver   `field:"optional" json:"resolvers" yaml:"resolvers"`
	YamlOutputType          YamlOutputType `field:"optional" json:"yamlOutputType" yaml:"yamlOutputType"`
}

type ChartProps struct {
	DisableResourceNameHashes *bool               `field:"optional" json:"disableResourceNameHashes" yaml:"disableResourceNameHashes"`
	Labels                    *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	Namespace                 *string             `field:"optional" json:"namespace" yaml:"namespace"`
}

type CronOptions struct {
	Day     *string `field:"optional" json:"day" yaml:"day"`
	Hour    *string `field:"optional" json:"hour" yaml:"hour"`
	Minute  *string `field:"optional" json:"minute" yaml:"minute"`
	Month   *string `field:"optional" json:"month" yaml:"month"`
	WeekDay *string `field:"optional" json:"weekDay" yaml:"weekDay"`
}

type GroupVersionKind struct {
	ApiVersion *string `field:"required" json:"apiVersion" yaml:"apiVersion"`
	Kind       *string `field:"required" json:"kind" yaml:"kind"`
}

type HelmProps struct {
	Chart          *string                 `field:"required" json:"chart" yaml:"chart"`
	HelmExecutable *string                 `field:"optional" json:"helmExecutable" yaml:"helmExecutable"`
	HelmFlags      *[]*string              `field:"optional" json:"helmFlags" yaml:"helmFlags"`
	Namespace      *string                 `field:"optional" json:"namespace" yaml:"namespace"`
	ReleaseName    *string                 `field:"optional" json:"releaseName" yaml:"releaseName"`
	Repo           *string                 `field:"optional" json:"repo" yaml:"repo"`
	Values         *map[string]interface{} `field:"optional" json:"values" yaml:"values"`
	Version        *string                 `field:"optional" json:"version" yaml:"version"`
}

type IncludeProps struct {
	Url *string `field:"required" json:"url" yaml:"url"`
}

type NameOptions struct {
	Delimiter   *string    `field:"optional" json:"delimiter" yaml:"delimiter"`
	Extra       *[]*string `field:"optional" json:"extra" yaml:"extra"`
	IncludeHash *bool      `field:"optional" json:"includeHash" yaml:"includeHash"`
	MaxLen      *float64   `field:"optional" json:"maxLen" yaml:"maxLen"`
}

type OwnerReference struct {
	ApiVersion         *string `field:"required" json:"apiVersion" yaml:"apiVersion"`
	Kind               *string `field:"required" json:"kind" yaml:"kind"`
	Name               *string `field:"required" json:"name" yaml:"name"`
	Uid                *string `field:"required" json:"uid" yaml:"uid"`
	BlockOwnerDeletion *bool   `field:"optional" json:"blockOwnerDeletion" yaml:"blockOwnerDeletion"`
	Controller         *bool   `field:"optional" json:"controller" yaml:"controller"`
}

type SizeConversionOptions struct {
	Rounding SizeRoundingBehavior `field:"optional" json:"rounding" yaml:"rounding"`
}

type TimeConversionOptions struct {
	Integral *bool `field:"optional" json:"integral" yaml:"integral"`
}

type YamlOutputType string

const (
	YamlOutputType_FILE_PER_APP                       YamlOutputType = "FILE_PER_APP"
	YamlOutputType_FILE_PER_CHART                     YamlOutputType = "FILE_PER_CHART"
	YamlOutputType_FILE_PER_RESOURCE                  YamlOutputType = "FILE_PER_RESOURCE"
	YamlOutputType_FOLDER_PER_CHART_FILE_PER_RESOURCE YamlOutputType = "FOLDER_PER_CHART_FILE_PER_RESOURCE"
)

type SizeRoundingBehavior string

const (
	SizeRoundingBehavior_FAIL  SizeRoundingBehavior = "FAIL"
	SizeRoundingBehavior_FLOOR SizeRoundingBehavior = "FLOOR"
	SizeRoundingBehavior_NONE  SizeRoundingBehavior = "NONE"
)

type ApiObject interface {
	constructs.Construct
	ApiGroup() *string
	ApiVersion() *string
	Chart() Chart
	Kind() *string
	Metadata() ApiObjectMetadataDefinition
	Name() *string
	Node() constructs.Node
	AddDependency(dependencies ...constructs.IConstruct)
	AddJsonPatch(ops ...JsonPatch)
	ToJson() interface{}
	ToString() *string
}

type ApiObjectMetadataDefinition interface {
	Name() *string
	Namespace() *string
	Add(key *string, value interface{})
	AddAnnotation(key *string, value *string)
	AddFinalizers(finalizers ...*string)
	AddLabel(key *string, value *string)
	AddOwnerReference(owner *OwnerReference)
	GetLabel(key *string) *string
	ToJson() interface{}
}

type App interface {
	constructs.Construct
	Charts() *[]Chart
	Node() constructs.Node
	Outdir() *string
	OutputFileExtension() *string
	Resolvers() *[]IResolver
	YamlOutputType() YamlOutputType
	Synth()
	SynthYaml() *string
	ToString() *string
}

type Chart interface {
	constructs.Construct
	ApiObjects() *[]ApiObject
	Labels() *map[string]*string
	Namespace() *string
	Node() constructs.Node
	AddDependency(dependencies ...constructs.IConstruct)
	GenerateObjectName(apiObject ApiObject) *string
	ToJson() *[]interface{}
	ToString() *string
}

type Cron interface {
	ExpressionString() *string
}

type DependencyGraph interface {
	Root() DependencyVertex
	Topology() *[]constructs.IConstruct
}

type DependencyVertex interface {
	Inbound() *[]DependencyVertex
	Outbound() *[]DependencyVertex
	Value() constructs.IConstruct
	AddChild(dep DependencyVertex)
	Topology() *[]constructs.IConstruct
}

type Duration interface {
	ToDays(opts *TimeConversionOptions) *float64
	ToHours(opts *TimeConversionOptions) *float64
	ToHumanString() *string
	ToIsoString() *string
	ToMilliseconds(opts *TimeConversionOptions) *float64
	ToMinutes(opts *TimeConversionOptions) *float64
	ToSeconds(opts *TimeConversionOptions) *float64
	UnitLabel() *string
}

type Helm interface {
	Include
	ApiObjects() *[]ApiObject
	Node() constructs.Node
	ReleaseName() *string
	ToString() *string
}

type IAnyProducer interface {
	Produce() interface{}
}

type IResolver interface {
	Resolve(context ResolutionContext)
}

type ImplicitTokenResolver interface {
	IResolver
	Resolve(context ResolutionContext)
}

type Include interface {
	constructs.Construct
	ApiObjects() *[]ApiObject
	Node() constructs.Node
	ToString() *string
}

type JsonPatch interface{}

type Lazy interface {
	Produce() interface{}
}

type LazyResolver interface {
	IResolver
	Resolve(context ResolutionContext)
}

type Names interface{}

type NumberStringUnionResolver interface {
	IResolver
	Resolve(context ResolutionContext)
}

type ResolutionContext interface {
	Key() *[]*string
	Obj() ApiObject
	Replaced() *bool
	SetReplaced(val *bool)
	ReplacedValue() interface{}
	SetReplacedValue(val interface{})
	Value() interface{}
	ReplaceValue(newValue interface{})
}

type Size interface {
	AsString() *string
	ToGibibytes(opts *SizeConversionOptions) *float64
	ToKibibytes(opts *SizeConversionOptions) *float64
	ToMebibytes(opts *SizeConversionOptions) *float64
	ToPebibytes(opts *SizeConversionOptions) *float64
	ToTebibytes(opts *SizeConversionOptions) *float64
}

type (
	Testing interface{}
	Yaml    interface{}
)
