package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tsukinoko-kun/kamehameha/khh/config"
)

var (
	create = flag.Bool("create", false, "create a new config file")
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, `usage: khh <file> [flags...]`)
		flag.PrintDefaults()

		os.Exit(1)
	}

	if len(os.Args) >= 3 {
		if err := flag.CommandLine.Parse(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "failed to parse flags: %v\n", err)
			os.Exit(1)
		}
	}

	p := os.Args[1]
	if *create {
		if err := config.New(p); err != nil {
			fmt.Fprintf(os.Stderr, "failed to create config: %v\n", err)
			os.Exit(1)
		} else {
			fmt.Printf("created config file %s\n", p)
		}
		return
	}

	c, err := config.Load(p)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%+v\n", c)
}
