package cdk8splus34_test

import (
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	plus "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

func TestServiceAccount(t *testing.T) {
	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/service-account.test.ts#L5
	t.Run("can grant permissions on imported", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		account := plus.ServiceAccount_FromServiceAccountName(chart, jsii.String("ServiceAccount"), jsii.String("service-account"), nil)
		role := plus.NewRole(chart, jsii.String("Role"), nil)
		role.AllowRead(account)
		rules := manifestOfKind(t, chart, "Role")["rules"].([]interface{})
		requireDeepEqual(t, rules, []interface{}{map[string]interface{}{
			"apiGroups":     []interface{}{string("")},
			"resourceNames": []interface{}{"service-account"},
			"resources":     []interface{}{"serviceaccounts"},
			"verbs":         []interface{}{"get", "list", "watch"},
		}})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/service-account.test.ts#L17
	t.Run("role can bind to imported", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		role := plus.NewRole(chart, jsii.String("Role"), nil)
		account := plus.ServiceAccount_FromServiceAccountName(chart, jsii.String("ServiceAccount"), jsii.String("sa-name"), &plus.FromServiceAccountNameOptions{NamespaceName: jsii.String("kube-system")})
		role.Bind(account)
		binding := manifestOfKind(t, chart, "RoleBinding")
		subjects := binding["subjects"].([]interface{})
		requireDeepEqual(t, subjects, []interface{}{map[string]interface{}{
			"apiGroup": "", "kind": "ServiceAccount", "name": "sa-name", "namespace": "kube-system",
		}})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/service-account.test.ts#L32
	t.Run("defaultChild", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		account := plus.NewServiceAccount(chart, jsii.String("ServiceAccount"), nil)
		if got := stringValue(cdk8s.ApiObject_Of(account).Kind()); got != "ServiceAccount" {
			t.Fatalf("default child kind = %q, want ServiceAccount", got)
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/service-account.test.ts#L42
	t.Run("minimal definition", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		account := plus.NewServiceAccount(chart, jsii.String("my-service-account"), nil)
		if boolValue(account.AutomountToken()) {
			t.Fatal("AutomountToken() = true, want false")
		}
		requireDeepEqual(t, synth(t, chart), []interface{}{map[string]interface{}{
			"apiVersion":                   "v1",
			"automountServiceAccountToken": false,
			"kind":                         "ServiceAccount",
			"metadata":                     map[string]interface{}{"name": "test-my-service-account-c84bb46b"},
		}})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/service-account.test.ts#L65
	t.Run("secrets can be added", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		secret1 := plus.Secret_FromSecretName(chart, jsii.String("Secret1"), jsii.String("my-secret-1"))
		secret2 := plus.Secret_FromSecretName(chart, jsii.String("Secret2"), jsii.String("my-secret-2"))
		account := plus.NewServiceAccount(chart, jsii.String("my-service-account"), &plus.ServiceAccountProps{Secrets: &[]plus.ISecret{secret1}})
		account.AddSecret(secret2)
		requireDeepEqual(t, manifestOfKind(t, chart, "ServiceAccount")["secrets"], []interface{}{
			map[string]interface{}{"name": "my-secret-1"},
			map[string]interface{}{"name": "my-secret-2"},
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/service-account.test.ts#L84
	t.Run("auto mounting token can be disabled", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		account := plus.NewServiceAccount(chart, jsii.String("my-service-account"), &plus.ServiceAccountProps{AutomountToken: jsii.Bool(false)})
		if boolValue(account.AutomountToken()) {
			t.Fatal("AutomountToken() = true, want false")
		}
		if got := manifestOfKind(t, chart, "ServiceAccount")["automountServiceAccountToken"]; got != false {
			t.Fatalf("automountServiceAccountToken = %#v, want false", got)
		}
	})
}
