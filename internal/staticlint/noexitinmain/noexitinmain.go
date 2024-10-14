// Package noexitinmain holds analyzer that prevents os.Exit calls from main function of main package
package noexitinmain

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer reports a call to os.Exit in the main function of the main package
var Analyzer = &analysis.Analyzer{
	Name: "noexitinmain",
	Doc:  "reports a call to os.Exit in the main function of the main package",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}
	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "main" && fn.Recv == nil {
				for _, stmt := range fn.Body.List {
					if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
						if callExpr, ok := exprStmt.X.(*ast.CallExpr); ok {
							if fun, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
								if pkgIdent, ok := fun.X.(*ast.Ident); ok && pkgIdent.Name == "os" && fun.Sel.Name == "Exit" {
									pass.Reportf(callExpr.Pos(), "os.Exit is not allowed in main function of main package")
								}
							}
						}
					}
				}
			}
			return true
		})
	}
	return nil, nil
}
