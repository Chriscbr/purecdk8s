package cdk8s_test

import (
	"strings"
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
)

func TestLegacyNamesDNSLabel(t *testing.T) {
	t.Setenv("CDK8S_LEGACY_HASH", "1")

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names-legacy.test.ts#L11
	t.Run("ignores default children", func(t *testing.T) {
		coreCheckNames(t, coreDNSName, nil, map[string]string{
			"hello/default/foo/world/default":   "hello-foo-world-5d193db9",
			"hello/resource/foo/world/resource": "hello-foo-world-f5dd971f",
			"hello/resource/foo/world/default":  "hello-foo-world-2f1cee85",
			"hello/Resource/foo/world/Default":  "hello-foo-world-857189b5",
			"hello/default/foo/world/resource":  "hello-foo-world-e89fdfae",
			"resource/default":                  "40b6bcd9",
		})
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names-legacy.test.ts#L20
	t.Run("normalize to dns_name", func(t *testing.T) {
		coreCheckNames(t, coreDNSName, nil, map[string]string{
			" ":           "36a9e7f1",
			"Hello":       "hello-185f8db3",
			"hey*":        "hey-96c05e6c",
			"not allowed": "notallowed-a26075ed",
		})
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names-legacy.test.ts#L27
	t.Run("maximum length for a single term", func(t *testing.T) {
		if got := coreDNSName("1234567890abcdef", &cdk8s.NameOptions{MaxLen: coreFloat(15)}); got != "123456-8e9916c5" {
			t.Errorf("short maximum name = %q", got)
		}
		if got := coreDNSName("x"+strings.Repeat("a", 64), nil); got != "xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-f69f4ba1" {
			t.Errorf("default maximum name = %q", got)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names-legacy.test.ts#L32
	t.Run("single term is not decorated with a hash", func(t *testing.T) {
		coreCheckNames(t, coreDNSName, nil, map[string]string{
			"foo":                   "foo",
			"foo-bar-123-455":       "foo-bar-123-455",
			strings.Repeat("z", 63): strings.Repeat("z", 63),
		})
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names-legacy.test.ts#L38
	t.Run("multiple terms are separated and a hash is appended", func(t *testing.T) {
		coreCheckNames(t, coreDNSName, nil, map[string]string{
			"hello-foo-world":       "hello-foo-world",
			"hello-hello-foo-world": "hello-hello-foo-world",
			"hello-foo/world":       "hello-foo-world-54700203",
			"hello-foo/foo":         "hello-foo-foo-e078a973",
			"hello/foo/world":       "hello-foo-world-4f6e4fd8",
		})
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names-legacy.test.ts#L46
	t.Run("invalid max length", func(t *testing.T) {
		for _, maxLen := range []float64{4, 7} {
			maxLen := maxLen
			coreRequirePanicContains(t, "minimum max length for object names is 8", func() {
				coreDNSName("foo", &cdk8s.NameOptions{MaxLen: &maxLen})
			})
		}
		for _, maxLen := range []float64{8, 9} {
			if got := coreDNSName("foo", &cdk8s.NameOptions{MaxLen: &maxLen}); got != "foo" {
				t.Errorf("name with max length %v = %q", maxLen, got)
			}
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names-legacy.test.ts#L56
	t.Run("omit duplicate components in names", func(t *testing.T) {
		coreCheckNames(t, coreDNSName, nil, map[string]string{
			"hello/hello/foo/world":         "hello-foo-world-1d4999d0",
			"hello/hello/hello/foo/world":   "hello-foo-world-d3ebcda3",
			"hello/hello/hello/hello/hello": "hello-456bb9d7",
			"hello/cool/cool/cool/cool":     "hello-cool-83150e81",
			"hello/world/world/world/cool":  "hello-world-cool-0148a798",
		})
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names-legacy.test.ts#L64
	t.Run("trimming prioritizes last component", func(t *testing.T) {
		cases := []struct {
			path   string
			maxLen float64
			want   string
		}{
			{"hello/world", 8, "761e91eb"},
			{"hello/world/this/is/cool", 8, "a7c39f00"},
			{"hello/world/this/is/cool", 12, "coo-a7c39f00"},
			{"hello/hello/this/is/cool", 12, "coo-8751188b"},
			{"hello/cool/cool/cool/cool", 15, "h-cool-83150e81"},
			{"hello/world/this/is/cool", 14, "cool-a7c39f00"},
			{"hello/world/this/is/cool", 15, "i-cool-a7c39f00"},
			{"hello/world/this/is/cool", 25, "wor-this-is-cool-a7c39f00"},
		}
		for _, testCase := range cases {
			if got := coreDNSName(testCase.path, &cdk8s.NameOptions{MaxLen: &testCase.maxLen}); got != testCase.want {
				t.Errorf("name for %q at %v = %q, want %q", testCase.path, testCase.maxLen, got, testCase.want)
			}
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names-legacy.test.ts#L75
	t.Run("filter empty components", func(t *testing.T) {
		coreCheckNames(t, coreDNSName, nil, map[string]string{
			"hello/world---this-is-cool---": "hello-world-this-is-cool-85209c22",
			"hello-world-this-is-cool":      "hello-world-this-is-cool",
			"hello/world-this/is-cool":      "hello-world-this-is-cool-9bdccb95",
		})
	})
}

func TestLegacyNamesLabelValue(t *testing.T) {
	t.Setenv("CDK8S_LEGACY_HASH", "1")

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names-legacy.test.ts#L83
	t.Run("ignores default children", func(t *testing.T) {
		coreCheckNames(t, coreLabelValue, nil, map[string]string{
			"hello/default/foo/world/default":   "hello-foo-world-5d193db9",
			"hello/resource/foo/world/resource": "hello-foo-world-f5dd971f",
			"hello/resource/foo/world/default":  "hello-foo-world-2f1cee85",
			"hello/Resource/foo/world/Default":  "hello-foo-world-857189b5",
			"hello/default/foo/world/resource":  "hello-foo-world-e89fdfae",
			"resource/default":                  "40b6bcd9",
		})
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names-legacy.test.ts#L92
	t.Run("normalize to dns_name", func(t *testing.T) {
		coreCheckNames(t, coreLabelValue, nil, map[string]string{
			" ":           "36a9e7f1",
			"Hello":       "Hello",
			"hey*":        "hey-96c05e6c",
			"not allowed": "notallowed-a26075ed",
		})
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names-legacy.test.ts#L99
	t.Run("maximum length for a single term", func(t *testing.T) {
		if got := coreLabelValue("1234567890abcdef", &cdk8s.NameOptions{MaxLen: coreFloat(15), Delimiter: coreString("-")}); got != "123456-8e9916c5" {
			t.Errorf("short maximum label = %q", got)
		}
		if got := coreLabelValue("x"+strings.Repeat("a", 64), nil); got != "xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-f69f4ba1" {
			t.Errorf("default maximum label = %q", got)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names-legacy.test.ts#L104
	t.Run("single term is not decorated with a hash", func(t *testing.T) {
		coreCheckNames(t, coreLabelValue, nil, map[string]string{
			"foo":                   "foo",
			"foo-bar-123-455":       "foo-bar-123-455",
			strings.Repeat("z", 63): strings.Repeat("z", 63),
		})
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names-legacy.test.ts#L110
	t.Run("multiple terms are separated and a hash is appended", func(t *testing.T) {
		coreCheckNames(t, coreLabelValue, nil, map[string]string{
			"hello-foo-world":       "hello-foo-world",
			"hello-hello-foo-world": "hello-hello-foo-world",
			"hello-foo/world":       "hello-foo-world-54700203",
			"hello-foo/foo":         "hello-foo-foo-e078a973",
			"hello/foo/world":       "hello-foo-world-4f6e4fd8",
		})
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names-legacy.test.ts#L118
	t.Run("invalid max length", func(t *testing.T) {
		for _, maxLen := range []float64{4, 7} {
			maxLen := maxLen
			coreRequirePanicContains(t, "minimum max length for label is 8", func() {
				coreLabelValue("foo", &cdk8s.NameOptions{MaxLen: &maxLen})
			})
		}
		for _, maxLen := range []float64{8, 9} {
			if got := coreLabelValue("foo", &cdk8s.NameOptions{MaxLen: &maxLen}); got != "foo" {
				t.Errorf("label with max length %v = %q", maxLen, got)
			}
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names-legacy.test.ts#L128
	t.Run("omit duplicate components in names", func(t *testing.T) {
		coreCheckNames(t, coreLabelValue, nil, map[string]string{
			"hello/hello/foo/world":         "hello-foo-world-1d4999d0",
			"hello/hello/hello/foo/world":   "hello-foo-world-d3ebcda3",
			"hello/hello/hello/hello/hello": "hello-456bb9d7",
			"hello/cool/cool/cool/cool":     "hello-cool-83150e81",
			"hello/world/world/world/cool":  "hello-world-cool-0148a798",
		})
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names-legacy.test.ts#L136
	t.Run("trimming prioritizes last component", func(t *testing.T) {
		cases := []struct {
			path   string
			maxLen float64
			want   string
		}{
			{"hello/world", 8, "761e91eb"},
			{"hello/world/this/is/cool", 8, "a7c39f00"},
			{"hello/world/this/is/cool", 12, "coo-a7c39f00"},
			{"hello/hello/this/is/cool", 12, "coo-8751188b"},
			{"hello/cool/cool/cool/cool", 15, "h-cool-83150e81"},
			{"hello/world/this/is/cool", 14, "cool-a7c39f00"},
			{"hello/world/this/is/cool", 15, "i-cool-a7c39f00"},
			{"hello/world/this/is/cool", 25, "wor-this-is-cool-a7c39f00"},
		}
		for _, testCase := range cases {
			if got := coreLabelValue(testCase.path, &cdk8s.NameOptions{MaxLen: &testCase.maxLen}); got != testCase.want {
				t.Errorf("label for %q at %v = %q, want %q", testCase.path, testCase.maxLen, got, testCase.want)
			}
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names-legacy.test.ts#L147
	t.Run("filter empty components", func(t *testing.T) {
		coreCheckNames(t, coreLabelValue, nil, map[string]string{
			"hello---this/is/cool/-":    "hello-this-is-cool-a30e4c1e",
			"hello---this/is---/cool/-": "hello-this-is-cool-a9b9d489",
		})
	})
}
