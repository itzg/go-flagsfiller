package flagsfiller_test

import (
	"flag"
	"fmt"
	"github.com/itzg/go-flagsfiller"
	"log"
	"time"
)

func Example() {
	type Config struct {
		Host      string        `default:"localhost" usage:"The remote host"`
		Enabled   bool          `default:"true" usage:"Turn it on"`
		Automatic bool          `default:"false" usage:"Make it automatic" aliases:"a"`
		Retries   int          	`default:"1" usage:"Retry" aliases:"r,t"`
		Timeout   time.Duration `default:"5s" usage:"How long to wait"`
	}

	var config Config

	flagset := flag.NewFlagSet("ExampleBasic", flag.ExitOnError)

	filler := flagsfiller.New()
	err := filler.Fill(flagset, &config)
	if err != nil {
		log.Fatal(err)
	}

	err = flagset.Parse([]string{"--host", "external.svc", "--timeout", "10m", "-a", "-t", "2"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", config)
	// Output:
	// {Host:external.svc Enabled:true Automatic:true Retries:2 Timeout:10m0s}
}
