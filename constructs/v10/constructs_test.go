package constructs

import "testing"

type customConstruct struct {
	Construct
}

type forwardingConstruct struct {
	IConstruct
}

func TestOverrideSupportsEmbeddedConstruct(t *testing.T) {
	rootID := ""
	root := NewRootConstruct(&rootID)
	id := "custom"
	custom := &customConstruct{}
	NewConstruct_Override(custom, root, &id)

	if got := *custom.Node().Path(); got != "custom" {
		t.Fatalf("path = %q", got)
	}
	if got := root.Node().TryFindChild(&id); got != custom {
		t.Fatalf("scope child = %#v, want custom construct", got)
	}
}

func TestAddressMatchesConstructsV10(t *testing.T) {
	rootID, chartID, resourceID := "", "getting-started", "deployment"
	root := NewRootConstruct(&rootID)
	chart := NewConstruct(root, &chartID)
	resource := NewConstruct(chart, &resourceID)
	if got, want := *resource.Node().Addr(), "c80c72572fe2d7984ea62b249ce2102ad5ed4b5c25"; got != want {
		t.Fatalf("addr = %q, want %q", got, want)
	}
}

func TestGetAllContextDefaultsOverrideTreeValues(t *testing.T) {
	rootID, childID, key := "root", "child", "shared"
	root := NewRootConstruct(&rootID)
	root.Node().SetContext(&key, "root")
	child := NewConstruct(root, &childID)
	child.Node().SetContext(&key, "child")

	defaults := map[string]interface{}{"shared": "default", "only-default": true}
	context := child.Node().GetAllContext(&defaults).(map[string]interface{})
	if got := context["shared"]; got != "default" {
		t.Fatalf("shared context = %#v, want defaults override", got)
	}
}

func TestTryRemoveChildUsesStoredUnsanitizedKey(t *testing.T) {
	rootID, slashID := "root", "a/b"
	root := NewRootConstruct(&rootID)
	NewConstruct(root, &slashID)

	if removed := root.Node().TryRemoveChild(&slashID); *removed {
		t.Fatal("TryRemoveChild unexpectedly sanitized the supplied id")
	}
	storedID := "a--b"
	if removed := root.Node().TryRemoveChild(&storedID); !*removed {
		t.Fatal("TryRemoveChild did not remove the stored child id")
	}
}

func TestDependenciesPreserveDuplicateRootsAcrossDependables(t *testing.T) {
	rootID, firstID, secondID := "root", "first", "second"
	root := NewRootConstruct(&rootID)
	first := NewConstruct(root, &firstID)
	second := NewConstruct(root, &secondID)
	first.Node().AddDependency(NewDependencyGroup(second), NewDependencyGroup(second))

	dependencies := *first.Node().Dependencies()
	if len(dependencies) != 2 || dependencies[0] != second || dependencies[1] != second {
		t.Fatalf("dependencies = %#v, want duplicate roots from distinct dependables", dependencies)
	}
}

func TestDependenciesAcceptForwardingConstructWrappers(t *testing.T) {
	rootID, sourceID, targetID := "root", "source", "target"
	root := NewRootConstruct(&rootID)
	source := NewConstruct(root, &sourceID)
	target := NewConstruct(root, &targetID)
	wrapper := &forwardingConstruct{IConstruct: target}

	source.Node().AddDependency(wrapper)
	dependencies := *source.Node().Dependencies()
	if len(dependencies) != 1 || dependencies[0] != target {
		t.Fatalf("dependencies = %#v, want underlying construct %#v", dependencies, target)
	}
}

func TestRemoveDependencyRequiresTheSameDependable(t *testing.T) {
	rootID, sourceID, targetID := "root", "source", "target"
	root := NewRootConstruct(&rootID)
	source := NewConstruct(root, &sourceID)
	target := NewConstruct(root, &targetID)

	dependency := NewDependencyGroup(target)
	source.Node().AddDependency(dependency)
	source.Node().RemoveDependency(NewDependencyGroup(target))
	if got := *source.Node().Dependencies(); len(got) != 1 || got[0] != target {
		t.Fatalf("dependencies after removing another dependable = %#v, want %#v", got, target)
	}

	source.Node().RemoveDependency(dependency)
	if got := *source.Node().Dependencies(); len(got) != 0 {
		t.Fatalf("dependencies after removal = %#v, want none", got)
	}
}

func TestChildrenMatchJavaScriptObjectKeyOrder(t *testing.T) {
	rootID := ""
	root := NewRootConstruct(&rootID)
	for _, id := range []string{"10", "2", "01", "4294967294", "4294967295"} {
		id := id
		NewConstruct(root, &id)
	}

	got := make([]string, 0)
	for _, child := range *root.Node().Children() {
		got = append(got, *child.Node().Id())
	}
	want := []string{"2", "10", "4294967294", "01", "4294967295"}
	if len(got) != len(want) {
		t.Fatalf("child ids = %#v, want %#v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("child ids = %#v, want %#v", got, want)
		}
	}
}
