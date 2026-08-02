package cdk8splus34_test

import (
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	kplus "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

func ingressSpec(t *testing.T, chart cdk8s.Chart) map[string]interface{} {
	t.Helper()
	return mapAt(t, manifestOfKind(t, chart, "Ingress"), "spec")
}

func ingressService(chart cdk8s.Chart, ports ...float64) kplus.Service {
	service := kplus.NewService(chart, jsii.String("my-service"), nil)
	for _, port := range ports {
		service.Bind(jsii.Number(port), nil)
	}
	return service
}

func ingressServiceBackend(service kplus.Service, port *float64) kplus.IngressBackend {
	if port == nil {
		return kplus.IngressBackend_FromService(service, nil)
	}
	return kplus.IngressBackend_FromService(service, &kplus.ServiceIngressBackendOptions{Port: port})
}

func ingressExpectedServiceBackend(service kplus.Service, port float64) map[string]interface{} {
	return map[string]interface{}{
		"service": map[string]interface{}{
			"name": stringValue(service.Name()),
			"port": map[string]interface{}{"number": port},
		},
	}
}

func ingressDefaultBackend(t *testing.T, chart cdk8s.Chart) map[string]interface{} {
	t.Helper()
	return mapAt(t, ingressSpec(t, chart), "defaultBackend")
}

func ingressRules(t *testing.T, chart cdk8s.Chart) []interface{} {
	t.Helper()
	return sliceAt(t, ingressSpec(t, chart), "rules")
}

func TestIngress(t *testing.T) {
	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/ingress.test.ts#L5
	t.Run("defaultChild", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart, 80)
		kplus.NewIngress(chart, jsii.String("Ingress"), &kplus.IngressProps{DefaultBackend: ingressServiceBackend(service, nil)})
		if got := manifestOfKind(t, chart, "Ingress")["kind"]; got != "Ingress" {
			t.Fatalf("kind = %#v, want Ingress", got)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/ingress.test.ts#L15
	t.Run("IngressClassName can be set", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart, 80)
		kplus.NewIngress(chart, jsii.String("my-ingress"), &kplus.IngressProps{
			DefaultBackend: ingressServiceBackend(service, nil),
			ClassName:      jsii.String("myIngressClassName"),
		})
		spec := ingressSpec(t, chart)
		if spec["ingressClassName"] != "myIngressClassName" {
			t.Fatalf("ingressClassName = %#v", spec["ingressClassName"])
		}
		requireDeepEqual(t, spec["defaultBackend"], ingressExpectedServiceBackend(service, 80))
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/ingress.test.ts#L47
	t.Run("IngressBackend fromService uses the exposed port", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart, 8899)
		kplus.NewIngress(chart, jsii.String("Ingress"), &kplus.IngressProps{DefaultBackend: ingressServiceBackend(service, nil)})
		requireDeepEqual(t, ingressDefaultBackend(t, chart), ingressExpectedServiceBackend(service, 8899))
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/ingress.test.ts#L64
	t.Run("IngressBackend fromService fails if the service does not expose a port", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart)
		requirePanicContains(t, "service does not expose any ports", func() {
			kplus.IngressBackend_FromService(service, nil)
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/ingress.test.ts#L74
	t.Run("IngressBackend fromService rejects a different explicit port", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart, 6011)
		requirePanicContains(t, "backend defines port 7766 but service exposes port 6011", func() {
			ingressServiceBackend(service, jsii.Number(7766))
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/ingress.test.ts#L86
	t.Run("IngressBackend fromService accepts the same explicit port", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart, 6011)
		kplus.NewIngress(chart, jsii.String("Ingress"), &kplus.IngressProps{DefaultBackend: ingressServiceBackend(service, jsii.Number(6011))})
		requireDeepEqual(t, ingressDefaultBackend(t, chart), ingressExpectedServiceBackend(service, 6011))
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/ingress.test.ts#L103
	t.Run("IngressBackend fromService accepts one of multiple ports", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart, 6011, 8899, 1011)
		kplus.NewIngress(chart, jsii.String("Ingress"), &kplus.IngressProps{DefaultBackend: ingressServiceBackend(service, jsii.Number(8899))})
		requireDeepEqual(t, ingressDefaultBackend(t, chart), ingressExpectedServiceBackend(service, 8899))
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/ingress.test.ts#L122
	t.Run("IngressBackend fromService requires a port for a multi-port service", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart, 6011, 1111)
		requirePanicContains(t, "unable to determine service port since service exposes multiple ports", func() {
			kplus.IngressBackend_FromService(service, nil)
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/ingress.test.ts#L135
	t.Run("IngressBackend fromService rejects a port absent from a multi-port service", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart, 6011, 1111)
		requirePanicContains(t, "service exposes ports 6011,1111 but backend is defined to use port 1234", func() {
			ingressServiceBackend(service, jsii.Number(1234))
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/ingress.test.ts#L149
	t.Run("IngressBackend fromResource", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart, 80)
		kplus.NewIngress(chart, jsii.String("Ingress"), &kplus.IngressProps{DefaultBackend: kplus.IngressBackend_FromResource(service)})
		requireDeepEqual(t, ingressDefaultBackend(t, chart), map[string]interface{}{
			"resource": map[string]interface{}{"apiGroup": "core", "kind": "Service", "name": stringValue(service.Name())},
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/ingress.test.ts#L171
	t.Run("Ingress defaultBackend property", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart, 80)
		kplus.NewIngress(chart, jsii.String("my-ingress"), &kplus.IngressProps{DefaultBackend: ingressServiceBackend(service, nil)})
		requireDeepEqual(t, ingressDefaultBackend(t, chart), ingressExpectedServiceBackend(service, 80))
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/ingress.test.ts#L199
	t.Run("Ingress addDefaultBackend", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart, 80)
		ingress := kplus.NewIngress(chart, jsii.String("my-ingress"), nil)
		ingress.AddDefaultBackend(ingressServiceBackend(service, nil))
		requireDeepEqual(t, ingressDefaultBackend(t, chart), ingressExpectedServiceBackend(service, 80))
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/ingress.test.ts#L228
	t.Run("Ingress addHostDefaultBackend", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart, 80)
		ingress := kplus.NewIngress(chart, jsii.String("my-ingress"), nil)
		ingress.AddHostDefaultBackend(jsii.String("my.host"), ingressServiceBackend(service, nil))
		rules := ingressRules(t, chart)
		requireDeepEqual(t, rules, []interface{}{map[string]interface{}{
			"host": "my.host",
			"http": map[string]interface{}{"paths": []interface{}{map[string]interface{}{
				"path": "/", "pathType": "Prefix", "backend": ingressExpectedServiceBackend(service, 80),
			}}},
		}})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/ingress.test.ts#L266
	t.Run("Ingress addHostRule", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart, 80)
		backend := ingressServiceBackend(service, nil)
		ingress := kplus.NewIngress(chart, jsii.String("my-ingress"), nil)
		ingress.AddHostRule(jsii.String("my.host"), jsii.String("/foo"), backend, "")
		ingress.AddHostRule(jsii.String("my.host"), jsii.String("/bar"), backend, "")
		ingress.AddHostRule(jsii.String("your.host"), jsii.String("/"), backend, "")
		rules := ingressRules(t, chart)
		if len(rules) != 2 {
			t.Fatalf("rules = %d, want 2", len(rules))
		}
		first := mapValue(t, rules[0])
		if first["host"] != "my.host" {
			t.Fatalf("first host = %#v", first["host"])
		}
		paths := sliceAt(t, first, "http", "paths")
		if mapValue(t, paths[0])["path"] != "/bar" || mapValue(t, paths[1])["path"] != "/foo" {
			t.Fatalf("sorted paths = %#v", paths)
		}
		if mapValue(t, rules[1])["host"] != "your.host" {
			t.Fatalf("second host = %#v", mapValue(t, rules[1])["host"])
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/ingress.test.ts#L335
	t.Run("Ingress addRule", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart, 80)
		backend := ingressServiceBackend(service, nil)
		ingress := kplus.NewIngress(chart, jsii.String("my-ingress"), nil)
		ingress.AddRule(jsii.String("/foo"), backend, "")
		ingress.AddRule(jsii.String("/foo/bar"), backend, "")
		rules := ingressRules(t, chart)
		if len(rules) != 1 {
			t.Fatalf("rules = %d, want 1", len(rules))
		}
		paths := sliceAt(t, mapValue(t, rules[0]), "http", "paths")
		if mapValue(t, paths[0])["path"] != "/foo" || mapValue(t, paths[1])["path"] != "/foo/bar" {
			t.Fatalf("paths = %#v", paths)
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/ingress.test.ts#L385
	t.Run("Ingress define rules upon initialization", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart, 4000)
		backend := ingressServiceBackend(service, nil)
		rules := []*kplus.IngressRule{
			{Backend: backend},
			{Host: jsii.String("foo.bar"), Backend: backend},
			{Path: jsii.String("/just/path"), Backend: backend},
			{Host: jsii.String("host.and"), Path: jsii.String("/path"), Backend: backend},
			{Host: jsii.String("host.and"), Path: jsii.String("/path/2"), Backend: backend},
		}
		kplus.NewIngress(chart, jsii.String("my-ingress"), &kplus.IngressProps{Rules: &rules})
		spec := ingressSpec(t, chart)
		requireDeepEqual(t, spec["defaultBackend"], ingressExpectedServiceBackend(service, 4000))
		gotRules := sliceAt(t, spec, "rules")
		if len(gotRules) != 3 {
			t.Fatalf("rules = %d, want 3", len(gotRules))
		}
		if mapValue(t, gotRules[0])["host"] != "foo.bar" || mapValue(t, gotRules[2])["host"] != "host.and" {
			t.Fatalf("rules grouped in unexpected order: %#v", gotRules)
		}
		if len(sliceAt(t, mapValue(t, gotRules[2]), "http", "paths")) != 2 {
			t.Fatalf("host.and paths = %#v", sliceAt(t, mapValue(t, gotRules[2]), "http", "paths"))
		}
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/ingress.test.ts#L446
	t.Run("Ingress rejects duplicate addDefaultBackend", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart, 4000)
		backend := ingressServiceBackend(service, nil)
		ingress := kplus.NewIngress(chart, jsii.String("ingress"), &kplus.IngressProps{DefaultBackend: backend})
		requirePanicContains(t, "a default backend is already defined for this ingress", func() { ingress.AddDefaultBackend(backend) })
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/ingress.test.ts#L462
	t.Run("Ingress rejects defaultBackend and default rule", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart, 4000)
		backend := ingressServiceBackend(service, nil)
		rules := []*kplus.IngressRule{{Backend: backend}}
		requirePanicContains(t, "a default backend is already defined for this ingress", func() {
			kplus.NewIngress(chart, jsii.String("ingress"), &kplus.IngressProps{DefaultBackend: backend, Rules: &rules})
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/ingress.test.ts#L475
	t.Run("Ingress rejects duplicate path without host", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart, 4000)
		backend := ingressServiceBackend(service, nil)
		ingress := kplus.NewIngress(chart, jsii.String("ingress"), nil)
		ingress.AddRule(jsii.String("/foo"), backend, "")
		requirePanicContains(t, "there is already an ingress rule for /foo", func() { ingress.AddRule(jsii.String("/foo"), backend, "") })
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/ingress.test.ts#L487
	t.Run("Ingress rejects duplicate path and host", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart, 4000)
		backend := ingressServiceBackend(service, nil)
		ingress := kplus.NewIngress(chart, jsii.String("ingress"), nil)
		ingress.AddHostRule(jsii.String("hello.io"), jsii.String("/foo"), backend, "")
		requirePanicContains(t, "there is already an ingress rule for hello.io/foo", func() {
			ingress.AddHostRule(jsii.String("hello.io"), jsii.String("/foo"), backend, "")
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/ingress.test.ts#L500
	t.Run("Ingress rejects a path without leading slash", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart, 4000)
		ingress := kplus.NewIngress(chart, jsii.String("ingress"), nil)
		requirePanicContains(t, "ingress paths must begin with a \"/\": bad/path", func() {
			ingress.AddRule(jsii.String("bad/path"), ingressServiceBackend(service, nil), "")
		})
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/ingress.test.ts#L511
	t.Run("Ingress fails if no rules or default backend are specified", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		kplus.NewIngress(chart, jsii.String("ingress"), nil)
		requirePanicContains(t, "ingress with no rules or default backend", func() { synth(t, chart) })
	})

	assertTLS := func(t *testing.T, chart cdk8s.Chart, secret kplus.Secret) {
		t.Helper()
		tls := sliceAt(t, ingressSpec(t, chart), "tls")
		requireDeepEqual(t, tls, []interface{}{map[string]interface{}{
			"hosts": []interface{}{"my.host"}, "secretName": stringValue(secret.Name()),
		}})
	}

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/ingress.test.ts#L522
	t.Run("Ingress addTls", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart, 80)
		secret := kplus.NewSecret(chart, jsii.String("my-tls-secret"), &kplus.SecretProps{Metadata: &cdk8s.ApiObjectMetadata{Name: jsii.String("my-tls-secret")}})
		ingress := kplus.NewIngress(chart, jsii.String("my-ingress"), nil)
		ingress.AddHostDefaultBackend(jsii.String("my.host"), ingressServiceBackend(service, nil))
		tls := []*kplus.IngressTls{{Hosts: &[]*string{jsii.String("my.host")}, Secret: secret}}
		ingress.AddTls(&tls)
		assertTLS(t, chart, secret)
	})

	// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/ingress.test.ts#L571
	t.Run("Ingress defines tls upon initialization", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		service := ingressService(chart, 80)
		secret := kplus.NewSecret(chart, jsii.String("my-tls-secret"), &kplus.SecretProps{Metadata: &cdk8s.ApiObjectMetadata{Name: jsii.String("my-tls-secret")}})
		tls := []*kplus.IngressTls{{Hosts: &[]*string{jsii.String("my.host")}, Secret: secret}}
		ingress := kplus.NewIngress(chart, jsii.String("my-ingress"), &kplus.IngressProps{Tls: &tls})
		ingress.AddHostDefaultBackend(jsii.String("my.host"), ingressServiceBackend(service, nil))
		assertTLS(t, chart, secret)
	})
}
