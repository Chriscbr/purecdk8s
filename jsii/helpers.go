// Package jsii provides the pointer helpers commonly used by generated cdk8s
// Go programs. It intentionally contains no JSII kernel or Node.js bridge.
package jsii

import (
	"fmt"
	"time"
)

type basicType interface {
	bool | string | float64 | time.Time
}

type numberType interface {
	~float32 | ~float64 | ~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

func Ptr[T basicType](v T) *T { return &v }

func PtrSlice[T basicType](values ...T) *[]*T {
	result := make([]*T, len(values))
	for i := range values {
		value := values[i]
		result[i] = &value
	}
	return &result
}

func String(v string) *string { return &v }

func Strings(values ...string) *[]*string { return PtrSlice(values...) }

func Bool(v bool) *bool { return &v }

func Bools(values ...bool) *[]*bool { return PtrSlice(values...) }

func Number[T numberType](v T) *float64 {
	value := float64(v)
	return &value
}

func Numbers[T numberType](values ...T) *[]*float64 {
	result := make([]*float64, len(values))
	for i := range values {
		value := float64(values[i])
		result[i] = &value
	}
	return &result
}

func AnyNumbers[T numberType](values ...T) *[]interface{} {
	result := make([]interface{}, len(values))
	for i := range values {
		result[i] = float64(values[i])
	}
	return &result
}

func AnyStrings(values ...string) *[]interface{} {
	result := make([]interface{}, len(values))
	for i := range values {
		result[i] = values[i]
	}
	return &result
}

func AnySlice[T any](values *[]*T) *[]interface{} {
	if values == nil {
		return nil
	}
	result := make([]interface{}, len(*values))
	for i := range *values {
		if (*values)[i] != nil {
			result[i] = *(*values)[i]
		}
	}
	return &result
}

func Time(v time.Time) *time.Time { return &v }

func Times(values ...time.Time) *[]*time.Time {
	result := make([]*time.Time, len(values))
	for i := range values {
		value := values[i]
		result[i] = &value
	}
	return &result
}

func Sprintf(format string, values ...interface{}) *string {
	result := fmt.Sprintf(format, values...)
	return &result
}

// Close is retained for source compatibility. There is no child runtime to
// close in purecdk8s.
func Close() {}
