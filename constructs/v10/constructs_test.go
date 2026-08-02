package constructs_test

import (
	"reflect"
	"strings"
	"testing"

	constructs "github.com/purecdk8s/purecdk8s/constructs/v10"
	"github.com/purecdk8s/purecdk8s/jsii"
)

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func constructIDs(constructsInTree []constructs.IConstruct) []string {
	ids := make([]string, 0, len(constructsInTree))
	for _, construct := range constructsInTree {
		ids = append(ids, stringValue(construct.Node().Id()))
	}
	return ids
}

func constructPaths(constructsInTree []constructs.IConstruct) []string {
	paths := make([]string, 0, len(constructsInTree))
	for _, construct := range constructsInTree {
		paths = append(paths, stringValue(construct.Node().Path()))
	}
	return paths
}

func requirePanicContains(t *testing.T, want string, callback func()) {
	t.Helper()
	defer func() {
		panicValue := recover()
		if panicValue == nil {
			t.Fatalf("expected panic containing %q", want)
		}
		if got := panicValue.(string); !strings.Contains(got, want) {
			t.Fatalf("panic = %q, want it to contain %q", got, want)
		}
	}()
	callback()
}

type constructTree struct {
	root       constructs.RootConstruct
	highChild  constructs.Construct
	child1     constructs.Construct
	child2     constructs.Construct
	child1_1   constructs.Construct
	child1_2   constructs.Construct
	child1_1_1 constructs.Construct
	child2_1   constructs.Construct
}

func newConstructTree(context map[string]interface{}) constructTree {
	root := constructs.NewRootConstruct(nil)
	highChild := constructs.NewConstruct(root, jsii.String("HighChild"))
	for key, value := range context {
		key, value := key, value
		highChild.Node().SetContext(&key, value)
	}

	child1 := constructs.NewConstruct(highChild, jsii.String("Child1"))
	child2 := constructs.NewConstruct(highChild, jsii.String("Child2"))
	child1_1 := constructs.NewConstruct(child1, jsii.String("Child11"))
	child1_2 := constructs.NewConstruct(child1, jsii.String("Child12"))
	child1_1_1 := constructs.NewConstruct(child1_1, jsii.String("Child111"))
	child2_1 := constructs.NewConstruct(child2, jsii.String("Child21"))

	return constructTree{
		root:       root,
		highChild:  highChild,
		child1:     child1,
		child2:     child2,
		child1_1:   child1_1,
		child1_2:   child1_2,
		child1_1_1: child1_1_1,
		child2_1:   child2_1,
	}
}

// Ported from:
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L8
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L16
func TestRootConstruct(t *testing.T) {
	root := constructs.NewRootConstruct(nil)
	node := root.Node()
	if got := stringValue(node.Id()); got != "" {
		t.Fatalf("root id = %q, want empty", got)
	}
	if node.Scope() != nil {
		t.Fatal("root scope is not nil")
	}
	if got := len(*node.Children()); got != 0 {
		t.Fatalf("root children = %d, want 0", got)
	}
	if got := stringValue(constructs.NewRootConstruct(jsii.String("")).Node().Id()); got != "" {
		t.Fatalf("explicit empty root id = %q, want empty", got)
	}

	empty := ""
	requirePanicContains(t, "Only root constructs may have an empty ID", func() {
		constructs.NewConstruct(root, &empty)
	})
}

// Ported from:
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L35
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L51
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L57
func TestConstructIDs(t *testing.T) {
	t.Run("accept ordinary characters", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		for _, id := range []string{"valid", "ValiD", "Va123lid", "v", "  invalid", "invalid   ", "123invalid", "in valid", "in_Valid", "in-Valid", "in\\Valid", "in.Valid"} {
			constructs.NewConstruct(root, jsii.String(id))
		}
	})

	t.Run("sanitize path and address separators", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		slash := constructs.NewConstruct(root, jsii.String("Boom/Boom/Bam"))
		if got := stringValue(slash.Node().Id()); got != "Boom--Boom--Bam" {
			t.Fatalf("slash id = %q", got)
		}

		newline := constructs.NewConstruct(constructs.NewRootConstruct(nil), jsii.String("Boom\nBoom\nBam"))
		if got := stringValue(newline.Node().Id()); got != "Boom--Boom--Bam" {
			t.Fatalf("newline id = %q", got)
		}
		if got := stringValue(newline.Node().Path()); got != "Boom--Boom--Bam" {
			t.Fatalf("newline path = %q", got)
		}
	})
}

