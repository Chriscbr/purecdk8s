package main

import (
	"testing"

	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func TestCncfDemoDefaultConfiguration(t *testing.T) {
	app := cdk8s.Testing_App(nil)
	chart := NewCncfDemoChart(app, "cncf-demo-test", nil)
	assertKinds(t, cdk8s.Testing_Synth(chart), "Deployment", "Service", "Ingress")
}

func TestCncfDemoOptionalResources(t *testing.T) {
	props := &AppProps{
		Name:     "cncf-demo",
		Image:    Image{Repository: "index.docker.io/vfarcic/cncf-demo", Tag: "latest"},
		Replicas: 2,
		Ingress:  Ingress{Host: "cncf-demo.example.test", ClassName: "traefik"},
		Tls:      Tls{Enabled: true},
		Db: Db{
			Id: "cncf-demo-db", Size: "small",
			Enabled: DbEnabled{Crossplane: DbCrossplane{AWS: true}},
		},
		SchemaHero: SchemaHero{Enabled: true},
		Otel:       Otel{Enabled: true, JaegerAddr: "http://jaeger:4318"},
	}
	chart := cdk8s.Testing_Chart()
	NewApp(chart, jsii.String("optional-resources"), props)
	assertKinds(t, cdk8s.Testing_Synth(chart),
		"Deployment", "Service", "Ingress", "Certificate", "SQLClaim", "ExternalSecret", "Database", "Table",
	)
}

func TestCncfDemoInsecureDatabaseUsesSecret(t *testing.T) {
	props := &AppProps{
		Name: "cncf-demo", Image: Image{Repository: "example.test/cncf-demo", Tag: "test"},
		Ingress: Ingress{Host: "cncf-demo.example.test", ClassName: "traefik"},
		Db: Db{
			Id: "cncf-demo-db", Insecure: true,
			Enabled: DbEnabled{Crossplane: DbCrossplane{Google: true}},
		},
	}
	chart := cdk8s.Testing_Chart()
	NewApp(chart, jsii.String("insecure-database"), props)
	assertKinds(t, cdk8s.Testing_Synth(chart), "Deployment", "Service", "Ingress", "SQLClaim", "Secret")
}

func assertKinds(t *testing.T, manifests *[]interface{}, kinds ...string) {
	t.Helper()
	if manifests == nil {
		t.Fatal("Testing_Synth() returned nil")
	}
	actual := make(map[string]bool, len(*manifests))
	for _, manifest := range *manifests {
		resource, ok := manifest.(map[string]interface{})
		if !ok {
			t.Fatalf("manifest has type %T, want map[string]interface{}", manifest)
		}
		kind, ok := resource["kind"].(string)
		if !ok {
			t.Fatalf("manifest kind has type %T, want string", resource["kind"])
		}
		actual[kind] = true
	}
	if len(actual) != len(kinds) {
		t.Fatalf("synthesized kinds %v, want %v", actual, kinds)
	}
	for _, kind := range kinds {
		if !actual[kind] {
			t.Fatalf("synthesized kinds %v, missing %s", actual, kind)
		}
	}
}
