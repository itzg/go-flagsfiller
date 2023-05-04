package flagsfiller

import (
	"flag"
	"fmt"
	"net"
	"strings"
)

type ipValue struct {
	addr *net.IP
}

func (v *ipValue) String() string {
	if v.addr == nil {
		return fmt.Sprint(nil)
	}
	return v.addr.String()
}

func (v *ipValue) Set(s string) error {
	*v.addr = net.ParseIP(s)
	if *v.addr == nil {
		return fmt.Errorf("invalid ip addr %v", s)
	}
	return nil
}

func (f *FlagSetFiller) processIP(fieldRef interface{}, hasDefaultTag bool, tagDefault string, flagSet *flag.FlagSet, renamed string, usage string, aliases string) (err error) {
	casted, ok := fieldRef.(*net.IP)
	if !ok {
		return f.processCustom(
			fieldRef,
			func(s string) (interface{}, error) {
				value := net.ParseIP(s)
				if value == nil {
					return nil, fmt.Errorf("invalid IP address %s", s)
				}
				return value, nil
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
		*casted = net.ParseIP(tagDefault)
		if *casted == nil {
			return fmt.Errorf("failed to parse default into net.IP: %s", tagDefault)
		}
	}
	flagSet.Var(&ipValue{casted}, renamed, usage)
	if aliases != "" {
		for _, alias := range strings.Split(aliases, ",") {
			flagSet.Var(&ipValue{casted}, alias, usage)
		}
	}
	return nil
}

type ipnetValue struct {
	prefix *net.IPNet
}

func (v *ipnetValue) String() string {
	if v.prefix == nil {
		return fmt.Sprint(nil)
	}
	return v.prefix.String()
}

func (v *ipnetValue) Set(s string) error {
	_, pr, err := net.ParseCIDR(s)
	if err != nil {
		return fmt.Errorf("invalid ip prefix %v", s)
	}
	*v.prefix = *pr
	return nil
}

func (f *FlagSetFiller) processIPNet(fieldRef interface{}, hasDefaultTag bool, tagDefault string, flagSet *flag.FlagSet, renamed string, usage string, aliases string) (err error) {
	casted, ok := fieldRef.(*net.IPNet)
	if !ok {
		return f.processCustom(
			fieldRef,
			func(s string) (interface{}, error) {
				_, value, err := net.ParseCIDR(s)
				if err != nil {
					return nil, fmt.Errorf("invalid IP prefix %s, %w", s, err)
				}
				return *value, nil
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
		_, casted, err = net.ParseCIDR(tagDefault)
		if err != nil {
			return fmt.Errorf("failed to parse default into net.IPNet: %s, %w", tagDefault, err)
		}
	}
	flagSet.Var(&ipnetValue{casted}, renamed, usage)
	if aliases != "" {
		for _, alias := range strings.Split(aliases, ",") {
			flagSet.Var(&ipnetValue{casted}, alias, usage)
		}
	}
	return nil
}

func init() {
	supportedStructList["net.IPNet"] = struct{}{}
}
