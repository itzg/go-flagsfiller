package flagsfiller_test

import (
	"bytes"
	"flag"
	"github.com/iancoleman/strcase"
	"github.com/itzg/go-flagsfiller"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestStringFields(t *testing.T) {
	type Config struct {
		Host          string
		MultiWordName string
	}

	var config Config

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	err = flagset.Parse([]string{"--host", "h1", "--multi-word-name", "val1"})
	require.NoError(t, err)

	assert.Equal(t, "h1", config.Host)
	assert.Equal(t, "val1", config.MultiWordName)
}

func TestUsage(t *testing.T) {
	type Config struct {
		MultiWordName string `usage:"usage goes here"`
	}

	var config Config

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	var buf bytes.Buffer
	buf.Write([]byte{'\n'}) // start with newline to make expected string nicer below
	flagset.SetOutput(&buf)
	flagset.PrintDefaults()

	assert.Equal(t, `
  -multi-word-name string
    	usage goes here
`, buf.String())
}

func TestRenamerOption(t *testing.T) {
	type Config struct {
		Host          string
		MultiWordName string
	}

	var config Config

	filler := flagsfiller.New(flagsfiller.WithFieldRenamer(strcase.ToSnake))

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	err = flagset.Parse([]string{"--host", "h1", "--multi_word_name", "val1"})
	require.NoError(t, err)

	assert.Equal(t, "h1", config.Host)
	assert.Equal(t, "val1", config.MultiWordName)
}

func TestNestedFields(t *testing.T) {
	type Config struct {
		Host         string
		SomeGrouping struct {
			SomeField string
		}
	}

	var config Config

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	err = flagset.Parse([]string{"--host", "h1", "--some-grouping-some-field", "val1"})
	require.NoError(t, err)

	assert.Equal(t, "h1", config.Host)
	assert.Equal(t, "val1", config.SomeGrouping.SomeField)
}

func TestNestedStructPtr(t *testing.T) {
	type Nested struct {
		SomeField string
	}
	type Config struct {
		Host         string
		SomeGrouping *Nested
	}

	var config Config

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	err = flagset.Parse([]string{"--host", "h1", "--some-grouping-some-field", "val1"})
	require.NoError(t, err)

	assert.Equal(t, "h1", config.Host)
	assert.Equal(t, "val1", config.SomeGrouping.SomeField)
}

func TestPtrField(t *testing.T) {
	type Config struct {
		// this should get ignored only inner struct pointers are supported
		Host *string
	}

	var config Config

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	var buf bytes.Buffer
	flagset.SetOutput(&buf)
	flagset.PrintDefaults()

	// not in usage
	assert.Equal(t, "", buf.String())
}

func TestDuration(t *testing.T) {
	type Config struct {
		Timeout time.Duration
	}

	var config Config

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	err = flagset.Parse([]string{"--timeout", "10s"})
	require.NoError(t, err)

	assert.Equal(t, 10*time.Second, config.Timeout)
}

func TestNumbers(t *testing.T) {
	type Config struct {
		ValFloat64 float64 `default:"3.14"`
		ValInt64   int64   `default:"43"`
		ValInt     int     `default:"44"`
		ValUint64  uint64  `default:"45"`
		ValUint    uint    `default:"46"`
	}

	var config Config

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	var buf bytes.Buffer
	buf.Write([]byte{'\n'}) // start with newline to make expected string nicer below
	flagset.SetOutput(&buf)
	flagset.PrintDefaults()

	assert.Equal(t, `
  -val-float-64 float
    	 (default 3.14)
  -val-int int
    	 (default 44)
  -val-int-64 int
    	 (default 43)
  -val-uint uint
    	 (default 46)
  -val-uint-64 uint
    	 (default 45)
`, buf.String())
}

func TestDefaultsViaLiteral(t *testing.T) {
	type Config struct {
		Host    string
		Enabled bool
		Timeout time.Duration
	}

	var config = Config{
		Host:    "h1",
		Enabled: true,
		Timeout: 5 * time.Second,
	}

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	var buf bytes.Buffer
	buf.Write([]byte{'\n'}) // start with newline to make expected string nicer below
	flagset.SetOutput(&buf)
	flagset.PrintDefaults()

	assert.Equal(t, `
  -enabled
    	 (default true)
  -host string
    	 (default "h1")
  -timeout duration
    	 (default 5s)
`, buf.String())
}

func TestDefaultsViaTag(t *testing.T) {
	type Config struct {
		Host    string        `default:"h1"`
		Enabled bool          `default:"true"`
		Timeout time.Duration `default:"5s"`
	}

	var config Config

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	var buf bytes.Buffer
	buf.Write([]byte{'\n'}) // start with newline to make expected string nicer below
	flagset.SetOutput(&buf)
	flagset.PrintDefaults()

	assert.Equal(t, `
  -enabled
    	 (default true)
  -host string
    	 (default "h1")
  -timeout duration
    	 (default 5s)
`, buf.String())
}

func TestBadDefaultsViaTag(t *testing.T) {
	type BadBoolConfig struct {
		Enabled bool `default:"wrong"`
	}
	type BadDurationConfig struct {
		Timeout time.Duration `default:"wrong"`
	}

	tests := []struct {
		Name   string
		Config interface{}
	}{
		{Name: "bool", Config: &BadBoolConfig{}},
		{Name: "duration", Config: &BadDurationConfig{}},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			filler := flagsfiller.New()

			var flagset flag.FlagSet
			err := filler.Fill(&flagset, tt.Config)
			require.Error(t, err)
		})
	}
}

