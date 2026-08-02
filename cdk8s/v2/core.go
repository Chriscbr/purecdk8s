package cdk8s

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
	purecdk8sserialization "github.com/Chriscbr/purecdk8s/serialization"
)

type constructBase struct {
	node constructs.Node
}

func (c *constructBase) Node() constructs.Node {
	return c.node
}

// SetNodeInternal is used by the native constructs implementation to complete
// the same host/node initialization that class inheritance provides upstream.
func (c *constructBase) SetNodeInternal(node constructs.Node) {
	c.node = node
}

func (c *constructBase) ToString() *string {
	if c.node == nil || stringValue(c.node.Path()) == "" {
		return jsii.String("<root>")
	}
	return c.node.Path()
}

func (c *constructBase) With(mixins ...constructs.IMixin) constructs.IConstruct {
	return c.node.With(mixins...)
}

type appImpl struct {
	constructBase
	outdir                  string
	outputFileExtension     string
	yamlOutputType          YamlOutputType
	resolvers               []IResolver
	recordConstructMetadata bool
}

func NewApp(props *AppProps) App {
	result := &appImpl{}
	initializeApp(result, props)
	return result
}

func NewApp_Override(app App, props *AppProps) {
	if app == nil {
		panic("parameter app is required, but nil was provided")
	}
	result, ok := app.(*appImpl)
	if ok {
		initializeApp(result, props)
		return
	}
	result = &appImpl{}
	if !setEmbeddedImplementation(app, result) {
		panic("cdk8s: App override must embed cdk8s.App")
	}
	initializeAppHost(result, app, props)
}

func initializeApp(app *appImpl, props *AppProps) {
	initializeAppHost(app, app, props)
}

func initializeAppHost(app *appImpl, host App, props *AppProps) {
	validateAppProps(props)
	if host == app {
		constructs.NewRootConstruct_Override(app, jsii.String(""))
	} else {
		app.node = constructs.NewNode(host, nil, jsii.String(""))
	}
	if props != nil && props.Outdir != nil {
		app.outdir = *props.Outdir
	} else if outdir, exists := os.LookupEnv("CDK8S_OUTDIR"); exists {
		app.outdir = outdir
	} else {
		app.outdir = "dist"
	}
	app.outputFileExtension = ".k8s.yaml"
	if props != nil && props.OutputFileExtension != nil {
		app.outputFileExtension = *props.OutputFileExtension
	}
	app.yamlOutputType = YamlOutputType_FILE_PER_CHART
	if props != nil && props.YamlOutputType != "" {
		app.yamlOutputType = props.YamlOutputType
	}
	if props != nil && props.Resolvers != nil {
		app.resolvers = append(app.resolvers, (*props.Resolvers)...)
	}
	app.resolvers = append(app.resolvers,
		NewLazyResolver(),
		NewImplicitTokenResolver(),
		NewNumberStringUnionResolver(),
	)
	app.recordConstructMetadata = os.Getenv("CDK8S_RECORD_CONSTRUCT_METADATA") == "true"
	if props != nil && props.RecordConstructMetadata != nil {
		app.recordConstructMetadata = *props.RecordConstructMetadata
	}
}

func validateAppProps(props *AppProps) {
	if props == nil {
		return
	}
	if props.Resolvers != nil {
		for _, resolver := range *props.Resolvers {
			if resolver == nil {
				panic("props.resolvers cannot contain nil")
			}
		}
	}
	if props.YamlOutputType != "" {
		switch props.YamlOutputType {
		case YamlOutputType_FILE_PER_APP,
			YamlOutputType_FILE_PER_CHART,
			YamlOutputType_FILE_PER_RESOURCE,
			YamlOutputType_FOLDER_PER_CHART_FILE_PER_RESOURCE:
		default:
			panic(fmt.Sprintf("invalid YamlOutputType: %s", props.YamlOutputType))
		}
	}
}

func (a *appImpl) Charts() *[]Chart {
	topology := NewDependencyGraph(a.Node()).Topology()
	result := make([]Chart, 0)
	if topology != nil {
		for _, item := range *topology {
			if chart, ok := item.(Chart); ok {
				result = append(result, chart)
			}
		}
	}
	return &result
}

func (a *appImpl) Outdir() *string {
	return &a.outdir
}

func (a *appImpl) OutputFileExtension() *string {
	return &a.outputFileExtension
}

func (a *appImpl) Resolvers() *[]IResolver {
	result := append([]IResolver(nil), a.resolvers...)
	return &result
}

func (a *appImpl) YamlOutputType() YamlOutputType {
	return a.yamlOutputType
}

