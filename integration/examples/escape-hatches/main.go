package main

import (
	"example.com/purecdk8s-integration/imports/k8s"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func main() {
	app := cdk8s.NewApp(nil)
	chart := cdk8s.NewChart(app, jsii.String("escape-hatches"), nil)
	data := map[string]*string{
		"message": jsii.String("before-patch"),
	}
	config := k8s.NewKubeConfigMap(chart, jsii.String("config"), &k8s.KubeConfigMapProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String("escape-hatch-config"),
		},
		Data: &data,
	})

	config.AddJsonPatch(
		cdk8s.JsonPatch_Add(jsii.String("/metadata/labels"), map[string]interface{}{
			"escape-hatch": "enabled",
		}),
		cdk8s.JsonPatch_Replace(jsii.String("/data/message"), "after-patch"),
	)

	app.Synth()
}
