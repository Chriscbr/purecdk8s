package cdk8splus34

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/purecdk8s/purecdk8s/cdk8s/v2"
	"github.com/purecdk8s/purecdk8s/jsii"
)

func TestDeploymentServiceConfigMapAndVolume(t *testing.T) {
	outdir := t.TempDir()
	app := cdk8s.NewApp(&cdk8s.AppProps{Outdir: jsii.String(outdir)})
	chart := cdk8s.NewChart(app, jsii.String("plus"), nil)

	config := NewConfigMap(chart, jsii.String("settings"), &ConfigMapProps{
		Data: &map[string]*string{"mode": jsii.String("native")},
	})
	volume := Volume_FromConfigMap(chart, jsii.String("settings-volume"), config, nil)
	deployment := NewDeployment(chart, jsii.String("web"), &DeploymentProps{
		Containers: &[]*ContainerProps{{
			Image:      jsii.String("nginx:1.27"),
			PortNumber: jsii.Number(8080),
		}},
	})
	container := (*deployment.Containers())[0]
	container.Mount(jsii.String("/etc/settings"), volume, nil)
	deployment.ExposeViaService(nil)
	secret := NewSecret(chart, jsii.String("credentials"), &SecretProps{
		StringData: &map[string]*string{"password": jsii.String("test")},
	})
	serviceAccount := NewServiceAccount(chart, jsii.String("runner"), &ServiceAccountProps{Secrets: &[]ISecret{secret}})
	NewNamespace(chart, jsii.String("workloads"), nil)
	NewPod(chart, jsii.String("worker"), &PodProps{
		Containers:     &[]*ContainerProps{{Image: jsii.String("busybox"), Command: &[]*string{jsii.String("sleep"), jsii.String("3600")}}},
		ServiceAccount: serviceAccount,
	})
	NewJob(chart, jsii.String("migration"), &JobProps{Containers: &[]*ContainerProps{{Image: jsii.String("busybox")}}})
	NewCronJob(chart, jsii.String("cleanup"), &CronJobProps{Schedule: jsii.String("0 * * * *"), Containers: &[]*ContainerProps{{Image: jsii.String("busybox")}}})

	app.Synth()
	data, err := os.ReadFile(filepath.Join(outdir, "plus.k8s.yaml"))
	if err != nil {
		t.Fatalf("read synthesized manifest: %v", err)
	}
	manifest := string(data)
	for _, expected := range []string{
		"kind: ConfigMap",
		"mode: native",
		"kind: Deployment",
		"mountPath: /etc/settings",
		"kind: Service",
		"targetPort: 8080",
		"kind: Secret",
		"kind: ServiceAccount",
		"kind: Namespace",
		"kind: Pod",
		"kind: Job",
		"kind: CronJob",
	} {
		if !strings.Contains(manifest, expected) {
			t.Errorf("manifest does not contain %q:\n%s", expected, manifest)
		}
	}
}