func (a *appImpl) Synth() {
	if err := os.MkdirAll(a.outdir, 0o755); err != nil {
		panic(err)
	}
	// App.synth validates before dependency inference. Chart.ToJson and
	// SynthYaml intentionally use the opposite order upstream.
	all := appConstructs(a)
	validateAppConstructs(all)
	hasDependentCharts := resolveAppDependencies(all)
	charts := *a.Charts()

	switch a.yamlOutputType {
	case YamlOutputType_FILE_PER_APP:
		documents := make([]interface{}, 0)
		for _, chart := range charts {
			documents = append(documents, (*chart.ToJson())...)
		}
		if len(charts) > 0 {
			Yaml_Save(jsii.String(filepath.Join(a.outdir, "app"+a.outputFileExtension)), &documents)
		}
	case YamlOutputType_FILE_PER_RESOURCE:
		for _, chart := range charts {
			for _, document := range *chart.ToJson() {
				writeResourceFile(a.outdir, a.outputFileExtension, document)
			}
		}
	case YamlOutputType_FOLDER_PER_CHART_FILE_PER_RESOURCE:
		for index, chart := range charts {
			name := stringValue(Names_ToDnsLabel(chart, nil))
			if hasDependentCharts {
				name = fmt.Sprintf("%04d-%s", index, name)
			}
			directory := filepath.Join(a.outdir, name)
			if err := os.MkdirAll(directory, 0o755); err != nil {
				panic(err)
			}
			for _, document := range *chart.ToJson() {
				writeResourceFile(directory, a.outputFileExtension, document)
			}
		}
	case YamlOutputType_FILE_PER_CHART:
		for index, chart := range charts {
			name := stringValue(Names_ToDnsLabel(chart, nil))
			if hasDependentCharts {
				name = fmt.Sprintf("%04d-%s", index, name)
			}
			documents := chart.ToJson()
			Yaml_Save(jsii.String(filepath.Join(a.outdir, name+a.outputFileExtension)), documents)
		}
	}

	if a.recordConstructMetadata {
		a.writeConstructMetadata()
	}
}

func (a *appImpl) SynthYaml() *string {
	prepareApp(a)
	documents := make([]interface{}, 0)
	for _, chart := range *a.Charts() {
		documents = append(documents, (*chart.ToJson())...)
	}
	return Yaml_Stringify(documents...)
}

func (a *appImpl) writeConstructMetadata() {
	resources := make(map[string]interface{})
	for _, chart := range *a.Charts() {
		for _, object := range chartObjects(chart) {
			resources[stringValue(object.Name())] = map[string]interface{}{
				"path": stringValue(object.Node().Path()),
			}
		}
	}
	data, err := json.Marshal(map[string]interface{}{
		"version":   "1.0.0",
		"resources": resources,
	})
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(filepath.Join(a.outdir, "construct-metadata.json"), data, 0o644); err != nil {
		panic(err)
	}
}

func App_Of(construct constructs.IConstruct) App {
	if construct == nil {
		panic("cdk8s: cannot find App for nil construct")
	}
	current := construct
	for current.Node().Scope() != nil {
		current = current.Node().Scope()
	}
	app, ok := current.(App)
	if !ok {
		panic("cdk8s: root construct is not an App")
	}
	return app
}

func App_IsConstruct(value interface{}) *bool {
	if value == nil {
		panic("parameter x is required, but nil was provided")
	}
	result := false
	if construct, ok := value.(constructs.IConstruct); ok {
		result = boolValue(constructs.Construct_IsConstruct(construct))
	}
	return &result
}

type chartImpl struct {
	constructBase
	self                      Chart
	namespace                 *string
	labels                    map[string]*string
	disableResourceNameHashes bool
}

func NewChart(scope constructs.Construct, id *string, props *ChartProps) Chart {
	if scope == nil {
		panic("scope is required")
	}
	if id == nil {
		panic("id is required")
	}
	result := &chartImpl{}
	initializeChart(result, scope, id, props)
	return result
}

func NewChart_Override(chart Chart, scope constructs.Construct, id *string, props *ChartProps) {
	if chart == nil {
		panic("parameter chart is required, but nil was provided")
	}
	if scope == nil {
		panic("parameter scope is required, but nil was provided")
	}
	if id == nil {
		panic("parameter id is required, but nil was provided")
	}
	result, ok := chart.(*chartImpl)
	if ok {
		initializeChart(result, scope, id, props)
		return
	}
	result = &chartImpl{}
	if !setEmbeddedImplementation(chart, result) {
		panic("cdk8s: Chart override must embed cdk8s.Chart")
	}
	initializeChartHost(result, chart, scope, id, props)
}

