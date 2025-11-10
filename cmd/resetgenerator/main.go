package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/oegegr/shortener/pkg/reset/generator"
)

func main() {

	var dir string
	flag.StringVar(&dir, "dir", ".", "root dir to find reset tags and then generate reset methods")
	flag.Parse()

	finder := reset.NewStructFinder(dir)
	packages, err := finder.Find()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	err = reset.NewResetGenerator(packages).GenerateReset()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
