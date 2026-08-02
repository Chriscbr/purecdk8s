package cdk8splus34

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// IConfigMap represents a config map resource.
type (
	// Represents a config map.
	IConfigMap interface{ IResource }
	// ConfigMap holds configuration data for pods to consume.
	ConfigMap interface {
		Resource
		IConfigMap
		// Whether or not this config map is immutable.
		Immutable() *bool
		// The data associated with this config map.
		//
		// Returns an copy. To add data records, use `addData()` or `addBinaryData()`.
		Data() *map[string]*string
		// The binary data associated with this config map.
		//
		// Returns a copy. To add data records, use `addBinaryData()` or `addData()`.
		BinaryData() *map[string]*string
		// Adds a data entry to the config map.
		AddData(key, value *string)
		// Adds a binary data entry to the config map.
		//
		// BinaryData can contain byte sequences that are not in the UTF-8 range.
		AddBinaryData(key, value *string)
		// Adds a file to the ConfigMap.
		AddFile(localFile, key *string)
		// Adds a directory to the ConfigMap.
		AddDirectory(localDir *string, options *AddDirectoryOptions)
	}
)

// Properties for initialization of `ConfigMap`.
type ConfigMapProps struct {
	// Metadata that all persisted resources must have, which includes all objects users must create.
	Metadata *cdk8s.ApiObjectMetadata `field:"optional" json:"metadata" yaml:"metadata"`
	// Data contains the configuration data.
	//
	// Each key must consist of alphanumeric characters, '-', '_' or '.'. Values with non-UTF-8 byte sequences must use the BinaryData field. The keys stored in Data must not overlap with the keys in the BinaryData field, this is enforced during validation process.
	//
	// You can also add data using `configMap.addData()`.
	Data *map[string]*string `field:"optional" json:"data" yaml:"data"`
	// BinaryData contains the binary data.
	//
	// Each key must consist of alphanumeric characters, '-', '_' or '.'. BinaryData can contain byte sequences that are not in the UTF-8 range. The keys stored in BinaryData must not overlap with the ones in the Data field, this is enforced during validation process.
	//
	// You can also add binary data using `configMap.addBinaryData()`.
	BinaryData *map[string]*string `field:"optional" json:"binaryData" yaml:"binaryData"`
	// If set to true, ensures that data stored in the ConfigMap cannot be updated (only object metadata can be modified).
	//
	// If not set to true, the field can be modified at any time. Default: false.
	Immutable *bool `field:"optional" json:"immutable" yaml:"immutable"`
}

// Options for `configmap.addDirectory()`.
type AddDirectoryOptions struct {
	// Glob patterns to exclude when adding files. Default: - include all files.
	Exclude *[]*string `field:"optional" json:"exclude" yaml:"exclude"`
	// A prefix to add to all keys in the config map. Default: "".
	KeyPrefix *string `field:"optional" json:"keyPrefix" yaml:"keyPrefix"`
}

type configMapImpl struct {
	resourceBase
	data       map[string]*string
	binaryData map[string]*string
	immutable  *bool
}

type importedConfigMap struct {
	node constructs.Node
	name *string
}

func (c *importedConfigMap) Node() constructs.Node {
	return c.node
}

func (c *importedConfigMap) SetNodeInternal(node constructs.Node) {
	c.node = node
}

func (c *importedConfigMap) ToString() *string {
	return c.node.Path()
}

func (c *importedConfigMap) With(mixins ...constructs.IMixin) constructs.IConstruct {
	return c.node.With(mixins...)
}

func (c *importedConfigMap) ApiVersion() *string {
	return jsii.String("v1")
}

func (c *importedConfigMap) ApiGroup() *string {
	return jsii.String("")
}

func (c *importedConfigMap) Kind() *string {
	return jsii.String("ConfigMap")
}

func (c *importedConfigMap) Name() *string {
	return c.name
}

func (c *importedConfigMap) ResourceName() *string {
	return c.name
}

func (c *importedConfigMap) ResourceType() *string {
	return jsii.String("configmaps")
}

// Represents a ConfigMap created elsewhere.
func ConfigMap_FromConfigMapName(scope constructs.Construct, id, name *string) IConfigMap {
	if scope == nil || id == nil || name == nil {
		panic("scope, id and name are required")
	}
	result := &importedConfigMap{name: name}
	constructs.NewConstruct_Override(result, scope, id)
	return result
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
	applyOverride(c, NewConfigMap(scope, id, props), "ConfigMap")
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct` instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on disk are seen as independent, completely different libraries. As a consequence, the class `Construct` in each copy of the `constructs` library is seen as a different class, and an instance of one class will not test as `instanceof` the other class. `npm install` will not create installations like this, but users may manually symlink construct libraries together or use a monorepo tool: in those cases, multiple copies of the `constructs` library can be accidentally installed, and `instanceof` will behave unpredictably. It is safest to avoid using `instanceof`, and using this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func ConfigMap_IsConstruct(x interface{}) *bool {
	return constructs.Construct_IsConstruct(x)
}

func (c *configMapImpl) Immutable() *bool {
	return c.immutable
}

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
