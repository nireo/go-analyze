package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/url"
	"sync"
)

type analyzer struct {
	fset          *token.FileSet
	functionCalls map[string]struct{} // doesn't work for methods.
	functions     []*ast.FuncDecl
	wg            sync.WaitGroup
}

// goal of this project is to implement some of the checks from https://staticcheck.io/docs/checks
// as a proof of concept.
func (a *analyzer) analyzeCalls(n *ast.CallExpr) {
	defer a.wg.Done()
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
		case "fmt":
			if x.Sel.Name != "Sprintf" {
				return
			}

			if len(n.Args) == 1 {
				fmt.Printf("useless call to fmt.Sprintf at: %v", a.fset.Position(n.Pos()))
			}
		}
	case *ast.Ident:
		a.functionCalls[x.Name] = struct{}{}
	}
}

func main() {
	fset := token.NewFileSet() // positions are relative to fset
	src := `package foo
import (
  "url"
)

func bar() {
}

func main() {
  bar()
  fmt.Sprintf("this is useless")
}
`
	f, err := parser.ParseFile(fset, "src.go", src, 0)
	if err != nil {
		panic(err)
	}

	a := analyzer{
		fset:          fset,
		functionCalls: make(map[string]struct{}),
		functions:     make([]*ast.FuncDecl, 0),
		wg:            sync.WaitGroup{},
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			a.wg.Add(1)
			go a.analyzeCalls(x)
		default:
		}
		return true
	})

	for _, d := range f.Decls {
		switch x := d.(type) {
		case *ast.FuncDecl:
			a.functions = append(a.functions, x)
		}
	}
	a.wg.Wait()

	for _, d := range a.functions {
		if d.Name.Name == "main" {
			continue
		}

		if _, ok := a.functionCalls[d.Name.Name]; !ok {
			fmt.Printf("function %s at %s is unused\n", d.Name.Name, a.fset.Position(d.Pos()))
		}
	}
}
