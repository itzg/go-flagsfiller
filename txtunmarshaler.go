// This file implements support for all types that support interface encoding.TextUnmarshaler
package flagsfiller

import (
	"encoding"
	"flag"
	"fmt"
	"reflect"
	"strings"
)

// RegisterTextUnmarshaler use is optional, since flagsfiller will automatically register the types implement encoding.TextUnmarshaler it encounters
func RegisterTextUnmarshaler(in any) {
	base := textUnmarshalerType{}
	extendedTypes[getTypeName(reflect.TypeOf(in).Elem())] = base.process
}

type textUnmarshalerType struct {
	val encoding.TextUnmarshaler
}

// String implements flag.Value interface
func (tv *textUnmarshalerType) String() string {
	if tv.val == nil {
		return fmt.Sprint(nil)
	}
	return fmt.Sprint(tv.val)
}

// Set implements flag.Value interface
func (tv *textUnmarshalerType) Set(s string) error {
	return tv.val.UnmarshalText([]byte(s))
}

func (tv *textUnmarshalerType) process(tag reflect.StructTag, fieldRef interface{},
	hasDefaultTag bool, tagDefault string,
	flagSet *flag.FlagSet, renamed string,
	usage string, aliases string) error {
	v, ok := fieldRef.(encoding.TextUnmarshaler)
	if !ok {
		return fmt.Errorf("can't cast %v into encoding.TextUnmarshaler", fieldRef)
	}
	newval := textUnmarshalerType{
		val: v,
	}
	if hasDefaultTag {
		err := newval.Set(tagDefault)
		if err != nil {
			return fmt.Errorf("failed to parse default value into %v: %w", reflect.TypeOf(fieldRef), err)
		}
	}
	flagSet.Var(&newval, renamed, usage)
	if aliases != "" {
		for _, alias := range strings.Split(aliases, ",") {
			flagSet.Var(&newval, alias, usage)
		}
	}
	return nil

}