// Ported from:
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L69
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L85
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L107
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L122
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L137
func TestNodeAddresses(t *testing.T) {
	t.Run("match known addresses", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		child1 := constructs.NewConstruct(root, jsii.String("This is the first child"))
		child2 := constructs.NewConstruct(child1, jsii.String("Second level"))
		first := constructs.NewConstruct(child2, jsii.String("My construct"))
		second := constructs.NewConstruct(child1, jsii.String("My construct"))

		if got := stringValue(first.Node().Path()); got != "This is the first child/Second level/My construct" {
			t.Fatalf("first path = %q", got)
		}
		if got := stringValue(second.Node().Path()); got != "This is the first child/My construct" {
			t.Fatalf("second path = %q", got)
		}
		for construct, want := range map[constructs.Construct]string{
			child1: "c8a0dfcbdc45cb728d75ebe6914d369e565dc3f61c",
			child2: "c825c5541e02ebd68e79ea636e370985b6c2de40a9",
			first:  "c83a2846e506bcc5f10682b564084bca2d275709ee",
			second: "c8003bcb3e82977712d0d7220b155cb69abd9ad383",
		} {
			if got := stringValue(construct.Node().Addr()); got != want {
				t.Errorf("address = %q, want %q", got, want)
			}
		}
	})

	t.Run("exclude Default exactly", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		plain := constructs.NewConstruct(root, jsii.String("c1"))
		hidden := constructs.NewConstruct(constructs.NewConstruct(root, jsii.String("Default")), jsii.String("c1"))
		visible := constructs.NewConstruct(constructs.NewConstruct(root, jsii.String("DeFAULt")), jsii.String("c1"))

		if got, want := stringValue(plain.Node().Addr()), "c86a34031367d11f4bef80afca42b7e7e5c6253b77"; got != want {
			t.Fatalf("plain address = %q, want %q", got, want)
		}
		if got, want := stringValue(hidden.Node().Addr()), stringValue(plain.Node().Addr()); got != want {
			t.Fatalf("hidden address = %q, want %q", got, want)
		}
		if got, want := stringValue(visible.Node().Addr()), "c8fa72abd28f794f6bacb100b26beb761d004572f5"; got != want {
			t.Fatalf("visible address = %q, want %q", got, want)
		}
	})

	t.Run("keep separator-containing IDs distinct", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		single := constructs.NewConstruct(root, jsii.String("a\nb"))
		nested := constructs.NewConstruct(constructs.NewConstruct(root, jsii.String("a")), jsii.String("b"))
		if stringValue(single.Node().Addr()) == stringValue(nested.Node().Addr()) {
			t.Fatal("a single sanitized ID and nested IDs have the same address")
		}

		root = constructs.NewRootConstruct(nil)
		left := constructs.NewConstruct(constructs.NewConstruct(root, jsii.String("a")), jsii.String("b\nc"))
		right := constructs.NewConstruct(constructs.NewConstruct(root, jsii.String("a\nb")), jsii.String("c"))
		if stringValue(left.Node().Addr()) == stringValue(right.Node().Addr()) {
			t.Fatal("separator position does not affect the address")
		}
	})

	t.Run("are deterministic for equivalent trees", func(t *testing.T) {
		first := constructs.NewConstruct(constructs.NewRootConstruct(nil), jsii.String("a"))
		second := constructs.NewConstruct(constructs.NewRootConstruct(nil), jsii.String("a"))
		if first == second {
			t.Fatal("separate constructs compare equal")
		}
		if got, want := stringValue(first.Node().Addr()), stringValue(second.Node().Addr()); got != want {
			t.Fatalf("address = %q, want %q", got, want)
		}
	})
}

// Ported from:
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L145
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L153
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L160
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L380
func TestNodeChildrenAndLookup(t *testing.T) {
	root := constructs.NewRootConstruct(nil)
	child := constructs.NewConstruct(root, jsii.String("Child1"))
	constructs.NewConstruct(root, jsii.String("Child2"))
	if got := len(*child.Node().Children()); got != 0 {
		t.Fatalf("child count = %d, want 0", got)
	}
	if got := len(*root.Node().Children()); got != 2 {
		t.Fatalf("root child count = %d, want 2", got)
	}

	childID := child.Node().Id()
	if got := root.Node().TryFindChild(childID); got != child {
		t.Fatal("TryFindChild did not return the child")
	}
	if got := root.Node().TryFindChild(jsii.String("NotFound")); got != nil {
		t.Fatalf("TryFindChild returned %#v for an absent child", got)
	}
	if got := root.Node().FindChild(childID); got != child {
		t.Fatal("FindChild did not return the child")
	}
	requirePanicContains(t, "No child with id: 'NotFound'", func() {
		root.Node().FindChild(jsii.String("NotFound"))
	})

	t.Run("allows multiple children with distinct IDs", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		for _, id := range []string{"mbc1", "mbc2", "mbc3", "mbc4"} {
			constructs.NewConstruct(root, jsii.String(id))
		}
		if got := len(*root.Node().Children()); got != 4 {
			t.Fatalf("child count = %d, want 4", got)
		}
	})
}

