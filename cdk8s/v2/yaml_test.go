package cdk8s_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
)

const yamlGuestbookExpectedJSON = `[{"apiVersion":"v1","kind":"Service","metadata":{"name":"redis-master","labels":{"app":"redis","tier":"backend","role":"master"}},"spec":{"ports":[{"port":6379,"targetPort":6379}],"selector":{"app":"redis","tier":"backend","role":"master"}}},{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"redis-master"},"spec":{"selector":{"matchLabels":{"app":"redis","role":"master","tier":"backend"}},"replicas":1,"template":{"metadata":{"labels":{"app":"redis","role":"master","tier":"backend"}},"spec":{"containers":[{"name":"master","image":"registry.k8s.io/redis:e2e","resources":{"requests":{"cpu":"100m","memory":"100Mi"}},"ports":[{"containerPort":6379}]}]}}}},{"apiVersion":"v1","kind":"Service","metadata":{"name":"redis-replica","labels":{"app":"redis","tier":"backend","role":"replica"}},"spec":{"ports":[{"port":6379}],"selector":{"app":"redis","tier":"backend","role":"replica"}}},{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"redis-replica"},"spec":{"selector":{"matchLabels":{"app":"redis","role":"replica","tier":"backend"}},"replicas":2,"template":{"metadata":{"labels":{"app":"redis","role":"replica","tier":"backend"}},"spec":{"containers":[{"name":"replica","image":"gcr.io/google_samples/gb-redisslave:v1","resources":{"requests":{"cpu":"100m","memory":"100Mi"}},"env":[{"name":"GET_HOSTS_FROM","value":"dns"}],"ports":[{"containerPort":6379}]}]}}}},{"apiVersion":"v1","kind":"Service","metadata":{"name":"frontend","labels":{"app":"guestbook","tier":"frontend"}},"spec":{"type":"NodePort","ports":[{"port":80}],"selector":{"app":"guestbook","tier":"frontend"}}},{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"frontend"},"spec":{"selector":{"matchLabels":{"app":"guestbook","tier":"frontend"}},"replicas":3,"template":{"metadata":{"labels":{"app":"guestbook","tier":"frontend"}},"spec":{"containers":[{"name":"php-redis","image":"gcr.io/google-samples/gb-frontend:v4","resources":{"requests":{"cpu":"100m","memory":"100Mi"}},"env":[{"name":"GET_HOSTS_FROM","value":"dns"}],"ports":[{"containerPort":80}]}]}}}}]`

func yamlString(value string) *string { return &value }

func yamlWriteFile(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "input.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func yamlTmp(t *testing.T, documents []interface{}) string {
	t.Helper()
	path := *cdk8s.Yaml_Tmp(&documents)
	t.Cleanup(func() { _ = os.RemoveAll(filepath.Dir(path)) })
	return path
}

func yamlReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func yamlAssertJSONEqual(t *testing.T, got, want interface{}) {
	t.Helper()
	normalize := func(value interface{}) interface{} {
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatalf("marshal value: %v", err)
		}
		var result interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("unmarshal value: %v", err)
		}
		return result
	}
	gotNormalized, wantNormalized := normalize(got), normalize(want)
	if !reflect.DeepEqual(gotNormalized, wantNormalized) {
		gotJSON, _ := json.MarshalIndent(gotNormalized, "", "  ")
		wantJSON, _ := json.MarshalIndent(wantNormalized, "", "  ")
		t.Fatalf("value mismatch\n--- got ---\n%s\n--- want ---\n%s", gotJSON, wantJSON)
	}
}

func yamlGuestbookDocuments(t *testing.T) []interface{} {
	t.Helper()
	var documents []interface{}
	if err := json.Unmarshal([]byte(yamlGuestbookExpectedJSON), &documents); err != nil {
		t.Fatal(err)
	}
	return documents
}

