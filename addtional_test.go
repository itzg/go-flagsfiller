package flagsfiller_test

import (
	"flag"
	"log/slog"
	"net"
	"net/netip"
	"testing"
	"time"

	"github.com/itzg/go-flagsfiller"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTime(t *testing.T) {
	type Config struct {
		T time.Time `default:"2010-Oct-01==10:02:03" layout:"2006-Jan-02==15:04:05"`
	}

	var config Config

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	//test default tag
	err = flagset.Parse([]string{})
	require.NoError(t, err)
	expeted, _ := time.Parse("2006-Jan-02==15:04:05", "2010-Oct-01==10:02:03")
	assert.Equal(t, expeted, config.T)

	err = flagset.Parse([]string{"-t", "2016-Dec-13==16:03:02"})
	require.NoError(t, err)
	expeted, _ = time.Parse("2006-01-02 15:04:05", "2016-12-13 16:03:02")
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

func TestTextUnmarshalerType(t *testing.T) {
	type Config struct {
		Addr netip.Addr `default:"9.9.9.9"`
	}

	var config Config

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	//test default tag
	err = flagset.Parse([]string{})
	require.NoError(t, err)
	assert.Equal(t, netip.AddrFrom4([4]byte{9, 9, 9, 9}), config.Addr)

	err = flagset.Parse([]string{"-addr", "1.2.3.4"})
	require.NoError(t, err)

	assert.Equal(t, netip.AddrFrom4([4]byte{1, 2, 3, 4}), config.Addr)
}

func TestSlogLevels(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected slog.Level
	}{
		{
			name:     "info",
			value:    "info",
			expected: slog.LevelInfo,
		},
		{
			name:     "error",
			value:    "error",
			expected: slog.LevelError,
		},
		{
			name: "numeric offset",
			// Borrowed from https://pkg.go.dev/log/slog#Level.UnmarshalText
			value:    "Error-8",
			expected: slog.LevelInfo,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var args struct {
				Level slog.Level
			}

			var flagset flag.FlagSet
			err := flagsfiller.New().Fill(&flagset, &args)
			require.NoError(t, err)

			err = flagset.Parse([]string{"--level", test.value})
			require.NoError(t, err)

			assert.Equal(t, test.expected, args.Level)
		})
	}
}

func TestSlogLevelWithDefault(t *testing.T) {
	var args struct {
		Level slog.Level `default:"info"`
	}

	var flagset flag.FlagSet
	err := flagsfiller.New().Fill(&flagset, &args)
	require.NoError(t, err)

	err = flagset.Parse([]string{})
	require.NoError(t, err)

	assert.Equal(t, slog.LevelInfo, args.Level)
}
