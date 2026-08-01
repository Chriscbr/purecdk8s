package importer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"path"
	"regexp"
	"sort"
	"strings"
)

const (
	defaultCoreImport          = "github.com/purecdk8s/purecdk8s/cdk8s/v2"
	defaultConstructsImport    = "github.com/purecdk8s/purecdk8s/constructs/v10"
	defaultSerializationImport = "github.com/purecdk8s/purecdk8s/serialization"
)

// GenerateOptions controls source generation independently of source
// acquisition. PackageName is required.
type GenerateOptions struct {
	PackageName            string
	PackagePrefix          string
	ClassNamePrefix        string
	DisableClassNamePrefix bool
	Excludes               []string
	CoreImport             string
	ConstructsImport       string
	SerializationImport    string
	excludeRegexps         []*regexp.Regexp
}

// Generation describes a generated Go package.
type Generation struct {
	PackageName string
	Group       string
	Code        []byte
	Types       []string
	Resources   []string
}

type resourceDefinition struct {
	FQN       string
	Group     string
	Version   string
	Kind      string
	Prefix    string
	Suffix    string
	Custom    bool
	Schema    *schema
	TypeName  string
	PropsName string
}

type generatedType struct {
	Name   string
	FQN    string
	Schema *schema
}

type goTypeCategory int

const (
	categoryAny goTypeCategory = iota
	categoryPrimitive
	categoryStruct
	categoryUnion
	categoryEnum
	categoryArray
	categoryMap
)

type goType struct {
	Name     string
	Category goTypeCategory
}

// GenerateKubernetes generates an upstream-shaped Go package from a cdk8s
// Kubernetes _definitions.json document.
func GenerateKubernetes(data []byte, options GenerateOptions) (*Generation, error) {
	if options.excludeRegexps == nil {
		excludes, err := compileExcludePatterns(options.Excludes)
		if err != nil {
			return nil, err
		}
		options.excludeRegexps = excludes
	}
	defs, err := decodeDefinitions(data)
	if err != nil {
		return nil, err
	}
	if options.PackageName == "" {
		options.PackageName = "k8s"
	}
	if options.ClassNamePrefix == "" && !options.DisableClassNamePrefix {
		options.ClassNamePrefix = "Kube"
	}
	resources := make([]resourceDefinition, 0)
	for _, fqn := range sortedKeys(defs) {
		item := defs[fqn]
		if item == nil || len(item.GroupVersionKind) == 0 || item.Properties["metadata"] == nil {
			continue
		}
		gvk := item.GroupVersionKind[0]
		if gvk.Version == "" || gvk.Kind == "" {
			continue
		}
		typeName := resourceTypeName(options.ClassNamePrefix, gvk.Kind, gvk.Version, false, "")
		resources = append(resources, resourceDefinition{
			FQN:       fqn,
			Group:     gvk.Group,
			Version:   gvk.Version,
			Kind:      gvk.Kind,
			Prefix:    options.ClassNamePrefix,
			Schema:    item,
			TypeName:  typeName,
			PropsName: normalizeTypeName(typeName + "Props"),
		})
	}
	return generatePackage(defs, resources, options, "")
}

func generatePackage(defs map[string]*schema, resources []resourceDefinition, options GenerateOptions, group string) (*Generation, error) {
	if options.PackageName == "" {
		return nil, fmt.Errorf("generate package: PackageName is required")
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
	generator := &packageGenerator{
		options:        options,
		definitions:    defs,
		resourceByFQN:  make(map[string]resourceDefinition),
		typeByFQN:      make(map[string]string),
		typeSchemas:    make(map[string]*schema),
		typeFQNs:       make(map[string]string),
		typeIndexes:    make(map[string]int),
		renderedTypes:  make(map[string]bool),
		excludeRegexps: append([]*regexp.Regexp(nil), options.excludeRegexps...),
	}

	// json2jsii de-duplicates normalized names. Sorting gives deterministic
	// behavior and matches the order of the published Kubernetes definitions.
	resourceIndexes := make(map[string]int)
	for _, resource := range resources {
		generator.resourceByFQN[resource.FQN] = resource
		if index, exists := resourceIndexes[resource.TypeName]; exists {
			// All constructs are registered before json2jsii renders them, so
			// duplicate resource names are overwritten by the later schema.
			// This is observable for core/v1 Event vs events.k8s.io/v1 Event.
			generator.resources[index] = resource
			continue
		}
		resourceIndexes[resource.TypeName] = len(generator.resources)
		generator.resources = append(generator.resources, resource)
	}
	for _, fqn := range sortedKeys(defs) {
		name := typeNameForDefinition(fqn)
		generator.typeByFQN[fqn] = name
	}
	for _, resource := range generator.resources {
		props := resourcePropsSchema(resource)
		generator.addType(resource.PropsName, resource.FQN, props)
	}

	code, err := generator.render()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(generator.resources))
	for _, resource := range generator.resources {
		names = append(names, resource.TypeName)
	}
	types := make([]string, 0, len(generator.types))
	for _, generatedType := range generator.types {
		types = append(types, generatedType.Name)
	}
	return &Generation{
		PackageName: options.PackageName,
		Group:       group,
		Code:        code,
		Types:       types,
		Resources:   names,
	}, nil
}

