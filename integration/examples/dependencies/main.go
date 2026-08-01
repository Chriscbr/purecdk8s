package main

import (
	"example.com/purecdk8s-integration/imports/k8s"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
)

func main() {
	app := cdk8s.NewApp(nil)

	namespaceChart := cdk8s.NewChart(app, jsii.String("namespace"), nil)
	applicationChart := cdk8s.NewChart(app, jsii.String("application"), nil)

	namespace := k8s.NewKubeNamespace(namespaceChart, jsii.String("namespace"), &k8s.KubeNamespaceProps{
		Metadata: &k8s.ObjectMeta{
			Name: jsii.String("integration"),
		},
	})
	service := k8s.NewKubeService(applicationChart, jsii.String("service"), &k8s.KubeServiceProps{
		Metadata: &k8s.ObjectMeta{
			Name:      jsii.String("application"),
			Namespace: namespace.Name(),
		},
	})

	// ApiObject dependencies determine resource order. Across charts, cdk8s
	// also infers the corresponding chart dependency.
	service.AddDependency(namespace)

	// Charts support explicit ordering through the same construct dependency
	// mechanism. This is intentionally redundant with the inferred ordering
	// above so the example exercises both APIs.
	applicationChart.AddDependency(namespaceChart)

	app.Synth()
}
