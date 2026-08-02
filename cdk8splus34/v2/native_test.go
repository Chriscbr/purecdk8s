package cdk8splus34_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	. "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
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
	NewCronJob(chart, jsii.String("cleanup"), &CronJobProps{Schedule: cdk8s.Cron_Hourly(), Containers: &[]*ContainerProps{{Image: jsii.String("busybox")}}})

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

func TestApiResourceDescriptors(t *testing.T) {
	deployment := ApiResource_DEPLOYMENTS()
	if got, want := stringValue(deployment.ApiGroup()), "apps"; got != want {
		t.Fatalf("api group = %q, want %q", got, want)
	}
	if got, want := stringValue(deployment.ResourceType()), "deployments"; got != want {
		t.Fatalf("resource type = %q, want %q", got, want)
	}
	if deployment.AsApiResource() != deployment || deployment.AsNonApiResource() != nil {
		t.Fatal("API resource endpoint conversion is incorrect")
	}
	nonAPI := NonApiResource_Of(jsii.String("/healthz"))
	if nonAPI.AsApiResource() != nil || stringValue(nonAPI.AsNonApiResource()) != "/healthz" {
		t.Fatal("non-API resource endpoint conversion is incorrect")
	}
}

func TestProbeAndLifecycleManifest(t *testing.T) {
	manifest := synthesizedContainer(t, &ContainerProps{
		Image:      jsii.String("app"),
		PortNumber: jsii.Number(8080),
		Liveness: Probe_FromHttpGet(jsii.String("/health"), &HttpGetProbeOptions{
			Scheme:      ConnectionScheme_HTTPS,
			HttpHeaders: &[]*HttpHeader{{Name: jsii.String("X-Test"), Value: jsii.String("ok")}},
		}),
		Readiness: Probe_FromGrpc(&GrpcProbeOptions{Service: jsii.String("readiness")}),
		Lifecycle: &ContainerLifecycle{
			PostStart: Handler_FromCommand(&[]*string{jsii.String("setup")}),
			PreStop:   Handler_FromTcpSocket(nil),
		},
	})
	liveness := manifest["livenessProbe"].(map[string]interface{})["httpGet"].(map[string]interface{})
	if got, want := liveness["path"], "/health"; got != want {
		t.Fatalf("HTTP probe path = %q, want %q", got, want)
	}
	if got, want := liveness["scheme"], "HTTPS"; got != want {
		t.Fatalf("HTTP probe scheme = %#v, want %q", got, want)
	}
	grpc := manifest["readinessProbe"].(map[string]interface{})["grpc"].(map[string]interface{})
	if got, want := grpc["service"], "readiness"; got != want {
		t.Fatalf("gRPC probe service = %q, want %q", got, want)
	}
	lifecycle := manifest["lifecycle"].(map[string]interface{})
	if lifecycle["postStart"] == nil || lifecycle["preStop"] == nil {
		t.Fatalf("lifecycle handlers missing from %#v", lifecycle)
	}
}

func TestPodDnsSecuritySchedulingAndDaemonSetManifest(t *testing.T) {
	outdir := t.TempDir()
	app := cdk8s.NewApp(&cdk8s.AppProps{Outdir: jsii.String(outdir)})
	chart := cdk8s.NewChart(app, jsii.String("options"), nil)
	pod := NewPod(chart, jsii.String("dns-pod"), &PodProps{
		Containers:      &[]*ContainerProps{{Image: jsii.String("busybox")}},
		Dns:             &PodDnsProps{Policy: DnsPolicy_NONE, Nameservers: &[]*string{jsii.String("1.1.1.1")}, Searches: &[]*string{jsii.String("example.test")}},
		SecurityContext: &PodSecurityContextProps{User: jsii.Number(1000), Sysctls: &[]*Sysctl{{Name: jsii.String("net.ipv4.ip_forward"), Value: jsii.String("1")}}},
	})
	pod.Scheduling().Attract(Node_Labeled(NodeLabelQuery_Is(jsii.String("disktype"), jsii.String("ssd"))), nil)
	NewDaemonSet(chart, jsii.String("agent"), &DaemonSetProps{Containers: &[]*ContainerProps{{Image: jsii.String("agent")}}, MinReadySeconds: jsii.Number(5)})
	app.Synth()
	data, err := os.ReadFile(filepath.Join(outdir, "options.k8s.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	manifest := string(data)
	for _, expected := range []string{"dnsPolicy: None", "nameservers:", "1.1.1.1", "runAsUser: 1000", "net.ipv4.ip_forward", "nodeAffinity:", "kind: DaemonSet", "minReadySeconds: 5"} {
		if !strings.Contains(manifest, expected) {
			t.Errorf("manifest does not contain %q:\n%s", expected, manifest)
		}
	}
}

func TestSourceBackedRbacHpaServiceAndEmptyDirOptions(t *testing.T) {
	outdir := t.TempDir()
	app := cdk8s.NewApp(&cdk8s.AppProps{Outdir: jsii.String(outdir)})
	chart := cdk8s.NewChart(app, jsii.String("surface"), nil)

	serviceAccount := ServiceAccount_FromServiceAccountName(chart, jsii.String("runner"), jsii.String("runner"), &FromServiceAccountNameOptions{NamespaceName: jsii.String("kube-system")})
	role := NewRole(chart, jsii.String("reader"), nil)
	role.AllowRead(serviceAccount)
	role.Bind(serviceAccount)

	clusterRole := NewClusterRole(chart, jsii.String("admin"), nil)
	clusterRole.BindInNamespace(jsii.String("apps"), serviceAccount)

	volume := Volume_FromEmptyDir(chart, jsii.String("cache"), jsii.String("cache"), &EmptyDirVolumeOptions{SizeLimit: cdk8s.Size_Mebibytes(jsii.Number(2))})
	deployment := NewDeployment(chart, jsii.String("web"), &DeploymentProps{Containers: &[]*ContainerProps{{Image: jsii.String("web"), VolumeMounts: &[]*VolumeMount{{Path: jsii.String("/cache"), Volume: volume}}}}})
	autoscaler := NewHorizontalPodAutoscaler(chart, jsii.String("autoscaler"), &HorizontalPodAutoscalerProps{Target: deployment, MaxReplicas: jsii.Number(3)})
	if autoscaler.ScaleDown() == nil || autoscaler.ScaleUp() == nil {
		t.Fatal("autoscaler scaling rules are not available")
	}
	NewService(chart, jsii.String("web-service"), &ServiceProps{
		Ports:                    &[]*ServicePort{{Port: jsii.Number(80)}},
		LoadBalancerSourceRanges: &[]*string{jsii.String("10.0.0.0/8")},
		PublishNotReadyAddresses: jsii.Bool(true),
	})

	app.Synth()
	data, err := os.ReadFile(filepath.Join(outdir, "surface.k8s.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	manifest := string(data)
	for _, expected := range []string{
		"namespace: kube-system",
		"namespace: apps",
		"sizeLimit: 2Mi",
		"loadBalancerSourceRanges:",
		"10.0.0.0/8",
		"publishNotReadyAddresses: true",
	} {
		if !strings.Contains(manifest, expected) {
			t.Errorf("manifest does not contain %q:\n%s", expected, manifest)
		}
	}
}
