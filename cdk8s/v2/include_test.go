package cdk8s_test

import (
	"encoding/json"
	"io"
	"os"
	"sort"
	"strings"
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
	"gopkg.in/yaml.v3"
)

func includeAssertEqual(t *testing.T, got, want interface{}) {
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

func includeParseYAML(t *testing.T, source string) []interface{} {
	t.Helper()
	decoder := yaml.NewDecoder(strings.NewReader(source))
	documents := make([]interface{}, 0)
	for {
		var document interface{}
		err := decoder.Decode(&document)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if document == nil {
			continue
		}
		encoded, err := json.Marshal(document)
		if err != nil {
			t.Fatal(err)
		}
		var normalized interface{}
		if err := json.Unmarshal(encoded, &normalized); err != nil {
			t.Fatal(err)
		}
		documents = append(documents, normalized)
	}
	return documents
}

const includeGuestbook = `apiVersion: v1
kind: Service
metadata:
  name: redis-master
  labels:
    app: redis
    tier: backend
    role: master
spec:
  ports:
  - port: 6379
    targetPort: 6379
  selector:
    app: redis
    tier: backend
    role: master
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis-master
spec:
  selector:
    matchLabels:
      app: redis
      role: master
      tier: backend
  replicas: 1
  template:
    metadata:
      labels:
        app: redis
        role: master
        tier: backend
    spec:
      containers:
      - name: master
        image: registry.k8s.io/redis:e2e
        resources:
          requests:
            cpu: 100m
            memory: 100Mi
        ports:
        - containerPort: 6379
---
apiVersion: v1
kind: Service
metadata:
  name: redis-slave
  labels:
    app: redis
    tier: backend
    role: slave
spec:
  ports:
  - port: 6379
  selector:
    app: redis
    tier: backend
    role: slave
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis-slave
spec:
  selector:
    matchLabels:
      app: redis
      role: slave
      tier: backend
  replicas: 2
  template:
    metadata:
      labels:
        app: redis
        role: slave
        tier: backend
    spec:
      containers:
      - name: slave
        image: gcr.io/google_samples/gb-redisslave:v1
        resources:
          requests:
            cpu: 100m
            memory: 100Mi
        env:
        - name: GET_HOSTS_FROM
          value: dns
        ports:
        - containerPort: 6379
---
apiVersion: v1
kind: Service
metadata:
  name: frontend
  labels:
    app: guestbook
    tier: frontend
spec:
  type: NodePort
  ports:
  - port: 80
  selector:
    app: guestbook
    tier: frontend
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
spec:
  selector:
    matchLabels:
      app: guestbook
      tier: frontend
  replicas: 3
  template:
    metadata:
      labels:
        app: guestbook
        tier: frontend
    spec:
      containers:
      - name: php-redis
        image: gcr.io/google-samples/gb-frontend:v4
        resources:
          requests:
            cpu: 100m
            memory: 100Mi
        env:
        - name: GET_HOSTS_FROM
          value: dns
        ports:
        - containerPort: 80
`

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/include.test.ts#L5
func TestIncludeCanLoadFromYAML(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	path := t.TempDir() + "/guestbook-all-in-one.yaml"
	if err := os.WriteFile(path, []byte(includeGuestbook), 0o644); err != nil {
		t.Fatal(err)
	}
	cdk8s.NewInclude(chart, jsii.String("guestbook"), &cdk8s.IncludeProps{Url: &path})

	expected := includeParseYAML(t, includeGuestbook)
	includeAssertEqual(t, *cdk8s.Testing_Synth(chart), expected)
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/include.test.ts#L21
func TestIncludeSkipsEmptyDocuments(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	documents := []interface{}{map[string]interface{}{}}
	file := cdk8s.Yaml_Tmp(&documents)
	include := cdk8s.NewInclude(chart, jsii.String("empty"), &cdk8s.IncludeProps{Url: file})

	if got := len(*include.Node().Children()); got != 0 {
		t.Fatalf("include children = %d, want 0", got)
	}
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/include.test.ts#L33
func TestIncludeSameNameDifferentKind(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	documents := []interface{}{
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Foo",
			"metadata":   map[string]interface{}{"name": "resource1"},
		},
		map[string]interface{}{},
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Bar",
			"metadata":   map[string]interface{}{"name": "resource1"},
		},
	}
	file := cdk8s.Yaml_Tmp(&documents)
	include := cdk8s.NewInclude(chart, jsii.String("foo"), &cdk8s.IncludeProps{Url: file})

	children := *include.Node().Children()
	ids := make([]string, 0, len(children))
	for _, child := range children {
		ids = append(ids, *child.Node().Id())
	}
	includeAssertEqual(t, ids, []string{"resource1-foo", "resource1-bar"})
}

// Ported from:
// https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/include.test.ts#L61
func TestIncludeApiObjectsReturnsAllObjects(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	documents := []interface{}{
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Foo",
			"metadata":   map[string]interface{}{"name": "resource1"},
		},
		map[string]interface{}{},
		map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Bar",
			"metadata":   map[string]interface{}{"name": "resource1"},
		},
	}
	file := cdk8s.Yaml_Tmp(&documents)
	include := cdk8s.NewInclude(chart, jsii.String("foo"), &cdk8s.IncludeProps{Url: file})

	objects := *include.ApiObjects()
	kinds := make([]string, 0, len(objects))
	for _, object := range objects {
		kinds = append(kinds, *object.Kind())
	}
	sort.Strings(kinds)
	includeAssertEqual(t, kinds, []string{"Bar", "Foo"})
}