func TestBadFieldErrorMessage(t *testing.T) {
	type BadBoolConfig struct {
		Enabled bool `default:"wrong"`
	}

	var config BadBoolConfig

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.Error(t, err)
	assert.Equal(t, "failed to process Enabled of flagsfiller_test.BadBoolConfig: failed to parse default into bool", err.Error())

}

func TestHiddenFields(t *testing.T) {
	type Config struct {
		hidden string
	}

	var config Config

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	var buf bytes.Buffer
	flagset.SetOutput(&buf)
	flagset.PrintDefaults()

	// not in usage
	assert.Equal(t, "", buf.String())
}

func TestStringSlice(t *testing.T) {
	type Config struct {
		NoDefault       []string
		InstanceDefault []string
		TagDefault      []string `default:"one,two"`
	}

	var config Config
	config.InstanceDefault = []string{"apple", "orange"}

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	var buf bytes.Buffer
	buf.Write([]byte{'\n'}) // start with newline to make expected string nicer below
	flagset.SetOutput(&buf)
	flagset.PrintDefaults()

	assert.Equal(t, `
  -instance-default value
    	 (default apple,orange)
  -no-default value
    	
  -tag-default value
    	 (default one,two)
`, buf.String())

	err = flagset.Parse([]string{"--no-default", "nd1", "--no-default", "nd2", "--no-default", "nd3,nd4"})
	require.NoError(t, err)

	assert.Equal(t, []string{"nd1", "nd2", "nd3", "nd4"}, config.NoDefault)
}

func TestStringToStringMap(t *testing.T) {
	type Config struct {
		NoDefault       map[string]string
		InstanceDefault map[string]string
		TagDefault      map[string]string `default:"fruit=apple,veggie=carrot"`
	}

	var config Config
	config.InstanceDefault = map[string]string{"fruit": "orange"}

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	var buf bytes.Buffer
	buf.Write([]byte{'\n'}) // start with newline to make expected string nicer below
	flagset.SetOutput(&buf)
	flagset.PrintDefaults()

	assert.Equal(t, `
  -instance-default value
    	 (default fruit=orange)
  -no-default value
    	
  -tag-default value
    	 (default fruit=apple,veggie=carrot)
`, buf.String())

	err = flagset.Parse([]string{"--no-default", "k1=v1", "--no-default", "k2=v2,k3=v3"})
	require.NoError(t, err)

	assert.Equal(t, map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"}, config.NoDefault)
	assert.Equal(t, map[string]string{"fruit": "apple", "veggie": "carrot"}, config.TagDefault)
}

func TestUsagePlaceholders(t *testing.T) {
	type Config struct {
		SomeUrl string `usage:"a [URL] to configure"`
	}

	var config Config

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	var buf bytes.Buffer
	buf.Write([]byte{'\n'}) // start with newline to make expected string nicer below
	flagset.SetOutput(&buf)
	flagset.PrintDefaults()

	assert.Equal(t, `
  -some-url URL
    	a URL to configure
`, buf.String())
}

func TestParse(t *testing.T) {
	type Config struct {
		Host string
	}

	var config Config
	os.Args = []string{"app", "--host", "host-a"}

	err := flagsfiller.Parse(&config)
	assert.NoError(t, err)

	require.Equal(t, "host-a", config.Host)
}

func TestParseError(t *testing.T) {
	type Config struct {
		BadDefault int `default:"not an int"`
	}

	var config Config
	os.Args = []string{"app", "--bad-default", "5"}

	err := flagsfiller.Parse(&config)
	assert.Error(t, err)
}
