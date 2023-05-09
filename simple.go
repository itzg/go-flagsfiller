package flagsfiller

import (
	"flag"
	"fmt"
	"reflect"
)

// RegisterSimpleType register a new type,
// should be called in init(),
// see time.go and net.go for implementation examples
func RegisterSimpleType[T any](c ConvertFunc[T]) {
	base := simpleType[T]{converter: c}
	supportedStructList[getTypeName(reflect.TypeOf(*new(T)))] = base.Process
}

// ConvertFunc is a function convert string s into a specific type T, the tag is the struct field tag, as addtional input.
// see time.go and net.go for implementation examples
type ConvertFunc[T any] func(s string, tag reflect.StructTag) (T, error)

type simpleType[T any] struct {
	val       *T
	tags      reflect.StructTag
	converter ConvertFunc[T]
}

func newSimpleType[T any](c ConvertFunc[T], tag reflect.StructTag) simpleType[T] {
	return simpleType[T]{val: new(T), converter: c, tags: tag}
}

func (v *simpleType[T]) String() string {
	if v.val == nil {
		return fmt.Sprint(nil)
	}
	return fmt.Sprintf("%v", *v.val)
}

func (v *simpleType[T]) StrConverter(s string) (T, error) {
	return v.converter(s, v.tags)
}

func (v *simpleType[T]) Set(s string) error {
	var err error
	*v.val, err = v.converter(s, v.tags)
	if err != nil {
		return fmt.Errorf("failed to parse %s into %T, %w", s, *(new(T)), err)
	}
	return nil
}

func (v *simpleType[T]) SetRef(t *T) {
	v.val = t
}

type handlerFunc func(tag reflect.StructTag, fieldRef interface{},
	hasDefaultTag bool, tagDefault string,
	flagSet *flag.FlagSet, renamed string,
	usage string, aliases string) error

func (v *simpleType[T]) Process(tag reflect.StructTag, fieldRef interface{},
	hasDefaultTag bool, tagDefault string,
	flagSet *flag.FlagSet, renamed string,
	usage string, aliases string) error {
	val := newSimpleType(v.converter, tag)
	return processGeneral[T](fieldRef, &val, hasDefaultTag, tagDefault, flagSet, renamed, usage, aliases)
}