// Ported from:
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L167
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L182
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L199
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L216
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L232
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L244
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L277
func TestContext(t *testing.T) {
	t.Run("looks up values through scopes", func(t *testing.T) {
		tree := newConstructTree(map[string]interface{}{"ctx1": 12, "ctx2": "hello"})
		if got := tree.child1_2.Node().GetContext(jsii.String("ctx1")); got != 12 {
			t.Fatalf("ctx1 = %#v", got)
		}
		if got := tree.child1_1_1.Node().GetContext(jsii.String("ctx2")); got != "hello" {
			t.Fatalf("ctx2 = %#v", got)
		}
		requirePanicContains(t, "No context value present for ctx3 key", func() {
			tree.child1_1_1.Node().GetContext(jsii.String("ctx3"))
		})
	})

	t.Run("returns all effective context", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		root.Node().SetContext(jsii.String("ctx1"), 12)
		root.Node().SetContext(jsii.String("ctx2"), "hello")
		if got, want := root.Node().GetAllContext(nil), map[string]interface{}{"ctx1": 12, "ctx2": "hello"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("root context = %#v, want %#v", got, want)
		}

		tree := newConstructTree(map[string]interface{}{"ctx1": 12, "ctx2": "hello"})
		tree.child1_1_1.Node().SetContext(jsii.String("ctx1"), 13)
		if got, want := tree.child1_2.Node().GetAllContext(nil), map[string]interface{}{"ctx1": 12, "ctx2": "hello"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("child context = %#v, want %#v", got, want)
		}
		if got, want := tree.child1_1_1.Node().GetAllContext(nil), map[string]interface{}{"ctx1": 13, "ctx2": "hello"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("nested context = %#v, want %#v", got, want)
		}
	})

	t.Run("uses the closest value", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		high := constructs.NewConstruct(root, jsii.String("highChild"))
		high.Node().SetContext(jsii.String("c1"), "root")
		high.Node().SetContext(jsii.String("c2"), "root")
		child1 := constructs.NewConstruct(high, jsii.String("child1"))
		child1.Node().SetContext(jsii.String("c2"), "child1")
		child1.Node().SetContext(jsii.String("c3"), "child1")
		child2 := constructs.NewConstruct(high, jsii.String("child2"))
		child3 := constructs.NewConstruct(child1, jsii.String("child1child1"))
		child3.Node().SetContext(jsii.String("c1"), "child3")
		child3.Node().SetContext(jsii.String("c4"), "child3")

		cases := []struct {
			node constructs.Node
			key  string
			want interface{}
		}{
			{high.Node(), "c1", "root"},
			{high.Node(), "c2", "root"},
			{high.Node(), "c3", nil},
			{child1.Node(), "c1", "root"},
			{child1.Node(), "c2", "child1"},
			{child1.Node(), "c3", "child1"},
			{child2.Node(), "c1", "root"},
			{child2.Node(), "c2", "root"},
			{child2.Node(), "c3", nil},
			{child3.Node(), "c1", "child3"},
			{child3.Node(), "c2", "child1"},
			{child3.Node(), "c3", "child1"},
			{child3.Node(), "c4", "child3"},
		}
		for _, test := range cases {
			if got := test.node.TryGetContext(&test.key); got != test.want {
				t.Errorf("TryGetContext(%q) = %#v, want %#v", test.key, got, test.want)
			}
		}
	})

	t.Run("cannot be set after children", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		constructs.NewConstruct(root, jsii.String("child1"))
		requirePanicContains(t, "Cannot set context after children have been added: child1", func() {
			root.Node().SetContext(jsii.String("k"), "v")
		})
	})
}

// Ported from:
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L283
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L290
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L296
func TestNodePathsAndSiblingNames(t *testing.T) {
	tree := newConstructTree(nil)
	if got := stringValue(tree.root.Node().Path()); got != "" {
		t.Fatalf("root path = %q", got)
	}
	if got := stringValue(tree.child1_1_1.Node().Path()); got != "HighChild/Child1/Child11/Child111" {
		t.Fatalf("deep path = %q", got)
	}
	if got := stringValue(tree.child2.Node().Path()); got != "HighChild/Child2" {
		t.Fatalf("child path = %q", got)
	}

	namedRoot := constructs.NewRootConstruct(jsii.String("Root"))
	namedChild := constructs.NewConstruct(namedRoot, jsii.String("Child"))
	if got := stringValue(namedChild.Node().Path()); got != "Root/Child" {
		t.Fatalf("named root child path = %q", got)
	}

	root := constructs.NewRootConstruct(nil)
	constructs.NewConstruct(root, jsii.String("SameName"))
	requirePanicContains(t, "There is already a Construct with name 'SameName'", func() {
		constructs.NewConstruct(root, jsii.String("SameName"))
	})
	parent := constructs.NewConstruct(root, jsii.String("c0"))
	constructs.NewConstruct(parent, jsii.String("SameName"))
	requirePanicContains(t, "There is already a Construct with name 'SameName' in Construct [c0]", func() {
		constructs.NewConstruct(parent, jsii.String("SameName"))
	})
}

func addMetadataWithTraceMarker(node constructs.Node) {
	node.AddMetadata(jsii.String("key"), "value", &constructs.MetadataOptions{StackTrace: jsii.Bool(true)})
}

