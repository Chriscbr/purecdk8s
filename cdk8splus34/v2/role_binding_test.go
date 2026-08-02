package cdk8splus34_test

import (
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	plus "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

func rbacTestSubjects(chart cdk8s.Chart) (plus.User, plus.Group) {
	return plus.User_FromName(chart, jsii.String("Alice"), jsii.String("alice@example.com")),
		plus.Group_FromName(chart, jsii.String("FrontendAdmins"), jsii.String("frontend-admins"))
}

func requireBindingSubjects(t *testing.T, manifest map[string]interface{}, want ...map[string]interface{}) {
	t.Helper()
	values := make([]interface{}, len(want))
	for i := range want {
		values[i] = want[i]
	}
	requireDeepEqual(t, manifest["subjects"], values)
}

func TestRoleBinding(t *testing.T) {
	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role-binding.test.ts#L4
	t.Run("can create from a Role", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		role := plus.NewRole(chart, jsii.String("pod-reader"), nil)
		role.AllowRead(plus.ApiResource_PODS())
		user, group := rbacTestSubjects(chart)
		binding := role.Bind(user, group)
		if stringValue(binding.Kind()) != "RoleBinding" || binding.Role() != role || binding.Metadata().Namespace() != nil {
			t.Fatalf("unexpected binding properties")
		}
		if got := *binding.Subjects(); len(got) != 2 || got[0] != user || got[1] != group {
			t.Fatalf("binding subjects = %#v", got)
		}
		manifest := manifestOfKind(t, chart, "RoleBinding")
		if mapAt(t, manifest, "roleRef")["kind"] != "Role" {
			t.Fatalf("roleRef = %#v", manifest["roleRef"])
		}
		requireBindingSubjects(t, manifest,
			map[string]interface{}{"apiGroup": "rbac.authorization.k8s.io", "kind": "User", "name": "alice@example.com"},
			map[string]interface{}{"apiGroup": "rbac.authorization.k8s.io", "kind": "Group", "name": "frontend-admins"},
		)
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role-binding.test.ts#L53
	t.Run("can create from a ClusterRole", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		role := plus.NewClusterRole(chart, jsii.String("pod-reader"), nil)
		role.AllowRead(plus.ApiResource_PODS())
		user, group := rbacTestSubjects(chart)
		binding := role.BindInNamespace(jsii.String("development"), user, group)
		if stringValue(binding.Kind()) != "RoleBinding" || binding.Role() != role || stringValue(binding.Metadata().Namespace()) != "development" {
			t.Fatalf("unexpected binding properties")
		}
		manifest := manifestOfKind(t, chart, "RoleBinding")
		if mapAt(t, manifest, "metadata")["namespace"] != "development" || mapAt(t, manifest, "roleRef")["kind"] != "ClusterRole" {
			t.Fatalf("binding manifest = %#v", manifest)
		}
		requireBindingSubjects(t, manifest,
			map[string]interface{}{"apiGroup": "rbac.authorization.k8s.io", "kind": "User", "name": "alice@example.com"},
			map[string]interface{}{"apiGroup": "rbac.authorization.k8s.io", "kind": "Group", "name": "frontend-admins"},
		)
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role-binding.test.ts#L104
	t.Run("can call bindInNamespace multiple times", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		role := plus.NewClusterRole(chart, jsii.String("pod-reader"), nil)
		role.AllowRead(plus.ApiResource_PODS())
		user1 := plus.User_FromName(chart, jsii.String("Alice"), jsii.String("alice@example.com"))
		user2 := plus.User_FromName(chart, jsii.String("Bob"), jsii.String("bob@example.com"))
		binding1 := role.BindInNamespace(jsii.String("staging"), user1)
		binding2 := role.BindInNamespace(jsii.String("development"), user2)
		if stringValue(binding1.Metadata().Namespace()) != "staging" || stringValue(binding2.Metadata().Namespace()) != "development" {
			t.Fatal("bindings did not retain their namespaces")
		}
		manifests := synth(t, chart)
		bindings := make([]map[string]interface{}, 0, 2)
		for _, value := range manifests {
			manifest := value.(map[string]interface{})
			if manifest["kind"] == "RoleBinding" {
				bindings = append(bindings, manifest)
			}
		}
		if len(bindings) != 2 {
			t.Fatalf("role bindings = %d, want 2", len(bindings))
		}
		if mapAt(t, bindings[0], "metadata")["namespace"] != "staging" || bindings[0]["subjects"].([]interface{})[0].(map[string]interface{})["name"] != "alice@example.com" {
			t.Fatalf("first binding = %#v", bindings[0])
		}
		if mapAt(t, bindings[1], "metadata")["namespace"] != "development" || bindings[1]["subjects"].([]interface{})[0].(map[string]interface{})["name"] != "bob@example.com" {
			t.Fatalf("second binding = %#v", bindings[1])
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role-binding.test.ts#L138
	t.Run("can create a ClusterRoleBinding", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		role := plus.NewClusterRole(chart, jsii.String("pod-reader"), nil)
		role.AllowRead(plus.ApiResource_PODS())
		user, group := rbacTestSubjects(chart)
		binding := role.Bind(user, group)
		if stringValue(binding.Kind()) != "ClusterRoleBinding" || binding.Role() != role || binding.Metadata().Namespace() != nil {
			t.Fatalf("unexpected binding properties")
		}
		manifest := manifestOfKind(t, chart, "ClusterRoleBinding")
		if mapAt(t, manifest, "roleRef")["kind"] != "ClusterRole" {
			t.Fatalf("roleRef = %#v", manifest["roleRef"])
		}
		requireBindingSubjects(t, manifest,
			map[string]interface{}{"apiGroup": "rbac.authorization.k8s.io", "kind": "User", "name": "alice@example.com"},
			map[string]interface{}{"apiGroup": "rbac.authorization.k8s.io", "kind": "Group", "name": "frontend-admins"},
		)
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role-binding.test.ts#L187
	t.Run("can bind a ServiceAccount", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		role := plus.NewRole(chart, jsii.String("pod-reader"), nil)
		role.AllowRead(plus.ApiResource_PODS())
		account := plus.NewServiceAccount(chart, jsii.String("my-service-account"), nil)
		binding := role.Bind(account)
		if binding.Role() != role || len(*binding.Subjects()) != 1 || (*binding.Subjects())[0] != account {
			t.Fatal("binding did not retain role and service account")
		}
		requireBindingSubjects(t, manifestOfKind(t, chart, "RoleBinding"), map[string]interface{}{
			"apiGroup": "", "kind": "ServiceAccount", "name": "test-my-service-account-c84bb46b",
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/role-binding.test.ts#L214
	t.Run("can add subjects after creation", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		role := plus.NewRole(chart, jsii.String("pod-reader"), nil)
		role.AllowRead(plus.ApiResource_PODS())
		user, group := rbacTestSubjects(chart)
		account := plus.NewServiceAccount(chart, jsii.String("ServiceAccount"), &plus.ServiceAccountProps{Metadata: &cdk8s.ApiObjectMetadata{Namespace: jsii.String("n1")}})
		binding := role.Bind()
		binding.AddSubjects(user, group, account)
		if len(*binding.Subjects()) != 3 {
			t.Fatalf("subjects = %d, want 3", len(*binding.Subjects()))
		}
		requireBindingSubjects(t, manifestOfKind(t, chart, "RoleBinding"),
			map[string]interface{}{"apiGroup": "rbac.authorization.k8s.io", "kind": "User", "name": "alice@example.com"},
			map[string]interface{}{"apiGroup": "rbac.authorization.k8s.io", "kind": "Group", "name": "frontend-admins"},
			map[string]interface{}{"apiGroup": "", "kind": "ServiceAccount", "name": "test-serviceaccount-c8f15383", "namespace": "n1"},
		)
	})
}
