package cdk8splus34_test

import (
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	plus "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

func TestProbe(t *testing.T) {
	probeManifest := func(t *testing.T, port float64, probe plus.Probe) map[string]interface{} {
		t.Helper()
		container := synthesizedContainer(t, &plus.ContainerProps{Image: jsii.String("foobar"), Port: jsii.Number(port), Liveness: probe})
		result, ok := container["livenessProbe"].(map[string]interface{})
		if !ok {
			t.Fatalf("livenessProbe has type %T, want map[string]interface{}", container["livenessProbe"])
		}
		return result
	}

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/probe.test.ts#L5
	t.Run("HTTP defaults to the container port", func(t *testing.T) {
		got := probeManifest(t, 5555, plus.Probe_FromHttpGet(jsii.String("/hello"), nil))
		requireDeepEqual(t, got, map[string]interface{}{
			"failureThreshold": float64(3),
			"httpGet":          map[string]interface{}{"path": "/hello", "port": float64(5555), "scheme": "HTTP"},
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/probe.test.ts#L27
	t.Run("HTTP specific port", func(t *testing.T) {
		got := probeManifest(t, 5555, plus.Probe_FromHttpGet(jsii.String("/hello"), &plus.HttpGetProbeOptions{Port: jsii.Number(1234)}))
		requireDeepEqual(t, got, map[string]interface{}{
			"failureThreshold": float64(3),
			"httpGet":          map[string]interface{}{"path": "/hello", "port": float64(1234), "scheme": "HTTP"},
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/probe.test.ts#L49
	t.Run("HTTP options", func(t *testing.T) {
		got := probeManifest(t, 5555, plus.Probe_FromHttpGet(jsii.String("/hello"), &plus.HttpGetProbeOptions{
			FailureThreshold:    jsii.Number(11),
			InitialDelaySeconds: cdk8s.Duration_Minutes(jsii.Number(1)),
			PeriodSeconds:       cdk8s.Duration_Seconds(jsii.Number(5)),
			SuccessThreshold:    jsii.Number(3),
			TimeoutSeconds:      cdk8s.Duration_Minutes(jsii.Number(2)),
			Host:                jsii.String("1.1.1.1"),
			HttpHeaders: &[]*plus.HttpHeader{{
				Name: jsii.String("A-Custom-Header"), Value: jsii.String("some-value"),
			}},
		}))
		requireDeepEqual(t, got, map[string]interface{}{
			"failureThreshold":    float64(11),
			"initialDelaySeconds": float64(60),
			"periodSeconds":       float64(5),
			"successThreshold":    float64(3),
			"timeoutSeconds":      float64(120),
			"httpGet": map[string]interface{}{
				"path": "/hello", "port": float64(5555), "scheme": "HTTP", "host": "1.1.1.1",
				"httpHeaders": []interface{}{map[string]interface{}{"name": "A-Custom-Header", "value": "some-value"}},
			},
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/probe.test.ts#L93
	t.Run("command minimal usage", func(t *testing.T) {
		got := probeManifest(t, 5555, plus.Probe_FromCommand(&[]*string{jsii.String("foo"), jsii.String("bar")}, nil))
		requireDeepEqual(t, got, map[string]interface{}{
			"exec":             map[string]interface{}{"command": []interface{}{"foo", "bar"}},
			"failureThreshold": float64(3),
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/probe.test.ts#L111
	t.Run("command options", func(t *testing.T) {
		got := probeManifest(t, 5555, plus.Probe_FromCommand(&[]*string{jsii.String("foo"), jsii.String("bar")}, &plus.CommandProbeOptions{
			FailureThreshold:    jsii.Number(11),
			InitialDelaySeconds: cdk8s.Duration_Minutes(jsii.Number(1)),
			PeriodSeconds:       cdk8s.Duration_Seconds(jsii.Number(5)),
			SuccessThreshold:    jsii.Number(3),
			TimeoutSeconds:      cdk8s.Duration_Minutes(jsii.Number(2)),
		}))
		requireDeepEqual(t, got, map[string]interface{}{
			"exec":                map[string]interface{}{"command": []interface{}{"foo", "bar"}},
			"failureThreshold":    float64(11),
			"initialDelaySeconds": float64(60),
			"periodSeconds":       float64(5),
			"successThreshold":    float64(3),
			"timeoutSeconds":      float64(120),
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/probe.test.ts#L139
	t.Run("TCP minimal usage", func(t *testing.T) {
		got := probeManifest(t, 5555, plus.Probe_FromTcpSocket(nil))
		requireDeepEqual(t, got, map[string]interface{}{
			"tcpSocket":        map[string]interface{}{"port": float64(5555)},
			"failureThreshold": float64(3),
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/probe.test.ts#L160
	t.Run("TCP specific port and hostname", func(t *testing.T) {
		got := probeManifest(t, 5555, plus.Probe_FromTcpSocket(&plus.TcpSocketProbeOptions{Port: jsii.Number(8080), Host: jsii.String("hostname")}))
		requireDeepEqual(t, got, map[string]interface{}{
			"tcpSocket":        map[string]interface{}{"port": float64(8080), "host": "hostname"},
			"failureThreshold": float64(3),
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/probe.test.ts#L184
	t.Run("TCP options", func(t *testing.T) {
		got := probeManifest(t, 5555, plus.Probe_FromTcpSocket(&plus.TcpSocketProbeOptions{
			FailureThreshold:    jsii.Number(11),
			InitialDelaySeconds: cdk8s.Duration_Minutes(jsii.Number(1)),
			PeriodSeconds:       cdk8s.Duration_Seconds(jsii.Number(5)),
			SuccessThreshold:    jsii.Number(3),
			TimeoutSeconds:      cdk8s.Duration_Minutes(jsii.Number(2)),
		}))
		requireDeepEqual(t, got, map[string]interface{}{
			"tcpSocket":           map[string]interface{}{"port": float64(5555)},
			"failureThreshold":    float64(11),
			"initialDelaySeconds": float64(60),
			"periodSeconds":       float64(5),
			"successThreshold":    float64(3),
			"timeoutSeconds":      float64(120),
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/probe.test.ts#L213
	t.Run("gRPC minimal usage", func(t *testing.T) {
		got := probeManifest(t, 5555, plus.Probe_FromGrpc(&plus.GrpcProbeOptions{Port: jsii.Number(5555)}))
		requireDeepEqual(t, got, map[string]interface{}{
			"grpc":             map[string]interface{}{"port": float64(5555)},
			"failureThreshold": float64(3),
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/probe.test.ts#L233
	t.Run("gRPC options", func(t *testing.T) {
		got := probeManifest(t, 5555, plus.Probe_FromGrpc(&plus.GrpcProbeOptions{
			Port:                jsii.Number(5555),
			FailureThreshold:    jsii.Number(11),
			InitialDelaySeconds: cdk8s.Duration_Minutes(jsii.Number(1)),
			PeriodSeconds:       cdk8s.Duration_Seconds(jsii.Number(5)),
			SuccessThreshold:    jsii.Number(3),
			TimeoutSeconds:      cdk8s.Duration_Minutes(jsii.Number(2)),
		}))
		requireDeepEqual(t, got, map[string]interface{}{
			"grpc":                map[string]interface{}{"port": float64(5555)},
			"failureThreshold":    float64(11),
			"initialDelaySeconds": float64(60),
			"periodSeconds":       float64(5),
			"successThreshold":    float64(3),
			"timeoutSeconds":      float64(120),
		})
	})
}
