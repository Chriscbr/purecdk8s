package cdk8splus34_test

import (
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	plus "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

func rbacRule(apiGroup, resource string, resourceNames []interface{}, verbs ...string) map[string]interface{} {
	result := map[string]interface{}{
		"apiGroups": []interface{}{apiGroup},
		"resources": []interface{}{resource},
	}
	verbValues := make([]interface{}, len(verbs))
	for i, verb := range verbs {
		verbValues[i] = verb
	}
	result["verbs"] = verbValues
	if resourceNames != nil {
		result["resourceNames"] = resourceNames
	}
	return result
}

func roleRules(t *testing.T, chart cdk8s.Chart, kind string) []interface{} {
	t.Helper()
	rules, ok := manifestOfKind(t, chart, kind)["rules"].([]interface{})
	if !ok {
		t.Fatalf("%s rules have unexpected type", kind)
	}
	return rules
}

func TestRole(t *testing.T) {
	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role.test.ts#L8
	t.Run("can grant permissions on imported", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		imported := plus.Role_FromRoleName(chart, jsii.String("R2"), jsii.String("r2"))
		role := plus.NewRole(chart, jsii.String("Role"), nil)
		role.AllowRead(imported)
		requireDeepEqual(t, roleRules(t, chart, "Role"), []interface{}{rbacRule("rbac.authorization.k8s.io", "roles", []interface{}{"r2"}, "get", "list", "watch")})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role.test.ts#L20
	t.Run("defaultChild", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		role := plus.NewRole(chart, jsii.String("pod-reader"), nil)
		if got := stringValue(cdk8s.ApiObject_Of(role).Kind()); got != "Role" {
			t.Fatalf("default child kind = %q, want Role", got)
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role.test.ts#L35
	t.Run("minimal definition", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewRole(chart, jsii.String("pod-reader"), nil)
		requireDeepEqual(t, synth(t, chart), []interface{}{map[string]interface{}{
			"apiVersion": "rbac.authorization.k8s.io/v1", "kind": "Role",
			"metadata": map[string]interface{}{"name": "test-pod-reader-c8ec1643"}, "rules": []interface{}{},
		}})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role.test.ts#L58
	t.Run("with a custom resource rule", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		role := plus.NewRole(chart, jsii.String("pod-reader"), nil)
		role.Allow(&[]*string{jsii.String("get"), jsii.String("watch"), jsii.String("list")}, plus.ApiResource_Custom(&plus.ApiResourceOptions{ApiGroup: jsii.String(""), ResourceType: jsii.String("pods")}))
		requireDeepEqual(t, roleRules(t, chart, "Role"), []interface{}{rbacRule("", "pods", nil, "get", "watch", "list")})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role.test.ts#L82
	t.Run("read access to a pod and secret", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("my-pod"), &plus.PodProps{Containers: &[]*plus.ContainerProps{{Image: jsii.String("nginx")}}})
		secret := plus.NewSecret(chart, jsii.String("Secret"), &plus.SecretProps{StringData: &map[string]*string{"key": jsii.String("value")}, Type: jsii.String("kubernetes.io/tls")})
		role := plus.NewRole(chart, jsii.String("pod-reader"), nil)
		role.AllowRead(pod, secret)
		requireDeepEqual(t, roleRules(t, chart, "Role"), []interface{}{
			rbacRule("", "pods", []interface{}{stringValue(pod.Name())}, "get", "list", "watch"),
			rbacRule("", "secrets", []interface{}{stringValue(secret.Name())}, "get", "list", "watch"),
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role.test.ts#L123
	t.Run("read access to a mix of resources", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("my-pod"), &plus.PodProps{Containers: &[]*plus.ContainerProps{{Image: jsii.String("nginx")}}})
		role := plus.NewRole(chart, jsii.String("my-role"), nil)
		role.AllowRead(plus.ApiResource_SECRETS(), pod)
		requireDeepEqual(t, roleRules(t, chart, "Role"), []interface{}{
			rbacRule("", "secrets", nil, "get", "list", "watch"),
			rbacRule("", "pods", []interface{}{"test-my-pod-c8a0e457"}, "get", "list", "watch"),
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role.test.ts#L157
	t.Run("specify access from props", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewRole(chart, jsii.String("my-role"), &plus.RoleProps{Rules: &[]*plus.RolePolicyRule{{
			Verbs:     &[]*string{jsii.String("get"), jsii.String("list"), jsii.String("watch")},
			Resources: &[]plus.IApiResource{plus.ApiResource_PODS()},
		}}})
		requireDeepEqual(t, roleRules(t, chart, "Role"), []interface{}{rbacRule("", "pods", nil, "get", "list", "watch")})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role.test.ts#L183
	t.Run("specific pod and all pods still permits all", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("my-pod"), &plus.PodProps{Containers: &[]*plus.ContainerProps{{Image: jsii.String("nginx")}}})
		role := plus.NewRole(chart, jsii.String("my-role"), nil)
		role.AllowRead(plus.ApiResource_PODS(), pod)
		rules := roleRules(t, chart, "Role")
		if len(rules) != 2 {
			t.Fatalf("rules = %d, want 2", len(rules))
		}
		requireDeepEqual(t, rules[0], rbacRule("", "pods", nil, "get", "list", "watch"))
		if _, exists := rules[0].(map[string]interface{})["resourceNames"]; exists {
			t.Fatal("all-pods rule unexpectedly restricts resourceNames")
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role.test.ts#L219
	t.Run("read all pods and secrets in a namespace", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		role := plus.NewRole(chart, jsii.String("pod-reader"), nil)
		role.AllowRead(plus.ApiResource_SECRETS(), plus.ApiResource_PODS())
		requireDeepEqual(t, roleRules(t, chart, "Role"), []interface{}{
			rbacRule("", "secrets", nil, "get", "list", "watch"),
			rbacRule("", "pods", nil, "get", "list", "watch"),
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role.test.ts#L245
	t.Run("read custom resource type", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		role := plus.NewRole(chart, jsii.String("pod-reader"), nil)
		role.AllowRead(plus.ApiResource_Custom(&plus.ApiResourceOptions{ApiGroup: jsii.String(""), ResourceType: jsii.String("pods/log")}))
		requireDeepEqual(t, roleRules(t, chart, "Role"), []interface{}{rbacRule("", "pods/log", nil, "get", "list", "watch")})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role.test.ts#L266
	t.Run("specific resource and resource type", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: &[]*plus.ContainerProps{{Image: jsii.String("image")}}})
		role := plus.NewRole(chart, jsii.String("pod-reader"), nil)
		role.Allow(&[]*string{jsii.String("get")}, pod, plus.ApiResource_PODS())
		requireDeepEqual(t, roleRules(t, chart, "Role"), []interface{}{
			rbacRule("", "pods", []interface{}{"test-pod-c890e1b8"}, "get"),
			rbacRule("", "pods", nil, "get"),
		})
	})
}

func TestClusterRole(t *testing.T) {
	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role.test.ts#L290
	t.Run("can grant permissions on imported", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		imported := plus.ClusterRole_FromClusterRoleName(chart, jsii.String("R2"), jsii.String("r2"))
		role := plus.NewRole(chart, jsii.String("Role"), nil)
		role.AllowRead(imported)
		requireDeepEqual(t, roleRules(t, chart, "Role"), []interface{}{rbacRule("rbac.authorization.k8s.io", "clusterroles", []interface{}{"r2"}, "get", "list", "watch")})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role.test.ts#L302
	t.Run("defaultChild", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		role := plus.NewClusterRole(chart, jsii.String("my-cluster-role"), nil)
		if got := stringValue(cdk8s.ApiObject_Of(role).Kind()); got != "ClusterRole" {
			t.Fatalf("default child kind = %q, want ClusterRole", got)
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role.test.ts#L317
	t.Run("minimal definition", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewClusterRole(chart, jsii.String("my-cluster-role"), nil)
		requireDeepEqual(t, synth(t, chart), []interface{}{map[string]interface{}{
			"apiVersion": "rbac.authorization.k8s.io/v1", "kind": "ClusterRole",
			"metadata": map[string]interface{}{"name": "test-my-cluster-role-c86cea4f"}, "rules": []interface{}{},
		}})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role.test.ts#L341
	t.Run("with a custom resource rule", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		role := plus.NewClusterRole(chart, jsii.String("pod-reader"), nil)
		role.Allow(&[]*string{jsii.String("get"), jsii.String("watch"), jsii.String("list")}, plus.ApiResource_Custom(&plus.ApiResourceOptions{ApiGroup: jsii.String(""), ResourceType: jsii.String("pods")}))
		requireDeepEqual(t, roleRules(t, chart, "ClusterRole"), []interface{}{rbacRule("", "pods", []interface{}{}, "get", "watch", "list")})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role.test.ts#L366
	t.Run("with custom non-resource rules", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		role := plus.NewClusterRole(chart, jsii.String("pod-reader"), nil)
		verbs := &[]*string{jsii.String("get"), jsii.String("post")}
		role.Allow(verbs, plus.NonApiResource_Of(jsii.String("/healthz")), plus.NonApiResource_Of(jsii.String("/healthz/*")))
		requireDeepEqual(t, roleRules(t, chart, "ClusterRole"), []interface{}{
			map[string]interface{}{"nonResourceURLs": []interface{}{string("/healthz")}, "verbs": []interface{}{"get", "post"}},
			map[string]interface{}{"nonResourceURLs": []interface{}{string("/healthz/*")}, "verbs": []interface{}{"get", "post"}},
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role.test.ts#L390
	t.Run("read access to a pod and secret", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("my-pod"), &plus.PodProps{Containers: &[]*plus.ContainerProps{{Image: jsii.String("nginx")}}})
		secret := plus.NewSecret(chart, jsii.String("Secret"), &plus.SecretProps{StringData: &map[string]*string{"key": jsii.String("value")}, Type: jsii.String("kubernetes.io/tls")})
		role := plus.NewClusterRole(chart, jsii.String("my-cluster-role"), nil)
		role.AllowRead(pod, secret)
		requireDeepEqual(t, roleRules(t, chart, "ClusterRole"), []interface{}{
			rbacRule("", "pods", []interface{}{stringValue(pod.Name())}, "get", "list", "watch"),
			rbacRule("", "secrets", []interface{}{stringValue(secret.Name())}, "get", "list", "watch"),
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role.test.ts#L430
	t.Run("read all pods and secrets in cluster", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		role := plus.NewClusterRole(chart, jsii.String("my-cluster-role"), nil)
		role.AllowRead(plus.ApiResource_SECRETS(), plus.ApiResource_PODS())
		requireDeepEqual(t, roleRules(t, chart, "ClusterRole"), []interface{}{
			rbacRule("", "secrets", []interface{}{}, "get", "list", "watch"),
			rbacRule("", "pods", []interface{}{}, "get", "list", "watch"),
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role.test.ts#L457
	t.Run("read custom resource type", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		role := plus.NewClusterRole(chart, jsii.String("pod-reader"), nil)
		role.AllowRead(plus.ApiResource_Custom(&plus.ApiResourceOptions{ApiGroup: jsii.String(""), ResourceType: jsii.String("pods/log")}))
		requireDeepEqual(t, roleRules(t, chart, "ClusterRole"), []interface{}{rbacRule("", "pods/log", []interface{}{}, "get", "list", "watch")})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role.test.ts#L479
	t.Run("specific resource type and non-resource endpoint", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		pod := plus.NewPod(chart, jsii.String("Pod"), &plus.PodProps{Containers: &[]*plus.ContainerProps{{Image: jsii.String("image")}}})
		role := plus.NewClusterRole(chart, jsii.String("pod-reader"), nil)
		role.Allow(&[]*string{jsii.String("get")}, pod, plus.ApiResource_PODS(), plus.NonApiResource_Of(jsii.String("/healthz")))
		requireDeepEqual(t, roleRules(t, chart, "ClusterRole"), []interface{}{
			rbacRule("", "pods", []interface{}{"test-pod-c890e1b8"}, "get"),
			rbacRule("", "pods", []interface{}{}, "get"),
			map[string]interface{}{"nonResourceURLs": []interface{}{string("/healthz")}, "verbs": []interface{}{"get"}},
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role.test.ts#L500
	t.Run("can be aggregated", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		role1 := plus.NewClusterRole(chart, jsii.String("pod-reader"), nil)
		role1.Allow(&[]*string{jsii.String("get"), jsii.String("watch"), jsii.String("list")}, plus.ApiResource_PODS())
		role2 := plus.NewClusterRole(chart, jsii.String("secrets-reader"), nil)
		role2.Allow(&[]*string{jsii.String("get"), jsii.String("watch"), jsii.String("list")}, plus.ApiResource_SECRETS())
		combined := plus.NewClusterRole(chart, jsii.String("combined-role"), nil)
		combined.Combine(role1)
		combined.Combine(role2)
		manifests := synth(t, chart)
		label := map[string]interface{}{"cdk8s.cluster-role/aggregate-to-test-combined-role-c8db37c0": "true"}
		requireDeepEqual(t, manifests[0].(map[string]interface{})["metadata"].(map[string]interface{})["labels"], label)
		requireDeepEqual(t, manifests[1].(map[string]interface{})["metadata"].(map[string]interface{})["labels"], label)
		requireDeepEqual(t, manifests[2].(map[string]interface{})["aggregationRule"], map[string]interface{}{
			"clusterRoleSelectors": []interface{}{map[string]interface{}{"matchLabels": label}},
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role.test.ts#L542
	t.Run("custom aggregation labels", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		role := plus.NewClusterRole(chart, jsii.String("secrets-reader"), &plus.ClusterRoleProps{AggregationLabels: &map[string]*string{
			"rbac.authorization.k8s.io/aggregate-to-view": jsii.String("true"),
		}})
		role.Aggregate(jsii.String("rbac.authorization.k8s.io/aggregate-to-admin"), jsii.String("true"))
		role.Allow(&[]*string{jsii.String("get"), jsii.String("watch"), jsii.String("list")}, plus.ApiResource_SECRETS())
		requireDeepEqual(t, manifestOfKind(t, chart, "ClusterRole")["aggregationRule"], map[string]interface{}{
			"clusterRoleSelectors": []interface{}{map[string]interface{}{"matchLabels": map[string]interface{}{
				"rbac.authorization.k8s.io/aggregate-to-admin": "true",
				"rbac.authorization.k8s.io/aggregate-to-view":  "true",
			}}},
		})
	})
}
