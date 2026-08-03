package cdk8s_test

import (
	"encoding/json"
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

func tokenAssertEqual(t *testing.T, got, want interface{}) {
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

type tokenProducer struct{ value interface{} }

func (p *tokenProducer) Produce() interface{} { return p.value }

type tokenImplicit struct{ value interface{} }

func (i *tokenImplicit) Resolve() interface{} { return i.value }

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/tokens.test.ts#L4
func TestTokenLazy(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	object := cdk8s.NewApiObjectWithManifest(chart, jsii.String("Pod"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("Pod"),
		Metadata:   &cdk8s.ApiObjectMetadata{Name: jsii.String("mypod")},
	}, map[string]interface{}{
		"spec": map[string]interface{}{
			"number":   cdk8s.Lazy_Any(&tokenProducer{value: 1234}),
			"string":   cdk8s.Lazy_Any(&tokenProducer{value: "hello"}),
			"implicit": &tokenImplicit{value: 908},
		},
	})

	tokenAssertEqual(t, object.ToJson(), map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata":   map[string]interface{}{"name": "mypod"},
		"spec": map[string]interface{}{
			"number":   float64(1234),
			"string":   "hello",
			"implicit": float64(908),
		},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/tokens.test.ts#L37
func TestTokenDoesNotResolveAWSCdkTokens(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	cdk8s.NewApiObjectWithManifest(chart, jsii.String("Pod"), &cdk8s.ApiObjectProps{
		ApiVersion: jsii.String("v1"),
		Kind:       jsii.String("Pod"),
		Metadata:   &cdk8s.ApiObjectMetadata{Name: jsii.String("mypod")},
	}, map[string]interface{}{
		"spec": map[string]interface{}{
			"bucketName":       "${Token[TOKEN.61]}",
			"someLazyProperty": cdk8s.Lazy_Any(&tokenProducer{value: "lazyValue"}),
		},
	})

	tokenAssertEqual(t, *chart.ToJson(), []interface{}{
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata":   map[string]interface{}{"name": "mypod"},
			"spec": map[string]interface{}{
				"bucketName":       "${Token[TOKEN.61]}",
				"someLazyProperty": "lazyValue",
			},
		},
	})
}
