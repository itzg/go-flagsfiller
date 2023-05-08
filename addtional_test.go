package flagsfiller_test

import (
	"flag"
	"net"
	"testing"
	"time"

	"github.com/itzg/go-flagsfiller"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTime(t *testing.T) {
	type Config struct {
		T time.Time `layout:"2006-Jan-02==15:04:05"`
	}

	var config Config

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	err = flagset.Parse([]string{"-t", "2016-Dec-13==16:03:02"})
	require.NoError(t, err)
	expeted, _ := time.Parse("2006-01-02 15:04:05", "2016-12-13 16:03:02")
	assert.Equal(t, expeted, config.T)
}

func TestNetIP(t *testing.T) {
	type Config struct {
		Addr net.IP
	}

	var config Config

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	err = flagset.Parse([]string{"-addr", "1.2.3.4"})
	require.NoError(t, err)

	assert.Equal(t, net.ParseIP("1.2.3.4"), config.Addr)
}

func TestMACAddr(t *testing.T) {
	type Config struct {
		Addr net.HardwareAddr
	}

	var config Config

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	err = flagset.Parse([]string{"-addr", "1c:2a:11:ce:23:45"})
	require.NoError(t, err)

	assert.Equal(t, net.HardwareAddr{0x1c, 0x2a, 0x11, 0xce, 0x23, 0x45}, config.Addr)
}

func TestIPNet(t *testing.T) {
	type Config struct {
		Prefix net.IPNet
	}

	var config Config

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	err = flagset.Parse([]string{"-prefix", "192.168.1.0/24"})
	require.NoError(t, err)
	_, expected, _ := net.ParseCIDR("192.168.1.0/24")
	assert.Equal(t, *expected, config.Prefix)
}