// Ported from:
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L313
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L333
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L345
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L361
func TestMetadata(t *testing.T) {
	t.Run("records data and requested traces", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		construct := constructs.NewConstruct(root, jsii.String("MyConstruct"))
		addMetadataWithTraceMarker(construct.Node())
		construct.Node().AddMetadata(jsii.String("number"), 103, nil)
		construct.Node().AddMetadata(jsii.String("array"), []int{123, 456}, nil)

		metadata := *construct.Node().Metadata()
		if got := len(metadata); got != 3 {
			t.Fatalf("metadata count = %d, want 3", got)
		}
		if got := stringValue(metadata[0].Type); got != "key" || metadata[0].Data != "value" {
			t.Fatalf("first metadata = %#v", metadata[0])
		}
		if metadata[0].Trace == nil || len(*metadata[0].Trace) == 0 || !strings.Contains(stringValue((*metadata[0].Trace)[0]), "addMetadataWithTraceMarker") {
			t.Fatalf("trace = %#v, want marker", metadata[0].Trace)
		}
		if metadata[1].Data != 103 || !reflect.DeepEqual(metadata[2].Data, []int{123, 456}) {
			t.Fatalf("metadata data = %#v", metadata)
		}
	})

	t.Run("honors trace options", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		construct := constructs.NewConstruct(root, jsii.String("Foo"))
		construct.Node().AddMetadata(jsii.String("foo"), "bar1", &constructs.MetadataOptions{StackTrace: jsii.Bool(true)})
		construct.Node().AddMetadata(jsii.String("foo"), "bar2", &constructs.MetadataOptions{StackTrace: jsii.Bool(false)})
		customTrace := []*string{jsii.String("custom/path/file.ts:10"), jsii.String("custom/path/other.ts:20")}
		construct.Node().AddMetadata(jsii.String("foo"), "bar3", &constructs.MetadataOptions{StackTraceOverride: &customTrace})
		construct.Node().AddMetadata(jsii.String("foo"), "bar4", &constructs.MetadataOptions{StackTrace: jsii.Bool(true), StackTraceOverride: &customTrace})
		emptyTrace := []*string{}
		construct.Node().AddMetadata(jsii.String("foo"), "bar5", &constructs.MetadataOptions{StackTraceOverride: &emptyTrace})
		construct.Node().AddMetadata(jsii.String("foo"), "bar6", &constructs.MetadataOptions{StackTrace: jsii.Bool(true)})

		metadata := *construct.Node().Metadata()
		if metadata[0].Trace == nil || len(*metadata[0].Trace) == 0 || metadata[1].Trace != nil {
			t.Fatalf("stack trace options were not honored: %#v", metadata[:2])
		}
		if !reflect.DeepEqual(metadata[2].Trace, &customTrace) || !reflect.DeepEqual(metadata[3].Trace, &customTrace) {
			t.Fatalf("trace overrides = %#v", metadata[2:4])
		}
		if metadata[4].Trace != nil || metadata[5].Trace == nil || len(*metadata[5].Trace) == 0 {
			t.Fatalf("empty and generated traces = %#v", metadata[4:])
		}
	})

	t.Run("ignores nil data", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		construct := constructs.NewConstruct(root, jsii.String("Foo"))
		construct.Node().AddMetadata(jsii.String("Null"), nil, nil)
		construct.Node().AddMetadata(jsii.String("True"), true, nil)
		construct.Node().AddMetadata(jsii.String("False"), false, nil)
		construct.Node().AddMetadata(jsii.String("Empty"), "", nil)

		metadata := *construct.Node().Metadata()
		if got := len(metadata); got != 3 {
			t.Fatalf("metadata count = %d, want 3", got)
		}
		for _, entry := range metadata {
			if stringValue(entry.Type) == "Null" {
				t.Fatal("nil metadata was retained")
			}
		}
	})
}

type staticValidation []string

func (validation staticValidation) Validate() *[]*string {
	values := make([]*string, 0, len(validation))
	for _, value := range validation {
		value := value
		values = append(values, &value)
	}
	return &values
}

func validationMessages(root constructs.IConstruct) []string {
	var visit func(constructs.IConstruct) []string
	visit = func(construct constructs.IConstruct) []string {
		messages := make([]string, 0)
		for _, child := range *construct.Node().Children() {
			messages = append(messages, visit(child)...)
		}
		for _, message := range *construct.Node().Validate() {
			messages = append(messages, stringValue(construct.Node().Path())+":"+stringValue(message))
		}
		return messages
	}
	return visit(root)
}

// Ported from:
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L390
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L455
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L463
func TestValidation(t *testing.T) {
	root := constructs.NewRootConstruct(nil)
	first := constructs.NewConstruct(root, jsii.String("MyConstruct"))
	first.Node().AddValidation(staticValidation{"my-error1", "my-error2"})
	second := constructs.NewConstruct(root, jsii.String("TheirConstruct"))
	child := constructs.NewConstruct(second, jsii.String("YourConstruct"))
	child.Node().AddValidation(staticValidation{"your-error1"})
	second.Node().AddValidation(staticValidation{"their-error"})
	root.Node().AddValidation(staticValidation{"stack-error"})

	if got, want := validationMessages(root), []string{"MyConstruct:my-error1", "MyConstruct:my-error2", "TheirConstruct/YourConstruct:your-error1", "TheirConstruct:their-error", ":stack-error"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("validation messages = %#v, want %#v", got, want)
	}

	empty := constructs.NewRootConstruct(nil)
	if got := *empty.Node().Validate(); len(got) != 0 {
		t.Fatalf("empty validation = %#v", got)
	}
	empty.Node().AddValidation(staticValidation{"error1", "error2"})
	empty.Node().AddValidation(staticValidation{"error3"})
	if got, want := *empty.Node().Validate(), []*string{jsii.String("error1"), jsii.String("error2"), jsii.String("error3")}; !reflect.DeepEqual(got, want) {
		t.Fatalf("validation = %#v, want %#v", got, want)
	}
}

