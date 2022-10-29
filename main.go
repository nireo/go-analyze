package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/url"
)

type analyzer struct {
	fset *token.FileSet
}

// goal of this project is to implement some of the checks from https://staticcheck.io/docs/checks
// as a proof of concept.
func (a *analyzer) analyzeCalls(n *ast.CallExpr) {
	if len(n.Args) == 0 {
		panic("no arguments")
	}

	switch x := n.Fun.(type) {
	case *ast.SelectorExpr:
		id, ok := x.X.(*ast.Ident)
		if !ok {
			panic("failed getting identifier")
		}

		switch id.Name {
		case "url":
			if x.Sel.Name != "Parse" {
				return
			}

			stringlit, ok := n.Args[0].(*ast.BasicLit)
			if !ok {
				return
			}

			_, err := url.ParseRequestURI(stringlit.Value)
			fmt.Println(err)
			if err != nil {
				fmt.Printf("invalid url to url.Parse() at: %v", a.fset.Position(n.Pos()))
			}
		}
	}
}

func main() {
	fset := token.NewFileSet() // positions are relative to fset
	src := `package foo
import (
  "url"
)

func bar() {
  url.Parse("hello")
}`
	f, err := parser.ParseFile(fset, "src.go", src, 0)
	if err != nil {
		panic(err)
	}

	a := analyzer{
		fset: fset,
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			a.analyzeCalls(x)
		default:
		}
		return true
	})
}
