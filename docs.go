/*
Package flagsfiller makes Go's flag package pleasant to use by mapping the fields of a given struct
into flags in a FlagSet.

Quick Start

A FlagSetFiller is created with the New constructor, passing it any desired FillerOptions.
With that, call Fill, passing it a flag.FlatSet, such as flag.CommandLine, and your struct to
be mapped.

Even a simple struct with no special changes can be used, such as:

	type Config struct {
		Host string
		Enabled bool
	}
	var config Config

	// create a FlagSetFiller
	filler := flagsfiller.New()
	// fill and map struct fields to flags
	filler.Fill(flag.CommandLine, &config)
	// parse command-line like usual
	flag.Parse()

After calling Parse on the flag.FlagSet, the corresponding fields of the mapped struct will
be populated with values passed from the command-line.

For an even quicker start, flagsfiller provides a convenience Parse function that does the same
as the snippet above in one call:

	type Config struct {
		Host string
		Enabled bool
	}
	var config Config

	flagsfiller.Parse(&config)

Flag Naming

By default, the flags are named by taking the field name and performing a word-wise conversion
to kebab-case. For example the field named "MyMultiWordField" becomes the flag named
"my-multi-word-field".

The naming strategy can be changed by passing a custom Renamer using the WithFieldRenamer
option in the constructor.

Nested Structs

FlagSetFiller supports nested structs and computes the flag names by prefixing the field
name of the struct to the names of the fields it contains. For example, the following maps to
the flags named remote-host, remote-auth-username, and remote-auth-password:

	type Config struct {
		Remote struct {
			Host string
			Auth struct {
				Username string
				Password string
			}
		}
	}

Flag Usage

To declare a flag's usage add a `usage:""` tag to the field, such as:

	type Config struct {
		Host string `usage:"the name of the host to access"`
	}

Since flag.UnquoteUsage normally uses back quotes to locate the argument placeholder name but
struct tags also use back quotes, flagsfiller will instead use [square brackets] to define the
placeholder name, such as:

	SomeUrl      string `usage:"a [URL] to configure"`

results in the rendered output:

	-some-url URL
		a URL to configure

Defaults

To declare the default value of a flag, you can either set a field's value before passing the
struct to process, such as:

	type Config struct {
		Host string
	}
	var config = Config{Host:"localhost"}

or add a `default:""` tag to the field. Be sure to provide a valid string that can be
converted into the field's type. For example,

	type Config struct {
		Host 	string `default:"localhost"`
		Timeout time.Duration `default:"1m"`
	}

String Slices

FlagSetFiller also includes support for []string fields.
Repetition of the argument appends to the slice and/or an argument value can contain a
comma-separated list of values.

For example:

	--arg one --arg two,three

results in a three element slice.

The default tag's value is provided as a comma-separated list, such as

	MultiValues []string `default:"one,two,three"`

Maps of String to String

FlagSetFiller also includes support for map[string]string fields.
Each argument entry is a key=value and/or repetition of the arguments adds to the map or
multiple entries can be comma-separated in a single argument value.

For example:

	--arg k1=v1 --arg k2=v2,k3=v3

results in a map with three entries.

The default tag's value is provided a comma-separate list of key=value entries, such as

	Mappings map[string]string `default:"k1=v1,k2=v2,k3=v3"`

Environment variable mapping

To activate the setting of flag values from environment variables, pass the WithEnv option to
flagsfiller.New or flagsfiller.Parse. That option takes a prefix that will be prepended to the
resolved field name and then the whole thing is converted to SCREAMING_SNAKE_CASE.

The environment variable name will be automatically included in the flag usage along with the
standard inclusion of the default value. For example, using the option WithEnv("App") along
with the following field declaration

	Host string `default:"localhost" usage:"the host to use"`

would render the following usage:

  -host string
    	the host to use (env APP_HOST) (default "localhost")

Per-field overrides

To override the naming of a flag, the field can be declared with the tag `flag:"name"` where
the given name will be used exactly as the flag name. An empty string for the name indicates
the field should be ignored and no flag is declared. For example,

	Host        string `flag:"server_address"
	GetsIgnored string `flag:""`

Environment variable naming and processing can be overridden with the `env:"name"` tag, where
the given name will be used exactly as the mapped environment variable name. If the WithEnv
or WithEnvRenamer options were enabled, a field can be excluded from environment variable
mapping by giving an empty string. Conversely, environment variable mapping can be enabled
per field with `env:"name"` even when the flagsfiller-wide option was not included. For example,

	Host 			string `env:"SERVER_ADDRESS"`
	NotEnvMapped 	string `env:""`

*/
package flagsfiller
