package flagsfiller

import (
	"errors"
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// FlagSetFiller is used to map the fields of a struct into flags of a flag.FlagSet
type FlagSetFiller struct {
	options *fillerOptions
}

// New creates a new FlagSetFiller with zero or more of the given FillerOption's
func New(options ...FillerOption) *FlagSetFiller {
	return &FlagSetFiller{options: newFillerOptions(options)}
}

// Fill populates the flagSet with a flag for each field in given struct passed in the 'from'
// argument which must be a struct reference.
func (f *FlagSetFiller) Fill(flagSet *flag.FlagSet, from interface{}) error {
	v := reflect.ValueOf(from)
	t := v.Type()
	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		return f.walkFields(flagSet, "", v.Elem(), t.Elem())
	} else {
		return fmt.Errorf("can only fill from struct pointer, but it was %s", t.Kind())
	}
}

func (f *FlagSetFiller) walkFields(flagSet *flag.FlagSet, prefix string,
	structVal reflect.Value, structType reflect.Type) error {

	for i := 0; i < structVal.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := structVal.Field(i)

		switch field.Type.Kind() {
		case reflect.Struct:
			err := f.walkFields(flagSet, field.Name, fieldValue, field.Type)
			if err != nil {
				return fmt.Errorf("failed to process %s of %s: %w", field.Name, structType.String(), err)
			}

		case reflect.Ptr:
			if field.Type.Elem().Kind() == reflect.Struct {
				// fill the pointer with a new struct of their type
				fieldValue.Set(reflect.New(field.Type.Elem()))

				err := f.walkFields(flagSet, field.Name, fieldValue.Elem(), field.Type.Elem())
				if err != nil {
					return fmt.Errorf("failed to process %s of %s: %w", field.Name, structType.String(), err)
				}
			}

		default:
			addr := fieldValue.Addr()
			// make sure it is exported/public
			if addr.CanInterface() {
				err := f.processField(flagSet, addr.Interface(), prefix+field.Name, field.Type, field.Tag)
				if err != nil {
					return fmt.Errorf("failed to process %s of %s: %w", field.Name, structType.String(), err)
				}
			}
		}
	}

	return nil
}

func (f *FlagSetFiller) processField(flagSet *flag.FlagSet, fieldRef interface{},
	name string, t reflect.Type, tag reflect.StructTag) error {

	usage := tag.Get("usage")
	tagDefault, hasDefaultTag := tag.Lookup("default")
	renamed := f.options.renameLongName(name)
	var err error

	switch {
	case t.Kind() == reflect.String:
		casted := fieldRef.(*string)
		var defaultVal string
		if hasDefaultTag {
			defaultVal = tagDefault
		} else {
			defaultVal = *casted
		}
		flagSet.StringVar(casted, renamed, defaultVal, usage)

	case t.Kind() == reflect.Bool:
		casted := fieldRef.(*bool)
		var defaultVal bool
		if hasDefaultTag {
			defaultVal, err = strconv.ParseBool(tagDefault)
			if err != nil {
				return errors.New("failed to parse default into bool")
			}
		} else {
			defaultVal = *casted
		}
		flagSet.BoolVar(casted, renamed, defaultVal, usage)

	case t == reflect.TypeOf(time.Duration(0)):
		casted := fieldRef.(*time.Duration)
		var defaultVal time.Duration
		if hasDefaultTag {
			defaultVal, err = time.ParseDuration(tagDefault)
			if err != nil {
				return errors.New("failed to parse default into time.Duration")
			}
		} else {
			defaultVal = *casted
		}
		flagSet.DurationVar(casted, renamed, defaultVal, usage)

	}
	return nil
}
