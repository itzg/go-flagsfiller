package flagsfiller

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

func init() {
	supportedStructList["time.Time"] = struct{}{}
}

// DefaultTimeLayout is the default layout string to parse time, following golang time.Parse() format,
// can be overridden per field by field tag "layout". Default value is "2006-01-02 15:04:05", which is
// the same as time.DateTime in Go 1.20
var DefaultTimeLayout = "2006-01-02 15:04:05"

type timeValue struct {
	t      *time.Time
	layout string
}

func (v *timeValue) String() string {
	if v.t == nil {
		return fmt.Sprint(nil)
	}
	return v.t.String()
}

func (v *timeValue) Set(s string) error {
	var err error
	*v.t, err = time.Parse(v.layout, s)
	if err != nil {
		return fmt.Errorf("failed to parse %s into time.Time, expect layout is %s", s, v.layout)
	}
	return nil
}

func (f *FlagSetFiller) processTime(fieldRef interface{},
	hasDefaultTag bool, tagDefault string,
	flagSet *flag.FlagSet, renamed string,
	usage string, aliases string, layout string) (err error) {
	if layout == "" {
		layout = DefaultTimeLayout
	}
	casted, ok := fieldRef.(*time.Time)
	if !ok {
		return f.processCustom(
			fieldRef,
			func(s string) (interface{}, error) {
				return time.Parse(layout, s)
			},
			hasDefaultTag,
			tagDefault,
			flagSet,
			renamed,
			usage,
			aliases,
		)
	}

	if hasDefaultTag {
		*casted, err = time.Parse(layout, tagDefault)
		if err != nil {
			return fmt.Errorf("failed to parse default into MAC(net.HardwareAddr): %w", err)
		}
	}
	val := &timeValue{t: casted, layout: layout}
	flagSet.Var(val, renamed, usage)
	if aliases != "" {
		for _, alias := range strings.Split(aliases, ",") {
			flagSet.Var(val, alias, usage)
		}
	}
	return nil
}
