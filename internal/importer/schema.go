package importer

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// schema is the subset of JSON Schema/OpenAPI used by the cdk8s Kubernetes
// importer. Unknown keywords are intentionally ignored.
type schema struct {
	Ref                  string             `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Description          string             `json:"description,omitempty" yaml:"description,omitempty"`
	Type                 interface{}        `json:"type,omitempty" yaml:"type,omitempty"`
	Properties           map[string]*schema `json:"properties,omitempty" yaml:"properties,omitempty"`
	Definitions          map[string]*schema `json:"definitions,omitempty" yaml:"definitions,omitempty"`
	Defs                 map[string]*schema `json:"$defs,omitempty" yaml:"$defs,omitempty"`
	Required             []string           `json:"required,omitempty" yaml:"required,omitempty"`
	Items                *schema            `json:"items,omitempty" yaml:"items,omitempty"`
	AdditionalProperties interface{}        `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	OneOf                []*schema          `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	AnyOf                []*schema          `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	AllOf                []*schema          `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	Enum                 []interface{}      `json:"enum,omitempty" yaml:"enum,omitempty"`
	Format               string             `json:"format,omitempty" yaml:"format,omitempty"`
	Pattern              string             `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	Nullable             bool               `json:"nullable,omitempty" yaml:"nullable,omitempty"`
	PreserveUnknown      bool               `json:"x-kubernetes-preserve-unknown-fields,omitempty" yaml:"x-kubernetes-preserve-unknown-fields,omitempty"`
	GroupVersionKind     []groupVersionKind `json:"x-kubernetes-group-version-kind,omitempty" yaml:"x-kubernetes-group-version-kind,omitempty"`
}

type definitionsDocument struct {
	Definitions map[string]*schema `json:"definitions"`
	Defs        map[string]*schema `json:"$defs"`
}

type groupVersionKind struct {
	Group   string `json:"group" yaml:"group"`
	Version string `json:"version" yaml:"version"`
	Kind    string `json:"kind" yaml:"kind"`
}

func decodeDefinitions(data []byte) (map[string]*schema, error) {
	var doc definitionsDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("decode Kubernetes definitions: %w", err)
	}
	defs := doc.Definitions
	if len(defs) == 0 {
		defs = doc.Defs
	}
	if len(defs) == 0 {
		return nil, fmt.Errorf("decode Kubernetes definitions: document has no definitions")
	}
	return defs, nil
}

func schemaType(s *schema) string {
	if s == nil {
		return ""
	}
	switch value := s.Type.(type) {
	case string:
		return value
	case []interface{}:
		for _, item := range value {
			if item, ok := item.(string); ok && item != "null" {
				return item
			}
		}
	}
	return ""
}

func requiredSet(s *schema) map[string]bool {
	result := make(map[string]bool, len(s.Required))
	for _, key := range s.Required {
		result[key] = true
	}
	return result
}

func cloneSchema(s *schema) *schema {
	if s == nil {
		return &schema{}
	}
	copy := *s
	copy.Properties = make(map[string]*schema, len(s.Properties))
	for key, value := range s.Properties {
		copy.Properties[key] = value
	}
	copy.Required = append([]string(nil), s.Required...)
	copy.GroupVersionKind = append([]groupVersionKind(nil), s.GroupVersionKind...)
	return &copy
}

func refName(ref string) string {
	for _, prefix := range []string{"#/definitions/", "#/$defs/"} {
		if strings.HasPrefix(ref, prefix) {
			return strings.TrimPrefix(ref, prefix)
		}
	}
	return ""
}

var apiVersionPattern = regexp.MustCompile(`^v[0-9]+(?:[a-z]+[0-9]+)?$`)

type apiTypeName struct {
	Basename string
	Version  string
}

func parseAPITypeName(fqn string) apiTypeName {
	parts := strings.Split(fqn, ".")
	if len(parts) == 0 {
		return apiTypeName{}
	}
	result := apiTypeName{Basename: parts[len(parts)-1]}
	if len(parts) > 1 && apiVersionPattern.MatchString(parts[len(parts)-2]) {
		result.Version = parts[len(parts)-2]
	}
	return result
}

func versionSuffix(version string) string {
	if version == "" {
		return ""
	}
	match := regexp.MustCompile(`^v([0-9]+)(?:([a-z]+)([0-9]+))?$`).FindStringSubmatch(version)
	if match == nil {
		return pascalCase(version)
	}
	result := "V" + match[1]
	if match[2] != "" {
		result += upperFirst(match[2]) + match[3]
	}
	return result
}

func typeNameForDefinition(fqn string) string {
	parsed := parseAPITypeName(fqn)
	name := parsed.Basename
	if parsed.Version != "" && parsed.Version != "v1" {
		name += versionSuffix(parsed.Version)
	}
	return normalizeTypeName(name)
}