func initializeChart(chart *chartImpl, scope constructs.Construct, id *string, props *ChartProps) {
	initializeChartHost(chart, chart, scope, id, props)
}

func initializeChartHost(chart *chartImpl, host Chart, scope constructs.Construct, id *string, props *ChartProps) {
	if host == chart {
		constructs.NewConstruct_Override(chart, scope, id)
	} else {
		chart.node = constructs.NewNode(host, scope, id)
	}
	chart.self = host
	chart.labels = make(map[string]*string)
	if props != nil {
		chart.namespace = props.Namespace
		chart.disableResourceNameHashes = boolValue(props.DisableResourceNameHashes)
		if props.Labels != nil {
			for key, item := range *props.Labels {
				chart.labels[key] = item
			}
		}
	}
}

func (c *chartImpl) ApiObjects() *[]ApiObject {
	result := make([]ApiObject, 0)
	for _, child := range *c.Node().Children() {
		if object, ok := child.(ApiObject); ok {
			result = append(result, object)
		}
	}
	return &result
}

func (c *chartImpl) Labels() *map[string]*string {
	result := make(map[string]*string, len(c.labels))
	for key, item := range c.labels {
		result[key] = item
	}
	return &result
}

func (c *chartImpl) Namespace() *string {
	return c.namespace
}

func (c *chartImpl) AddDependency(dependencies ...constructs.IConstruct) {
	values := make([]constructs.IDependable, len(dependencies))
	for index, dependency := range dependencies {
		values[index] = dependency
	}
	c.Node().AddDependency(values...)
}

func (c *chartImpl) GenerateObjectName(object ApiObject) *string {
	if object == nil {
		panic("parameter apiObject is required, but nil was provided")
	}
	maxLen := float64(63)
	if stringValue(object.Kind()) == "CronJob" {
		maxLen = 52
	}
	includeHash := !c.disableResourceNameHashes
	return Names_ToDnsLabel(object, &NameOptions{
		IncludeHash: &includeHash,
		MaxLen:      &maxLen,
	})
}

func (c *chartImpl) ToJson() *[]interface{} {
	app := App_Of(c)
	prepareApp(app)
	objects := chartObjects(c.self)
	result := make([]interface{}, 0, len(objects))
	for _, object := range objects {
		result = append(result, object.ToJson())
	}
	return &result
}

func Chart_Of(construct constructs.IConstruct) Chart {
	if construct == nil {
		panic("cannot find a parent chart (directly or indirectly)")
	}
	if chart, ok := construct.(Chart); ok {
		return chart
	}
	parent := construct.Node().Scope()
	if parent == nil {
		panic("cannot find a parent chart (directly or indirectly)")
	}
	return Chart_Of(parent)
}

func Chart_IsChart(value interface{}) *bool {
	if value == nil {
		panic("parameter x is required, but nil was provided")
	}
	_, ok := value.(Chart)
	return &ok
}

func Chart_IsConstruct(value interface{}) *bool {
	return App_IsConstruct(value)
}

type apiObjectImpl struct {
	constructBase
	apiVersion string
	apiGroup   string
	kind       string
	chart      Chart
	name       string
	metadata   *metadataDefinition
	manifest   interface{}
	patches    []JsonPatch
}

func NewApiObject(scope constructs.Construct, id *string, props *ApiObjectProps) ApiObject {
	return NewApiObjectWithManifest(scope, id, props, props)
}

// NewApiObjectWithManifest is the native hook used by generated purecdk8s
// bindings. It preserves the upstream ApiObject API while allowing generated
// resource-specific props to be retained without a JSII proxy.
func NewApiObjectWithManifest(scope constructs.Construct, id *string, props *ApiObjectProps, manifest interface{}) ApiObject {
	if scope == nil {
		panic("scope is required")
	}
	if id == nil {
		panic("id is required")
	}
	if props == nil || props.ApiVersion == nil || props.Kind == nil {
		panic("props.apiVersion and props.kind are required")
	}
	result := &apiObjectImpl{}
	initializeApiObject(result, result, scope, id, props, manifest)
	return result
}

