package cdk8splus34_test

import (
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	plus "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

func synthesizedContainer(t *testing.T, props *plus.ContainerProps) map[string]interface{} {
	t.Helper()
	chart := cdk8s.Testing_Chart()
	plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: &[]*plus.ContainerProps{props}})
	containers := sliceAt(t, manifestOfKind(t, chart, "Pod"), "spec", "containers")
	if len(containers) != 1 {
		t.Fatalf("synthesized containers = %d, want 1", len(containers))
	}
	container, ok := containers[0].(map[string]interface{})
	if !ok {
		t.Fatalf("container has type %T, want map[string]interface{}", containers[0])
	}
	return container
}

func TestHandler(t *testing.T) {
	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/handler.test.ts#L3
	t.Run("fromCommand", func(t *testing.T) {
		container := synthesizedContainer(t, &plus.ContainerProps{
			Image: jsii.String("image"),
			Lifecycle: &plus.ContainerLifecycle{
				PostStart: plus.Handler_FromCommand(&[]*string{jsii.String("hello")}),
			},
		})
		handler := mapAt(t, container, "lifecycle", "postStart")
		requireDeepEqual(t, handler["exec"], map[string]interface{}{"command": []interface{}{"hello"}})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/handler.test.ts#L9
	t.Run("fromHttpGet", func(t *testing.T) {
		container := synthesizedContainer(t, &plus.ContainerProps{
			Image: jsii.String("image"),
			Lifecycle: &plus.ContainerLifecycle{
				PostStart: plus.Handler_FromHttpGet(jsii.String("/path"), nil),
			},
		})
		handler := mapAt(t, container, "lifecycle", "postStart")
		requireDeepEqual(t, handler["httpGet"], map[string]interface{}{"path": "/path", "port": float64(80), "scheme": "HTTP"})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/handler.test.ts#L15
	t.Run("fromTcpSocket", func(t *testing.T) {
		container := synthesizedContainer(t, &plus.ContainerProps{
			Image: jsii.String("image"),
			Lifecycle: &plus.ContainerLifecycle{
				PostStart: plus.Handler_FromTcpSocket(&plus.HandlerFromTcpSocketOptions{Port: jsii.Number(8888)}),
			},
		})
		handler := mapAt(t, container, "lifecycle", "postStart")
		requireDeepEqual(t, handler["tcpSocket"], map[string]interface{}{"port": float64(8888)})
	})
}
