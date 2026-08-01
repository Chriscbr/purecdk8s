package importer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	helmChartFileName  = "Chart.yaml"
	helmValuesFileName = "values.schema.json"
	maxHelmPullOutput  = 10 * 1024 * 1024
)

// HelmSourceKind identifies how a Helm import is acquired and rendered.
type HelmSourceKind string

const (
	HelmSourceLocal HelmSourceKind = "local"
	HelmSourceHTTP  HelmSourceKind = "http"
	HelmSourceOCI   HelmSourceKind = "oci"
)

// HelmSource is the parsed, runtime-relevant part of a helm: import.
type HelmSource struct {
	Kind         HelmSourceKind
	Source       string
	ChartName    string
	ChartVersion string
	Chart        string
	Repo         string
	LocalPath    string
}

type helmChartDocument struct {
	Name         string `yaml:"name"`
	Version      string `yaml:"version"`
	Dependencies []struct {
		Name string `yaml:"name"`
	} `yaml:"dependencies"`
}

var helmSemanticVersionPattern = regexp.MustCompile(
	`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`,
)

// ParseHelmSource recognizes all upstream helm: import forms without
// performing I/O:
//   - helm:./local-chart
//   - helm:https://repo.example/charts/chart@1.2.3
//   - helm:oci://registry.example/charts/chart@1.2.3
func ParseHelmSource(source string) (*HelmSource, bool, error) {
	source = strings.TrimSpace(source)
	if !strings.HasPrefix(source, "helm:") {
		return nil, false, nil
	}
	value := strings.TrimPrefix(source, "helm:")
	if value == "" {
		return nil, true, fmt.Errorf("Invalid helm URL: %s. Must match a supported helm import format.", source)
	}
	if strings.HasPrefix(value, ".") || strings.HasPrefix(value, "/") {
		return &HelmSource{
			Kind:      HelmSourceLocal,
			Source:    source,
			Chart:     value,
			LocalPath: value,
		}, true, nil
	}

	reference, version, ok := splitHelmReferenceVersion(value)
	if !ok {
		if strings.HasPrefix(value, "oci://") {
			return nil, true, fmt.Errorf(
				"Invalid helm URL: %s. Must match the format: 'helm:<oci-registry-url>@<chart-version>'.",
				source,
			)
		}
		return nil, true, fmt.Errorf(
			"Invalid helm URL: %s. Must match the format: 'helm:<repo-url>/<chart-name>@<chart-version>'.",
			source,
		)
	}
	if !isValidHelmSemVer(version) {
		return nil, true, fmt.Errorf(
			"Invalid chart version (%s) in URL: %s. Must follow SemVer-2 (see https://semver.org/).",
			version,
			source,
		)
	}

	if strings.HasPrefix(reference, "oci://") {
		name := helmReferenceName(reference)
		if name == "" || !validHelmReference(reference) {
			return nil, true, fmt.Errorf(
				"Invalid helm URL: %s. Must match the format: 'helm:<oci-registry-url>@<chart-version>'.",
				source,
			)
		}
		return &HelmSource{
			Kind:         HelmSourceOCI,
			Source:       source,
			ChartName:    name,
			ChartVersion: version,
			Chart:        reference,
		}, true, nil
	}

	if !strings.HasPrefix(reference, "http://") && !strings.HasPrefix(reference, "https://") {
		return nil, true, fmt.Errorf(
			"Invalid helm URL: %s. Must match the format: 'helm:<repo-url>/<chart-name>@<chart-version>'.",
			source,
		)
	}
	lastSlash := strings.LastIndex(reference, "/")
	if lastSlash <= strings.Index(reference, "://")+2 || lastSlash == len(reference)-1 {
		return nil, true, fmt.Errorf(
			"Invalid helm URL: %s. Must match the format: 'helm:<repo-url>/<chart-name>@<chart-version>'.",
			source,
		)
	}
	repo, name := reference[:lastSlash], reference[lastSlash+1:]
	if !validHelmReference(repo) || !validHelmReference(name) {
		return nil, true, fmt.Errorf(
			"Invalid helm URL: %s. Must match the format: 'helm:<repo-url>/<chart-name>@<chart-version>'.",
			source,
		)
	}
	return &HelmSource{
		Kind:         HelmSourceHTTP,
		Source:       source,
		ChartName:    name,
		ChartVersion: version,
		Chart:        name,
		Repo:         repo,
	}, true, nil
}

func splitHelmReferenceVersion(value string) (reference, version string, ok bool) {
	index := strings.LastIndex(value, "@")
	if index <= 0 || index == len(value)-1 {
		return "", "", false
	}
	return value[:index], value[index+1:], true
}