// NewApiObjectWithManifest_Override is the native generated-binding
// counterpart to NewApiObject_Override. It initializes user-defined resource
// subclasses that anonymously embed their generated resource interface.
func NewApiObjectWithManifest_Override(object ApiObject, scope constructs.Construct, id *string, props *ApiObjectProps, manifest interface{}) {
	if object == nil || scope == nil || id == nil || props == nil || props.ApiVersion == nil || props.Kind == nil {
		panic("object, scope, id, props.apiVersion and props.kind are required")
	}
	if result, ok := object.(*apiObjectImpl); ok {
		initializeApiObject(result, result, scope, id, props, manifest)
		return
	}
	result := &apiObjectImpl{}
	if !setEmbeddedImplementation(object, result) {
		panic("cdk8s: ApiObject override must embed cdk8s.ApiObject")
	}
	initializeApiObject(result, object, scope, id, props, manifest)
}

func initializeApiObject(result *apiObjectImpl, host ApiObject, scope constructs.Construct, id *string, props *ApiObjectProps, manifest interface{}) {
	if host == result {
		constructs.NewConstruct_Override(result, scope, id)
	} else {
		result.node = constructs.NewNode(host, scope, id)
	}
	result.chart = Chart_Of(host)
	result.apiVersion = *props.ApiVersion
	result.kind = *props.Kind
	result.apiGroup = parseApiGroup(result.apiVersion)

	// ApiObjectProps.Metadata is part of the manifest contract even when the
	// caller supplies the remaining manifest separately. Start with it, then
	// let an explicit metadata section in that manifest take precedence.
	rawMetadata := make(map[string]interface{})
	if props.Metadata != nil {
		for key, value := range plainMap(props.Metadata) {
			rawMetadata[key] = value
		}
	}
	if manifestValues := manifestMetadata(manifest); manifestValues != nil {
		for key, value := range manifestValues {
			rawMetadata[key] = value
		}
	}
	if explicit, ok := rawMetadata["name"].(string); ok {
		result.name = explicit
	} else if props.Metadata != nil && props.Metadata.Name != nil {
		result.name = *props.Metadata.Name
	} else {
		result.name = stringValue(result.chart.GenerateObjectName(host))
	}

	// ApiObject passes an effective metadata object to the public metadata
	// constructor. Keep that merge here so constructing
	// ApiObjectMetadataDefinition directly does not unexpectedly inherit chart
	// defaults.
	effectiveMetadata := cloneStringMap(rawMetadata)
	effectiveMetadata["name"] = result.name
	if _, explicitNamespace := rawMetadata["namespace"].(string); !explicitNamespace {
		if namespace := result.chart.Namespace(); namespace != nil {
			effectiveMetadata["namespace"] = *namespace
		}
	}
	labels := make(map[string]interface{})
	for key, item := range *result.chart.Labels() {
		labels[key] = item
	}
	if manifestLabels, ok := rawMetadata["labels"].(map[string]interface{}); ok {
		for key, item := range manifestLabels {
			labels[key] = item
		}
	}
	effectiveMetadata["labels"] = labels

	result.metadata = newMetadataDefinition(host, effectiveMetadata)
	result.manifest = manifest
}

// ApiObjectManifest renders generated props as a standalone Kubernetes
// manifest. It is used by generated KubeX_Manifest functions.
func ApiObjectManifest(props *ApiObjectProps, manifest interface{}) interface{} {
	if props == nil || props.ApiVersion == nil || props.Kind == nil {
		panic("props.apiVersion and props.kind are required")
	}
	result := plainMap(manifest)
	result["apiVersion"] = *props.ApiVersion
	result["kind"] = *props.Kind
	return orderTopLevel(result)
}

func NewApiObject_Override(object ApiObject, scope constructs.Construct, id *string, props *ApiObjectProps) {
	NewApiObjectWithManifest_Override(object, scope, id, props, props)
}

func (a *apiObjectImpl) ApiGroup() *string {
	return &a.apiGroup
}

func (a *apiObjectImpl) ApiVersion() *string {
	return &a.apiVersion
}

func (a *apiObjectImpl) Chart() Chart {
	return a.chart
}

func (a *apiObjectImpl) Kind() *string {
	return &a.kind
}

func (a *apiObjectImpl) Metadata() ApiObjectMetadataDefinition {
	return a.metadata
}

func (a *apiObjectImpl) Name() *string {
	return &a.name
}

func (a *apiObjectImpl) AddDependency(dependencies ...constructs.IConstruct) {
	values := make([]constructs.IDependable, len(dependencies))
	for index, dependency := range dependencies {
		values[index] = dependency
	}
	a.Node().AddDependency(values...)
}

func (a *apiObjectImpl) AddJsonPatch(operations ...JsonPatch) {
	a.patches = append(a.patches, operations...)
}

