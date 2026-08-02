package cdk8splus34_test

import (
	"encoding/json"
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
	plus "github.com/Chriscbr/purecdk8s/cdk8splus34/v2"
	"github.com/Chriscbr/purecdk8s/jsii"
)

func requireSecretName(t *testing.T, chart cdk8s.Chart, secret plus.IResource, want string) {
	t.Helper()
	if got := stringValue(secret.Name()); got != want {
		t.Fatalf("secret name = %q, want %q", got, want)
	}
	if got := mapAt(t, manifestOfKind(t, chart, "Secret"), "metadata")["name"]; got != want {
		t.Fatalf("manifest secret name = %#v, want %q", got, want)
	}
}

func requireImmutableSecret(t *testing.T, chart cdk8s.Chart, secret plus.Secret) {
	t.Helper()
	if !boolValue(secret.Immutable()) {
		t.Fatal("Immutable() = false, want true")
	}
	if got := manifestOfKind(t, chart, "Secret")["immutable"]; got != true {
		t.Fatalf("immutable = %#v, want true", got)
	}
}

func TestSecret(t *testing.T) {
	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/secret.test.ts#L5
	t.Run("can grant permissions on imported", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		secret := plus.Secret_FromSecretName(chart, jsii.String("Secret"), jsii.String("secret"))
		role := plus.NewRole(chart, jsii.String("Role"), nil)
		role.AllowRead(secret)
		requireDeepEqual(t, roleRules(t, chart, "Role"), []interface{}{rbacRule("", "secrets", []interface{}{"secret"}, "get", "list", "watch")})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/secret.test.ts#L17
	t.Run("defaultChild", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		secret := plus.NewSecret(chart, jsii.String("Secret"), nil)
		if got := stringValue(cdk8s.ApiObject_Of(secret).Kind()); got != "Secret" {
			t.Fatalf("default child kind = %q, want Secret", got)
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/secret.test.ts#L26
	t.Run("can be imported from secret name", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		secret := plus.Secret_FromSecretName(chart, jsii.String("Secret"), jsii.String("secret"))
		if got := stringValue(secret.Name()); got != "secret" {
			t.Fatalf("secret name = %q, want secret", got)
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/secret.test.ts#L33
	t.Run("can create a new secret", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewSecret(chart, jsii.String("Secret"), &plus.SecretProps{
			StringData: &map[string]*string{"key": jsii.String("value")}, Type: jsii.String("kubernetes.io/tls"),
		})
		manifest := manifestOfKind(t, chart, "Secret")
		if manifest["type"] != "kubernetes.io/tls" || manifest["immutable"] != false {
			t.Fatalf("secret manifest = %#v", manifest)
		}
		requireDeepEqual(t, manifest["stringData"], map[string]interface{}{"key": "value"})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/secret.test.ts#L61
	t.Run("can add data", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		secret := plus.NewSecret(chart, jsii.String("Secret"), nil)
		secret.AddStringData(jsii.String("key"), jsii.String("value"))
		requireDeepEqual(t, manifestOfKind(t, chart, "Secret")["stringData"], map[string]interface{}{"key": "value"})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/secret.test.ts#L84
	t.Run("can create a basic auth secret", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewBasicAuthSecret(chart, jsii.String("BasicAuthSecret"), &plus.BasicAuthSecretProps{Username: jsii.String("admin"), Password: jsii.String("t0p-Secret")})
		manifest := manifestOfKind(t, chart, "Secret")
		if manifest["type"] != "kubernetes.io/basic-auth" {
			t.Fatalf("secret type = %#v", manifest["type"])
		}
		requireDeepEqual(t, manifest["stringData"], map[string]interface{}{"username": "admin", "password": "t0p-Secret"})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/secret.test.ts#L111
	t.Run("can override basic auth name", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		secret := plus.NewBasicAuthSecret(chart, jsii.String("BasicAuthSecret"), &plus.BasicAuthSecretProps{
			Metadata: &cdk8s.ApiObjectMetadata{Name: jsii.String("override-name")}, Username: jsii.String("admin"), Password: jsii.String("t0p-Secret"),
		})
		requireSecretName(t, chart, secret, "override-name")
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/secret.test.ts#L130
	t.Run("can create an SSH auth secret", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewSshAuthSecret(chart, jsii.String("SshAuthSecret"), &plus.SshAuthSecretProps{SshPrivateKey: jsii.String("fake-private-key")})
		manifest := manifestOfKind(t, chart, "Secret")
		if manifest["type"] != "kubernetes.io/ssh-auth" {
			t.Fatalf("secret type = %#v", manifest["type"])
		}
		requireDeepEqual(t, manifest["stringData"], map[string]interface{}{"ssh-privatekey": "fake-private-key"})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/secret.test.ts#L155
	t.Run("can override SSH auth name", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		secret := plus.NewSshAuthSecret(chart, jsii.String("SshAuthSecret"), &plus.SshAuthSecretProps{
			Metadata: &cdk8s.ApiObjectMetadata{Name: jsii.String("override-name")}, SshPrivateKey: jsii.String("fake-private-key"),
		})
		requireSecretName(t, chart, secret, "override-name")
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/secret.test.ts#L173
	t.Run("can create a service account token secret", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		account := plus.NewServiceAccount(chart, jsii.String("ServiceAccount"), nil)
		secret := plus.NewServiceAccountTokenSecret(chart, jsii.String("ServiceAccountToken"), &plus.ServiceAccountTokenSecretProps{ServiceAccount: account})
		secret.AddStringData(jsii.String("extra"), jsii.String("foo"))
		manifest := manifestOfKind(t, chart, "Secret")
		if manifest["type"] != "kubernetes.io/service-account-token" {
			t.Fatalf("secret type = %#v", manifest["type"])
		}
		requireDeepEqual(t, mapAt(t, manifest, "metadata", "annotations"), map[string]interface{}{"kubernetes.io/service-account.name": "test-serviceaccount-c8f15383"})
		requireDeepEqual(t, manifest["stringData"], map[string]interface{}{"extra": "foo"})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/secret.test.ts#L213
	t.Run("can override service account token name", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		account := plus.NewServiceAccount(chart, jsii.String("ServiceAccount"), nil)
		secret := plus.NewServiceAccountTokenSecret(chart, jsii.String("ServiceAccountToken"), &plus.ServiceAccountTokenSecretProps{
			Metadata: &cdk8s.ApiObjectMetadata{Name: jsii.String("override-name")}, ServiceAccount: account,
		})
		secret.AddStringData(jsii.String("extra"), jsii.String("foo"))
		requireSecretName(t, chart, secret, "override-name")
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/secret.test.ts#L233
	t.Run("can add annotations to a service account token", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		account := plus.NewServiceAccount(chart, jsii.String("ServiceAccount"), nil)
		secret := plus.NewServiceAccountTokenSecret(chart, jsii.String("ServiceAccountToken"), &plus.ServiceAccountTokenSecretProps{
			Metadata: &cdk8s.ApiObjectMetadata{Annotations: &map[string]*string{"cdk8s.io/hello": jsii.String("world")}}, ServiceAccount: account,
		})
		secret.AddStringData(jsii.String("extra"), jsii.String("foo"))
		requireDeepEqual(t, mapAt(t, manifestOfKind(t, chart, "Secret"), "metadata", "annotations"), map[string]interface{}{
			"kubernetes.io/service-account.name": "test-serviceaccount-c8f15383", "cdk8s.io/hello": "world",
		})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/secret.test.ts#L257
	t.Run("can create a TLS secret", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewTlsSecret(chart, jsii.String("TlsSecret"), &plus.TlsSecretProps{TlsCert: jsii.String("tls-cert-value"), TlsKey: jsii.String("tls-key-value")})
		manifest := manifestOfKind(t, chart, "Secret")
		if manifest["type"] != "kubernetes.io/tls" {
			t.Fatalf("secret type = %#v", manifest["type"])
		}
		requireDeepEqual(t, manifest["stringData"], map[string]interface{}{"tls.crt": "tls-cert-value", "tls.key": "tls-key-value"})
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/secret.test.ts#L284
	t.Run("can override TLS name", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		secret := plus.NewTlsSecret(chart, jsii.String("TlsSecret"), &plus.TlsSecretProps{
			Metadata: &cdk8s.ApiObjectMetadata{Name: jsii.String("override-name")}, TlsCert: jsii.String("tls-cert-value"), TlsKey: jsii.String("tls-key-value"),
		})
		requireSecretName(t, chart, secret, "override-name")
	})

	dockerData := func() *map[string]interface{} {
		return &map[string]interface{}{"auths": map[string]interface{}{"hub.xxx.com": map[string]interface{}{
			"username": "xxx", "password": "xxx", "email": "xxx", "auth": "xxx",
		}}}
	}

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/secret.test.ts#L303
	t.Run("can create a Docker config secret", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		plus.NewDockerConfigSecret(chart, jsii.String("DockerConfigSecret"), &plus.DockerConfigSecretProps{Data: dockerData()})
		manifest := manifestOfKind(t, chart, "Secret")
		if manifest["type"] != "kubernetes.io/dockerconfigjson" {
			t.Fatalf("secret type = %#v", manifest["type"])
		}
		encoded := manifest["stringData"].(map[string]interface{})[".dockerconfigjson"].(string)
		var got interface{}
		if err := json.Unmarshal([]byte(encoded), &got); err != nil {
			t.Fatal(err)
		}
		var want interface{}
		wantBytes, _ := json.Marshal(dockerData())
		_ = json.Unmarshal(wantBytes, &want)
		requireDeepEqual(t, got, want)
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/secret.test.ts#L337
	t.Run("can override Docker config name", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		secret := plus.NewDockerConfigSecret(chart, jsii.String("DockerConfigSecret"), &plus.DockerConfigSecretProps{
			Metadata: &cdk8s.ApiObjectMetadata{Name: jsii.String("override-name")}, Data: dockerData(),
		})
		requireSecretName(t, chart, secret, "override-name")
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/secret.test.ts#L364
	t.Run("default immutability", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		secret := plus.NewSecret(chart, jsii.String("Secret"), nil)
		if boolValue(secret.Immutable()) || manifestOfKind(t, chart, "Secret")["immutable"] != false {
			t.Fatal("secret is immutable by default")
		}
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/secret.test.ts#L377
	t.Run("immutable generic", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		requireImmutableSecret(t, chart, plus.NewSecret(chart, jsii.String("Secret"), &plus.SecretProps{Immutable: jsii.Bool(true)}))
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/secret.test.ts#L392
	t.Run("immutable basic auth", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		requireImmutableSecret(t, chart, plus.NewBasicAuthSecret(chart, jsii.String("Secret"), &plus.BasicAuthSecretProps{Username: jsii.String("user"), Password: jsii.String("pass"), Immutable: jsii.Bool(true)}))
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/secret.test.ts#L409
	t.Run("immutable SSH auth", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		requireImmutableSecret(t, chart, plus.NewSshAuthSecret(chart, jsii.String("Secret"), &plus.SshAuthSecretProps{SshPrivateKey: jsii.String("private"), Immutable: jsii.Bool(true)}))
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/secret.test.ts#L425
	t.Run("immutable service account token", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		account := plus.ServiceAccount_FromServiceAccountName(chart, jsii.String("SA"), jsii.String("sa"), nil)
		requireImmutableSecret(t, chart, plus.NewServiceAccountTokenSecret(chart, jsii.String("Secret"), &plus.ServiceAccountTokenSecretProps{ServiceAccount: account, Immutable: jsii.Bool(true)}))
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/secret.test.ts#L441
	t.Run("immutable TLS", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		requireImmutableSecret(t, chart, plus.NewTlsSecret(chart, jsii.String("Secret"), &plus.TlsSecretProps{TlsCert: jsii.String("cert"), TlsKey: jsii.String("key"), Immutable: jsii.Bool(true)}))
	})

	// Ported from:
	// https://github.com/cdk8s-team/cdk8s-plus/blob/fe0337f802a48b0fba3588b80b4afe82085225b0/test/secret.test.ts#L458
	t.Run("immutable Docker config", func(t *testing.T) {
		chart := cdk8s.Testing_Chart()
		data := map[string]interface{}{}
		requireImmutableSecret(t, chart, plus.NewDockerConfigSecret(chart, jsii.String("Secret"), &plus.DockerConfigSecretProps{Data: &data, Immutable: jsii.Bool(true)}))
	})
}
