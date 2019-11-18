package flagsfiller

import (
	"errors"
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	durationType          = reflect.TypeOf(time.Duration(0))
	stringSliceType       = reflect.TypeOf([]string{})
	stringToStringMapType = reflect.TypeOf(map[string]string{})
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
// Fill returns an error when a non-struct reference is passed as 'from' or a field has a
// default tag which could not converted to the field's type.
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

	usage := requoteUsage(tag.Get("usage"))
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

	case t.Kind() == reflect.Float64:
		casted := fieldRef.(*float64)
		var defaultVal float64
		if hasDefaultTag {
			defaultVal, err = strconv.ParseFloat(tagDefault, 64)
			if err != nil {
				return errors.New("failed to parse default into bool")
			}
		} else {
			defaultVal = *casted
		}
		flagSet.Float64Var(casted, renamed, defaultVal, usage)

	// NOTE check time.Duration before int64 since it is aliased from int64
	case t == durationType:
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

	case t.Kind() == reflect.Int64:
		casted := fieldRef.(*int64)
		var defaultVal int64
		if hasDefaultTag {
			defaultVal, err = strconv.ParseInt(tagDefault, 10, 64)
			if err != nil {
				return errors.New("failed to parse default into bool")
			}
		} else {
			defaultVal = *casted
		}
		flagSet.Int64Var(casted, renamed, defaultVal, usage)

	case t.Kind() == reflect.Int:
		casted := fieldRef.(*int)
		var defaultVal int
		if hasDefaultTag {
			defaultVal, err = strconv.Atoi(tagDefault)
			if err != nil {
				return errors.New("failed to parse default into bool")
			}
		} else {
			defaultVal = *casted
		}
		flagSet.IntVar(casted, renamed, defaultVal, usage)

	case t.Kind() == reflect.Uint64:
		casted := fieldRef.(*uint64)
		var defaultVal uint64
		if hasDefaultTag {
			defaultVal, err = strconv.ParseUint(tagDefault, 10, 64)
			if err != nil {
				return errors.New("failed to parse default into bool")
			}
		} else {
			defaultVal = *casted
		}
		flagSet.Uint64Var(casted, renamed, defaultVal, usage)

	case t.Kind() == reflect.Uint:
		casted := fieldRef.(*uint)
		var defaultVal uint
		if hasDefaultTag {
			var asInt int
			asInt, err = strconv.Atoi(tagDefault)
			defaultVal = uint(asInt)
			if err != nil {
				return errors.New("failed to parse default into bool")
			}
		} else {
			defaultVal = *casted
		}
		flagSet.UintVar(casted, renamed, defaultVal, usage)

	case t == stringSliceType:
		casted := fieldRef.(*[]string)
		if hasDefaultTag {
			*casted = parseStringSlice(tagDefault)
		}
		flagSet.Var(&strSliceVar{ref: casted}, renamed, usage)

	case t == stringToStringMapType:
		casted := fieldRef.(*map[string]string)
		var val map[string]string
		if hasDefaultTag {
			val = parseStringToStringMap(tagDefault)
			*casted = val
		} else if *casted == nil {
			val = make(map[string]string)
			*casted = val
		} else {
			val = *casted
		}
		flagSet.Var(&strToStrMapVar{val: val}, renamed, usage)

	}
	return nil
}

type strSliceVar struct {
	ref *[]string
}

func (s *strSliceVar) String() string {
	if s.ref == nil {
		return ""
	}
	return strings.Join(*s.ref, ",")
}

func (s *strSliceVar) Set(val string) error {
	parts := parseStringSlice(val)
	*s.ref = append(*s.ref, parts...)

	return nil
}

func parseStringSlice(val string) []string {
	return strings.Split(val, ",")
}

type strToStrMapVar struct {
	val map[string]string
}

func (s strToStrMapVar) String() string {
	if s.val == nil {
		return ""
	}

	var sb strings.Builder
	first := true
	for k, v := range s.val {
		if !first {
			sb.WriteString(",")
		} else {
			first = false
		}
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(v)
	}
	return sb.String()
}

func (s strToStrMapVar) Set(val string) error {
	content := parseStringToStringMap(val)
	for k, v := range content {
		s.val[k] = v
	}
	return nil
}

func parseStringToStringMap(val string) map[string]string {
	result := make(map[string]string)

	pairs := strings.Split(val, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = kv[1]
		} else {
			result[kv[0]] = ""
		}
	}

	return result
}

// requoteUsage converts a [name] quoted usage string into the back quote form processed by flag.UnquoteUsage
func requoteUsage(usage string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '[':
			return '`'
		case ']':
			return '`'
		default:
			return r
		}
	}, usage)
}