func (a *apiObjectImpl) ToJson() interface{} {
	defer func() {
		if recovered := recover(); recovered != nil {
			panic(fmt.Sprintf(
				"Failed serializing construct at path '%s' with name '%s': %v",
				stringValue(a.Node().Path()), a.name, recovered,
			))
		}
	}()
	data := plainMap(resolveValue(a.manifest, a))
	data["apiVersion"] = a.apiVersion
	data["kind"] = a.kind
	data["metadata"] = a.metadata.ToJson()
	result := sanitizeMap(data, false)
	if len(a.patches) > 0 {
		patched := JsonPatch_Apply(result, a.patches...)
		result = plainMap(patched)
	}
	return orderTopLevel(result)
}

func ApiObject_Of(construct constructs.IConstruct) ApiObject {
	if construct == nil {
		panic("parameter c is required, but nil was provided")
	}
	if object, ok := construct.(ApiObject); ok {
		return object
	}
	child := construct.Node().DefaultChild()
	if child == nil {
		panic(fmt.Sprintf(
			"cannot find a (direct or indirect) child of type ApiObject for construct %s",
			stringValue(construct.Node().Path()),
		))
	}
	return ApiObject_Of(child)
}

func ApiObject_IsApiObject(value interface{}) *bool {
	if value == nil {
		panic("parameter o is required, but nil was provided")
	}
	_, ok := value.(ApiObject)
	return &ok
}

func ApiObject_IsConstruct(value interface{}) *bool {
	return App_IsConstruct(value)
}

func parseApiGroup(apiVersion string) string {
	parts := strings.Split(apiVersion, "/")
	switch len(parts) {
	case 1:
		return "core"
	case 2:
		return parts[0]
	default:
		panic(fmt.Sprintf(
			"invalid apiVersion %s, expecting GROUP/VERSION. See https://kubernetes.io/docs/reference/using-api/api-overview/#api-groups",
			apiVersion,
		))
	}
}

type metadataDefinition struct {
	object          ApiObject
	name            *string
	namespace       *string
	labels          map[string]interface{}
	annotations     map[string]interface{}
	finalizers      []interface{}
	ownerReferences []interface{}
	additional      map[string]interface{}
}

func NewApiObjectMetadataDefinition(options *ApiObjectMetadataDefinitionOptions) ApiObjectMetadataDefinition {
	if options == nil || options.ApiObject == nil {
		panic("options.apiObject is required")
	}
	raw := plainMap(options)
	delete(raw, "apiObject")
	return newMetadataDefinition(options.ApiObject, raw)
}

func NewApiObjectMetadataDefinition_Override(definition ApiObjectMetadataDefinition, options *ApiObjectMetadataDefinitionOptions) {
	if definition == nil {
		panic("parameter definition is required, but nil was provided")
	}
	if options == nil || options.ApiObject == nil {
		panic("options.apiObject is required")
	}
	raw := plainMap(options)
	delete(raw, "apiObject")
	created := newMetadataDefinition(options.ApiObject, raw)
	if native, ok := definition.(*metadataDefinition); ok {
		*native = *created
		return
	}
	if !setEmbeddedImplementation(definition, created) {
		panic("cdk8s: metadata override must embed cdk8s.ApiObjectMetadataDefinition")
	}
}

func newMetadataDefinition(object ApiObject, raw map[string]interface{}) *metadataDefinition {
	result := &metadataDefinition{
		object:      object,
		labels:      make(map[string]interface{}),
		annotations: make(map[string]interface{}),
		additional:  cloneStringMap(raw),
	}
	if name, ok := raw["name"].(string); ok {
		result.name = &name
	}
	if namespace, ok := raw["namespace"].(string); ok {
		result.namespace = &namespace
	}
	if labels, ok := raw["labels"].(map[string]interface{}); ok {
		for key, item := range labels {
			result.labels[key] = item
		}
	}
	if annotations, ok := raw["annotations"].(map[string]interface{}); ok {
		for key, item := range annotations {
			result.annotations[key] = item
		}
	}
	if finalizers, ok := raw["finalizers"].([]interface{}); ok {
		result.finalizers = append(result.finalizers, finalizers...)
	}
	if owners, ok := raw["ownerReferences"].([]interface{}); ok {
		result.ownerReferences = append(result.ownerReferences, owners...)
	}
	for _, key := range []string{"name", "namespace", "labels", "annotations", "finalizers", "ownerReferences", "apiObject"} {
		delete(result.additional, key)
	}
	return result
}

func (m *metadataDefinition) Name() *string {
	return m.name
}

func (m *metadataDefinition) Namespace() *string {
	return m.namespace
}

func (m *metadataDefinition) Add(key *string, value interface{}) {
	if key == nil {
		panic("key is required")
	}
	if value == nil {
		panic("parameter value is required, but nil was provided")
	}
	m.additional[*key] = value
}

