package main

import (
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func main() {
	app := cdk8s.NewApp(nil)
	chart := cdk8s.NewChart(app, jsii.String("include"), nil)

	cdk8s.NewInclude(chart, jsii.String("checked-in-manifest"), &cdk8s.IncludeProps{
		Url: jsii.String("./manifest.yaml"),
	})

	app.Synth()
}