func helmReferenceName(reference string) string {
	index := strings.LastIndex(reference, "/")
	if index < 0 || index == len(reference)-1 {
		return ""
	}
	return reference[index+1:]
}

func validHelmReference(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		switch {
		case character >= 'a' && character <= 'z':
		case character >= 'A' && character <= 'Z':
		case character >= '0' && character <= '9':
		case strings.ContainsRune("_./:-", character):
		default:
			return false
		}
	}
	return true
}

func isValidHelmSemVer(value string) bool {
	if len(value) > 256 {
		return false
	}
	value = strings.TrimPrefix(strings.TrimSpace(value), "v")
	matches := helmSemanticVersionPattern.FindStringSubmatch(value)
	if matches == nil {
		return false
	}
	const maxSafeInteger = uint64(9_007_199_254_740_991)
	for _, component := range matches[1:4] {
		number, err := strconv.ParseUint(component, 10, 64)
		if err != nil || number > maxSafeInteger {
			return false
		}
	}
	if prerelease := matches[4]; prerelease != "" {
		for _, identifier := range strings.Split(prerelease, ".") {
			numeric := true
			for _, character := range identifier {
				if character < '0' || character > '9' {
					numeric = false
					break
				}
			}
			if numeric && len(identifier) > 1 && identifier[0] == '0' {
				return false
			}
		}
	}
	return true
}

// ParseLocalHelmSource is retained as a compatibility convenience for callers
// that only accept local charts.
func ParseLocalHelmSource(source string) (chartDirectory string, helm bool, err error) {
	parsed, helm, err := ParseHelmSource(source)
	if err != nil || !helm {
		return "", helm, err
	}
	if parsed.Kind != HelmSourceLocal {
		return "", true, fmt.Errorf("Helm import %q is not a local chart", source)
	}
	return parsed.LocalPath, true, nil
}

type helmRuntimeReference struct {
	Chart   string
	Repo    string
	Version string
}

// GenerateHelmSource parses and acquires a local, HTTP repository, or OCI
// Helm import and generates its ordinary Go package. workDirectory is used
// only to read relative local charts; their original ./ path is retained in
// generated HelmProps so generated packages remain portable.
func GenerateHelmSource(
	ctx context.Context,
	source string,
	workDirectory string,
	helmExecutable string,
	options GenerateOptions,
) (*Generation, *HelmSource, error) {
	parsed, helm, err := ParseHelmSource(source)
	if err != nil {
		return nil, nil, err
	}
	if !helm {
		return nil, nil, fmt.Errorf("source %q is not a Helm import", source)
	}
	if parsed.Kind == HelmSourceLocal {
		readDirectory := parsed.LocalPath
		if !filepath.IsAbs(readDirectory) && workDirectory != "" {
			readDirectory = filepath.Join(workDirectory, readDirectory)
		}
		chart, err := readHelmChartDocument(filepath.Clean(readDirectory))
		if err != nil {
			return nil, nil, err
		}
		if !isValidHelmSemVer(chart.Version) {
			return nil, nil, fmt.Errorf(
				"Invalid chart version (%s) in URL: %s. Must follow SemVer-2 (see https://semver.org/).",
				chart.Version,
				source,
			)
		}
		parsed.ChartName = chart.Name
		parsed.ChartVersion = chart.Version
		generated, err := generateHelmDirectory(
			filepath.Clean(readDirectory),
			chart,
			helmRuntimeReference{Chart: parsed.Chart, Version: chart.Version},
			options,
		)
		return generated, parsed, err
	}

	acquiredDirectory, cleanup, err := pullHelmChart(ctx, parsed, helmExecutable)
	if err != nil {
		return nil, nil, err
	}
	defer cleanup()
	chart, err := readHelmChartDocument(acquiredDirectory)
	if err != nil {
		return nil, nil, err
	}
	generated, err := generateHelmDirectory(
		acquiredDirectory,
		chart,
		helmRuntimeReference{
			Chart:   parsed.Chart,
			Repo:    parsed.Repo,
			Version: parsed.ChartVersion,
		},
		options,
	)
	return generated, parsed, err
}

