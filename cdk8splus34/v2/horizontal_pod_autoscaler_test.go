package cdk8splus34_test

import (
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	kplus "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

func hpaDeployment(chart cdk8s.Chart, props ...*kplus.ContainerProps) kplus.Deployment {
	containers := props
	if len(containers) == 0 {
		containers = []*kplus.ContainerProps{{Image: jsii.String("pod")}}
	}
	return kplus.NewDeployment(chart, jsii.String("Deployment"), &kplus.DeploymentProps{Containers: &containers})
}

func hpaSpec(t *testing.T, chart cdk8s.Chart) map[string]interface{} {
	t.Helper()
	return mapAt(t, manifestOfKind(t, chart, "HorizontalPodAutoscaler"), "spec")
}

func newHPA(chart cdk8s.Chart, target kplus.IScalable, metrics *[]kplus.Metric, options ...func(*kplus.HorizontalPodAutoscalerProps)) kplus.HorizontalPodAutoscaler {
	props := &kplus.HorizontalPodAutoscalerProps{Target: target, MaxReplicas: jsii.Number(10), Metrics: metrics}
	for _, option := range options {
		option(props)
	}
	return kplus.NewHorizontalPodAutoscaler(chart, jsii.String("Hpa"), props)
}

func hpaExpectedResourceMetric(name string, target map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"type": "Resource",
		"resource": map[string]interface{}{
			"name": name, "target": target,
		},
	}
}

func assertHPAMetric(t *testing.T, makeMetric func(kplus.Deployment) kplus.Metric, expected map[string]interface{}) {
	t.Helper()
	chart := cdk8s.Testing_Chart()
	deployment := hpaDeployment(chart)
	metrics := []kplus.Metric{makeMetric(deployment)}
	newHPA(chart, deployment, &metrics)
	spec := hpaSpec(t, chart)
	if spec["maxReplicas"] != float64(10) {
		t.Fatalf("maxReplicas = %#v, want 10", spec["maxReplicas"])
	}
	requireDeepEqual(t, spec["metrics"], []interface{}{expected})
}

func hpaScalingPolicy(replicas kplus.Replicas, duration cdk8s.Duration) *kplus.ScalingPolicy {
	return &kplus.ScalingPolicy{Replicas: replicas, Duration: duration}
}

func hpaDefaultScaleUp() map[string]interface{} {
	return map[string]interface{}{
		"policies": []interface{}{
			map[string]interface{}{"periodSeconds": float64(60), "type": "Pods", "value": float64(4)},
			map[string]interface{}{"periodSeconds": float64(60), "type": "Percent", "value": float64(200)},
		},
		"selectPolicy": "Max", "stabilizationWindowSeconds": float64(0),
	}
}

func hpaDefaultScaleDown(min float64) map[string]interface{} {
	return map[string]interface{}{
		"policies":     []interface{}{map[string]interface{}{"periodSeconds": float64(300), "type": "Pods", "value": min}},
		"selectPolicy": "Max", "stabilizationWindowSeconds": float64(300),
	}
}

