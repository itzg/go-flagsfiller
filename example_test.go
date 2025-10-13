package flagsfiller_test

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/itzg/go-flagsfiller"
)

func Example() {
	type Config struct {
		Host      string        `default:"localhost" usage:"The remote host"`
		Enabled   bool          `default:"true" usage:"Turn it on"`
		Automatic bool          `default:"false" usage:"Make it automatic" aliases:"a"`
		Retries   int           `default:"1" usage:"Retry" aliases:"r,t"`
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

func ExampleFlagSetFiller_Verify() {
	type Config struct {
		Host     string `required:"true" usage:"The remote host"`
		Port     int    `default:"8080" usage:"The port"`
		Username string `required:"true" usage:"Username for authentication"`
	}

	var config Config

	flagset := flag.NewFlagSet("ExampleVerify", flag.ContinueOnError)

	filler := flagsfiller.New()
	err := filler.Fill(flagset, &config)
	if err != nil {
		log.Fatal(err)
	}

	// Parse with required fields provided
	err = flagset.Parse([]string{"--host", "example.com", "--username", "admin"})
	if err != nil {
		log.Fatal(err)
	}

	// Verify all required fields are set
	err = filler.Verify()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Config validated: Host=%s, Port=%d, Username=%s\n", config.Host, config.Port, config.Username)
	// Output:
	// Config validated: Host=example.com, Port=8080, Username=admin
}