// Ported from:
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L472
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L497
func TestLockAndTraversal(t *testing.T) {
	t.Run("locks a subtree", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		left := constructs.NewConstruct(root, jsii.String("c0a"))
		right := constructs.NewConstruct(root, jsii.String("c0b"))
		leftChild := constructs.NewConstruct(left, jsii.String("c1a"))
		leftSibling := constructs.NewConstruct(left, jsii.String("c1b"))
		left.Node().Lock()
		constructs.NewConstruct(right, jsii.String("c1a"))
		requirePanicContains(t, `Cannot add children to "c0a" during synthesis`, func() { constructs.NewConstruct(left, jsii.String("fail1")) })
		requirePanicContains(t, `Cannot add children to "c0a/c1a" during synthesis`, func() { constructs.NewConstruct(leftChild, jsii.String("fail2")) })
		requirePanicContains(t, `Cannot add children to "c0a/c1b" during synthesis`, func() { constructs.NewConstruct(leftSibling, jsii.String("fail3")) })
		constructs.NewConstruct(root, jsii.String("c2"))
		root.Node().Lock()
		requirePanicContains(t, "Cannot add children during synthesis", func() { constructs.NewConstruct(root, jsii.String("test")) })
	})

	t.Run("finds all in depth-first order", func(t *testing.T) {
		root := constructs.NewRootConstruct(jsii.String("1"))
		child := constructs.NewConstruct(root, jsii.String("2"))
		constructs.NewConstruct(root, jsii.String("3"))
		constructs.NewConstruct(child, jsii.String("4"))
		constructs.NewConstruct(child, jsii.String("5"))
		if got, want := constructIDs(*root.Node().FindAll("")), []string{"1", "2", "4", "5", "3"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("default order = %#v, want %#v", got, want)
		}
		if got, want := constructIDs(*root.Node().FindAll(constructs.ConstructOrder_PREORDER)), []string{"1", "2", "4", "5", "3"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("preorder = %#v, want %#v", got, want)
		}
		if got, want := constructIDs(*root.Node().FindAll(constructs.ConstructOrder_POSTORDER)), []string{"4", "5", "2", "3", "1"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("postorder = %#v, want %#v", got, want)
		}
	})
}

// Ported from:
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L512
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L517
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L525
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L533
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L541
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L549
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L556
func TestScopesRootsAndDefaultChild(t *testing.T) {
	tree := newConstructTree(nil)
	if got, want := constructIDs(*tree.child1_1_1.Node().Scopes()), []string{"", "HighChild", "Child1", "Child11", "Child111"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("scopes = %#v, want %#v", got, want)
	}
	for _, construct := range []constructs.Construct{tree.child1, tree.child2, tree.child1_1_1} {
		if construct.Node().Root() != tree.root {
			t.Fatal("node root did not return the tree root")
		}
	}

	t.Run("recognizes Resource and Default", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		resource := constructs.NewConstruct(root, jsii.String("Resource"))
		if root.Node().DefaultChild() != resource {
			t.Fatal("Resource was not the default child")
		}
		root = constructs.NewRootConstruct(nil)
		defaultChild := constructs.NewConstruct(root, jsii.String("Default"))
		if root.Node().DefaultChild() != defaultChild {
			t.Fatal("Default was not the default child")
		}
	})

	t.Run("can be overridden or absent", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		constructs.NewConstruct(root, jsii.String("Resource"))
		override := constructs.NewConstruct(root, jsii.String("OtherResource"))
		root.Node().SetDefaultChild(override)
		if root.Node().DefaultChild() != override {
			t.Fatal("explicit default child was not returned")
		}
		root = constructs.NewRootConstruct(nil)
		constructs.NewConstruct(root, jsii.String("child1"))
		if root.Node().DefaultChild() != nil {
			t.Fatal("unexpected default child")
		}
	})

	t.Run("rejects both Resource and Default", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		constructs.NewConstruct(root, jsii.String("Default"))
		constructs.NewConstruct(root, jsii.String("Resource"))
		requirePanicContains(t, `There is both a child with id "Resource" and id "Default"`, func() {
			root.Node().DefaultChild()
		})
	})
}

type externalDependable struct{}

type externalDependableTrait struct{ roots []constructs.IConstruct }

func (trait externalDependableTrait) DependencyRoots() *[]constructs.IConstruct {
	roots := append([]constructs.IConstruct(nil), trait.roots...)
	return &roots
}