func resourcePropsSchema(resource resourceDefinition) *schema {
	result := cloneSchema(resource.Schema)
	delete(result.Properties, "apiVersion")
	delete(result.Properties, "kind")
	delete(result.Properties, "status")
	required := result.Required[:0]
	for _, key := range result.Required {
		if key != "apiVersion" && key != "kind" && key != "status" {
			required = append(required, key)
		}
	}
	result.Required = required
	result.GroupVersionKind = nil
	if resource.Custom {
		result.Properties["metadata"] = &schema{Ref: "#/definitions/ApiObjectMetadata"}
	}
	return result
}

type packageGenerator struct {
	options        GenerateOptions
	definitions    map[string]*schema
	resources      []resourceDefinition
	resourceByFQN  map[string]resourceDefinition
	typeByFQN      map[string]string
	typeSchemas    map[string]*schema
	typeFQNs       map[string]string
	typeIndexes    map[string]int
	renderedTypes  map[string]bool
	excludeRegexps []*regexp.Regexp
	types          []generatedType
}

func (g *packageGenerator) isRenderable(item *schema) bool {
	if item == nil {
		return false
	}
	return len(item.Properties) > 0 || len(item.Enum) > 0 || g.scalarUnionOptions(item) != nil
}

func (g *packageGenerator) addType(name, fqn string, item *schema) {
	if name == "" || !g.isRenderable(item) {
		return
	}
	if g.renderedTypes[name] {
		return
	}
	generated := generatedType{Name: name, FQN: fqn, Schema: item}
	if index, exists := g.typeIndexes[name]; exists {
		// json2jsii allows a later discovery to replace a type that is queued
		// but not rendered yet. Kubernetes has a few normalized-name
		// collisions (EndpointPort and ServiceReference) that rely on this.
		g.typeSchemas[name] = generated.Schema
		g.typeFQNs[name] = fqn
		g.types[index] = generated
		return
	}
	g.typeSchemas[name] = item
	g.typeFQNs[name] = fqn
	g.typeIndexes[name] = len(g.types)
	g.types = append(g.types, generated)
}

func (g *packageGenerator) render() ([]byte, error) {
	var out bytes.Buffer
	usesTime := g.usesDateTime()
	fmt.Fprintln(&out, "// Code generated by purecdk8s. DO NOT EDIT.")
	fmt.Fprintf(&out, "package %s\n\n", g.options.PackageName)
	fmt.Fprintln(&out, "import (")
	fmt.Fprintln(&out, "\t\"encoding/json\"")
	fmt.Fprintln(&out, "\t\"fmt\"")
	fmt.Fprintln(&out, "\t\"reflect\"")
	fmt.Fprintln(&out, "\t\"strings\"")
	if usesTime {
		fmt.Fprintln(&out, "\t\"time\"")
	}
	fmt.Fprintf(&out, "\tcdk8s %q\n", g.options.CoreImport)
	fmt.Fprintf(&out, "\tconstructs %q\n", g.options.ConstructsImport)
	fmt.Fprintf(&out, "\tpurecdk8sserialization %q\n", g.options.SerializationImport)
	fmt.Fprintln(&out, ")")
	fmt.Fprintln(&out)
	fmt.Fprintln(&out, "var _ = purecdk8sserialization.EnumWireValue")
	fmt.Fprintln(&out)
	if usesTime {
		fmt.Fprintln(&out, "var _ = time.Time{}")
		fmt.Fprintln(&out)
	}
	g.renderScalarUnionHelper(&out)

	for index := 0; index < len(g.types); index++ {
		item := g.types[index]
		if err := g.renderType(&out, item); err != nil {
			return nil, err
		}
		g.renderedTypes[item.Name] = true
	}
	for _, resource := range g.resources {
		g.renderResource(&out, resource)
	}
	g.renderHelpers(&out)

	formatted, err := format.Source(out.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated %s package: %w\n%s", g.options.PackageName, err, numberedSource(out.String()))
	}
	return formatted, nil
}