func pullHelmChart(
	ctx context.Context,
	source *HelmSource,
	helmExecutable string,
) (chartDirectory string, cleanup func(), err error) {
	if source == nil || (source.Kind != HelmSourceHTTP && source.Kind != HelmSourceOCI) {
		return "", nil, fmt.Errorf("remote Helm source is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if helmExecutable == "" {
		helmExecutable = "helm"
	}
	workDirectory, err := os.MkdirTemp("", "purecdk8s-helm-")
	if err != nil {
		return "", nil, fmt.Errorf("create Helm pull directory: %w", err)
	}
	remove := func() { _ = os.RemoveAll(workDirectory) }
	args := []string{"pull"}
	if source.Kind == HelmSourceHTTP {
		args = append(args, source.ChartName, "--repo", source.Repo)
	} else {
		args = append(args, source.Chart)
	}
	args = append(args, "--version", source.ChartVersion, "--untar", "--untardir", workDirectory)
	command := exec.CommandContext(ctx, helmExecutable, args...)
	output, commandErr := command.CombinedOutput()
	if len(output) > maxHelmPullOutput {
		remove()
		return "", nil, fmt.Errorf("pull Helm chart: command output exceeded %d bytes", maxHelmPullOutput)
	}
	if commandErr != nil {
		remove()
		if errors.Is(commandErr, exec.ErrNotFound) || errors.Is(commandErr, os.ErrNotExist) {
			return "", nil, fmt.Errorf(
				"Unable to execute '%s' to pull the Helm chart. Is helm installed on your system?",
				helmExecutable,
			)
		}
		if ctx.Err() != nil {
			return "", nil, fmt.Errorf("pull Helm chart: %w", ctx.Err())
		}
		message := strings.TrimSpace(string(output))
		if message != "" {
			return "", nil, errors.New(message)
		}
		return "", nil, fmt.Errorf("Failed pulling helm chart from URL (%s): %w", source.Chart, commandErr)
	}
	return filepath.Join(workDirectory, source.ChartName), remove, nil
}

// GenerateHelm generates an ordinary Go package for a local Helm chart.
//
// Chart.yaml supplies the construct name, chart version and dependency names.
// When values.schema.json is present, the generated package contains typed
// values DTOs and enums. Without a schema, the Values property remains an
// untyped map, matching the upstream cdk8s Helm import API.
func GenerateHelm(chartDirectory string, options GenerateOptions) (*Generation, error) {
	if strings.TrimSpace(chartDirectory) == "" {
		return nil, fmt.Errorf("generate Helm chart: chart directory is required")
	}
	runtimePath := chartDirectory
	readDirectory := filepath.Clean(chartDirectory)
	chart, err := readHelmChartDocument(readDirectory)
	if err != nil {
		return nil, err
	}
	if !isValidHelmSemVer(chart.Version) {
		return nil, fmt.Errorf("invalid Helm Chart.yaml: version %q must follow SemVer-2", chart.Version)
	}
	return generateHelmDirectory(
		readDirectory,
		chart,
		helmRuntimeReference{Chart: runtimePath, Version: chart.Version},
		options,
	)
}

func generateHelmDirectory(
	chartDirectory string,
	chart *helmChartDocument,
	runtimeSource helmRuntimeReference,
	options GenerateOptions,
) (*Generation, error) {
	if chart == nil {
		return nil, fmt.Errorf("generate Helm chart: Chart.yaml metadata is required")
	}
	if strings.TrimSpace(chart.Name) == "" {
		return nil, fmt.Errorf("generate Helm chart: chart name is required")
	}
	typeName := helmConstructTypeName(chart.Name)
	if typeName == "" {
		return nil, fmt.Errorf("generate Helm chart: chart name %q cannot be represented as a Go type", chart.Name)
	}
	if options.PackageName == "" {
		options.PackageName = prefixedPackageName(options.PackagePrefix, chart.Name)
	}
	if options.CoreImport == "" {
		options.CoreImport = defaultCoreImport
	}
	if options.ConstructsImport == "" {
		options.ConstructsImport = defaultConstructsImport
	}
	if options.SerializationImport == "" {
		options.SerializationImport = defaultSerializationImport
	}

	valuesSchema, schemaPresent, err := readHelmValuesSchema(chartDirectory)
	if err != nil {
		return nil, err
	}
	dependencies := make([]string, 0, len(chart.Dependencies))
	for _, dependency := range chart.Dependencies {
		if dependency.Name != "" {
			dependencies = append(dependencies, dependency.Name)
		}
	}
	if schemaPresent {
		if err := prepareHelmValuesSchema(valuesSchema, dependencies); err != nil {
			return nil, err
		}
	}

	generator := newHelmPackageGenerator(options, valuesSchema, schemaPresent)
	valuesType := typeName + "Values"
	if schemaPresent {
		// json2jsii emits the root values struct with the construct-derived
		// name, but uses the chart's raw FQN to name inline child types.
		// For example, sample-chart produces SamplechartValues while its
		// image and mode properties become SampleChartImage and
		// SampleChartMode.
		generator.addType(valuesType, chart.Name, valuesSchema)
	}
	code, err := renderHelmPackage(
		generator,
		typeName,
		runtimeSource.Chart,
		runtimeSource.Repo,
		runtimeSource.Version,
		valuesType,
		valuesSchema,
		schemaPresent,
	)
	if err != nil {
		return nil, err
	}

	types := make([]string, 0, len(generator.types)+1)
	for _, generatedType := range generator.types {
		types = append(types, generatedType.Name)
	}
	types = append(types, typeName+"Props")
	return &Generation{
		PackageName: options.PackageName,
		Code:        code,
		Types:       types,
		Resources:   []string{typeName},
	}, nil
}

func readHelmChartDocument(chartDirectory string) (*helmChartDocument, error) {
	filename := filepath.Join(chartDirectory, helmChartFileName)
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read Helm Chart.yaml: %w", err)
	}
	var chart helmChartDocument
	if err := yaml.Unmarshal(data, &chart); err != nil {
		return nil, fmt.Errorf("decode Helm Chart.yaml: %w", err)
	}
	if strings.TrimSpace(chart.Name) == "" || strings.TrimSpace(chart.Version) == "" {
		return nil, fmt.Errorf("invalid Helm Chart.yaml: name and version are required")
	}
	return &chart, nil
}

