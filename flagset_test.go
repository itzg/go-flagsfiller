package flagsfiller_test

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/itzg/go-flagsfiller"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
	"time"
)

func TestStringFields(t *testing.T) {
	type Config struct {
		Host          string
		MultiWordName string
	}

	var config Config

	filler := flagsfiller.NewFlagSetFiller()

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

	filler := flagsfiller.NewFlagSetFiller()

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

	filler := flagsfiller.NewFlagSetFiller(flagsfiller.WithFieldRenamer(strcase.ToSnake))

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

	filler := flagsfiller.NewFlagSetFiller()

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

	filler := flagsfiller.NewFlagSetFiller()

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

	filler := flagsfiller.NewFlagSetFiller()

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

	filler := flagsfiller.NewFlagSetFiller()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	err = flagset.Parse([]string{"--timeout", "10s"})
	require.NoError(t, err)

	assert.Equal(t, 10*time.Second, config.Timeout)
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

	filler := flagsfiller.NewFlagSetFiller()

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

	filler := flagsfiller.NewFlagSetFiller()

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
			filler := flagsfiller.NewFlagSetFiller()

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

	filler := flagsfiller.NewFlagSetFiller()

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

	filler := flagsfiller.NewFlagSetFiller()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	var buf bytes.Buffer
	flagset.SetOutput(&buf)
	flagset.PrintDefaults()

	// not in usage
	assert.Equal(t, "", buf.String())
}

func ExampleBasic() {
	type Config struct {
		Host    string        `default:"localhost" usage:"The remote host"`
		Enabled bool          `default:"true" usage:"Turn it on"`
		Timeout time.Duration `default:"5s" usage:"How long to wait"`
	}

	var config Config

	flagset := flag.NewFlagSet("ExampleBasic", flag.ExitOnError)

	filler := flagsfiller.NewFlagSetFiller()
	err := filler.Fill(flagset, &config)
	if err != nil {
		log.Fatal(err)
	}

	err = flagset.Parse([]string{"--host", "external.svc", "--timeout", "10m"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", config)
	// Output:
	// {Host:external.svc Enabled:true Timeout:10m0s}
}