func (g *packageGenerator) renderType(out *bytes.Buffer, item generatedType) error {
	s := item.Schema
	switch {
	case len(s.OneOf) > 0 || len(s.AnyOf) > 0:
		if g.scalarUnionOptions(s) != nil {
			g.renderUnion(out, item)
			return nil
		}
	case len(s.Enum) > 0:
		g.renderEnum(out, item)
		return nil
	case len(s.Properties) > 0:
		g.renderStruct(out, item)
		return nil
	case s.AdditionalProperties != nil:
		additional := additionalSchema(s.AdditionalProperties)
		element := g.typeForSchema(additional, item, item.FQN+"Value", false)
		fmt.Fprintf(out, "type %s map[string]%s\n\n", item.Name, element.Name)
		return nil
	}
	switch schemaType(s) {
	case "string":
		fmt.Fprintf(out, "type %s string\n\n", item.Name)
	case "number", "integer":
		fmt.Fprintf(out, "type %s float64\n\n", item.Name)
	case "boolean":
		fmt.Fprintf(out, "type %s bool\n\n", item.Name)
	case "array":
		element := g.typeForSchema(s.Items, item, item.FQN+"Item", false)
		fmt.Fprintf(out, "type %s []%s\n\n", item.Name, element.Name)
	default:
		// References to empty definitions are represented as interface{} and no
		// named declaration is emitted, matching json2jsii.
	}
	return nil
}

func (g *packageGenerator) renderStruct(out *bytes.Buffer, item generatedType) {
	renderDescription(out, item.Name, item.Schema.Description)
	fmt.Fprintf(out, "type %s struct {\n", item.Name)
	required := requiredSet(item.Schema)
	seen := make(map[string]bool)
	keys := sortedKeys(item.Schema.Properties)
	sort.SliceStable(keys, func(left, right int) bool {
		leftField, _ := goFieldName(keys[left])
		rightField, _ := goFieldName(keys[right])
		leftFolded := strings.ToLower(leftField)
		rightFolded := strings.ToLower(rightField)
		if leftFolded == rightFolded {
			return leftField < rightField
		}
		return leftFolded < rightFolded
	})
	orderedKeys := make([]string, 0, len(keys))
	// jsii emits required properties first and optional properties second,
	// sorting alphabetically within each group. Field order is observable to
	// Go callers that use unkeyed struct literals, so preserve it exactly.
	for _, requiredGroup := range []bool{true, false} {
		for _, schemaName := range keys {
			if required[schemaName] == requiredGroup {
				orderedKeys = append(orderedKeys, schemaName)
			}
		}
	}
	for _, schemaName := range orderedKeys {
		fieldName, normalized := goFieldName(schemaName)
		if seen[fieldName] {
			continue
		}
		seen[fieldName] = true
		property := item.Schema.Properties[schemaName]
		fieldType := g.typeForSchema(property, item, item.FQN+"."+normalized, true)
		if property.Ref != "" && fieldType.Category == categoryStruct && strings.TrimPrefix(fieldType.Name, "*") == item.Name {
			fieldType.Name = "*" + fieldType.Name
		}
		requirement := "optional"
		if required[schemaName] {
			requirement = "required"
		}
		renderDescription(out, "", property.Description)
		fmt.Fprintf(out, "\t%s %s `field:%q json:%q yaml:%q", fieldName, fieldType.Name, requirement, normalized, normalized)
		if normalized != schemaName {
			fmt.Fprintf(out, " k8s:%q", schemaName)
		}
		fmt.Fprintln(out, "`")
	}
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)
}

func (g *packageGenerator) renderEnum(out *bytes.Buffer, item generatedType) {
	renderDescription(out, item.Name, item.Schema.Description)
	// JSII Go enums are symbolic strings: the public value is the sanitized
	// member name, while serialization uses the original schema value.
	fmt.Fprintf(out, "type %s string\n\n", item.Name)
	fmt.Fprintln(out, "const (")
	seen := make(map[string]bool)
	for _, value := range item.Schema.Enum {
		member := enumMember(value)
		key := strings.ToLower(fmt.Sprint(value))
		if seen[key] {
			continue
		}
		seen[key] = true
		fmt.Fprintf(out, "\t%s_%s %s = %q\n", item.Name, member, item.Name, member)
	}
	fmt.Fprintln(out, ")")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "func init() {")
	fmt.Fprintf(out, "\tpurecdk8sserialization.RegisterEnumWireValues(map[%s]interface{}{\n", item.Name)
	seen = make(map[string]bool)
	for _, value := range item.Schema.Enum {
		member := enumMember(value)
		key := strings.ToLower(fmt.Sprint(value))
		if seen[key] {
			continue
		}
		seen[key] = true
		fmt.Fprintf(out, "\t\t%s_%s: %s,\n", item.Name, member, literal(value))
	}
	fmt.Fprintln(out, "\t})")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)
}