func readHelmValuesSchema(chartDirectory string) (*schema, bool, error) {
	filename := filepath.Join(chartDirectory, helmValuesFileName)
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("read Helm values.schema.json: %w", err)
	}
	var values schema
	if err := json.Unmarshal(data, &values); err != nil {
		return nil, false, fmt.Errorf("decode Helm values.schema.json: %w", err)
	}
	return &values, true, nil
}

func helmConstructTypeName(chartName string) string {
	var value strings.Builder
	for _, character := range chartName {
		switch {
		case character >= 'a' && character <= 'z':
			value.WriteRune(character)
		case character >= 'A' && character <= 'Z':
			value.WriteRune(character)
		case character >= '0' && character <= '9':
			value.WriteRune(character)
		case character == ' ':
			value.WriteRune(character)
		}
	}
	words := strings.Fields(value.String())
	for index, word := range words {
		words[index] = upperFirst(word)
	}
	return normalizeTypeName(strings.Join(words, ""))
}

func prepareHelmValuesSchema(values *schema, dependencies []string) error {
	if values == nil {
		return fmt.Errorf("invalid Helm values.schema.json: schema is required")
	}
	if valueType := schemaType(values); valueType != "" && valueType != "object" {
		return fmt.Errorf("invalid Helm values.schema.json: root type must be object, received %q", valueType)
	}
	if values.Properties == nil {
		values.Properties = make(map[string]*schema)
	}
	for _, dependency := range dependencies {
		values.Properties[dependency] = helmOpenValuesMapSchema()
	}
	values.Properties["global"] = helmOpenValuesMapSchema()
	addAdditionalValuesToHelmSchema(values, make(map[*schema]bool))
	for _, definition := range values.Definitions {
		addAdditionalValuesToHelmSchema(definition, make(map[*schema]bool))
	}
	for _, definition := range values.Defs {
		addAdditionalValuesToHelmSchema(definition, make(map[*schema]bool))
	}
	return nil
}

func helmOpenValuesMapSchema() *schema {
	return &schema{
		Type:                 "object",
		AdditionalProperties: &schema{Type: "object"},
	}
}

func addAdditionalValuesToHelmSchema(item *schema, seen map[*schema]bool) {
	if item == nil || seen[item] {
		return
	}
	seen[item] = true
	for _, name := range sortedKeys(item.Properties) {
		if name == "additionalValues" {
			continue
		}
		addAdditionalValuesToHelmSchema(item.Properties[name], seen)
	}
	addAdditionalValuesToHelmSchema(item.Items, seen)
	for _, option := range item.OneOf {
		addAdditionalValuesToHelmSchema(option, seen)
	}
	for _, option := range item.AnyOf {
		addAdditionalValuesToHelmSchema(option, seen)
	}
	for _, option := range item.AllOf {
		addAdditionalValuesToHelmSchema(option, seen)
	}
	switch item.AdditionalProperties.(type) {
	case *schema, map[string]interface{}:
		additional := additionalSchema(item.AdditionalProperties)
		addAdditionalValuesToHelmSchema(additional, seen)
		item.AdditionalProperties = additional
	}
	if item.Properties != nil {
		item.Properties["additionalValues"] = &schema{
			Type:        "object",
			Description: "Values that are not available in values.schema.json will not be code generated. You can add such values to this property.",
			AdditionalProperties: &schema{
				Type: "object",
			},
		}
	}
}

