package cdk8s

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const maximumYAMLDownload = 10 * 1024 * 1024

var (
	yaml11Boolean = regexp.MustCompile(`^(?:y|Y|yes|Yes|YES|n|N|no|No|NO|true|True|TRUE|false|False|FALSE|on|On|ON|off|Off|OFF)$`)
	yaml11Date    = regexp.MustCompile(`^\d{4}-\d\d?-\d\d?(?:[Tt]|[ \t]+)\d\d?:\d{2}:\d{2}|^\d{4}-\d\d?-\d\d?$`)
	yaml11Octal   = regexp.MustCompile(`^[+-]?0[0-7_]+$`)
)

// Yaml_FormatObjects is retained for API compatibility.
func Yaml_FormatObjects(docs *[]interface{}) *string {
	if docs == nil {
		panic("parameter docs is required, but nil was provided")
	}
	return Yaml_Stringify((*docs)...)
}

// Yaml_Load loads all non-empty documents from a local file or HTTP(S) URL.
func Yaml_Load(urlOrFile *string) *[]interface{} {
	if urlOrFile == nil {
		panic("parameter urlOrFile is required, but nil was provided")
	}
	body := loadYAMLBytes(*urlOrFile)
	decoder := yaml.NewDecoder(bytes.NewReader(body))
	result := make([]interface{}, 0)
	for {
		var document yaml.Node
		err := decoder.Decode(&document)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		if len(document.Content) == 0 {
			continue
		}
		value := decodeYAML11(document.Content[0])
		if emptyYAMLDocument(value) {
			continue
		}
		result = append(result, value)
	}
	return &result
}

// Yaml_Save writes docs as a multi-document YAML stream.
func Yaml_Save(filePath *string, docs *[]interface{}) {
	if filePath == nil {
		panic("parameter filePath is required, but nil was provided")
	}
	if docs == nil {
		panic("parameter docs is required, but nil was provided")
	}
	data := Yaml_Stringify((*docs)...)
	if err := os.WriteFile(*filePath, []byte(*data), 0o644); err != nil {
		panic(err)
	}
}

// Yaml_Stringify formats one or more values as YAML 1.1 documents.
func Yaml_Stringify(docs ...interface{}) *string {
	var output strings.Builder
	for index, document := range docs {
		if index > 0 {
			output.WriteString("---\n")
		}
		plain := plainValue(document)
		if plain == nil {
			output.WriteByte('\n')
			continue
		}
		var node yaml.Node
		if err := node.Encode(plain); err != nil {
			panic(err)
		}
		orderYAMLTopLevel(&node)
		quoteYAML11Strings(&node)

		var encoded bytes.Buffer
		encoder := yaml.NewEncoder(&encoded)
		encoder.SetIndent(2)
		if err := encoder.Encode(&node); err != nil {
			panic(err)
		}
		if err := encoder.Close(); err != nil {
			panic(err)
		}
		output.Write(encoded.Bytes())
	}
	result := output.String()
	return &result
}

func orderYAMLTopLevel(node *yaml.Node) {
	if node == nil {
		return
	}
	target := node
	if target.Kind == yaml.DocumentNode && len(target.Content) > 0 {
		target = target.Content[0]
	}
	if target.Kind != yaml.MappingNode || len(target.Content) < 2 {
		return
	}
	pairs := make(map[string][]*yaml.Node, len(target.Content)/2)
	order := make([]string, 0, len(target.Content)/2)
	for index := 0; index+1 < len(target.Content); index += 2 {
		key := target.Content[index].Value
		pairs[key] = []*yaml.Node{target.Content[index], target.Content[index+1]}
		order = append(order, key)
	}
	result := make([]*yaml.Node, 0, len(target.Content))
	for _, key := range []string{"apiVersion", "kind", "metadata"} {
		if pair, ok := pairs[key]; ok {
			result = append(result, pair...)
			delete(pairs, key)
		}
	}
	for _, key := range order {
		if pair, ok := pairs[key]; ok {
			result = append(result, pair...)
			delete(pairs, key)
		}
	}
	target.Content = result
}

