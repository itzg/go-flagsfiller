package flagsfiller_test

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/itzg/go-flagsfiller"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestCustomFields(t *testing.T) {
	type CustomStringType string
	type CustomBoolType bool
	type CustomFloat64 float64
	type CustomDuration time.Duration
	type CustomInt64 int64
	type CustomInt int
	type CustomUint64 uint64
	type CustomUint uint
	type CustomStringSlice []string
	type CustomStringMap map[string]string

	t.Run("Default values", func(t *testing.T) {
		type Config struct {
			String      CustomStringType  `default:"stringValue"`
			Bool        CustomBoolType    `default:"true"`
			Float64     CustomFloat64     `default:"1.234"`
			Duration    CustomDuration    `type:"duration" default:"2s"`
			Int64       CustomInt64       `default:"-1"`
			Int         CustomInt         `default:"-2"`
			Uint64      CustomUint64      `default:"1"`
			Uint        CustomUint        `default:"2"`
			StringSlice CustomStringSlice `type:"stringSlice" default:"one,two"`
			StringMap   CustomStringMap   `type:"stringMap" default:"one=value1,two=value2"`
		}

		var config Config

		filler := flagsfiller.New()

		var flagset flag.FlagSet
		err := filler.Fill(&flagset, &config)
		require.NoError(t, err)

		err = flagset.Parse([]string{})
		require.NoError(t, err)

		assert.Equal(t, CustomStringType("stringValue"), config.String)
		assert.Equal(t, CustomBoolType(true), config.Bool)
		assert.Equal(t, CustomFloat64(1.234), config.Float64)
		assert.Equal(t, CustomDuration(2*time.Second), config.Duration)
		assert.Equal(t, CustomInt64(-1), config.Int64)
		assert.Equal(t, CustomInt(-2), config.Int)
		assert.Equal(t, CustomUint64(1), config.Uint64)
		assert.Equal(t, CustomUint(2), config.Uint)
		assert.Equal(t, CustomStringSlice{"one", "two"}, config.StringSlice)
		assert.Equal(t, CustomStringMap{"one": "value1", "two": "value2"}, config.StringMap)
	})

	t.Run("Values set from arguments", func(t *testing.T) {
		type Config struct {
			String      CustomStringType
			Bool        CustomBoolType
			Float64     CustomFloat64
			Duration    CustomDuration `type:"duration"`
			Int64       CustomInt64
			Int         CustomInt
			Uint64      CustomUint64
			Uint        CustomUint
			StringSlice CustomStringSlice `type:"stringSlice"`
			StringMap   CustomStringMap   `type:"stringMap"`
		}

		var config Config

		filler := flagsfiller.New()

		var flagset flag.FlagSet
		err := filler.Fill(&flagset, &config)
		require.NoError(t, err)

		err = flagset.Parse([]string{
			"--string", "stringValue",
			"--bool", "true",
			"--float-64", "1.234",
			"--duration", "2s",
			"--int-64", "-1",
			"--int", "-2",
			"--uint-64", "1",
			"--uint", "2",
			"--string-slice", "one,two",
			"--string-map", "one=value1,two=value2",
		})
		require.NoError(t, err)

		assert.Equal(t, CustomStringType("stringValue"), config.String)
		assert.Equal(t, CustomBoolType(true), config.Bool)
		assert.Equal(t, CustomFloat64(1.234), config.Float64)
		assert.Equal(t, CustomDuration(2*time.Second), config.Duration)
		assert.Equal(t, CustomInt64(-1), config.Int64)
		assert.Equal(t, CustomInt(-2), config.Int)
		assert.Equal(t, CustomUint64(1), config.Uint64)
		assert.Equal(t, CustomUint(2), config.Uint)
		assert.Equal(t, CustomStringSlice{"one", "two"}, config.StringSlice)
		assert.Equal(t, CustomStringMap{"one": "value1", "two": "value2"}, config.StringMap)
	})
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

	buf := grabUsage(flagset)

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
		ALLCAPS struct {
			ALLCAPS string
		}
	}

	var config Config

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	err = flagset.Parse([]string{"--host", "h1", "--some-grouping-some-field", "val1", "--allcaps-allcaps", "val2"})
	require.NoError(t, err)

	assert.Equal(t, "h1", config.Host)
	assert.Equal(t, "val1", config.SomeGrouping.SomeField)
	assert.Equal(t, "val2", config.ALLCAPS.ALLCAPS)
}

