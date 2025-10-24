package service

import (
	"fmt"
	"testing"
)

func BenchmarkRandomShortCodeProvider_Get(b *testing.B) {
	provider := &RandomShortCodeProvider{}
	lengths := []int{5, 10, 20, 50}

	for _, length := range lengths {
		b.Run(fmt.Sprintf("length=%d", length), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				provider.Get(length)
			}
		})
	}
}