func newHelmPackageGenerator(options GenerateOptions, values *schema, schemaPresent bool) *packageGenerator {
	definitions := make(map[string]*schema)
	if schemaPresent && values != nil {
		for name, definition := range values.Definitions {
			definitions[name] = definition
		}
		for name, definition := range values.Defs {
			definitions[name] = definition
		}
	}
	generator := &packageGenerator{
		options:       options,
		definitions:   definitions,
		resourceByFQN: make(map[string]resourceDefinition),
		typeByFQN:     make(map[string]string),
		typeSchemas:   make(map[string]*schema),
		typeFQNs:      make(map[string]string),
		typeIndexes:   make(map[string]int),
		renderedTypes: make(map[string]bool),
	}
	for _, fqn := range sortedKeys(definitions) {
		generator.typeByFQN[fqn] = typeNameForDefinition(fqn)
	}
	return generator
}

func renderHelmPackage(
	generator *packageGenerator,
	typeName string,
	chart string,
	repo string,
	chartVersion string,
	valuesType string,
	valuesSchema *schema,
	schemaPresent bool,
) ([]byte, error) {
	var out bytes.Buffer
	usesTime := helmSchemaUsesDateTime(valuesSchema)
	fmt.Fprintln(&out, "// Code generated by purecdk8s. DO NOT EDIT.")
	fmt.Fprintf(&out, "package %s\n\n", generator.options.PackageName)
	fmt.Fprintln(&out, "import (")
	fmt.Fprintln(&out, "\t\"encoding/json\"")
	fmt.Fprintln(&out, "\t\"fmt\"")
	fmt.Fprintln(&out, "\t\"reflect\"")
	fmt.Fprintln(&out, "\t\"strings\"")
	if usesTime {
		fmt.Fprintln(&out, "\t\"time\"")
	}
	fmt.Fprintf(&out, "\tcdk8s %q\n", generator.options.CoreImport)
	fmt.Fprintf(&out, "\tconstructs %q\n", generator.options.ConstructsImport)
	fmt.Fprintf(&out, "\tpurecdk8sserialization %q\n", generator.options.SerializationImport)
	fmt.Fprintln(&out, ")")
	fmt.Fprintln(&out)
	if usesTime {
		fmt.Fprintln(&out, "var _ = time.Time{}")
		fmt.Fprintln(&out)
	}

	generator.renderScalarUnionHelper(&out)
	for index := 0; index < len(generator.types); index++ {
		item := generator.types[index]
		if err := generator.renderType(&out, item); err != nil {
			return nil, err
		}
		generator.renderedTypes[item.Name] = true
	}
	renderHelmProps(&out, typeName, valuesType, valuesSchema, schemaPresent)
	renderHelmConstruct(&out, typeName, chart, repo, chartVersion, schemaPresent)
	renderHelmValueHelpers(&out, typeName)
	generator.renderHelpers(&out)

	formatted, err := format.Source(out.Bytes())
	if err != nil {
		return nil, fmt.Errorf(
			"format generated %s Helm package: %w\n%s",
			generator.options.PackageName,
			err,
			numberedSource(out.String()),
		)
	}
	return formatted, nil
}

func helmSchemaUsesDateTime(root *schema) bool {
	seen := make(map[*schema]bool)
	var visit func(*schema) bool
	visit = func(item *schema) bool {
		if item == nil || seen[item] {
			return false
		}
		seen[item] = true
		if item.Format == "date-time" {
			return true
		}
		for _, property := range item.Properties {
			if visit(property) {
				return true
			}
		}
		if visit(item.Items) {
			return true
		}
		for _, option := range item.OneOf {
			if visit(option) {
				return true
			}
		}
		for _, option := range item.AnyOf {
			if visit(option) {
				return true
			}
		}
		for _, option := range item.AllOf {
			if visit(option) {
				return true
			}
		}
		if item.AdditionalProperties != nil && visit(additionalSchema(item.AdditionalProperties)) {
			return true
		}
		for _, definition := range item.Definitions {
			if visit(definition) {
				return true
			}
		}
		for _, definition := range item.Defs {
			if visit(definition) {
				return true
			}
		}
		return false
	}
	return visit(root)
}

func renderHelmProps(
	out *bytes.Buffer,
	typeName string,
	valuesType string,
	valuesSchema *schema,
	schemaPresent bool,
) {
	fmt.Fprintf(out, "type %sProps struct {\n", typeName)
	valuesRequired := schemaPresent && valuesSchema != nil && len(valuesSchema.Required) > 0
	renderValues := func() {
		if schemaPresent {
			requirement := "optional"
			if valuesRequired {
				requirement = "required"
			}
			fmt.Fprintf(
				out,
				"\tValues *%s `field:%q json:\"values\" yaml:\"values\"`\n",
				valuesType,
				requirement,
			)
			return
		}
		fmt.Fprintln(out, "\tValues *map[string]interface{} `field:\"optional\" json:\"values\" yaml:\"values\"`")
	}
	// jsii places required fields before optional fields, with each group
	// alphabetized. Values is therefore first only when the chart schema has
	// required properties; otherwise it follows ReleaseName.
	if valuesRequired {
		renderValues()
	}
	fmt.Fprintln(out, "\tHelmExecutable *string `field:\"optional\" json:\"helmExecutable\" yaml:\"helmExecutable\"`")
	fmt.Fprintln(out, "\tHelmFlags *[]*string `field:\"optional\" json:\"helmFlags\" yaml:\"helmFlags\"`")
	fmt.Fprintln(out, "\tNamespace *string `field:\"optional\" json:\"namespace\" yaml:\"namespace\"`")
	fmt.Fprintln(out, "\tReleaseName *string `field:\"optional\" json:\"releaseName\" yaml:\"releaseName\"`")
	if !valuesRequired {
		renderValues()
	}
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)
}

