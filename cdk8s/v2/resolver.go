package cdk8s

import (
	"encoding"
	"fmt"
	"reflect"
	"strings"
	"time"

	purecdk8sserialization "github.com/Chriscbr/purecdk8s/serialization"
)

type resolutionContextImpl struct {
	obj           ApiObject
	key           []*string
	value         interface{}
	replaced      bool
	replacedValue interface{}
}

func NewResolutionContext(obj ApiObject, key *[]*string, value interface{}) ResolutionContext {
	if obj == nil {
		panic("parameter obj is required, but nil was provided")
	}
	if key == nil {
		panic("parameter key is required, but nil was provided")
	}
	if value == nil {
		panic("parameter value is required, but nil was provided")
	}
	requireResolutionKey(*key)
	return newResolutionContext(obj, *key, value)
}

func NewResolutionContext_Override(context ResolutionContext, obj ApiObject, key *[]*string, value interface{}) {
	if context == nil {
		panic("parameter context is required, but nil was provided")
	}
	if obj == nil {
		panic("parameter obj is required, but nil was provided")
	}
	if key == nil {
		panic("parameter key is required, but nil was provided")
	}
	if value == nil {
		panic("parameter value is required, but nil was provided")
	}
	requireResolutionKey(*key)
	implementation := newResolutionContext(obj, *key, value)
	if target, ok := context.(*resolutionContextImpl); ok {
		*target = *implementation
		return
	}
	if !setEmbeddedImplementation(context, implementation) {
		panic("cdk8s: ResolutionContext override must embed cdk8s.ResolutionContext")
	}
}

func newResolutionContext(obj ApiObject, key []*string, value interface{}) *resolutionContextImpl {
	return &resolutionContextImpl{
		obj:   obj,
		key:   append([]*string(nil), key...),
		value: value,
	}
}

func (c *resolutionContextImpl) Key() *[]*string {
	result := append([]*string(nil), c.key...)
	return &result
}

func (c *resolutionContextImpl) Obj() ApiObject {
	return c.obj
}

func (c *resolutionContextImpl) Replaced() *bool {
	result := c.replaced
	return &result
}

func (c *resolutionContextImpl) SetReplaced(value *bool) {
	if value == nil {
		panic("parameter val is required, but nil was provided")
	}
	c.replaced = *value
}

func (c *resolutionContextImpl) ReplacedValue() interface{} {
	return c.replacedValue
}

func (c *resolutionContextImpl) SetReplacedValue(value interface{}) {
	if value == nil {
		panic("parameter val is required, but nil was provided")
	}
	c.replacedValue = value
}

func (c *resolutionContextImpl) Value() interface{} {
	return c.value
}

func (c *resolutionContextImpl) ReplaceValue(value interface{}) {
	if value == nil {
		panic("parameter newValue is required, but nil was provided")
	}
	c.replacedValue = value
	c.replaced = true
}

type lazyImpl struct {
	producer IAnyProducer
}

func Lazy_Any(producer IAnyProducer) interface{} {
	if producer == nil {
		panic("parameter producer is required, but nil was provided")
	}
	return &lazyImpl{producer: producer}
}

func (l *lazyImpl) Produce() interface{} {
	return l.producer.Produce()
}

type lazyResolverImpl struct{}

// undefinedLazyValue marks a lazy value that intentionally resolves to an
// omitted field. JavaScript cdk8s uses `undefined` for this; a nil value cannot
// be passed through ResolutionContext.ReplaceValue because that public API
// rejects nil arguments.
type undefinedLazyValue struct{}

func NewLazyResolver() LazyResolver {
	return &lazyResolverImpl{}
}

func NewLazyResolver_Override(resolver LazyResolver) {
	if resolver == nil {
		panic("parameter resolver is required, but nil was provided")
	}
	if _, native := resolver.(*lazyResolverImpl); native {
		return
	}
	if !setEmbeddedImplementation(resolver, &lazyResolverImpl{}) {
		panic("cdk8s: LazyResolver override must embed cdk8s.LazyResolver")
	}
}

func (r *lazyResolverImpl) Resolve(context ResolutionContext) {
	requireResolutionContext(context)
	if lazy, ok := context.Value().(*lazyImpl); ok {
		value := lazy.Produce()
		if value == nil {
			context.ReplaceValue(undefinedLazyValue{})
			return
		}
		context.ReplaceValue(value)
	}
}

