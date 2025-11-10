package main

import (
	"log"
	"os"
)

func main() {
	os.Exit(0)
}

func badHelper() {
	os.Exit(1) // want `log\.Fatal or os\.Exit called outside main function of main package`

	panic("helper panic") // want `usage of panic\(\) found`
}

func anotherHelper() {
	log.Fatal("fatal in helper") // want `log\.Fatal or os\.Exit called outside main function of main package`

}