// Yaml_Tmp saves docs to temp.yaml in a newly created cdk8s-* directory.
func Yaml_Tmp(docs *[]interface{}) *string {
	if docs == nil {
		panic("parameter docs is required, but nil was provided")
	}
	directory, err := os.MkdirTemp("", "cdk8s-")
	if err != nil {
		panic(err)
	}
	path := filepath.Join(directory, "temp.yaml")
	Yaml_Save(&path, docs)
	return &path
}

func loadYAMLBytes(location string) []byte {
	if info, err := os.Stat(location); err == nil && info.Mode().IsRegular() {
		data, err := os.ReadFile(location)
		if err != nil {
			panic(err)
		}
		if len(data) > maximumYAMLDownload {
			panic("YAML input exceeds the 10 MiB limit")
		}
		return data
	}

	parsed, err := url.Parse(location)
	if err != nil {
		panic(err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		if parsed.Scheme == "" {
			panic(fmt.Sprintf("unable to determine protocol from url: %s", location))
		}
		panic(fmt.Sprintf("unsupported protocol %q in url: %s", parsed.Scheme+":", location))
	}
	response, err := http.Get(location) // #nosec G107 -- loading a caller-supplied URL is this API's purpose.
	if err != nil {
		panic(fmt.Sprintf("http error: %v", err))
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("%d response from http get: %s", response.StatusCode, http.StatusText(response.StatusCode)))
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maximumYAMLDownload+1))
	if err != nil {
		panic(err)
	}
	if len(data) > maximumYAMLDownload {
		panic("YAML input exceeds the 10 MiB limit")
	}
	return data
}

func quoteYAML11Strings(node *yaml.Node) {
	if node == nil {
		return
	}
	if node.Kind == yaml.ScalarNode && node.Tag == "!!str" {
		if node.Value == "" || yaml11Boolean.MatchString(node.Value) || yaml11Date.MatchString(node.Value) {
			node.Style = yaml.DoubleQuotedStyle
		}
	}
	for _, child := range node.Content {
		quoteYAML11Strings(child)
	}
}

func decodeYAML11(node *yaml.Node) interface{} {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case yaml.MappingNode:
		result := make(map[string]interface{}, len(node.Content)/2)
		for index := 0; index+1 < len(node.Content); index += 2 {
			key := fmt.Sprint(decodeYAML11(node.Content[index]))
			result[key] = decodeYAML11(node.Content[index+1])
		}
		return result
	case yaml.SequenceNode:
		result := make([]interface{}, len(node.Content))
		for index, child := range node.Content {
			result[index] = decodeYAML11(child)
		}
		return result
	case yaml.AliasNode:
		return decodeYAML11(node.Alias)
	case yaml.ScalarNode:
		if node.Style == 0 && node.Tag == "!!str" {
			switch strings.ToLower(node.Value) {
			case "y", "yes", "true", "on":
				return true
			case "n", "no", "false", "off":
				return false
			}
			if yaml11Octal.MatchString(node.Value) {
				sign := float64(1)
				text := strings.ReplaceAll(node.Value, "_", "")
				if strings.HasPrefix(text, "-") {
					sign = -1
					text = text[1:]
				} else if strings.HasPrefix(text, "+") {
					text = text[1:]
				}
				value, err := strconv.ParseUint(strings.TrimPrefix(text, "0"), 8, 64)
				if err == nil {
					return sign * float64(value)
				}
			}
		}
		var value interface{}
		if err := node.Decode(&value); err != nil {
			panic(err)
		}
		switch scalar := value.(type) {
		case int:
			return float64(scalar)
		case int64:
			return float64(scalar)
		case uint64:
			return float64(scalar)
		case float32:
			return float64(scalar)
		case time.Time:
			// JavaScript's Date is marshalled through JSII as an ISO string.
			return scalar.UTC().Format("2006-01-02T15:04:05.000Z")
		default:
			return scalar
		}
	default:
		return nil
	}
}

func emptyYAMLDocument(value interface{}) bool {
	switch item := value.(type) {
	case nil:
		return true
	case map[string]interface{}:
		return len(item) == 0
	case []interface{}:
		return len(item) == 0
	default:
		return false
	}
}
