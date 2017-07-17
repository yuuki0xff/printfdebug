package main

import (
	"flag"
	"go/parser"
	"go/token"
	"fmt"
	"go/ast"
	"go/format"
	"os"
)

func parseFlag() (dirs []string) {
	flag.Parse()
	dirs = flag.Args()
	return
}

func getFuncStartStmt() *ast.ExprStmt {
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: ast.NewIdent("printfdebug"),
				Sel: ast.NewIdent("FuncStart"),
			},
			Args: make([]ast.Expr, 0),
		},
	}
}

func getFuncEndStmt() *ast.DeferStmt {
	return &ast.DeferStmt{
		Call: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: ast.NewIdent("printfdebug"),
				Sel: ast.NewIdent("FuncEnd"),
			},
			Args: make([]ast.Expr, 0),
		},
	}
}

func addPrintf(file *ast.File) error {
	fmt.Println("list of Decls")
	for _, decl := range file.Decls {
		switch d := decl.(type){
		default:
			fmt.Println("other", d)
		case *ast.FuncDecl:
			fmt.Println("func", d)

			stmtList := d.Body.List
			d.Body.List = make([]ast.Stmt, 0)
			d.Body.List = append(d.Body.List, getFuncStartStmt())
			d.Body.List = append(d.Body.List, getFuncEndStmt())
			for i := range stmtList {
				d.Body.List = append(d.Body.List, stmtList[i])
			}
		}
	}
	return nil
}

func main() {
	packageDirs := parseFlag()

	for _, dirname := range packageDirs {
		fset := token.NewFileSet()
		pkgs, error := parser.ParseDir(fset, dirname, nil, 0)
		if error != nil {
			fmt.Println(error)
			return
		}

		fmt.Println("pkgs len", len(pkgs))
		for pkgname, pkg := range pkgs {
			for _, file := range pkg.Files {
				fmt.Println(pkgname, ":", file.Name)
				if error := addPrintf(file); error != nil {
					fmt.Println(error)
					return
				}

				format.Node(os.Stdout, token.NewFileSet(), file)
			}
		}
	}
}
