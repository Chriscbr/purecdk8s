package main

import (
	"testing"

	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func TestChartWithTestingApp(t *testing.T) {
	app := cdk8s.Testing_App(nil)
	chart := NewChart(app, "testing-app")
	assertManifestCount(t, cdk8s.Testing_Synth(chart), 1)
}

func TestChartWithTestingChart(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	addFixture(chart)
	assertManifestCount(t, cdk8s.Testing_Synth(chart), 1)
}

func assertManifestCount(t *testing.T, manifests *[]interface{}, want int) {
	t.Helper()
	if manifests == nil {
		t.Fatal("Testing_Synth() returned nil")
	}
	if got := len(*manifests); got != want {
		t.Fatalf("Testing_Synth() returned %d manifests, want %d", got, want)
	}
}
