package cdk8splus34_test

import (
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	plus "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/job.test.ts#L5
func TestJobDefaultChild(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	job := plus.NewJob(chart, jsii.String("Job"), nil)
	child := job.Node().DefaultChild()
	if child == nil || stringValue(cdk8s.ApiObject_Of(child).Kind()) != "Job" {
		t.Fatal("Job default child is not a Job ApiObject")
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/job.test.ts#L15
func TestJobAllowsSettingAllOptions(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	job := plus.NewJob(chart, jsii.String("Job"), &plus.JobProps{
		Containers:       &[]*plus.ContainerProps{{Image: jsii.String("image")}},
		ActiveDeadline:   cdk8s.Duration_Seconds(jsii.Number(20)),
		BackoffLimit:     jsii.Number(4),
		TtlAfterFinished: cdk8s.Duration_Seconds(jsii.Number(1)),
	})
	spec := mapAt(t, manifestAt(t, chart, 0), "spec")
	if got := spec["activeDeadlineSeconds"]; got != float64(20) {
		t.Errorf("activeDeadlineSeconds = %#v, want 20", got)
	}
	if got := spec["backoffLimit"]; got != numberValue(job.BackoffLimit()) {
		t.Errorf("backoffLimit = %#v, want %v", got, numberValue(job.BackoffLimit()))
	}
	if got := spec["ttlSecondsAfterFinished"]; got != float64(1) {
		t.Errorf("ttlSecondsAfterFinished = %#v, want 1", got)
	}
	if got := job.RestartPolicy(); got != plus.RestartPolicy_NEVER {
		t.Errorf("RestartPolicy() = %q, want NEVER", got)
	}
	if got := numberValue(job.ActiveDeadline().ToSeconds(nil)); got != 20 {
		t.Errorf("ActiveDeadline() = %v seconds, want 20", got)
	}
	if got := numberValue(job.BackoffLimit()); got != 4 {
		t.Errorf("BackoffLimit() = %v, want 4", got)
	}
	if got := numberValue(job.TtlAfterFinished().ToSeconds(nil)); got != 1 {
		t.Errorf("TtlAfterFinished() = %v seconds, want 1", got)
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/job.test.ts#L41
func TestJobAppliesDefaultRestartPolicyToPodSpec(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	job := plus.NewJob(chart, jsii.String("Job"), &plus.JobProps{
		Containers:       &[]*plus.ContainerProps{{Image: jsii.String("image")}},
		TtlAfterFinished: cdk8s.Duration_Seconds(jsii.Number(1)),
	})
	if got := mapAt(t, manifestAt(t, chart, 0), "spec", "template", "spec")["restartPolicy"]; got != "Never" {
		t.Fatalf("pod restartPolicy = %#v, want Never", got)
	}
	if got := job.RestartPolicy(); got != plus.RestartPolicy_NEVER {
		t.Fatalf("Job RestartPolicy() = %q, want NEVER", got)
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/job.test.ts#L59
func TestJobDoesNotModifyExistingRestartPolicy(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	job := plus.NewJob(chart, jsii.String("Job"), &plus.JobProps{
		Containers:       &[]*plus.ContainerProps{{Image: jsii.String("image")}},
		RestartPolicy:    plus.RestartPolicy_ALWAYS,
		TtlAfterFinished: cdk8s.Duration_Seconds(jsii.Number(1)),
	})
	if got := mapAt(t, manifestAt(t, chart, 0), "spec", "template", "spec")["restartPolicy"]; got != "Always" {
		t.Fatalf("pod restartPolicy = %#v, want Always", got)
	}
	if got := job.RestartPolicy(); got != plus.RestartPolicy_ALWAYS {
		t.Fatalf("Job RestartPolicy() = %q, want ALWAYS", got)
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/job.test.ts#L78
func TestJobSynthesizesSpecLazily(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	job := plus.NewJob(chart, jsii.String("Job"), nil)
	job.AddContainer(&plus.ContainerProps{Image: jsii.String("image")})
	containers := sliceAt(t, manifestAt(t, chart, 0), "spec", "template", "spec", "containers")
	if got := mapAt(t, containers[0])["image"]; got != "image" {
		t.Fatalf("container image = %#v, want image", got)
	}
}

// Ported from: https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/job.test.ts#L95
func TestJobCanBeIsolated(t *testing.T) {
	chart := cdk8s.Testing_Chart()
	plus.NewJob(chart, jsii.String("Job"), &plus.JobProps{
		Containers: &[]*plus.ContainerProps{{Image: jsii.String("foobar")}}, Isolate: jsii.Bool(true),
	})
	manifests := synth(t, chart)
	requireSnapshotHash(t, manifests, "f1c651e8390cb400a3fafe28f05be584e84118a3da59945ef06771eb8a0f689e")
	policy := mapAt(t, manifests[1], "spec")
	if labels := mapAt(t, policy, "podSelector", "matchLabels"); len(labels) == 0 {
		t.Fatal("isolating Job produced no pod selector labels")
	}
	requireDeepEqual(t, policy["policyTypes"], []interface{}{"Egress", "Ingress"})
}