func (g *packageGenerator) scalarUnionOptions(s *schema) []*schema {
	options := s.OneOf
	if len(options) == 0 {
		options = s.AnyOf
	}
	if len(options) == 0 {
		return nil
	}
	for _, option := range options {
		if g.isExcludedReference(option.Ref) {
			return nil
		}
		resolved := option
		if option.Ref != "" {
			resolved = g.definitions[refName(option.Ref)]
		}
		switch schemaType(resolved) {
		case "string", "number", "integer", "boolean":
		default:
			return nil
		}
	}
	return options
}

func (g *packageGenerator) renderUnion(out *bytes.Buffer, item generatedType) {
	renderDescription(out, item.Name, item.Schema.Description)
	fmt.Fprintf(out, "type %s interface {\n\tValue() interface{}\n}\n\n", item.Name)
	seen := make(map[string]bool)
	for _, option := range g.scalarUnionOptions(item.Schema) {
		resolved := option
		if option.Ref != "" {
			resolved = g.definitions[refName(option.Ref)]
		}
		kind := schemaType(resolved)
		method := ""
		goType := ""
		switch kind {
		case "string":
			method, goType = "String", "string"
		case "number", "integer":
			method, goType = "Number", "float64"
		case "boolean":
			method, goType = "Boolean", "bool"
		}
		if seen[method] {
			continue
		}
		seen[method] = true
		fmt.Fprintf(out, "func %s_From%s(value *%s) %s {\n", item.Name, method, goType, item.Name)
		fmt.Fprintf(out, "\tif value == nil { panic(%q) }\n", item.Name+"_From"+method+": value is required")
		fmt.Fprintln(out, "\treturn &_purecdk8sScalarUnion{value: *value}")
		fmt.Fprintln(out, "}")
		fmt.Fprintln(out)
	}
}

func (g *packageGenerator) typeForSchema(s *schema, owner generatedType, propertyFQN string, property bool) goType {
	if s == nil {
		return goType{Name: "interface{}", Category: categoryAny}
	}
	if s.Ref != "" {
		if g.isExcludedReference(s.Ref) {
			return goType{Name: "interface{}", Category: categoryAny}
		}
		fqn := refName(s.Ref)
		if fqn == "ApiObjectMetadata" {
			return goType{Name: "*cdk8s.ApiObjectMetadata", Category: categoryStruct}
		}
		if resource, ok := g.resourceByFQN[fqn]; ok {
			return goType{Name: "*" + resource.PropsName, Category: categoryStruct}
		}
		target := g.definitions[fqn]
		if target == nil {
			return goType{Name: "interface{}", Category: categoryAny}
		}
		category := g.category(target)
		if category != categoryStruct && category != categoryUnion && category != categoryEnum {
			return g.typeForSchema(target, owner, propertyFQN, property)
		}
		name := g.typeByFQN[fqn]
		if name == "" {
			name = typeNameForDefinition(fqn)
			g.typeByFQN[fqn] = name
		}
		g.addType(name, fqn, target)
		if category == categoryStruct {
			return goType{Name: "*" + name, Category: category}
		}
		return goType{Name: name, Category: category}
	}
	if options := g.scalarUnionOptions(s); options != nil {
		name := normalizeTypeName(defaultInlineTypeName(propertyFQN))
		g.addType(name, propertyFQN, s)
		return goType{Name: name, Category: categoryUnion}
	}
	if len(s.Enum) > 0 {
		name := normalizeTypeName(defaultInlineTypeName(propertyFQN))
		g.addType(name, propertyFQN, s)
		return goType{Name: name, Category: categoryEnum}
	}
	if len(s.Properties) > 0 {
		name := normalizeTypeName(defaultInlineTypeName(propertyFQN))
		g.addType(name, propertyFQN, s)
		return goType{Name: "*" + name, Category: categoryStruct}
	}
	if s.Format == "date-time" {
		return goType{Name: "*time.Time", Category: categoryPrimitive}
	}
	switch schemaType(s) {
	case "string":
		return goType{Name: "*string", Category: categoryPrimitive}
	case "number", "integer":
		return goType{Name: "*float64", Category: categoryPrimitive}
	case "boolean":
		return goType{Name: "*bool", Category: categoryPrimitive}
	case "array":
		// json2jsii derives the Go name for an inline array element from the
		// array property itself. Appending "Item" produces types such as
		// ExternalSecretSpecDataFromItem, whereas upstream exposes
		// ExternalSecretSpecDataFrom.
		element := g.typeForSchema(s.Items, owner, propertyFQN, false)
		return goType{Name: "*[]" + element.Name, Category: categoryArray}
	case "object":
		if s.AdditionalProperties != nil {
			if flag, ok := s.AdditionalProperties.(bool); ok && !flag {
				return goType{Name: "interface{}", Category: categoryAny}
			}
			element := g.typeForSchema(additionalSchema(s.AdditionalProperties), owner, propertyFQN+"Value", false)
			return goType{Name: "*map[string]" + element.Name, Category: categoryMap}
		}
		return goType{Name: "interface{}", Category: categoryAny}
	}
	if s.AdditionalProperties != nil {
		element := g.typeForSchema(additionalSchema(s.AdditionalProperties), owner, propertyFQN+"Value", false)
		return goType{Name: "*map[string]" + element.Name, Category: categoryMap}
	}
	_ = property
	return goType{Name: "interface{}", Category: categoryAny}
}

