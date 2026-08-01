package main

import (
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func main() {
	app := cdk8s.NewApp(nil)
	chart := cdk8s.NewChart(app, jsii.String("helm"), nil)
	values := map[string]interface{}{
		"message": "hello from local helm",
	}

	cdk8s.NewHelm(chart, jsii.String("local-chart"), &cdk8s.HelmProps{
		Chart:       jsii.String("./chart"),
		ReleaseName: jsii.String("local-release"),
		Values:      &values,
	})

	app.Synth()
}
