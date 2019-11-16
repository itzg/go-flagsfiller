/*
Package flagsfiller makes Go's flag package pleasant to use by mapping the fields of a given struct
into flags in a FlagSet.

Quick Start

A FlagSetFiller is created with the New constructor passing it any desired FillerOption's.
With that call Fill passing it a flag.FlatSet, such as flag.CommandLine, and your struct to
be mapped.

Even a simple struct with no additional tags can be used, such as:

	type Config struct {
		Host string
		Enabled bool
	}
	var config Config

	filler := flagsfiller.New()
	filler.Fill(flag.CommandLine, &config)

Flag Naming

By default, the flags are named by taking the field name and performing a word-wise conversion
to kebab-case. For example the field named "MyMultiWordField" becomes the flag named
"my-multi-word-field".

The naming strategy can be changed by passing a custom Renamer using WithFieldRenamer in the
constructor.

Nested Structs

flagsfiller supports nested struct's and resolves the flag names by prefixing the field
name of the struct to the names of the fields it contains. For example, the following fills
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

Usage

To declare a flag usage's add a `usage:""` tag to the field, such as:

	type Config struct {
		Host string `usage:"the name of the host to access"`
	}

Defaults

To declare the default value of a flag, you can either set a field's value before passing the
struct to Fill, such as:

	type Config struct {
		Host string
	}
	var config = Config{Host:"localhost"}
	filler.Fill(flag.CommandLine, &config)

or add a `default:""` tag to the field; however, be sure to provide a valid string that can be
converted into the field's type. For example,

	type Config struct {
		Host 	string `default:"localhost"`
		Timeout time.Duration `default:"1m"`
	}

*/
package flagsfiller
