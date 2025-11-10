package reset

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"unicode"
)

type StructFinder struct {
	rootDir string
}

func NewStructFinder(rootDir string) *StructFinder {
	return &StructFinder{rootDir: rootDir}
}

func (f *StructFinder) Find() ([]*packageInfo, error) {

	// Собираем информацию о всех пакетах
	mapPackages := make(map[string]*packageInfo)
	packages := make([]*packageInfo, 0)

	err := filepath.WalkDir(f.rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			return nil
		}

		// Пропускаем скрытые директории и vendor
		if strings.Contains(path, "/.") || strings.Contains(path, "vendor") {
			return filepath.SkipDir
		}

		// Парсим пакет
		fset := token.NewFileSet()
		pkgs, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
		if err != nil {
			// Если не удалось распарсить (нет Go файлов), пропускаем
			return nil
		}

		for pkgName, pkg := range pkgs {
			info := &packageInfo{
				name:    pkgName,
				path:    path,
				structs: make([]*structInfo, 0),
				imports: make(map[string]string),
				files:   make(map[string]*ast.File),
			}

			// Анализируем каждый файл в пакете
			for filename, file := range pkg.Files {
				info.files[filename] = file
				analyzeFile(file, info)
			}

			if len(info.structs) > 0 {
				mapPackages[path] = info
			}
		}

		return nil
	})

	if err != nil {
		return packages, err
	}

	// Генерируем файлы reset.gen.go для каждого пакета со структурами
	for _, pkgInfo := range mapPackages {
		packages = append(packages, pkgInfo)
		// if err := generateResetFile(pkgInfo); err != nil {
		// 	return packages, fmt.Errorf("generating reset file for %s: %w", pkgInfo.path, err)
		// }
	}

	return packages, nil
}

func analyzeFile(file *ast.File, pkgInfo *packageInfo) {
	// Собираем импорты
	for _, imp := range file.Imports {
		if imp.Path != nil {
			path := strings.Trim(imp.Path.Value, `"`)
			var name string
			if imp.Name != nil {
				name = imp.Name.Name
			} else {
				// Извлекаем имя пакета из пути
				parts := strings.Split(path, "/")
				name = parts[len(parts)-1]
			}
			pkgInfo.imports[path] = name
		}
	}

	// Ищем структуры с комментарием // generate:reset
	ast.Inspect(file, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			return true
		}

		// Проверяем комментарии
		if genDecl.Doc == nil {
			return true
		}

		hasResetComment := false
		for _, comment := range genDecl.Doc.List {
			if strings.Contains(comment.Text, "generate:reset") {
				hasResetComment = true
				break
			}
		}

		if !hasResetComment {
			return true
		}

		// Обрабатываем каждую спецификацию типа
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			structInfo := &structInfo{
				name:   typeSpec.Name.Name,
				fields: make([]*fieldInfo, 0),
			}

			// Анализируем поля структуры
			for _, field := range structType.Fields.List {
				if len(field.Names) == 0 {
					continue // Пропускаем анонимные поля
				}

				fieldName := field.Names[0].Name
				if !unicode.IsUpper(rune(fieldName[0])) {
					continue // Пропускаем приватные поля
				}

				fieldInfo := analyzeFieldType(field.Type)
				fieldInfo.name = fieldName
				structInfo.fields = append(structInfo.fields, fieldInfo)
			}

			pkgInfo.structs = append(pkgInfo.structs, structInfo)
		}

		return true
	})
}

func analyzeFieldType(expr ast.Expr) *fieldInfo {
	info := &fieldInfo{}

	switch t := expr.(type) {
	case *ast.Ident:
		info.typeExpr = t.Name
		info.isStruct = isBuiltinStructType(t.Name)

	case *ast.StarExpr:
		info.isPtr = true
		innerInfo := analyzeFieldType(t.X)
		info.typeExpr = "*" + innerInfo.typeExpr
		info.isStruct = innerInfo.isStruct

	case *ast.ArrayType:
		info.isSlice = true
		innerInfo := analyzeFieldType(t.Elt)
		info.typeExpr = "[]" + innerInfo.typeExpr
		info.isStruct = innerInfo.isStruct

	case *ast.MapType:
		info.isMap = true
		keyInfo := analyzeFieldType(t.Key)
		valueInfo := analyzeFieldType(t.Value)
		info.typeExpr = fmt.Sprintf("map[%s]%s", keyInfo.typeExpr, valueInfo.typeExpr)
		info.isStruct = valueInfo.isStruct

	case *ast.SelectorExpr:
		// Для типов из других пакетов
		if x, ok := t.X.(*ast.Ident); ok {
			info.typeExpr = x.Name + "." + t.Sel.Name
		}
	}

	return info
}

func isBuiltinStructType(typeName string) bool {
	builtinTypes := map[string]bool{
		"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"float32": true, "float64": true, "complex64": true, "complex128": true,
		"string": true, "bool": true, "byte": true, "rune": true,
	}
	return !builtinTypes[typeName]
}