func (m *metadataDefinition) AddAnnotation(key *string, value *string) {
	if key == nil || value == nil {
		panic("key and value are required")
	}
	m.annotations[*key] = value
}

func (m *metadataDefinition) AddFinalizers(finalizers ...*string) {
	for _, finalizer := range finalizers {
		if finalizer == nil {
			panic("parameter finalizers is required, but nil was provided")
		}
		m.finalizers = append(m.finalizers, finalizer)
	}
}

func (m *metadataDefinition) AddLabel(key *string, value *string) {
	if key == nil || value == nil {
		panic("key and value are required")
	}
	m.labels[*key] = value
}

func (m *metadataDefinition) AddOwnerReference(owner *OwnerReference) {
	if owner == nil {
		panic("owner is required")
	}
	if owner.ApiVersion == nil || owner.Kind == nil || owner.Name == nil || owner.Uid == nil {
		panic("owner.apiVersion, owner.kind, owner.name and owner.uid are required")
	}
	m.ownerReferences = append(m.ownerReferences, owner)
}

func (m *metadataDefinition) GetLabel(key *string) *string {
	if key == nil {
		panic("parameter key is required, but nil was provided")
	}
	value, found := m.labels[*key]
	if !found {
		return nil
	}
	plain := plainValue(value)
	if text, ok := plain.(string); ok {
		return &text
	}
	return nil
}

func (m *metadataDefinition) ToJson() interface{} {
	data := cloneStringMap(m.additional)
	data["name"] = m.name
	if m.namespace != nil {
		data["namespace"] = *m.namespace
	}
	data["annotations"] = m.annotations
	data["finalizers"] = m.finalizers
	data["ownerReferences"] = m.ownerReferences
	data["labels"] = m.labels
	resolved := plainMap(resolveValueAt(appendResolutionKey(nil, "metadata"), data, m.object))
	return sanitizeMap(resolved, true)
}

func prepareApp(app App) bool {
	all := appConstructs(app)
	hasDependentCharts := resolveAppDependencies(all)
	validateAppConstructs(all)
	return hasDependentCharts
}

func appConstructs(app App) []constructs.IConstruct {
	return append([]constructs.IConstruct(nil), (*app.Node().FindAll(constructs.ConstructOrder_PREORDER))...)
}

func resolveAppDependencies(all []constructs.IConstruct) bool {
	hasDependentCharts := false

	for _, item := range all {
		parentChart, ok := item.(Chart)
		if !ok {
			continue
		}
		for _, child := range *parentChart.Node().Children() {
			if childChart, ok := child.(Chart); ok {
				parentChart.Node().AddDependency(childChart)
				hasDependentCharts = true
			}
		}
	}

	type dependency struct {
		source constructs.IConstruct
		target constructs.IConstruct
	}
	buildDependencies := func() []dependency {
		dependencies := make([]dependency, 0)
		for _, source := range all {
			for _, target := range *source.Node().Dependencies() {
				dependencies = append(dependencies, dependency{source: source, target: target})
			}
		}
		return dependencies
	}

	for _, item := range buildDependencies() {
		sourceChart := Chart_Of(item.source)
		targetChart := Chart_Of(item.target)
		if sourceChart != targetChart {
			sourceChart.Node().AddDependency(targetChart)
			hasDependentCharts = true
		}
	}

	// Upstream rebuilds the dependency list after chart dependencies have been
	// inferred. This makes a cross-chart relationship order all objects in the
	// source chart after all objects in the target chart.
	for _, item := range buildDependencies() {
		sourceChart := Chart_Of(item.source)
		targetChart := Chart_Of(item.target)
		sourceObjects := apiObjectsUnder(item.source, sourceChart)
		targetObjects := apiObjectsUnder(item.target, targetChart)
		for _, source := range sourceObjects {
			for _, target := range targetObjects {
				if source != target {
					source.Node().AddDependency(target)
				}
			}
		}
	}
	return hasDependentCharts
}

func validateAppConstructs(all []constructs.IConstruct) {
	errors := make([]string, 0)
	for _, child := range all {
		if validationErrors := child.Node().Validate(); validationErrors != nil {
			for _, item := range *validationErrors {
				if item != nil {
					errors = append(errors, fmt.Sprintf("[%s] %s", stringValue(child.Node().Path()), *item))
				}
			}
		}
	}
	if len(errors) > 0 {
		panic("Validation failed with the following errors:\n  " + strings.Join(errors, "\n  "))
	}
}

