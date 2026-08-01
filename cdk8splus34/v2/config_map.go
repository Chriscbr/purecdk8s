package cdk8splus34

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/purecdk8s/purecdk8s/cdk8s/v2"
	"github.com/purecdk8s/purecdk8s/constructs/v10"
	"github.com/purecdk8s/purecdk8s/jsii"
)

// IConfigMap represents a config map resource.
type IConfigMap interface{ IResource }
type ConfigMap interface {
	Resource
	IConfigMap
	Immutable() *bool
	Data() *map[string]*string
	BinaryData() *map[string]*string
	AddData(key, value *string)
	AddBinaryData(key, value *string)
	AddFile(localFile, key *string)
	AddDirectory(localDir *string, options *AddDirectoryOptions)
}
type ConfigMapProps struct {
	Metadata   *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	Data       *map[string]*string      `field:"optional" json:"data" yaml:"data"`
	BinaryData *map[string]*string      `field:"optional" json:"binaryData" yaml:"binaryData"`
	Immutable  *bool                    `field:"optional" json:"immutable" yaml:"immutable"`
}
type AddDirectoryOptions struct {
	Exclude   *[]*string `field:"optional" json:"exclude" yaml:"exclude"`
	KeyPrefix *string    `field:"optional" json:"keyPrefix" yaml:"keyPrefix"`
}
type configMapImpl struct {
	resourceBase
	data       map[string]*string
	binaryData map[string]*string
	immutable  *bool
}

func NewConfigMap(scope constructs.Construct, id *string, props *ConfigMapProps) ConfigMap {
	if props == nil {
		props = &ConfigMapProps{}
	}
	result := &configMapImpl{data: map[string]*string{}, binaryData: map[string]*string{}, immutable: props.Immutable}
	manifest := map[string]interface{}{}
	result.resourceBase.initialize(result, scope, id, "v1", "ConfigMap", "configmaps", props.Metadata, manifest)
	immutable := false
	if props.Immutable != nil {
		immutable = *props.Immutable
	}
	manifest["immutable"] = immutable
	if props.Data != nil {
		for key, value := range *props.Data {
			result.AddData(jsii.String(key), value)
		}
	}
	if props.BinaryData != nil {
		for key, value := range *props.BinaryData {
			result.AddBinaryData(jsii.String(key), value)
		}
	}
	return result
}
func NewConfigMap_Override(c ConfigMap, scope constructs.Construct, id *string, props *ConfigMapProps) {
	panic("native cdk8splus34 overrides are not implemented")
}
func ConfigMap_IsConstruct(x interface{}) *bool { return constructs.Construct_IsConstruct(x) }
func (c *configMapImpl) Immutable() *bool       { return c.immutable }
func (c *configMapImpl) Data() *map[string]*string {
	result := map[string]*string{}
	for k, v := range c.data {
		result[k] = v
	}
	return &result
}
func (c *configMapImpl) BinaryData() *map[string]*string {
	result := map[string]*string{}
	for k, v := range c.binaryData {
		result[k] = v
	}
	return &result
}
func (c *configMapImpl) verifyAvailable(key *string) {
	if key == nil {
		panic("key is required")
	}
	if _, ok := c.data[*key]; ok {
		panic(fmt.Sprintf("key %q already exists", *key))
	}
	if _, ok := c.binaryData[*key]; ok {
		panic(fmt.Sprintf("key %q already exists", *key))
	}
}
func (c *configMapImpl) AddData(key, value *string) {
	c.verifyAvailable(key)
	if value == nil {
		panic("value is required")
	}
	c.data[*key] = value
	c.manifest["data"] = c.data
}
func (c *configMapImpl) AddBinaryData(key, value *string) {
	c.verifyAvailable(key)
	if value == nil {
		panic("value is required")
	}
	c.binaryData[*key] = value
	c.manifest["binaryData"] = c.binaryData
}
func (c *configMapImpl) AddFile(localFile, key *string) {
	if localFile == nil {
		panic("localFile is required")
	}
	actualKey := key
	if actualKey == nil {
		actualKey = jsii.String(filepath.Base(*localFile))
	}
	value, err := os.ReadFile(*localFile)
	if err != nil {
		panic(err)
	}
	c.AddData(actualKey, jsii.String(string(value)))
}
func (c *configMapImpl) AddDirectory(localDir *string, options *AddDirectoryOptions) {
	if localDir == nil {
		panic("localDir is required")
	}
	entries, err := os.ReadDir(*localDir)
	if err != nil {
		panic(err)
	}
	prefix := ""
	if options != nil && options.KeyPrefix != nil {
		prefix = *options.KeyPrefix
	}
	excluded := map[string]bool{}
	if options != nil && options.Exclude != nil {
		for _, pattern := range *options.Exclude {
			if pattern != nil {
				excluded[*pattern] = true
			}
		}
	}
	for _, entry := range entries {
		if entry.IsDir() || excluded[entry.Name()] {
			continue
		}
		filename := filepath.Join(*localDir, entry.Name())
		c.AddFile(jsii.String(filename), jsii.String(prefix+entry.Name()))
	}
}
