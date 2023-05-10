package flagsfiller

import (
	"fmt"
	"net"
	"reflect"
)

func init() {
	RegisterSimpleType(ipConverter)
	RegisterSimpleType(ipnetConverter)
	RegisterSimpleType(macConverter)
}

func ipConverter(s string, tag reflect.StructTag) (net.IP, error) {
	addr := net.ParseIP(s)
	if addr == nil {
		return nil, fmt.Errorf("%s is not a valid IP address", s)
	}
	return addr, nil
}

func ipnetConverter(s string, tag reflect.StructTag) (net.IPNet, error) {
	_, prefix, err := net.ParseCIDR(s)
	if err != nil {
		return net.IPNet{}, err
	}
	return *prefix, nil
}

func macConverter(s string, tag reflect.StructTag) (net.HardwareAddr, error) {
	return net.ParseMAC(s)
}
