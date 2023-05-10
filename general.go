package flagsfiller

/*
The code in this file could be opened up in future if more complex implementation is needed
*/

import (
	"flag"
	"fmt"
	"reflect"
	"strings"
)

// this is a list of addtional supported types(include struct), like time.Time, that walkFields() won't walk into,
// the key is the is string returned by the getTypeName(<type>),
// each supported type need to be added in this map in init()
var extendedTypes = make(map[string]handlerFunc)

type handlerFunc func(tag reflect.StructTag, fieldRef interface{},
	hasDefaultTag bool, tagDefault string,
	flagSet *flag.FlagSet, renamed string,
	usage string, aliases string) error

type flagVal[T any] interface {
	flag.Value
	StrConverter(string) (T, error)
	SetRef(*T)
}

func processGeneral[T any](fieldRef interface{}, val flagVal[T],
	hasDefaultTag bool, tagDefault string,
	flagSet *flag.FlagSet, renamed string,
	usage string, aliases string) (err error) {

	casted := fieldRef.(*T)
	if hasDefaultTag {
		*casted, err = val.StrConverter(tagDefault)
		if err != nil {
			return fmt.Errorf("failed to parse default into %T: %w", *new(T), err)
		}
	}
	val.SetRef(casted)
	flagSet.Var(val, renamed, usage)
	if aliases != "" {
		for _, alias := range strings.Split(aliases, ",") {
			flagSet.Var(val, alias, usage)
		}
	}
	return nil

}
