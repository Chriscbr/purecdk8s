package cdk8s_test

import (
	"strings"
	"testing"

	cdk8s "github.com/Chriscbr/purecdk8s/cdk8s/v2"
)

func coreDNSName(path string, options *cdk8s.NameOptions) string {
	return coreStringValue(cdk8s.Names_ToDnsLabel(coreCreateTree(path), options))
}

func coreLabelValue(path string, options *cdk8s.NameOptions) string {
	return coreStringValue(cdk8s.Names_ToLabelValue(coreCreateTree(path), options))
}

func coreCheckNames(t *testing.T, name func(string, *cdk8s.NameOptions) string, options *cdk8s.NameOptions, cases map[string]string) {
	t.Helper()
	for path, want := range cases {
		if got := name(path, options); got != want {
			t.Errorf("name for %q = %q, want %q", path, got, want)
		}
	}
}

func TestNamesDNSLabel(t *testing.T) {
	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names.test.ts#L8
	t.Run("ignores default children", func(t *testing.T) {
		coreCheckNames(t, coreDNSName, nil, map[string]string{
			"hello/default/foo/world/default":   "hello-foo-world-c8ceb89a",
			"hello/resource/foo/world/resource": "hello-foo-world-c8c051a2",
			"hello/resource/foo/world/default":  "hello-foo-world-c8285558",
			"hello/Resource/foo/world/Default":  "hello-foo-world-c8455d08",
			"hello/default/foo/world/resource":  "hello-foo-world-c83a0f50",
			"resource/default":                  "c818ce2d",
		})
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names.test.ts#L17
	t.Run("normalize to dns_name", func(t *testing.T) {
		coreCheckNames(t, coreDNSName, nil, map[string]string{
			"Hello":       "hello-c8a347e4",
			"hey*":        "hey-c808ed9e",
			"not allowed": "notallowed-c82fed05",
		})
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names.test.ts#L23
	t.Run("maximum length for a single term", func(t *testing.T) {
		if got := coreDNSName("1234567890abcdef", &cdk8s.NameOptions{MaxLen: coreFloat(15)}); got != "123456-c85fab94" {
			t.Errorf("short maximum name = %q", got)
		}
		if got := coreDNSName("x"+strings.Repeat("a", 64), nil); got != "xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-c86953f2" {
			t.Errorf("default maximum name = %q", got)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names.test.ts#L28
	t.Run("single term is not decorated with a hash", func(t *testing.T) {
		coreCheckNames(t, coreDNSName, nil, map[string]string{
			"foo":                   "foo",
			"foo-bar-123-455":       "foo-bar-123-455",
			strings.Repeat("z", 63): strings.Repeat("z", 63),
		})
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names.test.ts#L34
	t.Run("multiple terms are separated and a hash is appended", func(t *testing.T) {
		coreCheckNames(t, coreDNSName, nil, map[string]string{
			"hello-foo-world":       "hello-foo-world",
			"hello-hello-foo-world": "hello-hello-foo-world",
			"hello-foo/world":       "hello-foo-world-c83c4a8a",
			"hello-foo/foo":         "hello-foo-foo-c884a60a",
			"hello/foo/world":       "hello-foo-world-c89b166b",
		})
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names.test.ts#L42
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

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names.test.ts#L52
	t.Run("omit duplicate components in names", func(t *testing.T) {
		coreCheckNames(t, coreDNSName, nil, map[string]string{
			"hello/hello/foo/world":         "hello-foo-world-c8538d75",
			"hello/hello/hello/foo/world":   "hello-foo-world-c815bea4",
			"hello/hello/hello/hello/hello": "hello-c830c284",
			"hello/cool/cool/cool/cool":     "hello-cool-c816948a",
			"hello/world/world/world/cool":  "hello-world-cool-c8e259cb",
		})
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names.test.ts#L60
	t.Run("trimming prioritizes last component", func(t *testing.T) {
		cases := []struct {
			path   string
			maxLen float64
			want   string
		}{
			{"hello/world", 8, "c85bc96a"},
			{"hello/world/this/is/cool", 8, "c80ec725"},
			{"hello/world/this/is/cool", 12, "coo-c80ec725"},
			{"hello/hello/this/is/cool", 12, "coo-c812c430"},
			{"hello/cool/cool/cool/cool", 15, "h-cool-c816948a"},
			{"hello/world/this/is/cool", 14, "cool-c80ec725"},
			{"hello/world/this/is/cool", 15, "i-cool-c80ec725"},
			{"hello/world/this/is/cool", 25, "wor-this-is-cool-c80ec725"},
		}
		for _, testCase := range cases {
			if got := coreDNSName(testCase.path, &cdk8s.NameOptions{MaxLen: &testCase.maxLen}); got != testCase.want {
				t.Errorf("name for %q at %v = %q, want %q", testCase.path, testCase.maxLen, got, testCase.want)
			}
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names.test.ts#L71
	t.Run("filter empty components", func(t *testing.T) {
		coreCheckNames(t, coreDNSName, nil, map[string]string{
			"hello/world---this-is-cool---": "hello-world-this-is-cool-c88665d5",
			"hello-world-this-is-cool":      "hello-world-this-is-cool",
			"hello/world-this/is-cool":      "hello-world-this-is-cool-c81c7478",
		})
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names.test.ts#L77
	t.Run("optional hash", func(t *testing.T) {
		if got := coreDNSName("hello/default/foo/world/resource", &cdk8s.NameOptions{IncludeHash: coreBool(false)}); got != "hello-foo-world" {
			t.Errorf("hashless name = %q", got)
		}
		if got := coreDNSName("hello/world/this/is/cool", &cdk8s.NameOptions{IncludeHash: coreBool(false), MaxLen: coreFloat(8)}); got != "is-cool" {
			t.Errorf("trimmed hashless name = %q", got)
		}
	})
}

func TestNamesLabelValue(t *testing.T) {
	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names.test.ts#L85
	t.Run("ignores default children", func(t *testing.T) {
		coreCheckNames(t, coreLabelValue, nil, map[string]string{
			"hello/default/foo/world/default":   "hello-foo-world-c8ceb89a",
			"hello/resource/foo/world/resource": "hello-foo-world-c8c051a2",
			"hello/resource/foo/world/default":  "hello-foo-world-c8285558",
			"hello/Resource/foo/world/Default":  "hello-foo-world-c8455d08",
			"hello/default/foo/world/resource":  "hello-foo-world-c83a0f50",
			"resource/default":                  "c818ce2d",
		})
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names.test.ts#L94
	t.Run("normalize to dns_name", func(t *testing.T) {
		coreCheckNames(t, coreLabelValue, nil, map[string]string{
			"Hello":       "Hello",
			"hey*":        "hey-c808ed9e",
			"not allowed": "notallowed-c82fed05",
		})
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names.test.ts#L100
	t.Run("maximum length for a single term", func(t *testing.T) {
		if got := coreLabelValue("1234567890abcdef", &cdk8s.NameOptions{MaxLen: coreFloat(15), Delimiter: coreString("-")}); got != "123456-c85fab94" {
			t.Errorf("short maximum label = %q", got)
		}
		if got := coreLabelValue("x"+strings.Repeat("a", 64), nil); got != "xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-c86953f2" {
			t.Errorf("default maximum label = %q", got)
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names.test.ts#L105
	t.Run("single term is not decorated with a hash", func(t *testing.T) {
		coreCheckNames(t, coreLabelValue, nil, map[string]string{
			"foo":                   "foo",
			"foo-bar-123-455":       "foo-bar-123-455",
			strings.Repeat("z", 63): strings.Repeat("z", 63),
		})
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names.test.ts#L111
	t.Run("multiple terms are separated and a hash is appended", func(t *testing.T) {
		coreCheckNames(t, coreLabelValue, nil, map[string]string{
			"hello-foo-world":       "hello-foo-world",
			"hello-hello-foo-world": "hello-hello-foo-world",
			"hello-foo/world":       "hello-foo-world-c83c4a8a",
			"hello-foo/foo":         "hello-foo-foo-c884a60a",
			"hello/foo/world":       "hello-foo-world-c89b166b",
		})
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names.test.ts#L119
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

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names.test.ts#L129
	t.Run("omit duplicate components in names", func(t *testing.T) {
		coreCheckNames(t, coreLabelValue, nil, map[string]string{
			"hello/hello/foo/world":         "hello-foo-world-c8538d75",
			"hello/hello/hello/foo/world":   "hello-foo-world-c815bea4",
			"hello/hello/hello/hello/hello": "hello-c830c284",
			"hello/cool/cool/cool/cool":     "hello-cool-c816948a",
			"hello/world/world/world/cool":  "hello-world-cool-c8e259cb",
		})
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names.test.ts#L137
	t.Run("trimming prioritizes last component", func(t *testing.T) {
		cases := []struct {
			path   string
			maxLen float64
			want   string
		}{
			{"hello/world", 8, "c85bc96a"},
			{"hello/world/this/is/cool", 8, "c80ec725"},
			{"hello/world/this/is/cool", 12, "coo-c80ec725"},
			{"hello/hello/this/is/cool", 12, "coo-c812c430"},
			{"hello/cool/cool/cool/cool", 15, "h-cool-c816948a"},
			{"hello/world/this/is/cool", 14, "cool-c80ec725"},
			{"hello/world/this/is/cool", 15, "i-cool-c80ec725"},
			{"hello/world/this/is/cool", 25, "wor-this-is-cool-c80ec725"},
		}
		for _, testCase := range cases {
			if got := coreLabelValue(testCase.path, &cdk8s.NameOptions{MaxLen: &testCase.maxLen}); got != testCase.want {
				t.Errorf("label for %q at %v = %q, want %q", testCase.path, testCase.maxLen, got, testCase.want)
			}
		}
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names.test.ts#L148
	t.Run("filter empty components", func(t *testing.T) {
		coreCheckNames(t, coreLabelValue, nil, map[string]string{
			"hello---this/is/cool/-":    "hello-this-is-cool-c83b900b",
			"hello---this/is---/cool/-": "hello-this-is-cool-c82d69dd",
		})
	})

	// Ported from https://github.com/cdk8s-team/cdk8s-core/blob/3c496b0cd654efe86628952af0d5e3bf7d4bb182/test/names.test.ts#L153
	t.Run("optional hash", func(t *testing.T) {
		if got := coreLabelValue("hello/default/foo/world/resource", &cdk8s.NameOptions{IncludeHash: coreBool(false)}); got != "hello-foo-world" {
			t.Errorf("hashless label = %q", got)
		}
		if got := coreLabelValue("hello/world/this/is/cool", &cdk8s.NameOptions{IncludeHash: coreBool(false), MaxLen: coreFloat(8)}); got != "is-cool" {
			t.Errorf("trimmed hashless label = %q", got)
		}
	})
}
