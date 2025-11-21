package linter

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestPanicAnalyzer запускает тесты для анализатора паники
func TestPanicAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, PanicAnalyzer, "a", "b")
}
