// generator_test.go
package reset

import (
	"strings"
	"testing"
)

func TestGenerateResetMethod(t *testing.T) {
	structInfo := &structInfo{
		name: "TestStruct",
		fields: []*fieldInfo{
			{name: "ID", typeExpr: "int", isPtr: false, isSlice: false, isMap: false, isStruct: false},
			{name: "Name", typeExpr: "string", isPtr: false, isSlice: false, isMap: false, isStruct: false},
			{name: "Tags", typeExpr: "[]string", isPtr: false, isSlice: true, isMap: false, isStruct: false},
			{name: "Data", typeExpr: "map[string]int", isPtr: false, isSlice: false, isMap: true, isStruct: false},
			{name: "Child", typeExpr: "*TestStruct", isPtr: true, isSlice: false, isMap: false, isStruct: true},
		},
	}

	generated := generateResetMethod(structInfo)

	// Проверяем базовую структуру метода
	if !strings.Contains(generated, "func (t *TestStruct) Reset() {") {
		t.Error("Method signature is incorrect")
	}

	if !strings.Contains(generated, "if t == nil {") {
		t.Error("Nil check is missing")
	}

	// Проверяем сброс примитивов
	if !strings.Contains(generated, "t.ID = 0") {
		t.Error("Int field reset is missing")
	}

	if !strings.Contains(generated, `t.Name = ""`) {
		t.Error("String field reset is missing")
	}

	// Проверяем сброс слайса
	if !strings.Contains(generated, "t.Tags = t.Tags[:0]") {
		t.Error("Slice reset is missing")
	}

	// Проверяем сброс мапы
	if !strings.Contains(generated, "clear(t.Data)") {
		t.Error("Map clear is missing")
	}

	// Проверяем сброс указателя на структуру
	if !strings.Contains(generated, "if t.Child != nil {") {
		t.Error("Pointer nil check is missing")
	}

	if !strings.Contains(generated, "if resetter, ok := interface{}(t.Child).(interface{ Reset() })") {
		t.Error("Interface type assertion for pointer is missing")
	}
}

func TestGenerateFieldReset(t *testing.T) {
	tests := []struct {
		name     string
		field    *fieldInfo
		expected string
	}{
		{
			name: "primitive int",
			field: &fieldInfo{
				name:     "Count",
				typeExpr: "int",
				isPtr:    false,
				isSlice:  false,
				isMap:    false,
				isStruct: false,
			},
			expected: "    t.Count = 0\n",
		},
		{
			name: "string",
			field: &fieldInfo{
				name:     "Name",
				typeExpr: "string",
				isPtr:    false,
				isSlice:  false,
				isMap:    false,
				isStruct: false,
			},
			expected: `    t.Name = ""` + "\n",
		},
		{
			name: "slice",
			field: &fieldInfo{
				name:     "Items",
				typeExpr: "[]string",
				isPtr:    false,
				isSlice:  true,
				isMap:    false,
				isStruct: false,
			},
			expected: "    t.Items = t.Items[:0]\n",
		},
		{
			name: "map",
			field: &fieldInfo{
				name:     "Data",
				typeExpr: "map[int]string",
				isPtr:    false,
				isSlice:  false,
				isMap:    true,
				isStruct: false,
			},
			expected: "    clear(t.Data)\n",
		},
		{
			name: "pointer to primitive",
			field: &fieldInfo{
				name:     "CountPtr",
				typeExpr: "*int",
				isPtr:    true,
				isSlice:  false,
				isMap:    false,
				isStruct: false,
			},
			expected: "    if t.CountPtr != nil {\n        *t.CountPtr = 0\n    }\n",
		},
		{
			name: "pointer to struct",
			field: &fieldInfo{
				name:     "Child",
				typeExpr: "*MyStruct",
				isPtr:    true,
				isSlice:  false,
				isMap:    false,
				isStruct: true,
			},
			expected: "    if t.Child != nil {\n        if resetter, ok := interface{}(t.Child).(interface{ Reset() }); ok {\n            resetter.Reset()\n        } else {\n            *t.Child = MyStruct{}\n        }\n    }\n",
		},
		{
			name: "embedded struct",
			field: &fieldInfo{
				name:     "Embedded",
				typeExpr: "MyStruct",
				isPtr:    false,
				isSlice:  false,
				isMap:    false,
				isStruct: true,
			},
			expected: "    if resetter, ok := interface{}(&t.Embedded).(interface{ Reset() }); ok {\n        resetter.Reset()\n    }\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateFieldReset("t", tt.field)
			if result != tt.expected {
				t.Errorf("generateFieldReset() mismatch:\nExpected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestGetZeroValue(t *testing.T) {
	tests := []struct {
		typeExpr string
		expected string
	}{
		{"int", "0"},
		{"string", `""`},
		{"bool", "false"},
		{"float64", "0"},
		{"MyStruct", "MyStruct{}"},
	}

	for _, tt := range tests {
		t.Run(tt.typeExpr, func(t *testing.T) {
			result := getZeroValue(tt.typeExpr)
			if result != tt.expected {
				t.Errorf("getZeroValue(%s): expected %s, got %s", tt.typeExpr, tt.expected, result)
			}
		})
	}
}

func TestResetGenerator_Integration(t *testing.T) {
	// Создаем тестовые данные
	pkgInfo := &packageInfo{
		name: "testpkg",
		path: "/tmp/testpkg",
		structs: []*structInfo{
			{
				name: "SimpleStruct",
				fields: []*fieldInfo{
					{name: "Value", typeExpr: "int"},
					{name: "Text", typeExpr: "string"},
				},
			},
		},
	}

	generator := NewResetGenerator([]*packageInfo{pkgInfo})

	// Этот тест в основном проверяет, что нет паники и ошибок
	// Реальная генерация файлов тестируется в отдельных тестах
	err := generator.GenerateReset()
	if err != nil {
		t.Errorf("GenerateReset() failed: %v", err)
	}
}
