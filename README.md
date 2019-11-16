# go-flagsfiller

[![](https://godoc.org/github.com/itzg/go-flagsfiller?status.svg)](http://godoc.org/github.com/itzg/go-flagsfiller)

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

- Populates Go's [flag.FagSet](https://golang.org/pkg/flag/#FlagSet) from a struct of your choosing
- By default, field names are converted to flag names using [kebab-case](https://en.wiktionary.org/wiki/kebab_case), but can be configured.
- Use nested structs where flag name is prefixed by the nesting struct field names
- Allows defaults to be given via struct tag `default`
- Falls back to using instance field values as declared default
- Declare flag usage via struct tag `usage`
- Easily combines with [jamiealquiza/envy](https://github.com/jamiealquiza/envy) for environment variable parsing and [google/subcommands](https://github.com/google/subcommands) for sub-command processing
- Beyond the standard types supported by flag.FlagSet also includes support for:
    - `[]string` where repetition of the argument appends to the slice and/or an argument value can contain a comma-separated list of values. For example: `--arg one --arg two,three`
    - `map[string]string` where each entry is a `key=value` and/or repetition of the arguments adds to the map or multiple entries can be comma-separated in a single argument value. For example: `--arg k1=v1 --arg k2=v2,k3=v3`

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

	filler := flagsfiller.New()
	err := filler.Fill(flag.CommandLine, &config)
	if err != nil {
		log.Fatal(err)
	}

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