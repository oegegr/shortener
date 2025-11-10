package reset

import "go/ast"

type packageInfo struct {
	name    string
	path    string
	structs []*structInfo
	imports map[string]string
	files   map[string]*ast.File
}

type structInfo struct {
	name   string
	fields []*fieldInfo
}

type fieldInfo struct {
	name     string
	typeExpr string
	isPtr    bool
	isSlice  bool
	isMap    bool
	isStruct bool
}
