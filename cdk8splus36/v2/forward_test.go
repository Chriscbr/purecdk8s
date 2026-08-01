package cdk8splus36

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/purecdk8s/purecdk8s/cdk8s/v2"
	"github.com/purecdk8s/purecdk8s/jsii"
)

func TestForwardPortCreatesDeployment(t *testing.T) {
	outdir := t.TempDir()
	app := cdk8s.NewApp(&cdk8s.AppProps{Outdir: jsii.String(outdir)})
	chart := cdk8s.NewChart(app, jsii.String("plus36"), nil)
	deployment := NewDeployment(chart, jsii.String("web"), &DeploymentProps{
		Containers: &[]*ContainerProps{{Image: jsii.String("nginx")}},
	})
	if deployment.ApiVersion() == nil || *deployment.ApiVersion() != "apps/v1" {
		t.Fatalf("unexpected deployment API version: %v", deployment.ApiVersion())
	}
	app.Synth()
	if _, err := os.Stat(filepath.Join(outdir, "plus36.k8s.yaml")); err != nil {
		t.Fatalf("synthesize 1.36 deployment: %v", err)
	}
}