func (g *packageGenerator) isExcludedReference(reference string) bool {
	if reference == "" {
		return false
	}
	for _, pattern := range g.excludeRegexps {
		if pattern.MatchString(reference) {
			return true
		}
	}
	return false
}

func compileExcludePatterns(patterns []string) ([]*regexp.Regexp, error) {
	result := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid Kubernetes exclude regular expression %q: %w", pattern, err)
		}
		result = append(result, compiled)
	}
	return result, nil
}

func (g *packageGenerator) category(s *schema) goTypeCategory {
	switch {
	case s == nil:
		return categoryAny
	case g.scalarUnionOptions(s) != nil:
		return categoryUnion
	case len(s.Enum) > 0:
		return categoryEnum
	case len(s.Properties) > 0:
		return categoryStruct
	case schemaType(s) == "array":
		return categoryArray
	case s.AdditionalProperties != nil:
		return categoryMap
	case schemaType(s) == "string" || schemaType(s) == "number" || schemaType(s) == "integer" || schemaType(s) == "boolean":
		return categoryPrimitive
	default:
		return categoryAny
	}
}

func additionalSchema(value interface{}) *schema {
	switch item := value.(type) {
	case *schema:
		return item
	case map[string]interface{}:
		result := &schema{}
		if data, err := json.Marshal(item); err == nil {
			if err := json.Unmarshal(data, result); err == nil {
				return result
			}
		}
		return result
	case nil:
		return &schema{}
	default:
		return &schema{}
	}
}

func (g *packageGenerator) usesDateTime() bool {
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
		for _, option := range append(append(append([]*schema{}, item.OneOf...), item.AnyOf...), item.AllOf...) {
			if visit(option) {
				return true
			}
		}
		if item.AdditionalProperties != nil && visit(additionalSchema(item.AdditionalProperties)) {
			return true
		}
		return false
	}
	for _, definition := range g.definitions {
		if visit(definition) {
			return true
		}
	}
	return false
}

func defaultInlineTypeName(fqn string) string {
	parts := strings.FieldsFunc(fqn, func(r rune) bool {
		return r == '.' || r == '/' || r == '#' || r == '$'
	})
	var out strings.Builder
	for _, part := range parts {
		if part == "definitions" || part == "defs" {
			continue
		}
		out.WriteString(pascalCase(part))
	}
	return out.String()
}

func (g *packageGenerator) renderScalarUnionHelper(out *bytes.Buffer) {
	fmt.Fprintln(out, "type _purecdk8sScalarUnion struct { value interface{} }")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "func (value *_purecdk8sScalarUnion) Value() interface{} {")
	fmt.Fprintln(out, "\tif value == nil { return nil }")
	fmt.Fprintln(out, "\treturn value.value")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "func (value *_purecdk8sScalarUnion) PureCDK8sScalarUnion() {}")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "func (value *_purecdk8sScalarUnion) MarshalJSON() ([]byte, error) {")
	fmt.Fprintln(out, "\treturn json.Marshal(value.Value())")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)
}