func TestNestedAdjacentFields(t *testing.T) {
	type SomeGrouping struct {
		SomeField  string
		EvenDeeper struct {
			Deepest string
		}
	}
	type Config struct {
		Host         string
		SomeGrouping SomeGrouping
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

	var buf bytes.Buffer
	flagset.SetOutput(&buf)
	flagset.PrintDefaults()

	assert.Equal(t, `  -host string
    	
  -some-grouping-even-deeper-deepest string
    	
  -some-grouping-some-field string
    	
`, buf.String())
}

func TestNestedUnexportedFields(t *testing.T) {
	type Config struct {
		Host        string
		hiddenField struct {
			SomeField    string
			anotherField string
		}
	}

	var config Config

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	var buf bytes.Buffer
	flagset.SetOutput(&buf)
	flagset.PrintDefaults()

	assert.Equal(t, `  -host string
    	
`, buf.String())
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

func TestNestedUnexportedStructPtr(t *testing.T) {
	type Nested struct {
		SomeField string
	}
	type Config struct {
		Host        string
		hiddenField *Nested
	}

	var config Config

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	var buf bytes.Buffer
	flagset.SetOutput(&buf)
	flagset.PrintDefaults()

	assert.Equal(t, `  -host string
    	
`, buf.String())
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

	buf := grabUsage(flagset)

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
	type Nested struct {
		Exported   string
		unExported string
	}
	type Config struct {
		Host    string
		Enabled bool
		Timeout time.Duration
		Nested  *Nested
	}

	var config = Config{
		Host:    "h1",
		Enabled: true,
		Timeout: 5 * time.Second,
		Nested: &Nested{
			Exported:   "exported",
			unExported: "un-exported",
		},
	}

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	buf := grabUsage(flagset)

	assert.Equal(t, "un-exported", config.Nested.unExported)

	assert.Equal(t, `
  -enabled
    	 (default true)
  -host string
    	 (default "h1")
  -nested-exported string
    	 (default "exported")
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

	buf := grabUsage(flagset)

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
	assert.Equal(t, "failed to process Enabled of flagsfiller_test.BadBoolConfig: failed to parse default into bool: strconv.ParseBool: parsing \"wrong\": invalid syntax", err.Error())

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
		TagOverride     []string `default:"one,two" override-value:"true"`
	}

	var config Config
	config.InstanceDefault = []string{"apple", "orange"}

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	buf := grabUsage(flagset)

	assert.Equal(t, `
  -instance-default value
    	 (default apple,orange)
  -no-default value
    	
  -tag-default value
    	 (default one,two)
  -tag-override value
    	 (default one,two)
`, buf.String())

	err = flagset.Parse([]string{
		"--no-default", "nd1",
		"--no-default", "nd2",
		"--no-default", "nd3,nd4",
		"--no-default", "nd5\nnd6",
		"--tag-default", "three",
		"--tag-override", "three",
	})
	require.NoError(t, err)

	assert.Equal(t, []string{"nd1", "nd2", "nd3", "nd4", "nd5", "nd6"}, config.NoDefault)
	assert.Equal(t, []string{"apple", "orange"}, config.InstanceDefault)
	assert.Equal(t, []string{"one", "two", "three"}, config.TagDefault)
	assert.Equal(t, []string{"three"}, config.TagOverride)
}

func TestStringSliceWithEmptyValuePattern(t *testing.T) {
	type Config struct {
		NoDefault  []string
		TagDefault []string `default:"one,two"`
	}

	var config Config
	filler := flagsfiller.New(flagsfiller.WithValueSplitPattern(""))

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	err = flagset.Parse([]string{
		"--no-default", "nd1,nd2",
		"--no-default", "nd3",
	})
	require.NoError(t, err)

	assert.Equal(t, []string{"nd1,nd2", "nd3"}, config.NoDefault)
	assert.Equal(t, []string{"one,two"}, config.TagDefault)
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

	buf := grabUsage(flagset)

	// using regexp assertion since -tag-default's map entries can be either order
	assert.Regexp(t, `
  -instance-default value
    	 \(default fruit=orange\)
  -no-default value
    	
  -tag-default value
    	 \(default (veggie=carrot,fruit=apple|fruit=apple,veggie=carrot)\)
`, buf.String())

	err = flagset.Parse([]string{"--no-default",
		"k1=v1",
		"--no-default",
		"k2=v2,k3=v3\nk4=v4\n",
	})
	require.NoError(t, err)

	assert.Equal(t, map[string]string{"k1": "v1", "k2": "v2", "k3": "v3", "k4": "v4"}, config.NoDefault)
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

	buf := grabUsage(flagset)

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

func TestIgnoreNonExportedFields(t *testing.T) {
	type Config struct {
		Host        string
		hiddenField string
	}

	var config Config
	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	buf := grabUsage(flagset)

	assert.Equal(t, `
  -host string
    	
`, buf.String())
}

func TestIgnoreNonExportedStructFields(t *testing.T) {
	type Config struct {
		Host   string
		nested struct {
			NotVisible string
		}
	}

	var config Config
	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	buf := grabUsage(flagset)

	assert.Equal(t, `
  -host string
    	
`, buf.String())
}

func TestWithEnv(t *testing.T) {
	type Config struct {
		Host          string `default:"localhost" usage:"the host to use"`
		MultiWordName string
	}

	var config Config

	assert.NoError(t, os.Setenv("APP_HOST", "host from env"))
	assert.NoError(t, os.Setenv("APP_MULTI_WORD_NAME", "value from env"))

	filler := flagsfiller.New(flagsfiller.WithEnv("App"))

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	buf := grabUsage(flagset)

	assert.Equal(t, `
  -host string
    	the host to use (env APP_HOST) (default "localhost")
  -multi-word-name string
    	 (env APP_MULTI_WORD_NAME)
`, buf.String())

	err = flagset.Parse([]string{"--host", "host from args"})
	require.NoError(t, err)

	assert.Equal(t, "host from args", config.Host)
	assert.Equal(t, "value from env", config.MultiWordName)
}

func TestWithEnvOverride(t *testing.T) {
	type Config struct {
		Host string `env:"SERVER_ADDRESS"`
	}

	var config Config

	filler := flagsfiller.New(flagsfiller.WithEnv("App"))

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	buf := grabUsage(flagset)

	assert.Equal(t, `
  -host string
    	 (env SERVER_ADDRESS)
`, buf.String())
}

func TestWithEnvOverrideDisable(t *testing.T) {
	type Config struct {
		Host string `env:"" usage:"arg only"`
	}

	var config Config

	filler := flagsfiller.New(flagsfiller.WithEnv("App"))

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	buf := grabUsage(flagset)

	assert.Equal(t, `
  -host string
    	arg only
`, buf.String())
}

func TestNoSetFromEnv(t *testing.T) {
	type Config struct {
		Host string `usage:"arg only"`
	}

	var config Config

	assert.NoError(t, os.Setenv("APP_HOST", "host from env"))

	filler := flagsfiller.New(
		flagsfiller.WithEnv("App"),
		flagsfiller.NoSetFromEnv(),
	)

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)

	buf := grabUsage(flagset)

	assert.Empty(t, config.Host)

	assert.Equal(t, `
  -host string
    	arg only (env APP_HOST)
`, buf.String())
}

func TestFlagNameOverride(t *testing.T) {
	type Config struct {
		Host        string `flag:"server_address" usage:"address of server"`
		GetsIgnored string `flag:""`
	}

	var config Config

	filler := flagsfiller.New()

	var flagset flag.FlagSet
	err := filler.Fill(&flagset, &config)
	require.NoError(t, err)
	buf := grabUsage(flagset)

	assert.Equal(t, `
  -server_address string
    	address of server
`, buf.String())

}

func grabUsage(flagset flag.FlagSet) *bytes.Buffer {
	var buf bytes.Buffer
	buf.Write([]byte{'\n'})
	// start with newline to make expected string nicer below
	flagset.SetOutput(&buf)
	flagset.PrintDefaults()
	return &buf
}

func ExampleWithEnv() {
	type Config struct {
		MultiWordName string
	}

	// simulate env variables from program invocation
	_ = os.Setenv("MY_APP_MULTI_WORD_NAME", "from env")

	var config Config

	// enable environment variable processing with given prefix
	filler := flagsfiller.New(flagsfiller.WithEnv("MyApp"))
	var flagset flag.FlagSet
	_ = filler.Fill(&flagset, &config)

	// simulate no args passed in
	_ = flagset.Parse([]string{})

	fmt.Println(config.MultiWordName)
	// Output:
	// from env
}