func renderHelmConstruct(
	out *bytes.Buffer,
	typeName string,
	chart string,
	repo string,
	chartVersion string,
	schemaPresent bool,
) {
	implementation := lowerFirst(typeName)
	fmt.Fprintf(out, "type %s interface {\n", typeName)
	fmt.Fprintln(out, "\tconstructs.Construct")
	fmt.Fprintln(out, "\tHelm() cdk8s.Helm")
	fmt.Fprintln(out, "\tSetHelm(value cdk8s.Helm)")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)

	fmt.Fprintf(out, "type %s struct {\n", implementation)
	fmt.Fprintln(out, "\tconstructs.Construct")
	fmt.Fprintln(out, "\thelm cdk8s.Helm")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)
	fmt.Fprintf(out, "func (value *%s) Helm() cdk8s.Helm { return value.helm }\n", implementation)
	fmt.Fprintf(out, "func (value *%s) SetHelm(helm cdk8s.Helm) { value.helm = helm }\n", implementation)
	fmt.Fprintln(out)

	fmt.Fprintf(out, "func New%s(scope constructs.Construct, id *string, props *%sProps) %s {\n", typeName, typeName, typeName)
	fmt.Fprintln(out, "\tif scope == nil || id == nil { panic(\"scope and id are required\") }")
	fmt.Fprintf(out, "\tresult := &%s{Construct: constructs.NewConstruct(scope, id)}\n", implementation)
	fmt.Fprintf(out, "\t_initialize%s(result, props)\n", typeName)
	fmt.Fprintln(out, "\treturn result")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)

	fmt.Fprintf(out, "func New%s_Override(value %s, scope constructs.Construct, id *string, props *%sProps) {\n", typeName, typeName, typeName)
	fmt.Fprintln(out, "\tif value == nil || scope == nil || id == nil { panic(\"value, scope and id are required\") }")
	fmt.Fprintf(out, "\tif result, ok := value.(*%s); ok {\n", implementation)
	fmt.Fprintln(out, "\t\tresult.Construct = constructs.NewConstruct(scope, id)")
	fmt.Fprintln(out, "\t} else {")
	fmt.Fprintf(out, "\t\tbase := &%s{Construct: constructs.NewConstruct(scope, id)}\n", implementation)
	fmt.Fprintf(out, "\t\tif !_purecdk8sSetEmbedded%s(value, base) { panic(%q) }\n", typeName, typeName+" override must embed "+typeName)
	fmt.Fprintln(out, "\t}")
	fmt.Fprintf(out, "\t_initialize%s(value, props)\n", typeName)
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)

	fmt.Fprintf(out, "func %s_IsConstruct(value interface{}) *bool {\n", typeName)
	fmt.Fprintln(out, "\t_, ok := value.(constructs.IConstruct)")
	fmt.Fprintln(out, "\treturn &ok")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)

	fmt.Fprintf(out, "func _initialize%s(value %s, props *%sProps) {\n", typeName, typeName, typeName)
	fmt.Fprintf(out, "\tif props == nil { props = &%sProps{} }\n", typeName)
	fmt.Fprintf(out, "\t_purecdk8sValidateRequired(props, %q)\n", typeName+"Props")
	if repo == "" {
		fmt.Fprintf(out, "\tchart, version := %q, %q\n", chart, chartVersion)
	} else {
		fmt.Fprintf(out, "\tchart, repo, version := %q, %q, %q\n", chart, repo, chartVersion)
	}
	fmt.Fprintln(out, "\tfinalProps := &cdk8s.HelmProps{")
	fmt.Fprintln(out, "\t\tChart: &chart,")
	if repo != "" {
		fmt.Fprintln(out, "\t\tRepo: &repo,")
	}
	fmt.Fprintln(out, "\t\tVersion: &version,")
	fmt.Fprintln(out, "\t\tHelmExecutable: props.HelmExecutable,")
	fmt.Fprintln(out, "\t\tHelmFlags: props.HelmFlags,")
	fmt.Fprintln(out, "\t\tNamespace: props.Namespace,")
	fmt.Fprintln(out, "\t\tReleaseName: props.ReleaseName,")
	fmt.Fprintln(out, "\t}")
	fmt.Fprintln(out, "\tif props.Values != nil {")
	if schemaPresent {
		fmt.Fprintf(out, "\t\t_purecdk8sValidateRequired(props.Values, %q)\n", typeName+"Values")
	}
	fmt.Fprintln(out, "\t\tplain := _purecdk8sHelmPlain(reflect.ValueOf(props.Values))")
	fmt.Fprintln(out, "\t\tflattened, ok := _purecdk8sFlattenAdditionalValues(plain).(map[string]interface{})")
	fmt.Fprintln(out, "\t\tif !ok { panic(\"Helm values must resolve to an object\") }")
	fmt.Fprintln(out, "\t\tfinalProps.Values = &flattened")
	fmt.Fprintln(out, "\t}")
	fmt.Fprintln(out, "\thelmID := \"Helm\"")
	fmt.Fprintln(out, "\tvalue.SetHelm(cdk8s.NewHelm(value, &helmID, finalProps))")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)
}