func (g *packageGenerator) renderResource(out *bytes.Buffer, resource resourceDefinition) {
	renderDescription(out, resource.TypeName, resource.Schema.Description)
	fmt.Fprintf(out, "type %s interface {\n\tcdk8s.ApiObject\n}\n\n", resource.TypeName)
	impl := lowerFirst(resource.TypeName)
	fmt.Fprintf(out, "type %s struct {\n\tcdk8s.ApiObject\n}\n\n", impl)
	fmt.Fprintf(out, "func New%s(scope constructs.Construct, id *string, props *%s) %s {\n", resource.TypeName, resource.PropsName, resource.TypeName)
	g.renderRequiredPropsCheck(out, resource, "props")
	fmt.Fprintf(out, "\t_purecdk8sValidateRequired(props, %q)\n", resource.PropsName)
	fmt.Fprintf(out, "\tobject := cdk8s.NewApiObjectWithManifest(scope, id, _%sApiObjectProps(), props)\n", impl)
	fmt.Fprintf(out, "\treturn &%s{ApiObject: object}\n", impl)
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)
	fmt.Fprintf(out, "func New%s_Override(value %s, scope constructs.Construct, id *string, props *%s) {\n", resource.TypeName, resource.TypeName, resource.PropsName)
	g.renderRequiredPropsCheck(out, resource, "props")
	fmt.Fprintf(out, "\t_purecdk8sValidateRequired(props, %q)\n", resource.PropsName)
	fmt.Fprintf(out, "\tcdk8s.NewApiObjectWithManifest_Override(value, scope, id, _%sApiObjectProps(), props)\n", impl)
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)
	fmt.Fprintf(out, "func %s_IsApiObject(value interface{}) *bool {\n", resource.TypeName)
	fmt.Fprintln(out, "\t_, ok := value.(cdk8s.ApiObject)")
	fmt.Fprintln(out, "\treturn &ok")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)
	fmt.Fprintf(out, "func %s_IsConstruct(value interface{}) *bool {\n", resource.TypeName)
	fmt.Fprintln(out, "\t_, ok := value.(constructs.IConstruct)")
	fmt.Fprintln(out, "\treturn &ok")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)
	fmt.Fprintf(out, "func %s_Manifest(props *%s) interface{} {\n", resource.TypeName, resource.PropsName)
	g.renderRequiredPropsCheck(out, resource, "props")
	fmt.Fprintf(out, "\t_purecdk8sValidateRequired(props, %q)\n", resource.PropsName)
	fmt.Fprintf(out, "\treturn cdk8s.ApiObjectManifest(_%sApiObjectProps(), props)\n", impl)
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)
	fmt.Fprintf(out, "func %s_Of(value constructs.IConstruct) cdk8s.ApiObject {\n", resource.TypeName)
	fmt.Fprintln(out, "\treturn _purecdk8sApiObjectOf(value)")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)
	fmt.Fprintf(out, "func %s_GVK() *cdk8s.GroupVersionKind {\n", resource.TypeName)
	apiVersion := resource.Version
	if resource.Group != "" {
		apiVersion = resource.Group + "/" + resource.Version
	}
	fmt.Fprintf(out, "\tapiVersion, kind := %q, %q\n", apiVersion, resource.Kind)
	fmt.Fprintln(out, "\treturn &cdk8s.GroupVersionKind{ApiVersion: &apiVersion, Kind: &kind}")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)
	fmt.Fprintf(out, "func _%sApiObjectProps() *cdk8s.ApiObjectProps {\n", impl)
	fmt.Fprintf(out, "\tgvk := %s_GVK()\n", resource.TypeName)
	fmt.Fprintln(out, "\treturn &cdk8s.ApiObjectProps{ApiVersion: gvk.ApiVersion, Kind: gvk.Kind}")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)
}

func (g *packageGenerator) renderRequiredPropsCheck(out *bytes.Buffer, resource resourceDefinition, variable string) {
	if len(resourcePropsSchema(resource).Required) == 0 {
		return
	}
	fmt.Fprintf(out, "\tif %s == nil { panic(%q) }\n", variable, "New"+resource.TypeName+": props is required")
}

