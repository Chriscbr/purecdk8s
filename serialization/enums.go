// Package serialization contains the small native runtime registry used by
// generated purecdk8s bindings when a Go enum's public symbolic value differs
// from its Kubernetes wire value.
package serialization

import (
	"reflect"
	"sync"
)

var enumWireValues = struct {
	sync.RWMutex
	byType map[reflect.Type]map[string]interface{}
}{
	byType: make(map[reflect.Type]map[string]interface{}),
}

// RegisterEnumWireValues associates the symbolic values of a generated Go
// enum with the original values from its source schema.
func RegisterEnumWireValues[T ~string](values map[T]interface{}) {
	enumType := reflect.TypeOf(*new(T))
	registered := make(map[string]interface{}, len(values))
	for symbol, wireValue := range values {
		registered[string(symbol)] = wireValue
	}

	enumWireValues.Lock()
	existing := enumWireValues.byType[enumType]
	if existing == nil {
		existing = make(map[string]interface{}, len(registered))
		enumWireValues.byType[enumType] = existing
	}
	for symbol, wireValue := range registered {
		existing[symbol] = wireValue
	}
	enumWireValues.Unlock()
}

// EnumWireValue returns the registered Kubernetes wire value for value.
//
// Values of a registered enum type that do not match a declared member retain
// their symbolic string. Ordinary strings and unrelated named string types are
// not treated as enums.
func EnumWireValue(value interface{}) (interface{}, bool) {
	if value == nil {
		return nil, false
	}
	reflected := reflect.ValueOf(value)
	for reflected.Kind() == reflect.Interface || reflected.Kind() == reflect.Pointer {
		if reflected.IsNil() {
			return nil, false
		}
		reflected = reflected.Elem()
	}
	if reflected.Kind() != reflect.String {
		return nil, false
	}

	enumWireValues.RLock()
	values, registered := enumWireValues.byType[reflected.Type()]
	if !registered {
		enumWireValues.RUnlock()
		return nil, false
	}
	wireValue, found := values[reflected.String()]
	enumWireValues.RUnlock()
	if found {
		return wireValue, true
	}
	return reflected.String(), true
}