// Ported from:
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L569
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L584
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L601
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L616
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L630
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L648
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L652
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L671
func TestDependencies(t *testing.T) {
	t.Run("add, deduplicate, and remove direct dependencies", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		consumer := constructs.NewConstruct(root, jsii.String("consumer"))
		producer1 := constructs.NewConstruct(root, jsii.String("producer1"))
		producer2 := constructs.NewConstruct(root, jsii.String("producer2"))
		consumer.Node().AddDependency(producer1, producer2, producer1, producer1)
		if got, want := constructPaths(*consumer.Node().Dependencies()), []string{"producer1", "producer2"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("dependencies = %#v, want %#v", got, want)
		}
		consumer.Node().RemoveDependency(producer1)
		if got, want := constructPaths(*consumer.Node().Dependencies()), []string{"producer2"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("dependencies after removal = %#v, want %#v", got, want)
		}
	})

	t.Run("groups expand lazily and can nest", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		consumer := constructs.NewConstruct(root, jsii.String("consumer"))
		groupA := constructs.NewDependencyGroup(constructs.NewConstruct(root, jsii.String("a1")), constructs.NewConstruct(root, jsii.String("a2")))
		groupB := constructs.NewDependencyGroup(constructs.NewConstruct(root, jsii.String("b1")), constructs.NewConstruct(root, jsii.String("b2")))
		composite := constructs.NewDependencyGroup(groupA)
		consumer.Node().AddDependency(composite)
		composite.Add(groupB)
		groupB.Add(constructs.NewConstruct(root, jsii.String("b3")))
		if got, want := constructPaths(*consumer.Node().Dependencies()), []string{"a1", "a2", "b1", "b2", "b3"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("group dependencies = %#v, want %#v", got, want)
		}
	})

	t.Run("flat groups expand after later additions", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		consumer := constructs.NewConstruct(root, jsii.String("consumer"))
		group := constructs.NewDependencyGroup(constructs.NewConstruct(root, jsii.String("producer1")), constructs.NewConstruct(root, jsii.String("producer2")))
		group.Add(constructs.NewConstruct(root, jsii.String("producer3")), constructs.NewConstruct(root, jsii.String("producer4")))
		consumer.Node().AddDependency(group)
		group.Add(constructs.NewConstruct(root, jsii.String("producer5")))
		if got, want := constructPaths(*consumer.Node().Dependencies()), []string{"producer1", "producer2", "producer3", "producer4", "producer5"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("flat group dependencies = %#v, want %#v", got, want)
		}
	})

	t.Run("traits can be attached to external values", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		producer := constructs.NewConstruct(root, jsii.String("producer"))
		consumer := constructs.NewConstruct(root, jsii.String("consumer"))
		value := &externalDependable{}
		constructs.Dependable_Implement(value, externalDependableTrait{roots: []constructs.IConstruct{producer}})
		consumer.Node().AddDependency(value)
		if got, want := constructPaths(*constructs.Dependable_Of(value).DependencyRoots()), []string{"producer"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("trait roots = %#v, want %#v", got, want)
		}
		if got, want := constructPaths(*consumer.Node().Dependencies()), []string{"producer"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("dependencies = %#v, want %#v", got, want)
		}
	})

	t.Run("rejects values without a trait", func(t *testing.T) {
		requirePanicContains(t, "does not implement IDependable", func() {
			constructs.Dependable_Of(struct{}{})
		})
	})
}

// Ported from:
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L691
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L706
func TestRemovingChildrenAndStringRepresentation(t *testing.T) {
	root := constructs.NewRootConstruct(nil)
	constructs.NewConstruct(root, jsii.String("child1"))
	constructs.NewConstruct(root, jsii.String("child2"))
	if removed := root.Node().TryRemoveChild(jsii.String("child1")); removed == nil || !*removed {
		t.Fatal("TryRemoveChild did not remove child1")
	}
	if removed := root.Node().TryRemoveChild(jsii.String("child-not-found")); removed == nil || *removed {
		t.Fatal("TryRemoveChild removed a missing child")
	}
	if got := len(*root.Node().Children()); got != 1 {
		t.Fatalf("child count = %d, want 1", got)
	}

	child := constructs.NewConstruct(root, jsii.String("child"))
	grandchild := constructs.NewConstruct(child, jsii.String("grand"))
	if got := stringValue(root.ToString()); got != "<root>" {
		t.Fatalf("root string = %q", got)
	}
	if got := stringValue(child.ToString()); got != "child" {
		t.Fatalf("child string = %q", got)
	}
	if got := stringValue(grandchild.ToString()); got != "child/grand" {
		t.Fatalf("grandchild string = %q", got)
	}
}

type embeddedConstruct struct{ constructs.Construct }

// Ported from:
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L718
func TestConstructIsConstruct(t *testing.T) {
	root := constructs.NewRootConstruct(nil)
	subclass := &embeddedConstruct{}
	constructs.NewConstruct_Override(subclass, root, jsii.String("subclass"))
	for _, value := range []interface{}{root, subclass} {
		if result := constructs.Construct_IsConstruct(value); result == nil || !*result {
			t.Fatalf("Construct_IsConstruct(%T) = %#v, want true", value, result)
		}
	}
	for _, value := range []interface{}{nil, "string", 1234, true, []int{1, 2, 3}, struct{}{}} {
		if result := constructs.Construct_IsConstruct(value); result == nil || *result {
			t.Errorf("Construct_IsConstruct(%#v) = %#v, want false", value, result)
		}
	}
}