func (g *packageGenerator) renderHelpers(out *bytes.Buffer) {
	fmt.Fprintln(out, "type _purecdk8sValidationVisit struct {")
	fmt.Fprintln(out, "\ttype_ reflect.Type")
	fmt.Fprintln(out, "\tpointer uintptr")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "func _purecdk8sValidateRequired(value interface{}, path string) {")
	fmt.Fprintln(out, "\t_purecdk8sValidateValue(reflect.ValueOf(value), path, make(map[_purecdk8sValidationVisit]bool))")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "func _purecdk8sValidateValue(value reflect.Value, path string, visited map[_purecdk8sValidationVisit]bool) {")
	fmt.Fprintln(out, "\tif !value.IsValid() { return }")
	fmt.Fprintln(out, "\tfor value.Kind() == reflect.Interface {")
	fmt.Fprintln(out, "\t\tif value.IsNil() { return }")
	fmt.Fprintln(out, "\t\tif value.CanInterface() {")
	fmt.Fprintln(out, "\t\t\tif _, ok := value.Interface().(interface{ PureCDK8sScalarUnion() }); ok { return }")
	fmt.Fprintln(out, "\t\t}")
	fmt.Fprintln(out, "\t\tvalue = value.Elem()")
	fmt.Fprintln(out, "\t}")
	fmt.Fprintln(out, "\tswitch value.Kind() {")
	fmt.Fprintln(out, "\tcase reflect.Ptr:")
	fmt.Fprintln(out, "\t\tif value.IsNil() { return }")
	fmt.Fprintln(out, "\t\tvisit := _purecdk8sValidationVisit{type_: value.Type(), pointer: value.Pointer()}")
	fmt.Fprintln(out, "\t\tif visited[visit] { return }")
	fmt.Fprintln(out, "\t\tvisited[visit] = true")
	fmt.Fprintln(out, "\t\t_purecdk8sValidateValue(value.Elem(), path, visited)")
	fmt.Fprintln(out, "\tcase reflect.Struct:")
	fmt.Fprintln(out, "\t\ttype_ := value.Type()")
	fmt.Fprintln(out, "\t\tfor index := 0; index < value.NumField(); index++ {")
	fmt.Fprintln(out, "\t\t\tfieldInfo := type_.Field(index)")
	fmt.Fprintln(out, "\t\t\trequirement := fieldInfo.Tag.Get(\"field\")")
	fmt.Fprintln(out, "\t\t\tif requirement == \"\" || fieldInfo.PkgPath != \"\" { continue }")
	fmt.Fprintln(out, "\t\t\tfield := value.Field(index)")
	fmt.Fprintln(out, "\t\t\tname := fieldInfo.Tag.Get(\"k8s\")")
	fmt.Fprintln(out, "\t\t\tif name == \"\" { name = fieldInfo.Tag.Get(\"json\") }")
	fmt.Fprintln(out, "\t\t\tif comma := strings.IndexByte(name, ','); comma >= 0 { name = name[:comma] }")
	fmt.Fprintln(out, "\t\t\tif name == \"\" { name = fieldInfo.Name }")
	fmt.Fprintln(out, "\t\t\tfieldPath := path + \".\" + name")
	fmt.Fprintln(out, "\t\t\tif requirement == \"required\" && _purecdk8sRequiredMissing(field) {")
	fmt.Fprintf(out, "%s", "\t\t\t\tpanic(fmt.Sprintf(\"%s is required\", fieldPath))\n")
	fmt.Fprintln(out, "\t\t\t}")
	fmt.Fprintln(out, "\t\t\t_purecdk8sValidateValue(field, fieldPath, visited)")
	fmt.Fprintln(out, "\t\t}")
	fmt.Fprintln(out, "\tcase reflect.Slice, reflect.Array:")
	fmt.Fprintln(out, "\t\tif value.Kind() == reflect.Slice {")
	fmt.Fprintln(out, "\t\t\tif value.IsNil() { return }")
	fmt.Fprintln(out, "\t\t\tvisit := _purecdk8sValidationVisit{type_: value.Type(), pointer: value.Pointer()}")
	fmt.Fprintln(out, "\t\t\tif visited[visit] { return }")
	fmt.Fprintln(out, "\t\t\tvisited[visit] = true")
	fmt.Fprintln(out, "\t\t}")
	fmt.Fprintln(out, "\t\tfor index := 0; index < value.Len(); index++ {")
	fmt.Fprintf(out, "%s", "\t\t\t_purecdk8sValidateValue(value.Index(index), fmt.Sprintf(\"%s[%d]\", path, index), visited)\n")
	fmt.Fprintln(out, "\t\t}")
	fmt.Fprintln(out, "\tcase reflect.Map:")
	fmt.Fprintln(out, "\t\tif value.IsNil() { return }")
	fmt.Fprintln(out, "\t\tvisit := _purecdk8sValidationVisit{type_: value.Type(), pointer: value.Pointer()}")
	fmt.Fprintln(out, "\t\tif visited[visit] { return }")
	fmt.Fprintln(out, "\t\tvisited[visit] = true")
	fmt.Fprintln(out, "\t\titerator := value.MapRange()")
	fmt.Fprintln(out, "\t\tfor iterator.Next() {")
	fmt.Fprintf(out, "%s", "\t\t\t_purecdk8sValidateValue(iterator.Value(), fmt.Sprintf(\"%s[%v]\", path, iterator.Key().Interface()), visited)\n")
	fmt.Fprintln(out, "\t\t}")
	fmt.Fprintln(out, "\t}")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "func _purecdk8sRequiredMissing(value reflect.Value) bool {")
	fmt.Fprintln(out, "\tif !value.IsValid() { return true }")
	fmt.Fprintln(out, "\tfor value.Kind() == reflect.Interface {")
	fmt.Fprintln(out, "\t\tif value.IsNil() { return true }")
	fmt.Fprintln(out, "\t\tvalue = value.Elem()")
	fmt.Fprintln(out, "\t}")
	fmt.Fprintln(out, "\tswitch value.Kind() {")
	fmt.Fprintln(out, "\tcase reflect.Ptr, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan:")
	fmt.Fprintln(out, "\t\treturn value.IsNil()")
	fmt.Fprintln(out, "\tdefault:")
	fmt.Fprintln(out, "\t\treturn value.IsZero()")
	fmt.Fprintln(out, "\t}")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "func _purecdk8sApiObjectOf(value constructs.IConstruct) cdk8s.ApiObject {")
	fmt.Fprintln(out, "\tif value == nil { panic(\"resource scope is nil\") }")
	fmt.Fprintln(out, "\tif object, ok := value.(cdk8s.ApiObject); ok { return object }")
	fmt.Fprintln(out, "\tfor _, id := range []string{\"Resource\", \"Default\"} {")
	fmt.Fprintln(out, "\t\tchild := value.Node().TryFindChild(&id)")
	fmt.Fprintln(out, "\t\tif child == nil { continue }")
	fmt.Fprintln(out, "\t\tif object, ok := child.(cdk8s.ApiObject); ok { return object }")
	fmt.Fprintf(out, "%s", "\t\tpanic(fmt.Sprintf(\"construct child %q is not an ApiObject\", id))\n")
	fmt.Fprintln(out, "\t}")
	fmt.Fprintln(out, "\tpanic(\"construct does not contain an ApiObject child named Resource or Default\")")
	fmt.Fprintln(out, "}")
	fmt.Fprintln(out)
}

