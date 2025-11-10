package linter

import (
	"go/ast"
	_ "go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/singlechecker"
)

var PanicAnalyzer = &analysis.Analyzer{
	Name:     "golinter",
	Doc:      "Reports usage of panic, log.Fatal and os.Exit outside main function",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      check,
}

func main() {
	singlechecker.Main(PanicAnalyzer)
}

func check(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// Пропускаем тестовые файлы
		if strings.HasSuffix(pass.Fset.File(file.Pos()).Name(), "_test.go") {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				inMainFunc := node.Name.Name == "main"
				ast.Inspect(node, func(n ast.Node) bool {
					switch call := n.(type) {
					case *ast.CallExpr:
						// Проверяем вызов panic
						if isPanicCall(call) {
							if !inMainFunc {
								pass.Reportf(call.Pos(), "usage of panic() found")
							}
						}

						// Проверяем вызов log.Fatal или os.Exit вне функции main
						if isForbiddenExitCall(call) {
							if !inMainFunc {
								pass.Reportf(call.Pos(), "log.Fatal or os.Exit called outside main function of main package")
							}
						}
					}
					return true
				})
			}
			return true
		})
	}
	return nil, nil
}

// isPanicCall проверяет, является ли вызов функцией panic
func isPanicCall(call *ast.CallExpr) bool {
	ident, ok := call.Fun.(*ast.Ident)
	return ok && ident.Name == "panic"
}

// isForbiddenExitCall проверяет, является ли вызов log.Fatal или os.Exit
func isForbiddenExitCall(call *ast.CallExpr) bool {
	switch fun := call.Fun.(type) {
	case *ast.SelectorExpr:
		// Проверяем os.Exit
		if ident, ok := fun.X.(*ast.Ident); ok {
			if ident.Name == "os" && fun.Sel.Name == "Exit" {
				return true
			}
		}

		// Проверяем log.Fatal
		if ident, ok := fun.X.(*ast.Ident); ok {
			if ident.Name == "log" && strings.HasPrefix(fun.Sel.Name, "Fatal") {
				return true
			}
		}

	case *ast.Ident:
		// Проверяем прямые вызовы Exit (если пакет импортирован с другим именем)
		if fun.Name == "Exit" {
			return true
		}
	}
	return false
}
