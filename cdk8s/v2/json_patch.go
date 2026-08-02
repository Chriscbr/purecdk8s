package cdk8s

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type jsonPatchOperation struct {
	op    string
	path  string
	from  string
	value interface{}
}

type jsonPatchImpl struct {
	operation jsonPatchOperation
}

// Adds a value to an object or inserts it into an array.
//
// In the case of an array, the value is inserted before the given index. The - character can be used instead of an index to insert at the end of an array.
//
// Example:
//
//	JsonPatch.add('/biscuits/1', { "name": "Ginger Nut" })
func JsonPatch_Add(path *string, value interface{}) JsonPatch {
	requirePatchPath(path, "path")
	requirePatchValue(value)
	return &jsonPatchImpl{operation: jsonPatchOperation{op: "add", path: *path, value: value}}
}

// Applies a set of JSON-Patch (RFC-6902) operations to `document` and returns the result.
//
// Returns: The result document.
func JsonPatch_Apply(document interface{}, ops ...JsonPatch) interface{} {
	if document == nil {
		panic("parameter document is required, but nil was provided")
	}
	result := cloneJSONValue(document)
	for _, item := range ops {
		patch, ok := item.(*jsonPatchImpl)
		if !ok || patch == nil {
			panic("invalid JsonPatch operation")
		}
		result = applyJSONPatchOperation(result, patch.operation)
	}
	return result
}

// Copies a value from one location to another within the JSON document.
//
// Both from and path are JSON Pointers.
//
// Example:
//
//	JsonPatch.copy('/biscuits/0', '/best_biscuit')
func JsonPatch_Copy(from *string, path *string) JsonPatch {
	requirePatchPath(from, "from")
	requirePatchPath(path, "path")
	return &jsonPatchImpl{operation: jsonPatchOperation{op: "copy", from: *from, path: *path}}
}

// Moves a value from one location to the other.
//
// Both from and path are JSON Pointers.
//
// Example:
//
//	JsonPatch.move('/biscuits', '/cookies')
func JsonPatch_Move(from *string, path *string) JsonPatch {
	requirePatchPath(from, "from")
	requirePatchPath(path, "path")
	return &jsonPatchImpl{operation: jsonPatchOperation{op: "move", from: *from, path: *path}}
}

// Removes a value from an object or array.
//
// Example:
//
//	JsonPatch.remove('/biscuits/0')
func JsonPatch_Remove(path *string) JsonPatch {
	requirePatchPath(path, "path")
	return &jsonPatchImpl{operation: jsonPatchOperation{op: "remove", path: *path}}
}

// Replaces a value.
//
// Equivalent to a “remove” followed by an “add”.
//
// Example:
//
//	JsonPatch.replace('/biscuits/0/name', 'Chocolate Digestive')
func JsonPatch_Replace(path *string, value interface{}) JsonPatch {
	requirePatchPath(path, "path")
	requirePatchValue(value)
	return &jsonPatchImpl{operation: jsonPatchOperation{op: "replace", path: *path, value: value}}
}

// Tests that the specified value is set in the document.
//
// If the test fails, then the patch as a whole should not apply.
//
// Example:
//
//	JsonPatch.test('/best_biscuit/name', 'Choco Leibniz')
func JsonPatch_Test(path *string, value interface{}) JsonPatch {
	requirePatchPath(path, "path")
	requirePatchValue(value)
	return &jsonPatchImpl{operation: jsonPatchOperation{op: "test", path: *path, value: value}}
}

func requirePatchPath(path *string, parameter string) {
	if path == nil {
		panic(fmt.Sprintf("parameter %s is required, but nil was provided", parameter))
	}
}

func requirePatchValue(value interface{}) {
	if value == nil {
		panic("parameter value is required, but nil was provided")
	}
}

func applyJSONPatchOperation(document interface{}, operation jsonPatchOperation) interface{} {
	switch operation.op {
	case "add", "replace":
		value := cloneJSONValue(operation.value)
		result, _ := patchJSONAt(document, parseJSONPointer(operation.path), operation.op, value)
		return result
	case "remove":
		result, _ := patchJSONAt(document, parseJSONPointer(operation.path), "remove", nil)
		return result
	case "test":
		actual, found := jsonPointerLookup(document, parseJSONPointer(operation.path))
		expected := cloneJSONValue(operation.value)
		if !found || !reflect.DeepEqual(actual, expected) {
			panic("Test operation failed")
		}
		return document
	case "copy":
		value := cloneJSONValue(jsonPointerValue(document, parseJSONPointer(operation.from)))
		result, _ := patchJSONAt(document, parseJSONPointer(operation.path), "add", value)
		return result
	case "move":
		from := parseJSONPointer(operation.from)
		value, found := jsonPointerLookup(document, from)
		without, _ := patchJSONAt(document, from, "remove", nil)
		if !found {
			// fast-json-patch assigns JavaScript undefined when the source is
			// absent; that property disappears when crossing the Go boundary.
			return without
		}
		result, _ := patchJSONAt(without, parseJSONPointer(operation.path), "add", value)
		return result
	default:
		return document
	}
}

