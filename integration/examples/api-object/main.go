package main

import (
	"example.com/purecdk8s-integration/imports/k8s"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func NewChart(scope constructs.Construct, id string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		DisableResourceNameHashes: jsii.Bool(true),
	})

	gvk := k8s.KubeConfigMap_GVK()
	object := cdk8s.NewApiObject(chart, jsii.String("configuration"), &cdk8s.ApiObjectProps{
		ApiVersion: gvk.ApiVersion,
		Kind:       gvk.Kind,
		Metadata: &cdk8s.ApiObjectMetadata{
			Name: jsii.String("application-config"),
		},
	})
	object.Metadata().AddLabel(jsii.String("app"), jsii.String("api-object-example"))
	object.Metadata().AddAnnotation(jsii.String("example.com/source"), jsii.String("cdk8s"))

	return chart
}

func main() {
	app := cdk8s.NewApp(nil)
	NewChart(app, "api-object")
	app.Synth()
}
