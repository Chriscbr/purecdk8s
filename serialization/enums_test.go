package serialization

import "testing"

type (
	testMode      string
	unrelatedMode string
)

const testModeDeleteOptions testMode = "DELETE_OPTIONS"

func TestEnumWireValue(t *testing.T) {
	RegisterEnumWireValues(map[testMode]interface{}{
		testModeDeleteOptions: "DeleteOptions",
	})

	if got, ok := EnumWireValue(testModeDeleteOptions); !ok || got != "DeleteOptions" {
		t.Fatalf("registered enum = (%#v, %v)", got, ok)
	}
	unknown := testMode("UNKNOWN")
	if got, ok := EnumWireValue(&unknown); !ok || got != "UNKNOWN" {
		t.Fatalf("unknown registered enum member = (%#v, %v)", got, ok)
	}
	if got, ok := EnumWireValue(unrelatedMode("DELETE_OPTIONS")); ok || got != nil {
		t.Fatalf("unregistered named string = (%#v, %v)", got, ok)
	}
	if got, ok := EnumWireValue("DELETE_OPTIONS"); ok || got != nil {
		t.Fatalf("ordinary string = (%#v, %v)", got, ok)
	}
}