func apiObjectsUnder(root constructs.IConstruct, chart Chart) []ApiObject {
	result := make([]ApiObject, 0)
	for _, item := range *root.Node().FindAll(constructs.ConstructOrder_PREORDER) {
		object, ok := item.(ApiObject)
		if ok && object.Chart() == chart {
			result = append(result, object)
		}
	}
	return result
}

func chartObjects(chart Chart) []ApiObject {
	topology := NewDependencyGraph(chart.Node()).Topology()
	result := make([]ApiObject, 0)
	if topology != nil {
		for _, item := range *topology {
			object, ok := item.(ApiObject)
			if ok && object.Chart() == chart {
				result = append(result, object)
			}
		}
	}
	return result
}

func writeResourceFile(directory, extension string, document interface{}) {
	manifest := plainMap(document)
	metadata, _ := manifest["metadata"].(map[string]interface{})
	kind := fmt.Sprint(manifest["kind"])
	name := fmt.Sprint(metadata["name"])
	fileName := regexp.MustCompile(`[^0-9a-zA-Z-_.]`).ReplaceAllString(kind+"."+name, "")
	documents := []interface{}{document}
	Yaml_Save(jsii.String(filepath.Join(directory, fileName+extension)), &documents)
}

func orderTopLevel(input map[string]interface{}) map[string]interface{} {
	// Go maps do not carry order, but retaining this helper centralizes the
	// upstream precedence. Yaml_Stringify creates an ordered node from it.
	result := make(map[string]interface{}, len(input))
	for _, key := range []string{"apiVersion", "kind", "metadata"} {
		if value, ok := input[key]; ok {
			result[key] = value
		}
	}
	keys := make([]string, 0, len(input))
	for key := range input {
		if key != "apiVersion" && key != "kind" && key != "metadata" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		result[key] = input[key]
	}
	return result
}

func plainMap(value interface{}) map[string]interface{} {
	plain := plainValue(value)
	if plain == nil {
		return make(map[string]interface{})
	}
	if result, ok := plain.(map[string]interface{}); ok {
		return result
	}
	panic(fmt.Sprintf("expected object, got %T", plain))
}

func manifestMetadata(manifest interface{}) map[string]interface{} {
	if manifest == nil {
		return nil
	}
	value := reflect.ValueOf(manifest)
	for value.IsValid() && (value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer) {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}
	if !value.IsValid() {
		return nil
	}
	switch value.Kind() {
	case reflect.Map:
		iterator := value.MapRange()
		for iterator.Next() {
			if fmt.Sprint(plainValue(iterator.Key().Interface())) == "metadata" {
				return plainMap(iterator.Value().Interface())
			}
		}
	case reflect.Struct:
		typ := value.Type()
		for index := 0; index < value.NumField(); index++ {
			field := typ.Field(index)
			name := field.Tag.Get("k8s")
			if name == "" {
				name = strings.Split(field.Tag.Get("json"), ",")[0]
			}
			if name == "" {
				name = strings.Split(field.Tag.Get("yaml"), ",")[0]
			}
			if name == "" {
				name = lowerFirst(field.Name)
			}
			if name != "metadata" {
				continue
			}
			fieldValue := value.Field(index)
			if !fieldValue.IsValid() || !fieldValue.CanInterface() {
				return nil
			}
			return plainMap(fieldValue.Interface())
		}
	}
	return nil
}

func plainValue(input interface{}) interface{} {
	if input == nil {
		return nil
	}
	value := reflect.ValueOf(input)
	return plainReflect(value)
}

