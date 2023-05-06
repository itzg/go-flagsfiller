package flagsfiller

import (
	"flag"
	"fmt"
	"net"
	"strings"
)

type macValue struct {
	mac *net.HardwareAddr
}

func (v *macValue) String() string {
	if v.mac == nil {
		return fmt.Sprint(nil)
	}
	return v.mac.String()
}

func (v *macValue) Set(s string) error {
	var err error
	*v.mac, err = net.ParseMAC(s)
	return err
}

func (f *FlagSetFiller) processMAC(fieldRef interface{}, hasDefaultTag bool, tagDefault string, flagSet *flag.FlagSet, renamed string, usage string, aliases string) (err error) {
	casted, ok := fieldRef.(*net.HardwareAddr)
	if !ok {
		return f.processCustom(
			fieldRef,
			func(s string) (interface{}, error) {
				return net.ParseMAC(s)
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
		*casted, err = net.ParseMAC(tagDefault)
		if err != nil {
			return fmt.Errorf("failed to parse default into MAC(net.HardwareAddr): %w", err)
		}
	}
	flagSet.Var(&macValue{casted}, renamed, usage)
	if aliases != "" {
		for _, alias := range strings.Split(aliases, ",") {
			flagSet.Var(&macValue{casted}, alias, usage)
		}
	}
	return nil
}