func renderDescription(out *bytes.Buffer, name, description string) {
	description = strings.TrimSpace(description)
	if description == "" {
		return
	}
	lines := strings.Split(description, "\n")
	for index, line := range lines {
		line = strings.TrimSpace(strings.ReplaceAll(line, "*/", "_/"))
		if index == 0 && name != "" {
			fmt.Fprintf(out, "// %s %s\n", name, line)
		} else if line == "" {
			fmt.Fprintln(out, "//")
		} else {
			fmt.Fprintf(out, "// %s\n", line)
		}
	}
}

func numberedSource(source string) string {
	lines := strings.Split(source, "\n")
	for index, line := range lines {
		lines[index] = fmt.Sprintf("%4d %s", index+1, line)
	}
	return strings.Join(lines, "\n")
}

func sanitizePackageName(value string) string {
	value = path.Base(strings.TrimSpace(value))
	var out strings.Builder
	for _, r := range value {
		if unicodeLetter(r) || (r >= '0' && r <= '9') {
			out.WriteRune(unicodeLower(r))
		}
	}
	result := out.String()
	if result == "" {
		return "imports"
	}
	if result[0] >= '0' && result[0] <= '9' {
		return "imports_" + result
	}
	return result
}

func sanitizePackagePrefix(value string) string {
	value = path.Base(strings.TrimSpace(value))
	var out strings.Builder
	for _, r := range value {
		if unicodeLetter(r) || r == '_' || (r >= '0' && r <= '9') {
			out.WriteRune(unicodeLower(r))
		}
	}
	result := strings.Trim(out.String(), "_")
	if result == "" {
		return "imports"
	}
	if result[0] >= '0' && result[0] <= '9' {
		return "imports_" + result
	}
	return result
}

func unicodeLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func unicodeLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}
