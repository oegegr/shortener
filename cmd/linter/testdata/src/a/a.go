package a

import (
	"log"
	"os"
)

func BadFunction() {
	panic("should not panic here") // want `usage of panic\(\) found`

	os.Exit(1) // want `log\.Fatal or os\.Exit called outside main function of main package`
}

func AnotherBadFunction() {
	log.Fatal("should not call log.Fatal here") // want `log\.Fatal or os\.Exit called outside main function of main package`
}

func GoodFunction() error {
	return nil
}
