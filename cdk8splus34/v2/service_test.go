package cdk8splus34_test

import (
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	plus "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

func TestService(t *testing.T) {
	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/service.test.ts#L5
	t.Run("defaultChild", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := plus.NewService(chart, jsii.String("Service"), nil)
		if got := stringValue(cdk8s.ApiObject_Of(service).Kind()); got != "Service" {
			t.Fatalf("default child kind = %q, want Service", got)
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/service.test.ts#L15
	t.Run("must have at least one port", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewService(chart, jsii.String("service"), nil)
		requirePanicContains(t, "A service must be configured with a port", func() { synth(t, chart) })
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/service.test.ts#L27
	t.Run("can provide cluster IP", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewService(chart, jsii.String("service"), &plus.ServiceProps{
			Ports:     &[]*plus.ServicePort{{Port: jsii.Number(9000)}},
			ClusterIP: jsii.String("3000"),
		})
		requireDeepEqual(t, manifestOfKind(t, chart, "Service")["spec"], map[string]interface{}{
			"clusterIP": "3000", "externalIPs": []interface{}{},
			"ports":    []interface{}{map[string]interface{}{"port": float64(9000)}},
			"selector": map[string]interface{}{}, "type": "ClusterIP",
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/service.test.ts#L52
	t.Run("can select a deployment", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		deployment := plus.NewDeployment(chart, jsii.String("Deployment"), &plus.DeploymentProps{Containers: &[]*plus.ContainerProps{{Image: jsii.String("image")}}})
		plus.NewService(chart, jsii.String("service"), &plus.ServiceProps{
			Ports:    &[]*plus.ServicePort{{Port: jsii.Number(9000)}},
			Selector: deployment,
		})
		service := manifestOfKind(t, chart, "Service")
		selector := mapAt(t, service, "spec", "selector")
		if got := selector["cdk8s.io/metadata.addr"]; got != "test-Deployment-c83f5e59" {
			t.Fatalf("deployment selector address = %#v", got)
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/service.test.ts#L69
	t.Run("can select by label", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := plus.NewService(chart, jsii.String("service"), &plus.ServiceProps{Ports: &[]*plus.ServicePort{{Port: jsii.Number(9000)}}})
		service.SelectLabel(jsii.String("key"), jsii.String("value"))
		requireDeepEqual(t, mapAt(t, manifestOfKind(t, chart, "Service"), "spec", "selector"), map[string]interface{}{"key": "value"})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/service.test.ts#L85
	t.Run("can serve by port", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := plus.NewService(chart, jsii.String("service"), nil)
		service.Bind(jsii.Number(9000), &plus.ServiceBindOptions{TargetPort: jsii.Number(80), NodePort: jsii.Number(30080)})
		want := []interface{}{map[string]interface{}{"port": float64(9000), "targetPort": float64(80), "nodePort": float64(30080)}}
		requireDeepEqual(t, mapAt(t, manifestOfKind(t, chart, "Service"), "spec")["ports"], want)
		ports := *service.Ports()
		if len(ports) != 1 || numberValue(ports[0].Port) != 9000 || numberValue(ports[0].TargetPort) != 80 || numberValue(ports[0].NodePort) != 30080 {
			t.Fatalf("service ports = %#v", ports)
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/service.test.ts#L102
	t.Run("synthesizes spec lazily", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := plus.NewService(chart, jsii.String("Service"), nil)
		service.Select(plus.Pods_Select(chart, jsii.String("Pods"), &plus.PodsSelectOptions{Labels: &map[string]*string{"key": jsii.String("value")}}))
		service.Bind(jsii.Number(9000), nil)
		spec := mapAt(t, manifestOfKind(t, chart, "Service"), "spec")
		requireDeepEqual(t, spec["selector"], map[string]interface{}{"key": "value"})
		requireDeepEqual(t, spec["ports"], []interface{}{map[string]interface{}{"port": float64(9000)}})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/service.test.ts#L117
	t.Run("sets external IPs", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		ips := &[]*string{jsii.String("1.1.1.1"), jsii.String("8.8.8.8")}
		service := plus.NewService(chart, jsii.String("service"), &plus.ServiceProps{ExternalIPs: ips})
		service.Bind(jsii.Number(53), nil)
		requireDeepEqual(t, mapAt(t, manifestOfKind(t, chart, "Service"), "spec")["externalIPs"], []interface{}{"1.1.1.1", "8.8.8.8"})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/service.test.ts#L130
	t.Run("external name type requires externalName", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := plus.NewService(chart, jsii.String("service"), &plus.ServiceProps{Type: plus.ServiceType_EXTERNAL_NAME})
		service.Bind(jsii.Number(5432), nil)
		requirePanicContains(t, "A service with type EXTERNAL_NAME requires an externalName prop", func() { synth(t, chart) })
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/service.test.ts#L146
	t.Run("type defaults to external name", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := plus.NewService(chart, jsii.String("service"), &plus.ServiceProps{ExternalName: jsii.String("test-external-name")})
		service.Bind(jsii.Number(5432), nil)
		if got := mapAt(t, manifestOfKind(t, chart, "Service"), "spec")["type"]; got != "ExternalName" {
			t.Fatalf("service type = %#v, want ExternalName", got)
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/service.test.ts#L162
	t.Run("can restrict load balancer source ranges", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewService(chart, jsii.String("service"), &plus.ServiceProps{
			Ports:                    &[]*plus.ServicePort{{Port: jsii.Number(80)}},
			Type:                     plus.ServiceType_LOAD_BALANCER,
			LoadBalancerSourceRanges: &[]*string{jsii.String("143.231.0.0/16")},
		})
		requireDeepEqual(t, mapAt(t, manifestOfKind(t, chart, "Service"), "spec")["loadBalancerSourceRanges"], []interface{}{"143.231.0.0/16"})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/service.test.ts#L177
	t.Run("can be exposed by an ingress", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := plus.NewService(chart, jsii.String("Service"), nil)
		service.Bind(jsii.Number(80), nil)
		service.ExposeViaIngress(jsii.String("/hello"), nil)
		ingress := manifestOfKind(t, chart, "Ingress")
		paths := sliceAt(t, ingress, "spec", "rules")
		path := paths[0].(map[string]interface{})["http"].(map[string]interface{})["paths"].([]interface{})[0].(map[string]interface{})
		if path["path"] != "/hello" || path["pathType"] != "Prefix" {
			t.Fatalf("ingress path = %#v", path)
		}
		serviceBackend := mapAt(t, path, "backend", "service")
		if serviceBackend["name"] != "test-service-c85b0531" || mapAt(t, serviceBackend, "port")["number"] != float64(80) {
			t.Fatalf("ingress backend = %#v", serviceBackend)
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/service.test.ts#L189
	t.Run("can set publishNotReadyAddresses", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewService(chart, jsii.String("service"), &plus.ServiceProps{
			Ports:                    &[]*plus.ServicePort{{Port: jsii.Number(80)}},
			PublishNotReadyAddresses: jsii.Bool(true),
		})
		if got := mapAt(t, manifestOfKind(t, chart, "Service"), "spec")["publishNotReadyAddresses"]; got != true {
			t.Fatalf("publishNotReadyAddresses = %#v, want true", got)
		}
	})
}