func resourceTypeName(prefix, kind, version string, custom bool, suffix string) string {
	name := kind
	if !custom && version != "v1" {
		name += versionSuffix(version)
	}
	return normalizeTypeName(prefix + name + suffix)
}

// normalizeTypeName follows json2jsii's TypeGenerator.normalizeTypeName. It is
// deliberately not Go's initialism convention: upstream names are CsiDriver,
// ApiService and KubeClusterCidrv1Alpha1.
func normalizeTypeName(value string) string {
	parts := strings.Split(value, "-")
	for index, part := range parts {
		parts[index] = upperFirst(part)
	}
	stage := strings.Join(parts, "")
	runes := []rune(stage)
	for index := 0; index < len(runes); {
		if !unicode.IsUpper(runes[index]) {
			index++
			continue
		}
		start := index
		for index < len(runes) && unicode.IsUpper(runes[index]) {
			index++
		}
		end := index
		// The regular expression used upstream only recognizes an all-caps
		// sequence when followed by a non-lowercase character or end-of-input.
		if end < len(runes) && unicode.IsLower(runes[end]) {
			end--
		}
		if end <= start {
			continue
		}
		for cursor := start + 1; cursor < end; cursor++ {
			runes[cursor] = unicode.ToLower(runes[cursor])
		}
		index = end
	}
	return upperFirst(string(runes))
}

// lowerCamel implements the property normalization used by the camelcase npm
// package. In particular, hostIP becomes hostIp while clusterIPs stays
// clusterIPs. The generated Go field is upperFirst(lowerCamel(name)).
func lowerCamel(value string) string {
	value = strings.TrimPrefix(value, "$")
	value = strings.ReplaceAll(value, "/", "_")
	var words []string
	var current []rune
	runes := []rune(value)
	flush := func() {
		if len(current) > 0 {
			words = append(words, string(current))
			current = nil
		}
	}
	for index, r := range runes {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			flush()
			continue
		}
		if len(current) > 0 && unicode.IsUpper(r) {
			previous := runes[index-1]
			nextIsLower := index+1 < len(runes) && unicode.IsLower(runes[index+1])
			if unicode.IsLower(previous) || unicode.IsDigit(previous) || (unicode.IsUpper(previous) && nextIsLower) {
				flush()
			}
		}
		current = append(current, r)
	}
	flush()
	if len(words) == 0 {
		return "_"
	}
	for index, word := range words {
		word = strings.ToLower(word)
		if index > 0 {
			word = upperFirst(word)
		}
		words[index] = word
	}
	result := strings.Join(words, "")
	switch strings.ToLower(result) {
	case "build", "equals", "hashcode":
		result += "_"
	}
	if result == "" {
		return "_"
	}
	if first, _ := utf8.DecodeRuneInString(result); unicode.IsDigit(first) {
		return "_" + result
	}
	return result
}

func goFieldName(schemaName string) (field string, normalized string) {
	normalized = lowerCamel(schemaName)
	return upperFirst(normalized), normalized
}

func pascalCase(value string) string {
	name := lowerCamel(value)
	runes := []rune(upperFirst(strings.TrimSuffix(name, "_")))
	for index := 1; index < len(runes); index++ {
		if unicode.IsDigit(runes[index-1]) && unicode.IsLetter(runes[index]) {
			runes[index] = unicode.ToUpper(runes[index])
		}
	}
	return string(runes)
}

func upperFirst(value string) string {
	if value == "" {
		return value
	}
	r, size := utf8.DecodeRuneInString(value)
	return string(unicode.ToUpper(r)) + value[size:]
}

func lowerFirst(value string) string {
	if value == "" {
		return value
	}
	r, size := utf8.DecodeRuneInString(value)
	return string(unicode.ToLower(r)) + value[size:]
}

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func enumMember(value interface{}) string {
	raw := fmt.Sprint(value)
	var words []string
	var current []rune
	runes := []rune(raw)
	flush := func() {
		if len(current) > 0 {
			words = append(words, strings.ToUpper(string(current)))
			current = nil
		}
	}
	for index, r := range runes {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			flush()
			continue
		}
		if len(current) > 0 && unicode.IsUpper(r) && index > 0 && unicode.IsLower(runes[index-1]) {
			flush()
		}
		current = append(current, r)
	}
	flush()
	result := strings.Join(words, "_")
	if result == "" {
		result = "VALUE"
	}
	if unicode.IsDigit([]rune(result)[0]) {
		result = "VALUE_" + result
	}
	return result
}

func literal(value interface{}) string {
	switch item := value.(type) {
	case string:
		return strconv.Quote(item)
	case float64:
		return strconv.FormatFloat(item, 'g', -1, 64)
	case bool:
		if item {
			return "true"
		}
		return "false"
	default:
		return strconv.Quote(fmt.Sprint(value))
	}
}