func parseJSONPointer(pointer string) []string {
	if pointer == "" {
		return nil
	}
	if !strings.HasPrefix(pointer, "/") {
		panic(fmt.Sprintf("invalid JSON pointer: %s", pointer))
	}
	raw := strings.Split(pointer[1:], "/")
	for index, token := range raw {
		token = strings.ReplaceAll(token, "~1", "/")
		token = strings.ReplaceAll(token, "~0", "~")
		if token == "__proto__" || (token == "prototype" && index > 0 && raw[index-1] == "constructor") {
			panic("JSON-Patch: modifying `__proto__` or `constructor/prototype` prop is banned for security reasons")
		}
		raw[index] = token
	}
	return raw
}

func jsonPointerValue(document interface{}, path []string) interface{} {
	value, _ := jsonPointerLookup(document, path)
	return value
}

func jsonPointerLookup(document interface{}, path []string) (interface{}, bool) {
	current := document
	for _, token := range path {
		switch value := current.(type) {
		case map[string]interface{}:
			item, ok := value[token]
			if !ok {
				return nil, false
			}
			current = item
		case []interface{}:
			index := patchArrayIndex(token, len(value), false)
			if index < 0 || index >= len(value) {
				return nil, false
			}
			current = value[index]
		default:
			panic("Cannot perform operation at the desired path")
		}
	}
	return current, true
}

func patchJSONAt(document interface{}, path []string, operation string, value interface{}) (interface{}, interface{}) {
	if len(path) == 0 {
		switch operation {
		case "add", "replace":
			return value, document
		case "remove":
			return nil, document
		default:
			return document, nil
		}
	}

	token := path[0]
	if len(path) == 1 {
		switch target := document.(type) {
		case map[string]interface{}:
			removed := target[token]
			switch operation {
			case "add", "replace":
				target[token] = value
			case "remove":
				delete(target, token)
			}
			return target, removed
		case []interface{}:
			index := patchArrayIndex(token, len(target), operation == "add")
			switch operation {
			case "add":
				if index > len(target) {
					index = len(target)
				}
				target = append(target, nil)
				copy(target[index+1:], target[index:])
				target[index] = value
				return target, nil
			case "replace":
				if index >= len(target) {
					extended := make([]interface{}, index+1)
					copy(extended, target)
					target = extended
				}
				removed := target[index]
				target[index] = value
				return target, removed
			case "remove":
				if index >= len(target) {
					return target, nil
				}
				removed := target[index]
				copy(target[index:], target[index+1:])
				target[len(target)-1] = nil
				target = target[:len(target)-1]
				return target, removed
			}
		default:
			panic("Cannot perform operation at the desired path")
		}
	}

	switch target := document.(type) {
	case map[string]interface{}:
		child, ok := target[token]
		if !ok || child == nil {
			panic("Cannot perform operation at the desired path")
		}
		updated, removed := patchJSONAt(child, path[1:], operation, value)
		target[token] = updated
		return target, removed
	case []interface{}:
		index := patchArrayIndex(token, len(target), false)
		if index < 0 || index >= len(target) || target[index] == nil {
			panic("Cannot perform operation at the desired path")
		}
		updated, removed := patchJSONAt(target[index], path[1:], operation, value)
		target[index] = updated
		return target, removed
	default:
		panic("Cannot perform operation at the desired path")
	}
}

func patchArrayIndex(token string, length int, adding bool) int {
	if token == "-" {
		if adding {
			return length
		}
		return length
	}
	if token == "" {
		return 0
	}
	for _, character := range token {
		if character < '0' || character > '9' {
			panic(fmt.Sprintf("invalid JSON array index: %s", token))
		}
	}
	index, err := strconv.ParseUint(token, 10, 31)
	if err != nil {
		panic(fmt.Sprintf("invalid JSON array index: %s", token))
	}
	return int(index)
}

func cloneJSONValue(value interface{}) interface{} {
	switch value.(type) {
	case string, bool, float64, nil:
		return value
	}
	data, err := json.Marshal(plainValue(value))
	if err != nil {
		panic(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	var result interface{}
	if err := decoder.Decode(&result); err != nil {
		panic(err)
	}
	return result
}