func TestHorizontalPodAutoscaler(t *testing.T) {
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L6
	t.Run("targets a deployment that has containers with volume mounts", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		deployment := kplus.NewDeployment(chart, jsii.String("deployment"), nil)
		volume := kplus.Volume_FromEmptyDir(chart, jsii.String("test-volume"), jsii.String("test"), nil)
		mounts := []*kplus.VolumeMount{{Volume: volume, Path: jsii.String("./test")}}
		deployment.AddContainer(&kplus.ContainerProps{Image: jsii.String("ubuntu"), Name: jsii.String("test"), VolumeMounts: &mounts})
		kplus.NewHorizontalPodAutoscaler(chart, jsii.String("HPA"), &kplus.HorizontalPodAutoscalerProps{Target: deployment, MaxReplicas: jsii.Number(5)})
		deploymentSpec := mapAt(t, manifestOfKind(t, chart, "Deployment"), "spec")
		containers := sliceAt(t, deploymentSpec, "template", "spec", "containers")
		mount := mapValue(t, sliceAt(t, mapValue(t, containers[0]), "volumeMounts")[0])
		if mount["mountPath"] != "./test" || mount["name"] != "test" {
			t.Fatalf("volume mount = %#v", mount)
		}
		if _, exists := deploymentSpec["replicas"]; exists {
			t.Fatal("an autoscaled deployment must not synthesize default replicas")
		}
		target := mapAt(t, hpaSpec(t, chart), "scaleTargetRef")
		if target["kind"] != "Deployment" || target["name"] != stringValue(deployment.Name()) {
			t.Fatalf("scale target = %#v", target)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L30
	t.Run("defaultChild", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		deployment := hpaDeployment(chart)
		newHPA(chart, deployment, nil)
		if manifestOfKind(t, chart, "HorizontalPodAutoscaler")["kind"] != "HorizontalPodAutoscaler" {
			t.Fatal("default child did not synthesize a HorizontalPodAutoscaler")
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L43
	t.Run("default configuration", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		deployment := hpaDeployment(chart)
		newHPA(chart, deployment, nil)
		spec := hpaSpec(t, chart)
		if spec["maxReplicas"] != float64(10) || spec["minReplicas"] != float64(1) {
			t.Fatalf("replica bounds = %#v/%#v", spec["minReplicas"], spec["maxReplicas"])
		}
		if _, exists := spec["metrics"]; exists {
			t.Fatalf("default metrics = %#v, want absent", spec["metrics"])
		}
		behavior := mapAt(t, spec, "behavior")
		requireDeepEqual(t, behavior["scaleUp"], hpaDefaultScaleUp())
		requireDeepEqual(t, behavior["scaleDown"], hpaDefaultScaleDown(1))
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L83
	t.Run("creates default policies when all other scaling options are provided", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		deployment := hpaDeployment(chart)
		newHPA(chart, deployment, nil, func(props *kplus.HorizontalPodAutoscalerProps) {
			props.ScaleUp = &kplus.ScalingRules{StabilizationWindow: cdk8s.Duration_Minutes(jsii.Number(5)), Strategy: kplus.ScalingStrategy_MAX_CHANGE}
			props.ScaleDown = &kplus.ScalingRules{StabilizationWindow: cdk8s.Duration_Minutes(jsii.Number(5)), Strategy: kplus.ScalingStrategy_MAX_CHANGE}
		})
		behavior := mapAt(t, hpaSpec(t, chart), "behavior")
		scaleUp := hpaDefaultScaleUp()
		scaleUp["stabilizationWindowSeconds"] = float64(300)
		requireDeepEqual(t, behavior["scaleUp"], scaleUp)
		requireDeepEqual(t, behavior["scaleDown"], hpaDefaultScaleDown(1))
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L132
	t.Run("provided policies default to 15 second periods", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		deployment := hpaDeployment(chart)
		policies := []*kplus.ScalingPolicy{
			hpaScalingPolicy(kplus.Replicas_Absolute(jsii.Number(3)), nil),
			hpaScalingPolicy(kplus.Replicas_Percent(jsii.Number(30)), nil),
		}
		newHPA(chart, deployment, nil, func(props *kplus.HorizontalPodAutoscalerProps) {
			props.ScaleUp = &kplus.ScalingRules{Policies: &policies}
			props.ScaleDown = &kplus.ScalingRules{Policies: &policies}
		})
		expectedPolicies := []interface{}{
			map[string]interface{}{"periodSeconds": float64(15), "type": "Pods", "value": float64(3)},
			map[string]interface{}{"periodSeconds": float64(15), "type": "Percent", "value": float64(30)},
		}
		behavior := mapAt(t, hpaSpec(t, chart), "behavior")
		requireDeepEqual(t, mapAt(t, behavior, "scaleUp")["policies"], expectedPolicies)
		requireDeepEqual(t, mapAt(t, behavior, "scaleDown")["policies"], expectedPolicies)
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L183
	t.Run("supports different scale-up and scale-down strategies", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		deployment := hpaDeployment(chart)
		metrics := []kplus.Metric{kplus.Metric_ResourceCpu(kplus.MetricTarget_AverageUtilization(jsii.Number(50)))}
		upPolicies := []*kplus.ScalingPolicy{hpaScalingPolicy(kplus.Replicas_Absolute(jsii.Number(3)), cdk8s.Duration_Minutes(jsii.Number(3)))}
		downPolicies := []*kplus.ScalingPolicy{hpaScalingPolicy(kplus.Replicas_Absolute(jsii.Number(3)), cdk8s.Duration_Minutes(jsii.Number(3)))}
		newHPA(chart, deployment, &metrics, func(props *kplus.HorizontalPodAutoscalerProps) {
			props.MinReplicas = jsii.Number(2)
			props.ScaleUp = &kplus.ScalingRules{Policies: &upPolicies, StabilizationWindow: cdk8s.Duration_Minutes(jsii.Number(5)), Strategy: kplus.ScalingStrategy_MAX_CHANGE}
			props.ScaleDown = &kplus.ScalingRules{Policies: &downPolicies, StabilizationWindow: cdk8s.Duration_Minutes(jsii.Number(5)), Strategy: kplus.ScalingStrategy_MIN_CHANGE}
		})
		behavior := mapAt(t, hpaSpec(t, chart), "behavior")
		if mapAt(t, behavior, "scaleUp")["selectPolicy"] != "Max" || mapAt(t, behavior, "scaleDown")["selectPolicy"] != "Min" {
			t.Fatalf("scaling strategies = %#v", behavior)
		}
	})

	containerMetric := func(name string, makeMetric func(*kplus.MetricContainerResourceOptions) kplus.Metric) func(kplus.Deployment) kplus.Metric {
		return func(deployment kplus.Deployment) kplus.Metric {
			container := (*deployment.Containers())[0]
			return makeMetric(&kplus.MetricContainerResourceOptions{Container: container, Target: kplus.MetricTarget_AverageUtilization(jsii.Number(50))})
		}
	}
	expectedContainerMetric := func(name string) map[string]interface{} {
		return map[string]interface{}{
			"type": "ContainerResource",
			"containerResource": map[string]interface{}{
				"container": "main", "name": name,
				"target": map[string]interface{}{"averageUtilization": float64(50), "type": "Utilization"},
			},
		}
	}

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L249
	t.Run("creates CPU ContainerResource metric", func(t *testing.T) {
		assertHPAMetric(t, containerMetric("cpu", kplus.Metric_ContainerCpu), expectedContainerMetric("cpu"))
	})
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L282
	t.Run("creates Memory ContainerResource metric", func(t *testing.T) {
		assertHPAMetric(t, containerMetric("memory", kplus.Metric_ContainerMemory), expectedContainerMetric("memory"))
	})
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L314
	t.Run("creates Storage ContainerResource metric", func(t *testing.T) {
		assertHPAMetric(t, containerMetric("storage", kplus.Metric_ContainerStorage), expectedContainerMetric("storage"))
	})
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L346
	t.Run("creates Ephemeral Storage ContainerResource metric", func(t *testing.T) {
		assertHPAMetric(t, containerMetric("ephemeral-storage", kplus.Metric_ContainerEphemeralStorage), expectedContainerMetric("ephemeral-storage"))
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L378
	t.Run("creates external metric", func(t *testing.T) {
		assertHPAMetric(t, func(kplus.Deployment) kplus.Metric {
			labels := map[string]*string{"app": jsii.String("scraper")}
			return kplus.Metric_External(&kplus.MetricOptions{
				LabelSelector: kplus.LabelSelector_Of(&kplus.LabelSelectorOptions{Labels: &labels}),
				Name:          jsii.String("sqs-queue"), Target: kplus.MetricTarget_AverageUtilization(jsii.Number(50)),
			})
		}, map[string]interface{}{
			"type": "External",
			"external": map[string]interface{}{
				"metric": map[string]interface{}{"name": "sqs-queue", "selector": map[string]interface{}{"matchLabels": map[string]interface{}{"app": "scraper"}}},
				"target": map[string]interface{}{"averageUtilization": float64(50), "type": "Utilization"},
			},
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L417
	t.Run("creates object metric", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		deployment := hpaDeployment(chart)
		service := ingressService(chart, 80)
		ingress := kplus.NewIngress(chart, jsii.String("my-ingress"), &kplus.IngressProps{DefaultBackend: ingressServiceBackend(service, nil)})
		metric := kplus.Metric_Object(&kplus.MetricObjectOptions{Object: ingress, Name: jsii.String("requests-per-second"), Target: kplus.MetricTarget_AverageUtilization(jsii.Number(50))})
		metrics := []kplus.Metric{metric}
		newHPA(chart, deployment, &metrics)
		requireDeepEqual(t, hpaSpec(t, chart)["metrics"], []interface{}{map[string]interface{}{
			"type": "Object",
			"object": map[string]interface{}{
				"describedObject": map[string]interface{}{"apiVersion": "networking.k8s.io/v1", "kind": "Ingress", "name": stringValue(ingress.Name())},
				"metric":          map[string]interface{}{"name": "requests-per-second"},
				"target":          map[string]interface{}{"averageUtilization": float64(50), "type": "Utilization"},
			},
		}})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L460
	t.Run("creates pods metric", func(t *testing.T) {
		assertHPAMetric(t, func(kplus.Deployment) kplus.Metric {
			labels := map[string]*string{"app": jsii.String("scraper")}
			return kplus.Metric_Pods(&kplus.MetricOptions{
				Name:          jsii.String("requests-per-second"),
				Target:        kplus.MetricTarget_AverageUtilization(jsii.Number(50)),
				LabelSelector: kplus.LabelSelector_Of(&kplus.LabelSelectorOptions{Labels: &labels}),
			})
		}, map[string]interface{}{
			"type": "Pods",
			"pods": map[string]interface{}{
				"metric": map[string]interface{}{"name": "requests-per-second", "selector": map[string]interface{}{"matchLabels": map[string]interface{}{"app": "scraper"}}},
				"target": map[string]interface{}{"averageUtilization": float64(50), "type": "Utilization"},
			},
		})
	})

	resourceMetric := func(name string, target kplus.MetricTarget) func(kplus.Deployment) kplus.Metric {
		return func(kplus.Deployment) kplus.Metric {
			switch name {
			case "cpu":
				return kplus.Metric_ResourceCpu(target)
			case "memory":
				return kplus.Metric_ResourceMemory(target)
			case "storage":
				return kplus.Metric_ResourceStorage(target)
			default:
				return kplus.Metric_ResourceEphemeralStorage(target)
			}
		}
	}
	utilization50 := map[string]interface{}{"averageUtilization": float64(50), "type": "Utilization"}

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L499
	t.Run("creates Resource CPU metric", func(t *testing.T) {
		assertHPAMetric(t, resourceMetric("cpu", kplus.MetricTarget_AverageUtilization(jsii.Number(50))), hpaExpectedResourceMetric("cpu", utilization50))
	})
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L527
	t.Run("creates Resource Memory metric", func(t *testing.T) {
		assertHPAMetric(t, resourceMetric("memory", kplus.MetricTarget_AverageUtilization(jsii.Number(50))), hpaExpectedResourceMetric("memory", utilization50))
	})
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L555
	t.Run("creates Resource Storage metric", func(t *testing.T) {
		assertHPAMetric(t, resourceMetric("storage", kplus.MetricTarget_AverageUtilization(jsii.Number(50))), hpaExpectedResourceMetric("storage", utilization50))
	})
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L583
	t.Run("creates Resource Ephemeral Storage metric", func(t *testing.T) {
		assertHPAMetric(t, resourceMetric("ephemeral-storage", kplus.MetricTarget_AverageUtilization(jsii.Number(50))), hpaExpectedResourceMetric("ephemeral-storage", utilization50))
	})
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L611
	t.Run("creates Resource CPU metric targeting 70 percent average utilization", func(t *testing.T) {
		target := map[string]interface{}{"averageUtilization": float64(70), "type": "Utilization"}
		assertHPAMetric(t, resourceMetric("cpu", kplus.MetricTarget_AverageUtilization(jsii.Number(70))), hpaExpectedResourceMetric("cpu", target))
	})
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L639
	t.Run("creates Resource CPU metric targeting 47.2 average value", func(t *testing.T) {
		target := map[string]interface{}{"averageValue": float64(47.2), "type": "AverageValue"}
		assertHPAMetric(t, resourceMetric("cpu", kplus.MetricTarget_AverageValue(jsii.Number(47.2))), hpaExpectedResourceMetric("cpu", target))
	})
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L667
	t.Run("creates Resource CPU metric targeting exact value 29.5", func(t *testing.T) {
		target := map[string]interface{}{"value": float64(29.5), "type": "Value"}
		assertHPAMetric(t, resourceMetric("cpu", kplus.MetricTarget_Value(jsii.Number(29.5))), hpaExpectedResourceMetric("cpu", target))
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L697
	t.Run("creates HPA when target is a deployment", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		deployment := hpaDeployment(chart)
		newHPA(chart, deployment, nil)
		if mapAt(t, hpaSpec(t, chart), "scaleTargetRef")["kind"] != "Deployment" {
			t.Fatalf("scaleTargetRef = %#v", mapAt(t, hpaSpec(t, chart), "scaleTargetRef"))
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L712
	t.Run("creates HPA when target is a StatefulSet", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart, 80)
		containers := []*kplus.ContainerProps{{Image: jsii.String("foobar")}}
		statefulSet := kplus.NewStatefulSet(chart, jsii.String("StatefulSet"), &kplus.StatefulSetProps{Select: jsii.Bool(false), Containers: &containers, Service: service})
		newHPA(chart, statefulSet, nil)
		if mapAt(t, hpaSpec(t, chart), "scaleTargetRef")["kind"] != "StatefulSet" {
			t.Fatalf("scaleTargetRef = %#v", mapAt(t, hpaSpec(t, chart), "scaleTargetRef"))
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L732
	t.Run("creates HPA when minReplicas equals maxReplicas", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		deployment := hpaDeployment(chart)
		newHPA(chart, deployment, nil, func(props *kplus.HorizontalPodAutoscalerProps) { props.MinReplicas = jsii.Number(10) })
		spec := hpaSpec(t, chart)
		if spec["minReplicas"] != float64(10) || spec["maxReplicas"] != float64(10) {
			t.Fatalf("bounds = %#v/%#v", spec["minReplicas"], spec["maxReplicas"])
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L749
	t.Run("creates expected spec when all options are configured", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		deployment := hpaDeployment(chart)
		metrics := []kplus.Metric{kplus.Metric_ResourceCpu(kplus.MetricTarget_AverageUtilization(jsii.Number(50)))}
		policies := []*kplus.ScalingPolicy{
			hpaScalingPolicy(kplus.Replicas_Absolute(jsii.Number(3)), cdk8s.Duration_Minutes(jsii.Number(3))),
			hpaScalingPolicy(kplus.Replicas_Percent(jsii.Number(30)), nil),
		}
		newHPA(chart, deployment, &metrics, func(props *kplus.HorizontalPodAutoscalerProps) {
			props.MinReplicas = jsii.Number(2)
			props.ScaleUp = &kplus.ScalingRules{Policies: &policies, StabilizationWindow: cdk8s.Duration_Minutes(jsii.Number(5)), Strategy: kplus.ScalingStrategy_MAX_CHANGE}
			props.ScaleDown = &kplus.ScalingRules{Policies: &policies, StabilizationWindow: cdk8s.Duration_Minutes(jsii.Number(5)), Strategy: kplus.ScalingStrategy_MAX_CHANGE}
		})
		spec := hpaSpec(t, chart)
		requireDeepEqual(t, spec["metrics"], []interface{}{hpaExpectedResourceMetric("cpu", utilization50)})
		expectedRules := map[string]interface{}{
			"policies": []interface{}{
				map[string]interface{}{"periodSeconds": float64(180), "type": "Pods", "value": float64(3)},
				map[string]interface{}{"periodSeconds": float64(15), "type": "Percent", "value": float64(30)},
			},
			"selectPolicy": "Max", "stabilizationWindowSeconds": float64(300),
		}
		behavior := mapAt(t, spec, "behavior")
		requireDeepEqual(t, behavior["scaleUp"], expectedRules)
		requireDeepEqual(t, behavior["scaleDown"], expectedRules)
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L821
	t.Run("creates HPA when one of two containers has resource constraints", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		deployment := hpaDeployment(chart,
			&kplus.ContainerProps{Image: jsii.String("pod1"), Resources: &kplus.ContainerResources{}},
			&kplus.ContainerProps{Image: jsii.String("pod2"), Resources: &kplus.ContainerResources{Cpu: &kplus.CpuResources{Request: kplus.Cpu_Millis(jsii.Number(256))}}},
		)
		newHPA(chart, deployment, nil)
		if hpaSpec(t, chart)["maxReplicas"] != float64(10) {
			t.Fatalf("HPA spec = %#v", hpaSpec(t, chart))
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L857
	t.Run("rejects missing metrics when target container has no resource constraints", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		deployment := hpaDeployment(chart, &kplus.ContainerProps{Image: jsii.String("pod"), Resources: &kplus.ContainerResources{}})
		newHPA(chart, deployment, nil)
		requirePanicContains(t, "every container in the target must have a CPU or memory resource constraint defined", func() { synth(t, chart) })
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L879
	t.Run("rejects minReplicas greater than maxReplicas", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		deployment := hpaDeployment(chart)
		requirePanicContains(t, "'minReplicas' (11) must be less than or equal to 'maxReplicas' (10)", func() {
			newHPA(chart, deployment, nil, func(props *kplus.HorizontalPodAutoscalerProps) { props.MinReplicas = jsii.Number(11) })
		})
	})

	stabilizationFailure := func(t *testing.T, direction string, duration cdk8s.Duration) {
		t.Helper()
		chart := cdk8s.Testing_Chart()
		deployment := hpaDeployment(chart)
		requirePanicContains(t, "must be 0 seconds or more with a max of 1 hour", func() {
			newHPA(chart, deployment, nil, func(props *kplus.HorizontalPodAutoscalerProps) {
				rules := &kplus.ScalingRules{StabilizationWindow: duration}
				if direction == "scaleUp" {
					props.ScaleUp = rules
				} else {
					props.ScaleDown = rules
				}
			})
		})
	}

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L892
	t.Run("rejects scaleUp stabilizationWindow over one hour", func(t *testing.T) { stabilizationFailure(t, "scaleUp", cdk8s.Duration_Hours(jsii.Number(2))) })
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L908
	t.Run("rejects scaleDown stabilizationWindow over one hour", func(t *testing.T) { stabilizationFailure(t, "scaleDown", cdk8s.Duration_Hours(jsii.Number(2))) })

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L923
	t.Run("rejects negative scaleUp stabilizationWindow", func(t *testing.T) {
		requirePanicContains(t, "Duration amounts cannot be negative. Received: -1", func() { cdk8s.Duration_Seconds(jsii.Number(-1)) })
	})
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L938
	t.Run("rejects negative scaleDown stabilizationWindow", func(t *testing.T) {
		requirePanicContains(t, "Duration amounts cannot be negative. Received: -1", func() { cdk8s.Duration_Seconds(jsii.Number(-1)) })
	})

	policyDurationFailure := func(t *testing.T, direction string, duration cdk8s.Duration, expected string) {
		t.Helper()
		chart := cdk8s.Testing_Chart()
		deployment := hpaDeployment(chart)
		policies := []*kplus.ScalingPolicy{hpaScalingPolicy(kplus.Replicas_Absolute(jsii.Number(3)), duration)}
		requirePanicContains(t, expected, func() {
			newHPA(chart, deployment, nil, func(props *kplus.HorizontalPodAutoscalerProps) {
				rules := &kplus.ScalingRules{Policies: &policies}
				if direction == "scaleUp" {
					props.ScaleUp = rules
				} else {
					props.ScaleDown = rules
				}
			})
		})
	}

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L953
	t.Run("rejects scaleUp policy duration over 30 minutes", func(t *testing.T) {
		policyDurationFailure(t, "scaleUp", cdk8s.Duration_Minutes(jsii.Number(31)), "duration (31 minutes) is outside of the allowed range")
	})
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L973
	t.Run("rejects scaleDown policy duration over 30 minutes", func(t *testing.T) {
		policyDurationFailure(t, "scaleDown", cdk8s.Duration_Minutes(jsii.Number(31)), "duration (31 minutes) is outside of the allowed range")
	})
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L993
	t.Run("rejects zero scaleUp policy duration", func(t *testing.T) {
		policyDurationFailure(t, "scaleUp", cdk8s.Duration_Minutes(jsii.Number(0)), "duration (0 minutes) is outside of the allowed range")
	})
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L1013
	t.Run("rejects zero scaleDown policy duration", func(t *testing.T) {
		policyDurationFailure(t, "scaleDown", cdk8s.Duration_Minutes(jsii.Number(0)), "duration (0 minutes) is outside of the allowed range")
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L1033
	t.Run("rejects negative scaleUp policy duration", func(t *testing.T) {
		requirePanicContains(t, "Duration amounts cannot be negative. Received: -10", func() { cdk8s.Duration_Minutes(jsii.Number(-10)) })
	})
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L1053
	t.Run("rejects negative scaleDown policy duration", func(t *testing.T) {
		requirePanicContains(t, "Duration amounts cannot be negative. Received: -10", func() { cdk8s.Duration_Minutes(jsii.Number(-10)) })
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L1073
	t.Run("rejects Deployment target with fixed replicas", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		containers := []*kplus.ContainerProps{{Image: jsii.String("pod")}}
		deployment := kplus.NewDeployment(chart, jsii.String("Deployment"), &kplus.DeploymentProps{Containers: &containers, Replicas: jsii.Number(3)})
		newHPA(chart, deployment, nil)
		requirePanicContains(t, "target cannot have a fixed number of replicas (3)", func() { synth(t, chart) })
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/horizontal-pod-autoscaler.test.ts#L1089
	t.Run("rejects StatefulSet target with fixed replicas", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart, 80)
		containers := []*kplus.ContainerProps{{Image: jsii.String("foobar")}}
		statefulSet := kplus.NewStatefulSet(chart, jsii.String("StatefulSet"), &kplus.StatefulSetProps{Select: jsii.Bool(false), Containers: &containers, Service: service, Replicas: jsii.Number(5)})
		newHPA(chart, statefulSet, nil)
		requirePanicContains(t, "target cannot have a fixed number of replicas (5)", func() { synth(t, chart) })
	})
}
