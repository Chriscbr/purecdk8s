package cdk8splus34_test

import (
	"os"
	"path/filepath"
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	plus "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

func configMapFixtureDirectories(t *testing.T) (flat, nested string) {
	t.Helper()
	root := t.TempDir()
	flat = filepath.Join(root, "flat-directory")
	nested = filepath.Join(root, "nested-directory")
	if err := os.MkdirAll(filepath.Join(nested, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(flat, 0o755); err != nil {
		t.Fatal(err)
	}
	for path, contents := range map[string]string{
		filepath.Join(flat, "file1.txt"):             "Hello, world!",
		filepath.Join(flat, "file2.html"):            "<html>Hey</html>",
		filepath.Join(nested, "file1.txt"):           "Hello, world!",
		filepath.Join(nested, "nested", "file2.txt"): "Hello, world!",
	} {
		if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return flat, nested
}

func TestConfigMap(t *testing.T) {
	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/config-map.test.ts#L5
	t.Run("can grant permissions on imported", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		configMap := plus.ConfigMap_FromConfigMapName(chart, jsii.String("ConfigMap"), jsii.String("name"))
		role := plus.NewRole(chart, jsii.String("Role"), nil)
		role.AllowRead(configMap)

		manifest := manifestOfKind(t, chart, "Role")
		rules := manifest["rules"].([]interface{})
		requireDeepEqual(t, rules, []interface{}{map[string]interface{}{
			"apiGroups":     []interface{}{string("")},
			"resourceNames": []interface{}{"name"},
			"resources":     []interface{}{"configmaps"},
			"verbs":         []interface{}{"get", "list", "watch"},
		}})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/config-map.test.ts#L17
	t.Run("defaultChild", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		configMap := plus.NewConfigMap(chart, jsii.String("ConfigMap"), nil)
		if got := stringValue(cdk8s.ApiObject_Of(configMap).Kind()); got != "ConfigMap" {
			t.Fatalf("default child kind = %q, want ConfigMap", got)
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/config-map.test.ts#L27
	t.Run("minimal", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewConfigMap(chart, jsii.String("my-config-map"), nil)
		requireDeepEqual(t, synth(t, chart), []interface{}{map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata":   map[string]interface{}{"name": "test-my-config-map-c8eaefa4"},
			"immutable":  false,
		}})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/config-map.test.ts#L47
	t.Run("with data", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewConfigMap(chart, jsii.String("my-config-map"), &plus.ConfigMapProps{Data: &map[string]*string{
			"key1": jsii.String("foo"),
			"key2": jsii.String("bar"),
		}})
		manifest := manifestAt(t, chart, 0)
		requireDeepEqual(t, manifest["data"], map[string]interface{}{"key1": "foo", "key2": "bar"})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/config-map.test.ts#L76
	t.Run("with binaryData", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewConfigMap(chart, jsii.String("my-config-map"), &plus.ConfigMapProps{BinaryData: &map[string]*string{
			"key1": jsii.String("foo"),
			"key2": jsii.String("bar"),
		}})
		manifest := manifestAt(t, chart, 0)
		requireDeepEqual(t, manifest["binaryData"], map[string]interface{}{"key1": "foo", "key2": "bar"})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/config-map.test.ts#L105
	t.Run("with binaryData and data", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewConfigMap(chart, jsii.String("my-config-map"), &plus.ConfigMapProps{
			Data:       &map[string]*string{"hello": jsii.String("world")},
			BinaryData: &map[string]*string{"key1": jsii.String("foo"), "key2": jsii.String("bar")},
		})
		manifest := manifestAt(t, chart, 0)
		requireDeepEqual(t, manifest["data"], map[string]interface{}{"hello": "world"})
		requireDeepEqual(t, manifest["binaryData"], map[string]interface{}{"key1": "foo", "key2": "bar"})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/config-map.test.ts#L140
	t.Run("binaryData and data cannot share keys", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		requirePanicContains(t, "key1", func() {
			plus.NewConfigMap(chart, jsii.String("my-config-map"), &plus.ConfigMapProps{
				Data:       &map[string]*string{"key1": jsii.String("world")},
				BinaryData: &map[string]*string{"key1": jsii.String("foo"), "key2": jsii.String("bar")},
			})
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/config-map.test.ts#L161
	t.Run("addData and addBinaryData can add data", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		configMap := plus.NewConfigMap(chart, jsii.String("my-config-map"), &plus.ConfigMapProps{
			Data:       &map[string]*string{"hello": jsii.String("world")},
			BinaryData: &map[string]*string{"key1": jsii.String("foo"), "key2": jsii.String("bar")},
		})
		configMap.AddData(jsii.String("world"), jsii.String("oh yeah!"))
		configMap.AddBinaryData(jsii.String("key3"), jsii.String("baz"))
		manifest := manifestAt(t, chart, 0)
		requireDeepEqual(t, manifest["data"], map[string]interface{}{"hello": "world", "world": "oh yeah!"})
		requireDeepEqual(t, manifest["binaryData"], map[string]interface{}{"key1": "foo", "key2": "bar", "key3": "baz"})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/config-map.test.ts#L200
	t.Run("addData and addBinaryData throw if key already used", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		configMap := plus.NewConfigMap(chart, jsii.String("my-config-map"), &plus.ConfigMapProps{Data: &map[string]*string{"key": jsii.String("value")}})
		requirePanicContains(t, "key", func() { configMap.AddData(jsii.String("key"), jsii.String("value2")) })
		requirePanicContains(t, "key", func() { configMap.AddBinaryData(jsii.String("key"), jsii.String("value2")) })
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/config-map.test.ts#L218
	t.Run("addFile adds local files", func(t *testing.T) {
		flat, _ := configMapFixtureDirectories(t)
		chart := cdk8s.Testing_Chart()
		configMap := plus.NewConfigMap(chart, jsii.String("my-config-map"), nil)
		configMap.AddFile(jsii.String(filepath.Join(flat, "file1.txt")), nil)
		configMap.AddFile(jsii.String(filepath.Join(flat, "file2.html")), jsii.String("hey-there"))
		manifest := manifestAt(t, chart, 0)
		requireDeepEqual(t, manifest["data"], map[string]interface{}{"file1.txt": "Hello, world!", "hey-there": "<html>Hey</html>"})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/config-map.test.ts#L233
	t.Run("addDirectory skips sub-directories", func(t *testing.T) {
		_, nested := configMapFixtureDirectories(t)
		chart := cdk8s.Testing_Chart()
		configMap := plus.NewConfigMap(chart, jsii.String("my-config-map"), nil)
		configMap.AddDirectory(jsii.String(nested), nil)
		requireDeepEqual(t, manifestAt(t, chart, 0)["data"], map[string]interface{}{"file1.txt": "Hello, world!"})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/config-map.test.ts#L245
	t.Run("addDirectory keys use file names", func(t *testing.T) {
		flat, _ := configMapFixtureDirectories(t)
		chart := cdk8s.Testing_Chart()
		configMap := plus.NewConfigMap(chart, jsii.String("my-config-map"), nil)
		configMap.AddDirectory(jsii.String(flat), nil)
		requireDeepEqual(t, manifestAt(t, chart, 0)["data"], map[string]interface{}{"file1.txt": "Hello, world!", "file2.html": "<html>Hey</html>"})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/config-map.test.ts#L257
	t.Run("addDirectory prefixes keys", func(t *testing.T) {
		flat, _ := configMapFixtureDirectories(t)
		chart := cdk8s.Testing_Chart()
		configMap := plus.NewConfigMap(chart, jsii.String("my-config-map"), nil)
		configMap.AddDirectory(jsii.String(flat), &plus.AddDirectoryOptions{KeyPrefix: jsii.String("prefix.")})
		requireDeepEqual(t, manifestAt(t, chart, 0)["data"], map[string]interface{}{"prefix.file1.txt": "Hello, world!", "prefix.file2.html": "<html>Hey</html>"})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/config-map.test.ts#L271
	t.Run("addDirectory excludes with globs", func(t *testing.T) {
		flat, _ := configMapFixtureDirectories(t)
		chart := cdk8s.Testing_Chart()
		configMap := plus.NewConfigMap(chart, jsii.String("my-config-map"), nil)
		configMap.AddDirectory(jsii.String(flat), &plus.AddDirectoryOptions{Exclude: &[]*string{jsii.String("*.html")}})
		requireDeepEqual(t, manifestAt(t, chart, 0)["data"], map[string]interface{}{"file1.txt": "Hello, world!"})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/config-map.test.ts#L286
	t.Run("metadata is synthesized", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewConfigMap(chart, jsii.String("my-config-map"), &plus.ConfigMapProps{Metadata: &cdk8s.ApiObjectMetadata{Name: jsii.String("my-name")}})
		requireDeepEqual(t, manifestAt(t, chart, 0)["metadata"], map[string]interface{}{"name": "my-name"})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/config-map.test.ts#L309
	t.Run("can configure an immutable config map", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		configMap := plus.NewConfigMap(chart, jsii.String("my-config-map"), &plus.ConfigMapProps{Immutable: jsii.Bool(true)})
		if !boolValue(configMap.Immutable()) {
			t.Fatal("Immutable() = false, want true")
		}
		if got := manifestAt(t, chart, 0)["immutable"]; got != true {
			t.Fatalf("immutable = %#v, want true", got)
		}
	})
}
