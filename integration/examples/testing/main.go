package main

import (
	"example.com/purecdk8s-integration/imports/k8s"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func addFixture(scope constructs.Construct) {
	data := map[string]*string{
		"message": jsii.String("hello from testing"),
	}
	k8s.NewKubeConfigMap(scope, jsii.String("config"), &k8s.KubeConfigMapProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String("testing-config"),
		},
		Data: &data,
	})
}

func NewChart(scope constructs.Construct, id string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), nil)
	addFixture(chart)
	return chart
}

func main() {
	app := cdk8s.NewApp(nil)
	NewChart(app, "testing")
	app.Synth()
}