type testMixin struct {
	supports func(constructs.IConstruct) *bool
	apply    func(constructs.IConstruct)
}

func (mixin testMixin) Supports(construct constructs.IConstruct) *bool {
	return mixin.supports(construct)
}

func (mixin testMixin) ApplyTo(construct constructs.IConstruct) {
	mixin.apply(construct)
}

// Ported from:
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L768
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L774
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L787
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L806
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L823
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L839
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L854
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L873
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L885
// https://github.com/aws/constructs/blob/9f11c0801cd9623e8edddd0f27b3dbc2118d541b/test/construct.test.ts#L904
func TestMixins(t *testing.T) {
	t.Run("return the construct and apply supported mixins", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		var applied []constructs.IConstruct
		mixin := testMixin{
			supports: func(constructs.IConstruct) *bool { return jsii.Bool(true) },
			apply:    func(construct constructs.IConstruct) { applied = append(applied, construct) },
		}
		if got := root.With(mixin); got != root {
			t.Fatal("With did not return the root")
		}
		if len(applied) != 1 || applied[0] != root {
			t.Fatalf("applied = %#v, want root", applied)
		}
	})

	t.Run("walk supported constructs in order", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		child1 := constructs.NewConstruct(root, jsii.String("child1"))
		constructs.NewConstruct(root, jsii.String("child2"))
		constructs.NewConstruct(child1, jsii.String("grandchild"))
		var first, second []string
		all := testMixin{
			supports: func(constructs.IConstruct) *bool { return jsii.Bool(true) },
			apply: func(construct constructs.IConstruct) {
				id := stringValue(construct.Node().Id())
				if id == "" {
					id = "root"
				}
				first = append(first, id)
			},
		}
		onlyChild1 := testMixin{
			supports: func(construct constructs.IConstruct) *bool {
				return jsii.Bool(stringValue(construct.Node().Id()) == "child1")
			},
			apply: func(construct constructs.IConstruct) { second = append(second, stringValue(construct.Node().Id())) },
		}
		root.With(all, onlyChild1)
		if got, want := first, []string{"root", "child1", "grandchild", "child2"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("first mixin = %#v, want %#v", got, want)
		}
		if got, want := second, []string{"child1"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("second mixin = %#v, want %#v", got, want)
		}
	})

	t.Run("complete each mixin before the next", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		constructs.NewConstruct(root, jsii.String("child"))
		var order []string
		newMixin := func(name string) testMixin {
			return testMixin{
				supports: func(constructs.IConstruct) *bool { return jsii.Bool(true) },
				apply: func(construct constructs.IConstruct) {
					id := stringValue(construct.Node().Id())
					if id == "" {
						id = "root"
					}
					order = append(order, name+":"+id)
				},
			}
		}
		root.With(newMixin("m1"), newMixin("m2"))
		if got, want := order, []string{"m1:root", "m1:child", "m2:root", "m2:child"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("mixin order = %#v, want %#v", got, want)
		}
	})

	t.Run("honor false support, chaining, and metadata changes", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		var first, second []constructs.IConstruct
		mixin1 := testMixin{supports: func(constructs.IConstruct) *bool { return jsii.Bool(true) }, apply: func(construct constructs.IConstruct) { first = append(first, construct) }}
		mixin2 := testMixin{supports: func(constructs.IConstruct) *bool { return jsii.Bool(true) }, apply: func(construct constructs.IConstruct) { second = append(second, construct) }}
		root.With(mixin1).With(mixin2)
		if len(first) != 1 || len(second) != 1 {
			t.Fatalf("chained mixins = %#v, %#v", first, second)
		}
		root.With(testMixin{supports: func(constructs.IConstruct) *bool { return jsii.Bool(false) }, apply: func(constructs.IConstruct) { t.Fatal("unsupported mixin applied") }})
		root.With(testMixin{
			supports: func(constructs.IConstruct) *bool { return jsii.Bool(true) },
			apply: func(construct constructs.IConstruct) {
				construct.Node().AddMetadata(jsii.String("mixin-applied"), true, nil)
			},
		})
		metadata := *root.Node().Metadata()
		if len(metadata) != 1 || stringValue(metadata[0].Type) != "mixin-applied" || metadata[0].Data != true || metadata[0].Trace != nil {
			t.Fatalf("mixin metadata = %#v", metadata)
		}
	})

	t.Run("capture the tree before applying", func(t *testing.T) {
		root := constructs.NewRootConstruct(nil)
		adding := testMixin{
			supports: func(construct constructs.IConstruct) *bool {
				return jsii.Bool(stringValue(construct.Node().Id()) == "")
			},
			apply: func(construct constructs.IConstruct) {
				constructs.NewConstruct(construct.(constructs.Construct), jsii.String("added-by-mixin"))
			},
		}
		var applied []string
		tracking := testMixin{
			supports: func(constructs.IConstruct) *bool { return jsii.Bool(true) },
			apply:    func(construct constructs.IConstruct) { applied = append(applied, stringValue(construct.Node().Id())) },
		}
		root.With(adding, tracking)
		if got, want := applied, []string{""}; !reflect.DeepEqual(got, want) {
			t.Fatalf("applied = %#v, want %#v", got, want)
		}
		if root.Node().FindChild(jsii.String("added-by-mixin")) == nil {
			t.Fatal("mixin-added child was not found")
		}
	})
}

