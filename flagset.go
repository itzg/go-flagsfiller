package flagsfiller

import (
	"encoding"
	"flag"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	durationType          = reflect.TypeOf(time.Duration(0))
	stringSliceType       = reflect.TypeOf([]string{})
	stringToStringMapType = reflect.TypeOf(map[string]string{})
)

const (
	TagDefault       = "default"
	TagEnv           = "env"
	TagFlag          = "flag"
	TagFlatten       = "flatten"
	TagOverrideValue = "override-value"
	TagType          = "type"
)

// FlagSetFiller is used to map the fields of a struct into flags of a flag.FlagSet
type FlagSetFiller struct {
	options *fillerOptions
}

// Parse is a convenience function that creates a FlagSetFiller with the given options,
// fills and maps the flags from the given struct reference into flag.CommandLine, and uses
// flag.Parse to parse the os.Args.
// Returns an error if the given struct could not be used for filling flags.
func Parse(from interface{}, options ...FillerOption) error {
	filler := New(options...)
	err := filler.Fill(flag.CommandLine, from)
	if err != nil {
		return err
	}

	flag.Parse()
	return nil
}

// New creates a new FlagSetFiller with zero or more of the given FillerOption's
func New(options ...FillerOption) *FlagSetFiller {
	return &FlagSetFiller{options: newFillerOptions(options...)}
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

// isSupportedStruct checks if the given field reference is a registered extended type or implements
// encoding.TextUnmarshaler
func isSupportedStruct(in any) bool {
	t := reflect.TypeOf(in)
	_, ok := extendedTypes[getTypeName(t)]
	if ok {
		return true
	}
	if t.Kind() != reflect.Pointer {
		val := reflect.ValueOf(in)
		t = val.Addr().Type()
	}
	if t.Implements(reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()) {
		RegisterTextUnmarshaler(in)
		return true
	}
	return false
}

func getTypeName(t reflect.Type) string {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return fmt.Sprint(t)
}

func (f *FlagSetFiller) walkFields(flagSet *flag.FlagSet, prefix string,
	structVal reflect.Value, structType reflect.Type) error {

	if prefix != "" {
		prefix += "-"
	}
	handleDefault := func(field reflect.StructField, fieldValue reflect.Value) error {
		addr := fieldValue.Addr()
		// make sure it is exported/public
		ftype := field.Type
		if field.Type.Kind() == reflect.Ptr {
			ftype = field.Type.Elem()
		}
		if addr.CanInterface() {
			err := f.processField(flagSet, addr.Interface(), prefix+field.Name, ftype, field.Tag)
			if err != nil {
				return fmt.Errorf("failed to process %s of %s: %w", field.Name, structType.String(), err)
			}
		}
		return nil
	}
	for i := 0; i < structVal.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := structVal.Field(i)

		if flagTag, ok := field.Tag.Lookup(TagFlag); ok {
			if flagTag == "" {
				continue
			}
		}

		switch field.Type.Kind() {
		case reflect.Struct:
			if field.IsExported() {
				if isSupportedStruct(fieldValue.Addr().Interface()) {
					err := handleDefault(field, fieldValue)
					if err != nil {
						return err
					}
					continue
				}
			}

			err := f.walkFields(flagSet,
				qualifiedNameForStructField(field, prefix),
				fieldValue,
				field.Type)
			if err != nil {
				return fmt.Errorf("failed to process %s of %s: %w", field.Name, structType.String(), err)
			}

		case reflect.Ptr:
			if fieldValue.CanSet() && field.Type.Elem().Kind() == reflect.Struct {
				// fieldTypeName := getTypeName(field.Type.Elem())
				// fill the pointer with a new struct of their type if it is nil
				if fieldValue.IsNil() {
					fieldValue.Set(reflect.New(field.Type.Elem()))
				}
				if field.IsExported() {
					if isSupportedStruct(fieldValue.Interface()) {
						err := handleDefault(field, fieldValue.Elem())
						if err != nil {
							return err
						}
						continue
					}
				}

				err := f.walkFields(flagSet,
					qualifiedNameForStructField(field, prefix),
					fieldValue.Elem(),
					field.Type.Elem())
				if err != nil {
					return fmt.Errorf("failed to process %s of %s: %w", field.Name, structType.String(), err)
				}
			}

		default:
			err := handleDefault(field, fieldValue)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func qualifiedNameForStructField(field reflect.StructField, prefix string) string {
	if !shouldFlatten(field) {
		return prefix + field.Name
	} else {
		return prefix
	}
}

func shouldFlatten(field reflect.StructField) bool {
	value, ok := field.Tag.Lookup(TagFlatten)
	if !ok {
		return false
	}
	return value == "true"
}

func (f *FlagSetFiller) processField(flagSet *flag.FlagSet, fieldRef interface{},
	name string, t reflect.Type, tag reflect.StructTag) (err error) {

	var envName string
	if override, exists := tag.Lookup(TagEnv); exists {
		envName = override
	} else if len(f.options.envRenamer) > 0 {
		envName = name
		for _, renamer := range f.options.envRenamer {
			envName = renamer(envName)
		}
	}

	aliases := tag.Get("aliases")
	usage := requoteUsage(tag.Get("usage"))
	if envName != "" {
		usage = fmt.Sprintf("%s (env %s)", usage, envName)
	}

	tagDefault, hasDefaultTag := tag.Lookup(TagDefault)

	fieldType, _ := tag.Lookup(TagType)

	var renamed string
	if override, exists := tag.Lookup(TagFlag); exists {
		if override == "" {
			// empty flag override signal to skip this field
			return nil
		}
		renamed = override
	} else {
		renamed = f.options.renameLongName(name)
	}
	// go through all supported structs
	if isSupportedStruct(fieldRef) {
		handler := extendedTypes[getTypeName(t)]
		err = handler(tag, fieldRef, hasDefaultTag, tagDefault, flagSet, renamed, usage, aliases)
		return err
	}

	switch {
	case t.Kind() == reflect.String:
		f.processString(fieldRef, hasDefaultTag, tagDefault, flagSet, renamed, usage, aliases)

	case t.Kind() == reflect.Bool:
		err = f.processBool(fieldRef, hasDefaultTag, tagDefault, flagSet, renamed, usage, aliases)

	case t.Kind() == reflect.Float64:
		err = f.processFloat64(fieldRef, hasDefaultTag, tagDefault, flagSet, renamed, usage, aliases)

	// NOTE check time.Duration before int64 since it is aliasesed from int64
	case t == durationType, fieldType == "duration":
		err = f.processDuration(fieldRef, hasDefaultTag, tagDefault, flagSet, renamed, usage, aliases)

	case t.Kind() == reflect.Int64:
		err = f.processInt64(fieldRef, hasDefaultTag, tagDefault, flagSet, renamed, usage, aliases)

	case t.Kind() == reflect.Int:
		err = f.processInt(fieldRef, hasDefaultTag, tagDefault, flagSet, renamed, usage, aliases)

	case t.Kind() == reflect.Uint64:
		err = f.processUint64(fieldRef, hasDefaultTag, tagDefault, flagSet, renamed, usage, aliases)

	case t.Kind() == reflect.Uint:
		err = f.processUint(fieldRef, hasDefaultTag, tagDefault, flagSet, renamed, usage, aliases)

	case t == stringSliceType, fieldType == "stringSlice":
		var override bool
		if overrideValue, exists := tag.Lookup(TagOverrideValue); exists {
			if value, err := strconv.ParseBool(overrideValue); err == nil {
				override = value
			}
		}
		f.processStringSlice(fieldRef, hasDefaultTag, tagDefault, flagSet, renamed, usage, override, aliases)

	case t == stringToStringMapType, fieldType == "stringMap":
		f.processStringToStringMap(fieldRef, hasDefaultTag, tagDefault, flagSet, renamed, usage, aliases)

		// ignore any other types
	}

	if err != nil {
		return err
	}

	if !f.options.noSetFromEnv && envName != "" {
		if val, exists := os.LookupEnv(envName); exists {
			err := flagSet.Lookup(renamed).Value.Set(val)
			if err != nil {
				return fmt.Errorf("failed to set from environment variable %s: %w",
					envName, err)
			}
		}
	}

	return nil
}

func (f *FlagSetFiller) processStringToStringMap(fieldRef interface{}, hasDefaultTag bool, tagDefault string, flagSet *flag.FlagSet, renamed string, usage string, aliases string) {
	casted, ok := fieldRef.(*map[string]string)
	if !ok {
		_ = f.processCustom(
			fieldRef,
			func(s string) (interface{}, error) {
				return parseStringToStringMap(s), nil
			},
			hasDefaultTag,
			tagDefault,
			flagSet,
			renamed,
			usage,
			aliases,
		)
		return
	}
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
	if aliases != "" {
		for _, alias := range strings.Split(aliases, ",") {
			flagSet.Var(&strToStrMapVar{val: val}, alias, usage)
		}
	}
}

func (f *FlagSetFiller) processStringSlice(fieldRef interface{}, hasDefaultTag bool, tagDefault string, flagSet *flag.FlagSet, renamed string, usage string, override bool, aliases string) {
	casted, ok := fieldRef.(*[]string)
	if !ok {
		_ = f.processCustom(
			fieldRef,
			func(s string) (interface{}, error) {
				return parseStringSlice(s, f.options.valueSplitPattern), nil
			},
			hasDefaultTag,
			tagDefault,
			flagSet,
			renamed,
			usage,
			aliases,
		)
		return
	}
	if hasDefaultTag {
		*casted = parseStringSlice(tagDefault, f.options.valueSplitPattern)
	}
	flagSet.Var(&strSliceVar{
		ref:               casted,
		override:          override,
		valueSplitPattern: f.options.valueSplitPattern,
	}, renamed, usage)
	if aliases != "" {
		for _, alias := range strings.Split(aliases, ",") {
			flagSet.Var(&strSliceVar{
				ref:               casted,
				override:          override,
				valueSplitPattern: f.options.valueSplitPattern,
			}, alias, usage)
		}
	}
}

func (f *FlagSetFiller) processUint(fieldRef interface{}, hasDefaultTag bool, tagDefault string, flagSet *flag.FlagSet, renamed string, usage string, aliases string) (err error) {
	casted, ok := fieldRef.(*uint)
	if !ok {
		return f.processCustom(
			fieldRef,
			func(s string) (interface{}, error) {
				value, err := strconv.Atoi(s)
				return value, err
			},
			hasDefaultTag,
			tagDefault,
			flagSet,
			renamed,
			usage,
			aliases,
		)
	}
	var defaultVal uint
	if hasDefaultTag {
		var asInt int
		asInt, err = strconv.Atoi(tagDefault)
		defaultVal = uint(asInt)
		if err != nil {
			return fmt.Errorf("failed to parse default into uint: %w", err)
		}
	} else {
		defaultVal = *casted
	}
	flagSet.UintVar(casted, renamed, defaultVal, usage)
	if aliases != "" {
		for _, alias := range strings.Split(aliases, ",") {
			flagSet.UintVar(casted, alias, defaultVal, usage)
		}
	}
	return err
}

func (f *FlagSetFiller) processUint64(fieldRef interface{}, hasDefaultTag bool, tagDefault string, flagSet *flag.FlagSet, renamed string, usage string, aliases string) (err error) {
	casted, ok := fieldRef.(*uint64)
	if !ok {
		return f.processCustom(
			fieldRef,
			func(s string) (interface{}, error) {
				value, err := strconv.ParseUint(s, 10, 64)
				return value, err
			},
			hasDefaultTag,
			tagDefault,
			flagSet,
			renamed,
			usage,
			aliases,
		)
	}
	var defaultVal uint64
	if hasDefaultTag {
		defaultVal, err = strconv.ParseUint(tagDefault, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse default into uint64: %w", err)
		}
	} else {
		defaultVal = *casted
	}
	flagSet.Uint64Var(casted, renamed, defaultVal, usage)
	if aliases != "" {
		for _, alias := range strings.Split(aliases, ",") {
			flagSet.Uint64Var(casted, alias, defaultVal, usage)
		}
	}
	return err
}

func (f *FlagSetFiller) processInt(fieldRef interface{}, hasDefaultTag bool, tagDefault string, flagSet *flag.FlagSet, renamed string, usage string, aliases string) (err error) {
	casted, ok := fieldRef.(*int)
	if !ok {
		return f.processCustom(
			fieldRef,
			func(s string) (interface{}, error) {
				value, err := strconv.Atoi(s)
				return value, err
			},
			hasDefaultTag,
			tagDefault,
			flagSet,
			renamed,
			usage,
			aliases,
		)
	}
	var defaultVal int
	if hasDefaultTag {
		defaultVal, err = strconv.Atoi(tagDefault)
		if err != nil {
			return fmt.Errorf("failed to parse default into int: %w", err)
		}
	} else {
		defaultVal = *casted
	}
	flagSet.IntVar(casted, renamed, defaultVal, usage)
	if aliases != "" {
		for _, alias := range strings.Split(aliases, ",") {
			flagSet.IntVar(casted, alias, defaultVal, usage)
		}
	}
	return err
}

func (f *FlagSetFiller) processInt64(fieldRef interface{}, hasDefaultTag bool, tagDefault string, flagSet *flag.FlagSet, renamed string, usage string, aliases string) (err error) {
	casted, ok := fieldRef.(*int64)
	if !ok {
		return f.processCustom(
			fieldRef,
			func(s string) (interface{}, error) {
				value, err := strconv.ParseInt(s, 10, 64)
				return value, err
			},
			hasDefaultTag,
			tagDefault,
			flagSet,
			renamed,
			usage,
			aliases,
		)
	}
	var defaultVal int64
	if hasDefaultTag {
		defaultVal, err = strconv.ParseInt(tagDefault, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse default into int64: %w", err)
		}
	} else {
		defaultVal = *casted
	}
	flagSet.Int64Var(casted, renamed, defaultVal, usage)
	if aliases != "" {
		for _, alias := range strings.Split(aliases, ",") {
			flagSet.Int64Var(casted, alias, defaultVal, usage)
		}
	}
	return nil
}

func (f *FlagSetFiller) processDuration(fieldRef interface{}, hasDefaultTag bool, tagDefault string, flagSet *flag.FlagSet, renamed string, usage string, aliases string) (err error) {
	casted, ok := fieldRef.(*time.Duration)
	if !ok {
		return f.processCustom(
			fieldRef,
			func(s string) (interface{}, error) {
				value, err := time.ParseDuration(s)
				return value, err
			},
			hasDefaultTag,
			tagDefault,
			flagSet,
			renamed,
			usage,
			aliases,
		)
	}
	var defaultVal time.Duration
	if hasDefaultTag {
		defaultVal, err = time.ParseDuration(tagDefault)
		if err != nil {
			return fmt.Errorf("failed to parse default into time.Duration: %w", err)
		}
	} else {
		defaultVal = *casted
	}
	flagSet.DurationVar(casted, renamed, defaultVal, usage)
	if aliases != "" {
		for _, alias := range strings.Split(aliases, ",") {
			flagSet.DurationVar(casted, alias, defaultVal, usage)
		}
	}
	return nil
}

func (f *FlagSetFiller) processFloat64(fieldRef interface{}, hasDefaultTag bool, tagDefault string, flagSet *flag.FlagSet, renamed string, usage string, aliases string) (err error) {
	casted, ok := fieldRef.(*float64)
	if !ok {
		return f.processCustom(
			fieldRef,
			func(s string) (interface{}, error) {
				value, err := strconv.ParseFloat(s, 64)
				return value, err
			},
			hasDefaultTag,
			tagDefault,
			flagSet,
			renamed,
			usage,
			aliases,
		)
	}
	var defaultVal float64
	if hasDefaultTag {
		defaultVal, err = strconv.ParseFloat(tagDefault, 64)
		if err != nil {
			return fmt.Errorf("failed to parse default into float64: %w", err)
		}
	} else {
		defaultVal = *casted
	}
	flagSet.Float64Var(casted, renamed, defaultVal, usage)
	if aliases != "" {
		for _, alias := range strings.Split(aliases, ",") {
			flagSet.Float64Var(casted, alias, defaultVal, usage)
		}
	}
	return nil
}

func (f *FlagSetFiller) processBool(fieldRef interface{}, hasDefaultTag bool, tagDefault string, flagSet *flag.FlagSet, renamed string, usage string, aliases string) (err error) {
	casted, ok := fieldRef.(*bool)
	if !ok {
		return f.processCustom(
			fieldRef,
			func(s string) (interface{}, error) {
				value, err := strconv.ParseBool(s)
				return value, err
			},
			hasDefaultTag,
			tagDefault,
			flagSet,
			renamed,
			usage,
			aliases,
		)
	}
	var defaultVal bool
	if hasDefaultTag {
		defaultVal, err = strconv.ParseBool(tagDefault)
		if err != nil {
			return fmt.Errorf("failed to parse default into bool: %w", err)
		}
	} else {
		defaultVal = *casted
	}
	flagSet.BoolVar(casted, renamed, defaultVal, usage)
	if aliases != "" {
		for _, alias := range strings.Split(aliases, ",") {
			flagSet.BoolVar(casted, alias, defaultVal, usage)
		}
	}
	return nil
}

func (f *FlagSetFiller) processString(fieldRef interface{}, hasDefaultTag bool, tagDefault string, flagSet *flag.FlagSet, renamed string, usage string, aliases string) {
	casted, ok := fieldRef.(*string)
	if !ok {
		_ = f.processCustom(
			fieldRef,
			func(s string) (interface{}, error) {
				return s, nil
			},
			hasDefaultTag,
			tagDefault,
			flagSet,
			renamed,
			usage,
			aliases,
		)
		return
	}
	var defaultVal string
	if hasDefaultTag {
		defaultVal = tagDefault
	} else {
		defaultVal = *casted
	}
	flagSet.StringVar(casted, renamed, defaultVal, usage)
	if aliases != "" {
		for _, alias := range strings.Split(aliases, ",") {
			flagSet.StringVar(casted, alias, defaultVal, usage)
		}
	}
}

func (f *FlagSetFiller) processCustom(fieldRef interface{}, converter func(string) (interface{}, error), hasDefaultTag bool, tagDefault string, flagSet *flag.FlagSet, renamed string, usage string, aliases string) error {
	if hasDefaultTag {
		value, err := converter(tagDefault)
		if err != nil {
			return fmt.Errorf("failed to parse default into custom type: %w", err)
		}
		reflect.ValueOf(fieldRef).Elem().Set(reflect.ValueOf(value).Convert(reflect.TypeOf(fieldRef).Elem()))
	}
	flagSet.Func(renamed, usage, func(s string) error {
		value, err := converter(s)
		if err != nil {
			return err
		}
		reflect.ValueOf(fieldRef).Elem().Set(reflect.ValueOf(value).Convert(reflect.TypeOf(fieldRef).Elem()))
		return nil
	})
	if aliases != "" {
		for _, alias := range strings.Split(aliases, ",") {
			flagSet.Func(alias, usage, func(s string) error {
				value, err := converter(s)
				if err != nil {
					return err
				}
				reflect.ValueOf(fieldRef).Elem().Set(reflect.ValueOf(value).Convert(reflect.TypeOf(fieldRef).Elem()))
				return nil
			})
		}
	}
	return nil
}

type strSliceVar struct {
	ref               *[]string
	override          bool
	valueSplitPattern string
}

func (s *strSliceVar) String() string {
	if s.ref == nil {
		return ""
	}
	return strings.Join(*s.ref, ",")
}

func (s *strSliceVar) Set(val string) error {
	parts := parseStringSlice(val, s.valueSplitPattern)

	if s.override {
		*s.ref = parts
		return nil
	}

	*s.ref = append(*s.ref, parts...)

	return nil
}

func (s *strSliceVar) Type() string {
	return getTypeName(reflect.TypeOf(s.ref))
}

func parseStringSlice(val string, valueSplitPattern string) []string {
	if valueSplitPattern == "" {
		return []string{val}
	}

	splitter := regexp.MustCompile(valueSplitPattern)
	parts := splitter.Split(val, -1)

	// trim out blank parts
	result := make([]string, 0, len(parts))
	for _, s := range parts {
		s = strings.TrimSpace(s)
		if s != "" {
			result = append(result, s)
		}
	}
	return result
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

func (s strToStrMapVar) Type() string {
	return getTypeName(reflect.TypeOf(s.val))
}

func parseStringToStringMap(val string) map[string]string {
	result := make(map[string]string)

	splitter := regexp.MustCompile("[\n,]")

	pairs := splitter.Split(val, -1)
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)

		if pair != "" {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				result[kv[0]] = kv[1]
			} else {
				result[kv[0]] = ""
			}
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