type implicitTokenResolverImpl struct{}

func NewImplicitTokenResolver() ImplicitTokenResolver {
	return &implicitTokenResolverImpl{}
}

func NewImplicitTokenResolver_Override(resolver ImplicitTokenResolver) {
	if resolver == nil {
		panic("parameter resolver is required, but nil was provided")
	}
	if _, native := resolver.(*implicitTokenResolverImpl); native {
		return
	}
	if !setEmbeddedImplementation(resolver, &implicitTokenResolverImpl{}) {
		panic("cdk8s: ImplicitTokenResolver override must embed cdk8s.ImplicitTokenResolver")
	}
}

func (r *implicitTokenResolverImpl) Resolve(context ResolutionContext) {
	requireResolutionContext(context)
	value := context.Value()
	if value == nil {
		return
	}
	if token, ok := value.(interface{ Resolve() interface{} }); ok {
		context.ReplaceValue(token.Resolve())
		return
	}

	reflected := reflect.ValueOf(value)
	method := reflected.MethodByName("Resolve")
	if !method.IsValid() {
		return
	}
	typ := method.Type()
	if typ.NumIn() != 0 || typ.NumOut() != 1 {
		return
	}
	result := method.Call(nil)
	context.ReplaceValue(result[0].Interface())
}

type numberStringUnionResolverImpl struct{}

func NewNumberStringUnionResolver() NumberStringUnionResolver {
	return &numberStringUnionResolverImpl{}
}

func NewNumberStringUnionResolver_Override(resolver NumberStringUnionResolver) {
	if resolver == nil {
		panic("parameter resolver is required, but nil was provided")
	}
	if _, native := resolver.(*numberStringUnionResolverImpl); native {
		return
	}
	if !setEmbeddedImplementation(resolver, &numberStringUnionResolverImpl{}) {
		panic("cdk8s: NumberStringUnionResolver override must embed cdk8s.NumberStringUnionResolver")
	}
}

func (r *numberStringUnionResolverImpl) Resolve(context ResolutionContext) {
	requireResolutionContext(context)
	value := context.Value()
	if value == nil {
		return
	}
	// Plain dictionaries are deliberately ignored, as in the upstream
	// resolver. Number/string union wrappers expose Value() instead.
	reflected := reflect.ValueOf(value)
	if reflected.Kind() == reflect.Map {
		return
	}
	union, ok := value.(interface{ Value() interface{} })
	if !ok {
		return
	}
	scalar, ok := numberStringScalar(union.Value())
	if !ok {
		return
	}
	if _, native := value.(interface{ PureCDK8sScalarUnion() }); native {
		context.ReplaceValue(scalar)
		return
	}
	context.ReplaceValue(map[string]interface{}{"value": scalar})
}

func requireResolutionContext(context ResolutionContext) {
	if context == nil {
		panic("parameter context is required, but nil was provided")
	}
}

func requireResolutionKey(key []*string) {
	for _, component := range key {
		if component == nil {
			panic("parameter key cannot contain nil")
		}
	}
}

func numberStringScalar(value interface{}) (interface{}, bool) {
	if value == nil {
		return nil, false
	}
	reflected := reflect.ValueOf(value)
	for reflected.Kind() == reflect.Pointer || reflected.Kind() == reflect.Interface {
		if reflected.IsNil() {
			return nil, false
		}
		reflected = reflected.Elem()
	}
	switch reflected.Kind() {
	case reflect.String:
		return reflected.String(), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflected.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflected.Uint(), true
	case reflect.Float32, reflect.Float64:
		return reflected.Float(), true
	default:
		return nil, false
	}
}

// resolveValue resolves all values attached to obj using the owning App's
// resolver chain. It mirrors cdk8s' recursive resolution entry point.
func resolveValue(value interface{}, obj ApiObject) interface{} {
	return resolveValueAt(nil, value, obj)
}

func resolveValueAt(key []*string, value interface{}, obj ApiObject) interface{} {
	if value == nil {
		return nil
	}
	if obj == nil {
		panic("cdk8s: cannot resolve a value without an ApiObject")
	}

	resolverValue := resolverScalarValue(value)
	if resolverValue == nil {
		return nil
	}
	context := newResolutionContext(obj, key, resolverValue)
	resolvers := App_Of(obj).Resolvers()
	if resolvers != nil {
		for _, resolver := range *resolvers {
			if resolver == nil {
				continue
			}
			resolver.Resolve(context)
			if context.replaced {
				return resolveValueAt(key, context.replacedValue, obj)
			}
		}
	}
	return resolveNestedValue(key, resolverValue, obj)
}

