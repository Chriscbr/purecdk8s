package cdk8s_test

import (
	"encoding/json"
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

func metadataAssertEqual(t *testing.T, got, want interface{}) {
	t.Helper()
	gotJSON, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		t.Fatalf("marshal actual value: %v", err)
	}
	wantJSON, err := json.MarshalIndent(want, "", "  ")
	if err != nil {
		t.Fatalf("marshal expected value: %v", err)
	}
	if string(gotJSON) != string(wantJSON) {
		t.Fatalf("value mismatch\n--- got ---\n%s\n--- want ---\n%s", gotJSON, wantJSON)
	}
}

func metadataCreateApiObject() cdk8s.ApiObject {
	chart := cdk8s.Testing_Chart()
	return cdk8s.NewApiObject(chart, jsii.String("ApiObject"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("Service"),
	})
}

type metadataProducer struct{ value interface{} }

func (p *metadataProducer) Produce() interface{} { return p.value }

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/metadata.test.ts#L9
func TestMetadataCanAddLabel(t *testing.T) {
	metadata := cdk8s.NewApiObjectMetadataDefinition(&cdk8s.ApiObjectMetadataDefinitionOptions{
		ApiObject: metadataCreateApiObject(),
	})
	metadata.AddLabel(jsii.String("key"), jsii.String("value"))

	metadataAssertEqual(t, metadata.ToJson(), map[string]interface{}{
		"labels": map[string]interface{}{"key": "value"},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/metadata.test.ts#L23
func TestMetadataCanAddAnnotation(t *testing.T) {
	metadata := cdk8s.NewApiObjectMetadataDefinition(&cdk8s.ApiObjectMetadataDefinitionOptions{
		ApiObject: metadataCreateApiObject(),
	})
	metadata.AddAnnotation(jsii.String("key"), jsii.String("value"))

	metadataAssertEqual(t, metadata.ToJson(), map[string]interface{}{
		"annotations": map[string]interface{}{"key": "value"},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/metadata.test.ts#L37
func TestMetadataCanAddFinalizer(t *testing.T) {
	metadata := cdk8s.NewApiObjectMetadataDefinition(&cdk8s.ApiObjectMetadataDefinitionOptions{
		ApiObject: metadataCreateApiObject(),
	})
	metadata.AddFinalizers(jsii.String("my-finalizer"))

	metadataAssertEqual(t, metadata.ToJson(), map[string]interface{}{
		"finalizers": []interface{}{"my-finalizer"},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/metadata.test.ts#L49
func TestMetadataCanAddOwnerReference(t *testing.T) {
	metadata := cdk8s.NewApiObjectMetadataDefinition(&cdk8s.ApiObjectMetadataDefinitionOptions{
		ApiObject: metadataCreateApiObject(),
	})
	metadata.AddOwnerReference(&cdk8s.OwnerReference{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("Pod"),
		Name:       jsii.String("mypod"),
		Uid:        jsii.String("abcdef12-3456-7890-abcd-ef1234567890"),
	})

	metadataAssertEqual(t, metadata.ToJson(), map[string]interface{}{
		"ownerReferences": []interface{}{
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"name":       "mypod",
				"uid":        "abcdef12-3456-7890-abcd-ef1234567890",
			},
		},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/metadata.test.ts#L73
func TestMetadataInstantiationPropertiesAreRespected(t *testing.T) {
	labels := map[string]*string{"key": jsii.String("value")}
	annotations := map[string]*string{"key": jsii.String("value")}
	metadata := cdk8s.NewApiObjectMetadataDefinition(&cdk8s.ApiObjectMetadataDefinitionOptions{
		ApiObject:   metadataCreateApiObject(),
		Labels:      &labels,
		Annotations: &annotations,
		Name:        jsii.String("name"),
		Namespace:   jsii.String("namespace"),
	})

	metadataAssertEqual(t, metadata.ToJson(), map[string]interface{}{
		"name":        "name",
		"namespace":   "namespace",
		"annotations": map[string]interface{}{"key": "value"},
		"labels":      map[string]interface{}{"key": "value"},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/metadata.test.ts#L98
func TestMetadataLazyPropertiesAreResolved(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	object := cdk8s.NewApiObjectWithManifest(chart, jsii.String("ApiObject"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("Service"),
	}, map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": map[string]interface{}{"key": "value"},
			"annotations": map[string]interface{}{
				"key": "value",
				"lazy": cdk8s.Lazy_Any(&metadataProducer{value: map[string]interface{}{
					"uiMeta": "is good",
				}}),
			},
			"name":      "name",
			"namespace": "namespace",
		},
	})

	metadataAssertEqual(t, object.Metadata().ToJson(), map[string]interface{}{
		"name":      "name",
		"namespace": "namespace",
		"annotations": map[string]interface{}{
			"key":  "value",
			"lazy": map[string]interface{}{"uiMeta": "is good"},
		},
		"labels": map[string]interface{}{"key": "value"},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/metadata.test.ts#L133
func TestMetadataCanIncludeArbitraryKeyValueOptions(t *testing.T) {
	// The upstream index-signature constructor properties are @jsii ignore and
	// therefore have no Go struct fields. A raw manifest is the public Go API
	// that exercises the equivalent constructor flow without exposing internals.
	chart := cdk8s.Testing_Chart()
	object := cdk8s.NewApiObjectWithManifest(chart, jsii.String("ApiObject"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("Service"),
	}, map[string]interface{}{
		"metadata": map[string]interface{}{
			"foo": 123,
			"bar": map[string]interface{}{"helloL": "world"},
		},
	})
	metadata := object.Metadata()
	metadata.Add(jsii.String("bar"), "baz")

	actual, ok := metadata.ToJson().(map[string]interface{})
	if !ok {
		t.Fatalf("metadata ToJson returned %T, want map[string]interface{}", metadata.ToJson())
	}
	// ApiObject contributes its generated name on this public construction path;
	// it is orthogonal to the arbitrary constructor attributes under test.
	delete(actual, "name")
	metadataAssertEqual(t, actual, map[string]interface{}{
		"bar": "baz",
		"foo": float64(123),
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/metadata.test.ts#L150
func TestMetadataLabelsAreCloned(t *testing.T) {
	shared := map[string]*string{"foo": jsii.String("bar")}
	metadata1 := cdk8s.NewApiObjectMetadataDefinition(&cdk8s.ApiObjectMetadataDefinitionOptions{
		ApiObject: metadataCreateApiObject(),
		Labels:    &shared,
	})
	metadata1.AddLabel(jsii.String("bar"), jsii.String("baz"))
	metadata2 := cdk8s.NewApiObjectMetadataDefinition(&cdk8s.ApiObjectMetadataDefinitionOptions{
		ApiObject: metadataCreateApiObject(),
		Labels:    &shared,
	})

	metadataAssertEqual(t, metadata2.ToJson(), map[string]interface{}{
		"labels": map[string]interface{}{"foo": "bar"},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/metadata.test.ts#L173
func TestMetadataAnnotationsAreCloned(t *testing.T) {
	shared := map[string]*string{"foo": jsii.String("bar")}
	metadata1 := cdk8s.NewApiObjectMetadataDefinition(&cdk8s.ApiObjectMetadataDefinitionOptions{
		ApiObject:   metadataCreateApiObject(),
		Annotations: &shared,
	})
	metadata1.AddAnnotation(jsii.String("bar"), jsii.String("baz"))
	metadata2 := cdk8s.NewApiObjectMetadataDefinition(&cdk8s.ApiObjectMetadataDefinitionOptions{
		ApiObject:   metadataCreateApiObject(),
		Annotations: &shared,
	})

	metadataAssertEqual(t, metadata2.ToJson(), map[string]interface{}{
		"annotations": map[string]interface{}{"foo": "bar"},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/metadata.test.ts#L196
func TestMetadataFinalizersAreCloned(t *testing.T) {
	shared := []*string{jsii.String("foo")}
	metadata1 := cdk8s.NewApiObjectMetadataDefinition(&cdk8s.ApiObjectMetadataDefinitionOptions{
		ApiObject:  metadataCreateApiObject(),
		Finalizers: &shared,
	})
	metadata1.AddFinalizers(jsii.String("bar"), jsii.String("baz"))
	metadata2 := cdk8s.NewApiObjectMetadataDefinition(&cdk8s.ApiObjectMetadataDefinitionOptions{
		ApiObject:  metadataCreateApiObject(),
		Finalizers: &shared,
	})

	metadataAssertEqual(t, metadata2.ToJson(), map[string]interface{}{
		"finalizers": []interface{}{"foo"},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/metadata.test.ts#L219
func TestMetadataOwnerReferencesAreCloned(t *testing.T) {
	shared := []*cdk8s.OwnerReference{{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("Kind"),
		Name:       jsii.String("name1"),
		Uid:        jsii.String("uid1"),
	}}
	metadata1 := cdk8s.NewApiObjectMetadataDefinition(&cdk8s.ApiObjectMetadataDefinitionOptions{
		ApiObject:       metadataCreateApiObject(),
		OwnerReferences: &shared,
	})
	metadata1.AddOwnerReference(&cdk8s.OwnerReference{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("Kind"),
		Name:       jsii.String("name2"),
		Uid:        jsii.String("uid2"),
	})
	metadata2 := cdk8s.NewApiObjectMetadataDefinition(&cdk8s.ApiObjectMetadataDefinitionOptions{
		ApiObject:       metadataCreateApiObject(),
		OwnerReferences: &shared,
	})

	metadataAssertEqual(t, metadata2.ToJson(), map[string]interface{}{
		"ownerReferences": []interface{}{
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Kind",
				"name":       "name1",
				"uid":        "uid1",
			},
		},
	})
}