func yamlGuestbookServer(t *testing.T) *httptest.Server {
	t.Helper()
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate YAML test source")
	}
	fixture, err := os.ReadFile(filepath.Join(filepath.Dir(sourceFile), "testdata", "guestbook-all-in-one.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/yaml")
		_, _ = writer.Write(fixture)
	}))
	t.Cleanup(server.Close)
	return server
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/yaml.test.ts#L7
func TestYamlLoadFromFile(t *testing.T) {
	path := yamlWriteFile(t, `hello: 1234
world: 111
---
foo:
  - bar
  - zoo
  - goo
---
---
- hello
- world
`)

	want := []interface{}{
		map[string]interface{}{"hello": 1234, "world": 111},
		map[string]interface{}{"foo": []interface{}{"bar", "zoo", "goo"}},
		[]interface{}{"hello", "world"},
	}
	yamlAssertJSONEqual(t, *cdk8s.Yaml_Load(&path), want)
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/yaml.test.ts#L11
func TestYamlLoadFromURL(t *testing.T) {
	server := yamlGuestbookServer(t)
	want := yamlGuestbookDocuments(t)

	yamlAssertJSONEqual(t, *cdk8s.Yaml_Load(yamlString(server.URL)), want)
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/yaml.test.ts#L15
func TestYamlLoadFiltersEmptyDocuments(t *testing.T) {
	path := yamlTmp(t, []interface{}{
		map[string]interface{}{"doc": 1},
		nil,
		"str_doc",
		map[string]interface{}{},
		"",
		nil,
		0,
		[]interface{}{},
		map[string]interface{}{"doc": 2},
	})

	want := []interface{}{
		map[string]interface{}{"doc": 1},
		"str_doc",
		"",
		0,
		map[string]interface{}{"doc": 2},
	}
	yamlAssertJSONEqual(t, *cdk8s.Yaml_Load(&path), want)
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/yaml.test.ts#L33
func TestYamlSaveSingleDocument(t *testing.T) {
	path := yamlTmp(t, []interface{}{
		map[string]interface{}{"foo": "bar", "hello": []interface{}{1, 2, 3}},
	})

	want := "foo: bar\nhello:\n  - 1\n  - 2\n  - 3\n"
	if got := yamlReadFile(t, path); got != want {
		t.Fatalf("YAML = %q, want %q", got, want)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/yaml.test.ts#L38
func TestYamlSaveMultipleDocuments(t *testing.T) {
	path := yamlTmp(t, []interface{}{
		map[string]interface{}{"foo": "bar", "hello": []interface{}{1, 2, 3}},
		map[string]interface{}{"number": 2},
	})

	want := "foo: bar\nhello:\n  - 1\n  - 2\n  - 3\n---\nnumber: 2\n"
	if got := yamlReadFile(t, path); got != want {
		t.Fatalf("YAML = %q, want %q", got, want)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/yaml.test.ts#L47
func TestYamlSavePreservesEmptyValues(t *testing.T) {
	// Go has no distinct undefined value, so nil represents both upstream
	// undefined and null values here. Go maps also do not retain insertion
	// order, so YAML snapshots use purecdk8s's canonical lexical map order.
	path := yamlTmp(t, []interface{}{
		map[string]interface{}{
			"i_am_undefined":  nil,
			"i_am_null":       nil,
			"empty_array":     []interface{}{},
			"empty_object":    map[string]interface{}{},
			"typed_nil_map":   map[string]string(nil),
			"typed_nil_slice": []string(nil),
		},
	})

	want := "empty_array: []\nempty_object: {}\ni_am_null: null\ni_am_undefined: null\ntyped_nil_map: null\ntyped_nil_slice: null\n"
	if got := yamlReadFile(t, path); got != want {
		t.Fatalf("YAML = %q, want %q", got, want)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/yaml.test.ts#L60
func TestYamlSaveRespectsEmptyDocuments(t *testing.T) {
	path := yamlTmp(t, []interface{}{
		map[string]interface{}{},
		map[string]interface{}{},
		nil,
		map[string]interface{}{"empty": true},
		map[string]interface{}{},
	})

	want := "{}\n---\n{}\n---\n\n---\nempty: true\n---\n{}\n"
	if got := yamlReadFile(t, path); got != want {
		t.Fatalf("YAML = %q, want %q", got, want)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/yaml.test.ts#L72
func TestYamlSaveQuotesStrings(t *testing.T) {
	path := yamlTmp(t, []interface{}{
		map[string]interface{}{
			"foo":          "on",
			"bar":          `this has a "big quote"`,
			"not_a_string": true,
		},
	})

	want := "bar: this has a \"big quote\"\nfoo: \"on\"\nnot_a_string: true\n"
	if got := yamlReadFile(t, path); got != want {
		t.Fatalf("YAML = %q, want %q", got, want)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/yaml.test.ts#L82
func TestYamlEscapedCharacterDoesNotCrossLineBoundaries(t *testing.T) {
	longStringList := []interface{}{
		`^(((d*(.d*)?h)|(d*(.d*)?m)|(d*(.d*)?s)|(d*(.d*)?ms)|(d*(.d*)?us)|(d*(.d*)?µs)|(d*(.d*)?ns))+|infinity|infinite)$`,
	}
	path := yamlTmp(t, longStringList)

	if got := strings.TrimRight(yamlReadFile(t, path), "\n"); strings.Contains(got, "\n") {
		t.Fatalf("long escaped string crossed a line boundary: %q", got)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/yaml.test.ts#L93
func TestYAML11OctalNumbersParsedCorrectly(t *testing.T) {
	path := yamlWriteFile(t, "foo: 0755")

	yamlAssertJSONEqual(t, *cdk8s.Yaml_Load(&path), []interface{}{
		map[string]interface{}{"foo": 493},
	})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/yaml.test.ts#L101
func TestYamlMultilineTextWithLongLineKeepsLineBreak(t *testing.T) {
	value := map[string]interface{}{
		"foo": "[section]\n    abc: s\n    def: 012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789",
		"bar": "something",
	}

	want := "bar: something\nfoo: |-\n  [section]\n      abc: s\n      def: 012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789\n"
	if got := *cdk8s.Yaml_Stringify(value); got != want {
		t.Fatalf("YAML = %q, want %q", got, want)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/yaml.test.ts#L111
func TestYamlStringifyAcceptsMultipleDocuments(t *testing.T) {
	got := *cdk8s.Yaml_Stringify(
		map[string]interface{}{"foo": 123},
		map[string]interface{}{"bar": []interface{}{"hi", "there"}, "jam": 12},
	)
	want := "foo: 123\n---\nbar:\n  - hi\n  - there\njam: 12\n"
	if got != want {
		t.Fatalf("YAML = %q, want %q", got, want)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/yaml.test.ts#L124
func TestYamlStringsDoNotBecomeBooleans(t *testing.T) {
	got := *cdk8s.Yaml_Stringify(map[string]interface{}{
		"a_yes":   "yes",
		"a_no":    "no",
		"a_true":  "true",
		"a_false": "false",
	})
	want := "a_false: \"false\"\na_no: \"no\"\na_true: \"true\"\na_yes: \"yes\"\n"
	if got != want {
		t.Fatalf("YAML = %q, want %q", got, want)
	}
}