func resolveNestedValue(key []*string, input interface{}, obj ApiObject) interface{} {
	if _, omitted := input.(undefinedLazyValue); omitted {
		return nil
	}
	// Value objects are classes rather than plain dictionaries upstream. Keep
	// them intact so sanitizeValue can reject unresolved instances instead of
	// silently turning their private Go representation into an empty object.
	if isCDK8sValueObject(input) {
		return input
	}
	value := reflect.ValueOf(input)
	for value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}
	if value.CanInterface() {
		if timestamp, ok := value.Interface().(time.Time); ok {
			return timestamp.Format(time.RFC3339)
		}
		if wireValue, ok := purecdk8sserialization.EnumWireValue(value.Interface()); ok {
			return resolveNestedValue(key, wireValue, obj)
		}
		if text, ok := value.Interface().(encoding.TextMarshaler); ok {
			encoded, err := text.MarshalText()
			if err != nil {
				panic(err)
			}
			return string(encoded)
		}
	}

	switch value.Kind() {
	case reflect.Bool:
		return value.Bool()
	case reflect.String:
		return value.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(value.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(value.Uint())
	case reflect.Float32, reflect.Float64:
		return value.Float()
	case reflect.Slice, reflect.Array:
		result := make([]interface{}, value.Len())
		for index := 0; index < value.Len(); index++ {
			result[index] = resolveValueAt(appendResolutionKey(key, strconvKey(index)), value.Index(index).Interface(), obj)
		}
		return result
	case reflect.Map:
		result := make(map[string]interface{}, value.Len())
		iterator := value.MapRange()
		for iterator.Next() {
			mapKey := fmt.Sprint(plainValue(iterator.Key().Interface()))
			item := iterator.Value()
			var raw interface{}
			if item.IsValid() {
				raw = item.Interface()
			}
			resolved := resolveValueAt(appendResolutionKey(key, mapKey), raw, obj)
			if resolved != nil {
				result[mapKey] = resolved
			}
		}
		return result
	case reflect.Struct:
		result := make(map[string]interface{})
		typ := value.Type()
		for index := 0; index < value.NumField(); index++ {
			field := typ.Field(index)
			if field.PkgPath != "" {
				continue
			}
			fieldValue := value.Field(index)
			if field.Tag.Get("field") == "optional" && fieldValue.IsZero() {
				continue
			}
			name := field.Tag.Get("k8s")
			if name == "" {
				name = strings.Split(field.Tag.Get("json"), ",")[0]
			}
			if name == "" {
				name = strings.Split(field.Tag.Get("yaml"), ",")[0]
			}
			if name == "" {
				name = lowerFirst(field.Name)
			}
			if name == "-" {
				continue
			}
			var raw interface{}
			if fieldValue.IsValid() && fieldValue.CanInterface() {
				raw = fieldValue.Interface()
			}
			resolved := resolveValueAt(appendResolutionKey(key, name), raw, obj)
			if resolved != nil {
				result[name] = resolved
			}
		}
		return result
	default:
		if value.CanInterface() {
			return value.Interface()
		}
		return input
	}
}

func isCDK8sValueObject(value interface{}) bool {
	switch value.(type) {
	case Cron, Duration, Size:
		return true
	default:
		return false
	}
}

// JSII represents primitive properties as pointers in Go, while resolvers in
// the source implementation receive the logical JavaScript primitive. Unwrap
// only primitive pointers; object pointers must retain their concrete identity
// so custom resolvers can recognize them.
func resolverScalarValue(input interface{}) interface{} {
	value := reflect.ValueOf(input)
	if !value.IsValid() || value.Kind() != reflect.Pointer {
		return input
	}
	for value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}
	switch value.Kind() {
	case reflect.Bool, reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return plainValue(input)
	default:
		return input
	}
}

func appendResolutionKey(key []*string, component string) []*string {
	result := append([]*string(nil), key...)
	value := component
	return append(result, &value)
}

func strconvKey(index int) string {
	return fmt.Sprintf("%d", index)
}
