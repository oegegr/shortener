package main

import (
	"log"
	"os"
)

// Ожидаем, что тут нет проблем
func main() {
	os.Exit(0)
}

// Ожидаем, что анализатор сообщит об использовании panic
func badHelper() {
	os.Exit(1) // want `log\.Fatal or os\.Exit called outside main function of main package`

	panic("helper panic") // want `usage of panic\(\) found`
}

// Ожидаем, что анализатор сообщит об использовании panic
func anotherHelper() {
	log.Fatal("fatal in helper") // want `log\.Fatal or os\.Exit called outside main function of main package`

}