func plainReflect(value reflect.Value) interface{} {
	if !value.IsValid() {
		return nil
	}
	// JSII enums use a sanitized public symbol (for example
	// "DELETE_OPTIONS") and an original Kubernetes wire value
	// ("DeleteOptions"). Generated packages register this mapping without
	// adding methods to the enum's upstream-compatible method set.
	if value.CanInterface() {
		if wireValue, ok := purecdk8sserialization.EnumWireValue(value.Interface()); ok {
			return plainValue(wireValue)
		}
	}
	// Generated number/string unions implement Value on a pointer receiver.
	// Consult that interface before dereferencing so their concrete wrapper
	// fields never leak into a manifest.
	if (value.Kind() != reflect.Pointer && value.Kind() != reflect.Interface) || !value.IsNil() {
		if value.CanInterface() {
			if producer, ok := value.Interface().(interface{ Value() interface{} }); ok {
				produced := producer.Value()
				if reflect.TypeOf(produced) != value.Type() {
					return plainValue(produced)
				}
			}
		}
	}
	for value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}
	if value.CanInterface() {
		if timeValue, ok := value.Interface().(time.Time); ok {
			return timeValue.Format(time.RFC3339)
		}
		if producer, ok := value.Interface().(interface{ Value() interface{} }); ok {
			return plainValue(producer.Value())
		}
	}
	switch value.Kind() {
	case reflect.Bool:
		return value.Bool()
	case reflect.String:
		return value.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(value.Uint())
	case reflect.Float32, reflect.Float64:
		return value.Float()
	case reflect.Slice, reflect.Array:
		result := make([]interface{}, 0, value.Len())
		for index := 0; index < value.Len(); index++ {
			result = append(result, plainReflect(value.Index(index)))
		}
		return result
	case reflect.Map:
		result := make(map[string]interface{}, value.Len())
		iterator := value.MapRange()
		for iterator.Next() {
			key := fmt.Sprint(plainReflect(iterator.Key()))
			item := plainReflect(iterator.Value())
			if item != nil {
				result[key] = item
			}
		}
		return result
	case reflect.Struct:
		result := make(map[string]interface{})
		typ := value.Type()
		for index := 0; index < value.NumField(); index++ {
			field := typ.Field(index)
			if field.PkgPath != "" {
				continue
			}
			fieldValue := value.Field(index)
			if field.Tag.Get("field") == "optional" && fieldValue.IsZero() {
				continue
			}
			name := field.Tag.Get("k8s")
			if name == "" {
				name = strings.Split(field.Tag.Get("json"), ",")[0]
			}
			if name == "" {
				name = strings.Split(field.Tag.Get("yaml"), ",")[0]
			}
			if name == "" {
				name = lowerFirst(field.Name)
			}
			if name == "-" {
				continue
			}
			item := plainReflect(fieldValue)
			if item != nil {
				result[name] = item
			}
		}
		return result
	default:
		if value.CanInterface() {
			return value.Interface()
		}
		return nil
	}
}

func sanitizeMap(input map[string]interface{}, filterEmpty bool) map[string]interface{} {
	result := make(map[string]interface{})
	for key, item := range input {
		sanitized, keep := sanitizeValue(item, filterEmpty)
		if keep {
			result[key] = sanitized
		}
	}
	return result
}

func sanitizeValue(input interface{}, filterEmpty bool) (interface{}, bool) {
	if input == nil {
		return nil, false
	}
	switch value := input.(type) {
	case map[string]interface{}:
		result := sanitizeMap(value, filterEmpty)
		if filterEmpty && len(result) == 0 {
			return nil, false
		}
		return result, true
	case []interface{}:
		if filterEmpty && len(value) == 0 {
			return nil, false
		}
		result := make([]interface{}, len(value))
		for index, item := range value {
			sanitized, keep := sanitizeValue(item, filterEmpty)
			if keep {
				result[index] = sanitized
			}
		}
		return result, true
	default:
		return value, true
	}
}

func cloneStringMap(input map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(input))
	for key, item := range input {
		result[key] = item
	}
	return result
}

func lowerFirst(value string) string {
	if value == "" {
		return value
	}
	return strings.ToLower(value[:1]) + value[1:]
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func boolValue(value *bool) bool {
	return value != nil && *value
}

func setEmbeddedImplementation(target interface{}, implementation interface{}) bool {
	value := reflect.ValueOf(target)
	if !value.IsValid() || value.Kind() != reflect.Pointer || value.IsNil() {
		return false
	}
	value = value.Elem()
	if value.Kind() != reflect.Struct {
		return false
	}
	implementationValue := reflect.ValueOf(implementation)
	for index := 0; index < value.NumField(); index++ {
		fieldInfo := value.Type().Field(index)
		field := value.Field(index)
		if !fieldInfo.Anonymous || !field.CanSet() || field.Kind() != reflect.Interface {
			continue
		}
		if implementationValue.Type().Implements(field.Type()) {
			field.Set(implementationValue)
			return true
		}
	}
	return false
}

func Testing_App(props *AppProps) App {
	directory, err := os.MkdirTemp("", "cdk8s.outdir.")
	if err != nil {
		panic(err)
	}
	cloned := AppProps{}
	if props != nil {
		cloned = *props
	}
	if cloned.Outdir == nil {
		cloned.Outdir = &directory
	}
	return NewApp(&cloned)
}

func Testing_Chart() Chart {
	return NewChart(Testing_App(nil), jsii.String("test"), nil)
}

func Testing_Synth(chart Chart) *[]interface{} {
	if chart == nil {
		panic("parameter chart is required, but nil was provided")
	}
	return chart.ToJson()
}
