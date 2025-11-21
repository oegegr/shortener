package a

import (
	"log"
	"os"
)

// Ожидаем, что анализатор сообщит об использовании panic
func BadFunction() {
	panic("should not panic here") // want `usage of panic\(\) found`

	os.Exit(1) // want `log\.Fatal or os\.Exit called outside main function of main package`
}

// Ожидаем, что анализатор сообщит об log.Fatal вне main
func AnotherBadFunction() {
	log.Fatal("should not call log.Fatal here") // want `log\.Fatal or os\.Exit called outside main function of main package`
}

// Ожидаем, что анализатор не найдет проблем
func GoodFunction() error {
	return nil
}
