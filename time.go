package flagsfiller

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

// layout string to parse time, following golang time.Parse() format, change this var to the layout needed
var TimeLayout = "2006-01-02_15:04:05"

type timeValue struct {
	t *time.Time
}

func (v *timeValue) String() string {
	if v.t == nil {
		return fmt.Sprint(nil)
	}
	return v.t.String()
}

func (v *timeValue) Set(s string) error {
	var err error
	*v.t, err = time.Parse(TimeLayout, s)
	return err
}

func (f *FlagSetFiller) processTime(fieldRef interface{}, hasDefaultTag bool, tagDefault string, flagSet *flag.FlagSet, renamed string, usage string, aliases string) (err error) {
	casted, ok := fieldRef.(*time.Time)
	if !ok {
		return f.processCustom(
			fieldRef,
			func(s string) (interface{}, error) {
				return time.Parse(TimeLayout, s)
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
		*casted, err = time.Parse(TimeLayout, tagDefault)
		if err != nil {
			return fmt.Errorf("failed to parse default into MAC(net.HardwareAddr): %w", err)
		}
	}
	flagSet.Var(&timeValue{casted}, renamed, usage)
	if aliases != "" {
		for _, alias := range strings.Split(aliases, ",") {
			flagSet.Var(&timeValue{casted}, alias, usage)
		}
	}
	return nil
}

func init() {
	supportedStructList["time.Time"] = struct{}{}
}
