# go-flagsfiller

[![](https://godoc.org/github.com/itzg/go-flagsfiller?status.svg)](https://godoc.org/github.com/itzg/go-flagsfiller)
[![](https://img.shields.io/badge/go.dev-module-007D9C)](https://pkg.go.dev/github.com/itzg/go-flagsfiller)

Bring your own struct and make Go's flag package pleasant to use.

## Install

```
go get github.com/itzg/go-flagsfiller
```

## Import

```go
import "github.com/itzg/go-flagsfiller"
```

## Features

- Populates Go's [flag.FlagSet](https://golang.org/pkg/flag/#FlagSet) from a struct of your choosing
- By default, field names are converted to flag names using [kebab-case](https://en.wiktionary.org/wiki/kebab_case), but can be configured.
- Use nested structs where flag name is prefixed by the nesting struct field names
- Allows defaults to be given via struct tag `default`
- Falls back to using instance field values as declared default
- Declare flag usage via struct tag `usage`
- Can be combined with other modules, such as [google/subcommands](https://github.com/google/subcommands) for sub-command processing. Can also be integrated with [spf13/cobra](https://github.com/spf13/cobra) by using pflag's [AddGoFlagSet](https://godoc.org/github.com/spf13/pflag#FlagSet.AddGoFlagSet)
- Beyond the standard types supported by flag.FlagSet also includes support for:
    - `[]string` where repetition of the argument appends to the slice and/or an argument value can contain a comma-separated list of values. For example: `--arg one --arg two,three`
    - `map[string]string` where each entry is a `key=value` and/or repetition of the arguments adds to the map or multiple entries can be comma-separated in a single argument value. For example: `--arg k1=v1 --arg k2=v2,k3=v3`
- Optionally set flag values from environment variables. Similar to flag names, environment variable names are derived automatically from the field names

## Quick example

```go
package main

import (
	"flag"
	"fmt"
	"github.com/itzg/go-flagsfiller"
	"log"
	"time"
)

type Config struct {
	Host         string        `default:"localhost" usage:"The remote host"`
	DebugEnabled bool          `default:"true" usage:"Show debugs"`
	MaxTimeout   time.Duration `default:"5s" usage:"How long to wait"`
	Feature      struct {
		Faster         bool `usage:"Go faster"`
		LudicrousSpeed bool `usage:"Go even faster"`
	}
}

func main() {
	var config Config
    
    // create a FlagSetFiller
	filler := flagsfiller.New()
    // fill and map struct fields to flags
	err := filler.Fill(flag.CommandLine, &config)
	if err != nil {
		log.Fatal(err)
	}

    // parse command-line like usual
	flag.Parse()

	fmt.Printf("Loaded: %+v\n", config)
}
```

The following shows an example of the usage provided when passing `--help`:
```
  -debug-enabled
    	Show debugs (default true)
  -feature-faster
    	Go faster
  -feature-ludicrous-speed
    	Go even faster
  -host string
    	The remote host (default "localhost")
  -max-timeout duration
    	How long to wait (default 5s)
```

## Real world example

[saml-auth-proxy](https://github.com/itzg/saml-auth-proxy) shows an end-to-end usage of flagsfiller where the main function fills the flags, maps those to environment variables with [envy](https://github.com/jamiealquiza/envy), and parses the command line:

```go
func main() {
	var serverConfig server.Config

	filler := flagsfiller.New()
	err := filler.Fill(flag.CommandLine, &serverConfig)
	if err != nil {
		log.Fatal(err)
	}

	envy.Parse("SAML_PROXY")
	flag.Parse()
```

where `server.Config` is declared as

```go
type Config struct {
	Version                 bool              `usage:"show version and exit"`
	Bind                    string            `default:":8080" usage:"host:port to bind for serving HTTP"`
	BaseUrl                 string            `usage:"External URL of this proxy"`
	BackendUrl              string            `usage:"URL of the backend being proxied"`
	IdpMetadataUrl          string            `usage:"URL of the IdP's metadata XML"`
	IdpCaPath               string            `usage:"Optional path to a CA certificate PEM file for the IdP"`
    // ...see https://github.com/itzg/saml-auth-proxy/blob/master/server/server.go for full set
}
```

## Using with google/subcommands

Flagsfiller can be used in combination with [google/subcommands](https://github.com/google/subcommands) to fill both global command-line flags and subcommand flags.

For the global flags, it is best to declare a struct type, such as

```go
type GlobalConfig struct {
	Debug bool `usage:"enable debug logging"`
}
```

Prior to calling `Execute` on the subcommands' `Commander`, fill and parse the global flags like normal:

```go
func main() {
    //... register subcommands here

	var globalConfig GlobalConfig

	err := flagsfiller.Parse(&globalConfig)
	if err != nil {
		log.Fatal(err)
	}

    //... execute subcommands but pass global config
	os.Exit(int(subcommands.Execute(context.Background(), &globalConfig)))
}
```

Each of your subcommand struct types should contain the flag fields to fill and parse, such as:

```go
type connectCmd struct {
	Host string `usage:"the hostname of the server" env:"GITHUB_TOKEN"`
	Port int `usage:"the port of the server" default:"8080"`
}
```

Your implementation of `SetFlags` will use flagsfiller to fill the definition of the subcommand's flagset, such as:

```go
func (c *connectCmd) SetFlags(f *flag.FlagSet) {
	filler := flagsfiller.New()
	err := filler.Fill(f, c)
	if err != nil {
		log.Fatal(err)
	}
}
```

Finally, your subcommand's `Execute` function can accept the global config passed from the main `Execute` call and access its own fields populated from the subcommand flags:

```go
func (c *loadFromGitCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	globalConfig := args[0].(*GlobalConfig)
    if globalConfig.Debug {
        //... enable debug logs
    }

    // ...operate on subcommand flags, such as
    conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port))
}
```
## More information

[Refer to the GoDocs](https://godoc.org/github.com/itzg/go-flagsfiller) for more information about this module.
