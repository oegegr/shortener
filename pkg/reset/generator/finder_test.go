package reset 

import (
	"go/parser"
	_ "go/token"
	"os"
	"path/filepath"
	"testing"
)

func TestStructFinder_Find(t *testing.T) {
	// Создаем временную директорию с тестовыми файлами
	tmpDir := t.TempDir()

	// Создаем тестовую структуру Go
	testPkgDir := filepath.Join(tmpDir, "testpkg")
	err := os.MkdirAll(testPkgDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Создаем тестовый Go файл с resetable структурами
	testFile := `package testpkg

// generate:reset
type TestStruct struct {
    ID    int
    Name  string
    Tags  []string
    Data  map[string]interface{}
    Child *TestStruct
}

// generate:reset  
type AnotherStruct struct {
    Active bool
    Scores []float64
}

// Эта структура не должна быть найдена (нет комментария)
type IgnoredStruct struct {
    Value int
}
`

	err = os.WriteFile(filepath.Join(testPkgDir, "test.go"), []byte(testFile), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Тестируем поиск
	finder := NewStructFinder(tmpDir)
	packages, err := finder.Find()
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}

	if len(packages) != 1 {
		t.Fatalf("Expected 1 package, got %d", len(packages))
	}

	pkg := packages[0]
	if len(pkg.structs) != 2 {
		t.Fatalf("Expected 2 resetable structs, got %d", len(pkg.structs))
	}

	// Проверяем первую структуру
	struct1 := pkg.structs[0]
	if struct1.name != "TestStruct" {
		t.Errorf("Expected first struct to be TestStruct, got %s", struct1.name)
	}

	if len(struct1.fields) != 5 {
		t.Errorf("Expected 5 fields in TestStruct, got %d", len(struct1.fields))
	}

	// Проверяем типы полей
	expectedFields := []struct {
		name     string
		isPtr    bool
		isSlice  bool
		isMap    bool
		isStruct bool
	}{
		{"ID", false, false, false, false},
		{"Name", false, false, false, false},
		{"Tags", false, true, false, false},
		{"Data", false, false, true, false},
		{"Child", true, false, false, true},
	}

	for i, expected := range expectedFields {
		if i >= len(struct1.fields) {
			break
		}
		field := struct1.fields[i]
		if field.name != expected.name {
			t.Errorf("Field %d: expected name %s, got %s", i, expected.name, field.name)
		}
		if field.isPtr != expected.isPtr {
			t.Errorf("Field %s: expected isPtr %v, got %v", field.name, expected.isPtr, field.isPtr)
		}
		if field.isSlice != expected.isSlice {
			t.Errorf("Field %s: expected isSlice %v, got %v", field.name, expected.isSlice, field.isSlice)
		}
		if field.isMap != expected.isMap {
			t.Errorf("Field %s: expected isMap %v, got %v", field.name, expected.isMap, field.isMap)
		}
	}
}

func TestAnalyzeFieldType(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected fieldInfo
	}{
		{
			name: "primitive int",
			code: "int",
			expected: fieldInfo{
				typeExpr: "int",
				isPtr:    false,
				isSlice:  false,
				isMap:    false,
				isStruct: false,
			},
		},
		{
			name: "pointer to string",
			code: "*string",
			expected: fieldInfo{
				typeExpr: "*string",
				isPtr:    true,
				isSlice:  false,
				isMap:    false,
				isStruct: false,
			},
		},
		{
			name: "slice of int",
			code: "[]int",
			expected: fieldInfo{
				typeExpr: "[]int",
				isPtr:    false,
				isSlice:  true,
				isMap:    false,
				isStruct: false,
			},
		},
		{
			name: "map",
			code: "map[string]int",
			expected: fieldInfo{
				typeExpr: "map[string]int",
				isPtr:    false,
				isSlice:  false,
				isMap:    true,
				isStruct: false,
			},
		},
		{
			name: "custom struct",
			code: "MyStruct",
			expected: fieldInfo{
				typeExpr: "MyStruct",
				isPtr:    false,
				isSlice:  false,
				isMap:    false,
				isStruct: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Парсим выражение типа
			expr, err := parser.ParseExpr(tt.code)
			if err != nil {
				t.Fatalf("Failed to parse expression %s: %v", tt.code, err)
			}

			result := analyzeFieldType(expr)

			if result.typeExpr != tt.expected.typeExpr {
				t.Errorf("typeExpr: expected %s, got %s", tt.expected.typeExpr, result.typeExpr)
			}
			if result.isPtr != tt.expected.isPtr {
				t.Errorf("isPtr: expected %v, got %v", tt.expected.isPtr, result.isPtr)
			}
			if result.isSlice != tt.expected.isSlice {
				t.Errorf("isSlice: expected %v, got %v", tt.expected.isSlice, result.isSlice)
			}
			if result.isMap != tt.expected.isMap {
				t.Errorf("isMap: expected %v, got %v", tt.expected.isMap, result.isMap)
			}
			if result.isStruct != tt.expected.isStruct {
				t.Errorf("isStruct: expected %v, got %v", tt.expected.isStruct, result.isStruct)
			}
		})
	}
}

func TestIsBuiltinStructType(t *testing.T) {
	tests := []struct {
		typeName string
		expected bool
	}{
		{"int", false},
		{"string", false},
		{"bool", false},
		{"float64", false},
		{"MyStruct", true},
		{"CustomType", true},
		{"User", true},
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			result := isBuiltinStructType(tt.typeName)
			if result != tt.expected {
				t.Errorf("isBuiltinStructType(%s): expected %v, got %v", tt.typeName, tt.expected, result)
			}
		})
	}
}
