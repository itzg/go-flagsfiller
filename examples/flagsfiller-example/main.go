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
