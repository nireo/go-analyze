package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

func analyzeCalls(n *ast.CallExpr) {
	switch x := n.Fun.(type) {
	case *ast.SelectorExpr:
		fmt.Println(x.X.(*ast.Ident).Name, x.Sel.Name)
	}
}

func main() {
	fset := token.NewFileSet() // positions are relative to fset

	src := `package foo

import (
	"fmt"
	"time"
)

func bar() {
	fmt.Println(time.Now())
}`
	f, err := parser.ParseFile(fset, "src.go", src, 0)
	if err != nil {
		panic(err)
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			analyzeCalls(x)
		default:
		}
		return true
	})
}
