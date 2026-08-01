package main

import (
	"github.com/aws/jsii-runtime-go"
	"github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2"
	cdk8splus34 "github.com/cdk8s-team/cdk8s-plus-go/cdk8splus34/v2"
)

func main() {
	app := cdk8s.NewApp(nil)
	chart := cdk8s.NewChart(app, jsii.String("plus"), nil)
	settings := cdk8splus34.NewConfigMap(chart, jsii.String("settings"), &cdk8splus34.ConfigMapProps{
		Data: &map[string]*string{"mode": jsii.String("native")},
	})
	settingsVolume := cdk8splus34.Volume_FromConfigMap(chart, jsii.String("settings-volume"), settings, nil)

	deployment := cdk8splus34.NewDeployment(chart, jsii.String("web"), &cdk8splus34.DeploymentProps{
		Replicas: jsii.Number(3),
		Containers: &[]*cdk8splus34.ContainerProps{{
			Image:      jsii.String("nginx:1.27"),
			PortNumber: jsii.Number(8080),
		}},
	})
	(*deployment.Containers())[0].Mount(jsii.String("/etc/settings"), settingsVolume, nil)

	deployment.ExposeViaService(&cdk8splus34.DeploymentExposeViaServiceOptions{
		ServiceType: cdk8splus34.ServiceType_LOAD_BALANCER,
	})

	app.Synth()
}