func renderHelmValueHelpers(out *bytes.Buffer, typeName string) {
	fmt.Fprintln(out, "func _purecdk8sHelmPlain(value reflect.Value) interface{} {")
	fmt.Fprintln(out, "\tif !value.IsValid() { return nil }")
	fmt.Fprintln(out, "\tfor value.Kind() == reflect.Interface || value.Kind() == reflect.Ptr {")
	fmt.Fprintln(out, "\t\tif value.IsNil() { return nil }")
	fmt.Fprintln(out, "\t\tif value.CanInterface() {")
	fmt.Fprintln(out, "\t\t\tif union, ok := value.Interface().(interface{ Value() interface{} }); ok {")
	fmt.Fprintln(out, "\t\t\t\treturn _purecdk8sHelmPlain(reflect.ValueOf(union.Value()))")
	fmt.Fprintln(out, "\t\t\t}")
	fmt.Fprintln(out, "\t\t}")
	fmt.Fprintln(out, "\t\tvalue = value.Elem()")
	fmt.Fprintln(out, "\t}")
	fmt.Fprintln(out, "\tif value.CanInterface() {")
	fmt.Fprintln(out, "\t\tif wireValue, ok := purecdk8sserialization.EnumWireValue(value.Interface()); ok {")
	fmt.Fprintln(out, "\t\t\treturn _purecdk8sHelmPlain(reflect.ValueOf(wireValue))")
	fmt.Fprintln(out, "\t\t}")
	fmt.Fprintln(out, "\t}")
	fmt.Fprintln(out, "\tif value.CanInterface() && value.Kind() == reflect.Struct && value.Type().PkgPath() == \"time\" && value.Type().Name() == \"Time\" {")
	fmt.Fprintln(out, "\t\tdata, err := json.Marshal(value.Interface())")
	fmt.Fprintln(out, "\t\tif err != nil { panic(err) }")
	fmt.Fprintln(out, "\t\tvar result interface{}")
	fmt.Fprintln(out, "\t\tif err := json.Unmarshal(data, &result); err != nil { panic(err) }")
	fmt.Fprintln(out, "\t\treturn result")
	fmt.Fprintln(out, "\t}")
	fmt.Fprintln(out, "\tswitch value.Kind() {")
	fmt.Fprintln(out, "\tcase reflect.Struct:")
	fmt.Fprintln(out, "\t\tresult := make(map[string]interface{})")
	fmt.Fprintln(out, "\t\ttype_ := value.Type()")
	fmt.Fprintln(out, "\t\tfor index := 0; index < value.NumField(); index++ {")
	fmt.Fprintln(out, "\t\t\tfieldInfo := type_.Field(index)")
	fmt.Fprintln(out, "\t\t\tif fieldInfo.PkgPath != \"\" { continue }")
	fmt.Fprintln(out, "\t\t\tfield := value.Field(index)")
	fmt.Fprintln(out, "\t\t\tif fieldInfo.Tag.Get(\"field\") == \"optional\" && field.IsZero() { continue }")
	fmt.Fprintln(out, "\t\t\tname := fieldInfo.Tag.Get(\"k8s\")")
	fmt.Fprintln(out, "\t\t\tif name == \"\" { name = fieldInfo.Tag.Get(\"json\") }")
	fmt.Fprintln(out, "\t\t\tif comma := strings.IndexByte(name, ','); comma >= 0 { name = name[:comma] }")
	fmt.Fprintln(out, "\t\t\tif name == \"\" { name = fieldInfo.Name }")
	fmt.Fprintln(out, "\t\t\tif name == \"-\" { continue }")
	fmt.Fprintln(out, "\t\t\tresult[name] = _purecdk8sHelmPlain(field)")
	fmt.Fprintln(out, "\t\t}")
	fmt.Fprintln(out, "\t\treturn result")
	fmt.Fprintln(out, "\tcase reflect.Map:")
	fmt.Fprintln(out, "\t\tif value.IsNil() { return nil }")
	fmt.Fprintln(out, "\t\tresult := make(map[string]interface{}, value.Len())")
	fmt.Fprintln(out, "\t\titerator := value.MapRange()")
	fmt.Fprintln(out, "\t\tfor iterator.Next() {")
	fmt.Fprintln(out, "\t\t\tresult[fmt.Sprint(iterator.Key().Interface())] = _purecdk8sHelmPlain(iterator.Value())")
	fmt.Fprintln(out, "\t\t}")
	fmt.Fprintln(out, "\t\treturn result")
	fmt.Fprintln(out, "\tcase reflect.Slice, reflect.Array:")
	fmt.Fprintln(out, "\t\tif value.Kind() == reflect.Slice && value.IsNil() { return nil }")
	fmt.Fprintln(out, "\t\tresult := make([]interface{}, value.Len())")
	fmt.Fprintln(out, "\t\tfor index := 0; index < value.Len(); index++ { result[index] = _purecdk8sHelmPlain(value.Index(index)) }")
	fmt.Fprintln(out, "\t\treturn result")
	fmt.Fprintln(out, "\tdefault:")
	fmt.Fprintln(out, "\t\tif value.CanInterface() { return value.Interface() }")
	fmt.Fprintln(out, "\t\treturn nil")
	fmt.Fprintln(out, "\t}")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)

	fmt.Fprintln(out, "func _purecdk8sFlattenAdditionalValues(value interface{}) interface{} {")
	fmt.Fprintln(out, "\tswitch item := value.(type) {")
	fmt.Fprintln(out, "\tcase map[string]interface{}:")
	fmt.Fprintln(out, "\t\tresult := make(map[string]interface{}, len(item))")
	fmt.Fprintln(out, "\t\tfor key, nested := range item {")
	fmt.Fprintln(out, "\t\t\tif key == \"additionalValues\" { continue }")
	fmt.Fprintln(out, "\t\t\tresult[key] = _purecdk8sFlattenAdditionalValues(nested)")
	fmt.Fprintln(out, "\t\t}")
	fmt.Fprintln(out, "\t\tif additional, ok := item[\"additionalValues\"].(map[string]interface{}); ok {")
	fmt.Fprintln(out, "\t\t\tfor key, nested := range additional { result[key] = _purecdk8sFlattenAdditionalValues(nested) }")
	fmt.Fprintln(out, "\t\t}")
	fmt.Fprintln(out, "\t\treturn result")
	fmt.Fprintln(out, "\tcase []interface{}:")
	fmt.Fprintln(out, "\t\tresult := make([]interface{}, len(item))")
	fmt.Fprintln(out, "\t\tfor index, nested := range item { result[index] = _purecdk8sFlattenAdditionalValues(nested) }")
	fmt.Fprintln(out, "\t\treturn result")
	fmt.Fprintln(out, "\tdefault:")
	fmt.Fprintln(out, "\t\treturn item")
	fmt.Fprintln(out, "\t}")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)

	fmt.Fprintf(out, "func _purecdk8sSetEmbedded%s(target %s, implementation %s) bool {\n", typeName, typeName, typeName)
	fmt.Fprintln(out, "\tvalue := reflect.ValueOf(target)")
	fmt.Fprintln(out, "\tif value.Kind() != reflect.Ptr || value.IsNil() { return false }")
	fmt.Fprintln(out, "\tvalue = value.Elem()")
	fmt.Fprintln(out, "\tif value.Kind() != reflect.Struct { return false }")
	fmt.Fprintln(out, "\timplementationValue := reflect.ValueOf(implementation)")
	fmt.Fprintln(out, "\ttype_ := value.Type()")
	fmt.Fprintln(out, "\tfor index := 0; index < value.NumField(); index++ {")
	fmt.Fprintln(out, "\t\tfield := value.Field(index)")
	fmt.Fprintln(out, "\t\tfieldInfo := type_.Field(index)")
	fmt.Fprintln(out, "\t\tif fieldInfo.Anonymous && field.CanSet() && implementationValue.Type().AssignableTo(field.Type()) {")
	fmt.Fprintln(out, "\t\t\tfield.Set(implementationValue)")
	fmt.Fprintln(out, "\t\t\treturn true")
	fmt.Fprintln(out, "\t\t}")
	fmt.Fprintln(out, "\t}")
	fmt.Fprintln(out, "\treturn false")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)
}