type customConstruct struct {
	constructs.Construct
}

type forwardingConstruct struct {
	constructs.IConstruct
}

func TestOverrideSupportsEmbeddedConstruct(t *testing.T) {
	root := constructs.NewRootConstruct(jsii.String(""))
	custom := &customConstruct{}
	constructs.NewConstruct_Override(custom, root, jsii.String("custom"))

	if got := stringValue(custom.Node().Path()); got != "custom" {
		t.Fatalf("path = %q", got)
	}
	if got := root.Node().TryFindChild(jsii.String("custom")); got != custom {
		t.Fatalf("scope child = %#v, want custom construct", got)
	}
}

func TestGetAllContextTreeValuesOverrideDefaults(t *testing.T) {
	root := constructs.NewRootConstruct(jsii.String("root"))
	root.Node().SetContext(jsii.String("shared"), "root")
	child := constructs.NewConstruct(root, jsii.String("child"))
	child.Node().SetContext(jsii.String("shared"), "child")

	defaults := map[string]interface{}{"shared": "default", "only-default": true}
	context := child.Node().GetAllContext(&defaults).(map[string]interface{})
	if got := context["shared"]; got != "child" {
		t.Fatalf("shared context = %#v, want closest tree value", got)
	}
	if got := context["only-default"]; got != true {
		t.Fatalf("only-default context = %#v, want default", got)
	}
}

func TestTryRemoveChildUsesStoredUnsanitizedKey(t *testing.T) {
	root := constructs.NewRootConstruct(jsii.String("root"))
	constructs.NewConstruct(root, jsii.String("a/b"))

	if removed := root.Node().TryRemoveChild(jsii.String("a/b")); *removed {
		t.Fatal("TryRemoveChild unexpectedly sanitized the supplied id")
	}
	if removed := root.Node().TryRemoveChild(jsii.String("a--b")); !*removed {
		t.Fatal("TryRemoveChild did not remove the stored child id")
	}
}

func TestDependenciesPreserveDuplicateRootsAcrossDependables(t *testing.T) {
	root := constructs.NewRootConstruct(jsii.String("root"))
	first := constructs.NewConstruct(root, jsii.String("first"))
	second := constructs.NewConstruct(root, jsii.String("second"))
	first.Node().AddDependency(constructs.NewDependencyGroup(second), constructs.NewDependencyGroup(second))

	dependencies := *first.Node().Dependencies()
	if len(dependencies) != 2 || dependencies[0] != second || dependencies[1] != second {
		t.Fatalf("dependencies = %#v, want duplicate roots from distinct dependables", dependencies)
	}
}

func TestDependenciesAcceptForwardingConstructWrappers(t *testing.T) {
	root := constructs.NewRootConstruct(jsii.String("root"))
	source := constructs.NewConstruct(root, jsii.String("source"))
	target := constructs.NewConstruct(root, jsii.String("target"))
	wrapper := &forwardingConstruct{IConstruct: target}

	source.Node().AddDependency(wrapper)
	dependencies := *source.Node().Dependencies()
	if len(dependencies) != 1 || dependencies[0] != target {
		t.Fatalf("dependencies = %#v, want underlying construct %#v", dependencies, target)
	}
}

func TestRemoveDependencyRequiresTheSameDependable(t *testing.T) {
	root := constructs.NewRootConstruct(jsii.String("root"))
	source := constructs.NewConstruct(root, jsii.String("source"))
	target := constructs.NewConstruct(root, jsii.String("target"))

	dependency := constructs.NewDependencyGroup(target)
	source.Node().AddDependency(dependency)
	source.Node().RemoveDependency(constructs.NewDependencyGroup(target))
	if got := *source.Node().Dependencies(); len(got) != 1 || got[0] != target {
		t.Fatalf("dependencies after removing another dependable = %#v, want %#v", got, target)
	}

	source.Node().RemoveDependency(dependency)
	if got := *source.Node().Dependencies(); len(got) != 0 {
		t.Fatalf("dependencies after removal = %#v, want none", got)
	}
}

func TestChildrenMatchJavaScriptObjectKeyOrder(t *testing.T) {
	root := constructs.NewRootConstruct(jsii.String(""))
	for _, id := range []string{"10", "2", "01", "4294967294", "4294967295"} {
		constructs.NewConstruct(root, jsii.String(id))
	}

	if got, want := constructIDs(*root.Node().Children()), []string{"2", "10", "4294967294", "01", "4294967295"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("child ids = %#v, want %#v", got, want)
	}
}
