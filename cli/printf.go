package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"reflect"
)

const (
	defaultBufferSize = 128
	thisPackagePath   = `"github.com/yuuki0xff/printfdebug"`
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
				X:   ast.NewIdent("printfdebug"),
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
				X:   ast.NewIdent("printfdebug"),
				Sel: ast.NewIdent("FuncEnd"),
			},
			Args: make([]ast.Expr, 0),
		},
	}
}

func getImportSpec() *ast.ImportSpec {
	return &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: thisPackagePath,
		},
	}
}

func findAllFuncsByCallExper(caller *ast.CallExpr, funcs []*ast.FuncLit) []*ast.FuncLit {
	if funcs == nil {
		panic(errors.New(`"funcs" must be not nil`))
	}

	fmt.Println("           Call", reflect.TypeOf(caller), caller)
	fmt.Println("      Call.Func", reflect.TypeOf(caller.Fun), caller.Fun)
	switch f := caller.Fun.(type) {
	default:
	case *ast.FuncLit:
		funcs = append(funcs, f)
		funcs = findAllFuncsByStmt(f.Body.List, funcs)
		fmt.Println()
	}
	return funcs
}

func findAllFuncsByDecl(decls []ast.Decl, funcs []*ast.FuncLit) []*ast.FuncLit {
	if funcs == nil {
		panic(errors.New(`"funcs" must be not nil`))
	}

	for _, decl := range decls {
		switch d := decl.(type) {
		default:
			fmt.Println("+other", reflect.TypeOf(d), d)
		}
	}
	return funcs
}

func findAllFuncsByStmt(stmts []ast.Stmt, funcs []*ast.FuncLit) []*ast.FuncLit {
	if funcs == nil {
		panic(errors.New(`"funcs" must be not nil`))
	}

	for _, stmt := range stmts {
		switch s := stmt.(type) {
		default:
			fmt.Println("      stmt", reflect.TypeOf(s), s)
		case *ast.DeferStmt:
			fmt.Println(" deferstmt", reflect.TypeOf(s), s)
			funcs = findAllFuncsByCallExper(s.Call, funcs)
		case *ast.GoStmt:
			fmt.Println("    gostmt", reflect.TypeOf(s), s)
			funcs = findAllFuncsByCallExper(s.Call, funcs)
		case *ast.IfStmt:
			fmt.Println("    ifstmt", reflect.TypeOf(s), s)
			funcs = findAllFuncsByStmt(s.Body.List, funcs)
		case *ast.ForStmt:
			fmt.Println("   forstmt", reflect.TypeOf(s), s)
			funcs = findAllFuncsByStmt(s.Body.List, funcs)
		case *ast.SwitchStmt:
			fmt.Println("switchstmt", reflect.TypeOf(s), s)
			funcs = findAllFuncsByStmt(s.Body.List, funcs)
		case *ast.TypeSwitchStmt:
			fmt.Println("typeswitchstmt", reflect.TypeOf(s), s)
			funcs = findAllFuncsByStmt(s.Body.List, funcs)
		case *ast.DeclStmt:
			fmt.Println("  declstmt", reflect.TypeOf(s), s)
			decls := make([]ast.Decl, 1)
			decls[0] = s.Decl
			funcs = findAllFuncsByDecl(decls, funcs)
		}
	}
	return funcs
}

func findAllFuncs(decls []ast.Decl) ([]*ast.FuncDecl, []*ast.FuncLit, *ast.GenDecl) {
	funcs := make([]*ast.FuncDecl, 0, defaultBufferSize)
	innerFuncs := make([]*ast.FuncLit, 0, defaultBufferSize)
	var importDecl *ast.GenDecl

	for _, decl := range decls {
		switch d := decl.(type) {
		default:
			fmt.Println("other", reflect.TypeOf(d), d)
		case *ast.GenDecl:
			fmt.Println("gendecl", reflect.TypeOf(d), d)
			if d.Tok == token.IMPORT {
				importDecl = d
			}
		case *ast.FuncDecl:
			fmt.Println("funcdecl", d)
			funcs = append(funcs, d)

			innerFuncs = findAllFuncsByStmt(d.Body.List, innerFuncs)
		}
	}
	return funcs, innerFuncs, importDecl
}

func addPrintf(file *ast.File) error {
	funcs, innerFuncs, importDecl := findAllFuncs(file.Decls)

	blocks := make([]*ast.BlockStmt, len(funcs)+len(innerFuncs))
	{
		// initialize blocks
		i := 0
		for _, f := range funcs {
			blocks[i] = f.Body
			i++
		}
		for _, f := range innerFuncs {
			blocks[i] = f.Body
			i++
		}
	}

	isAddedDebugLog := false

	for _, s := range blocks {
		stmtList := s.List
		s.List = make([]ast.Stmt, 0)
		s.List = append(s.List, getFuncStartStmt())
		s.List = append(s.List, getFuncEndStmt())
		for i := range stmtList {
			s.List = append(s.List, stmtList[i])
		}
		isAddedDebugLog = true
	}

	if isAddedDebugLog {
		importDecl.Specs = append(importDecl.Specs, getImportSpec())
		file.Imports = append(file.Imports, getImportSpec())
		for _, imp := range file.Imports {
			fmt.Printf(
				"import path=%s, name=%s, comment=%s, doc=%s\n",
				imp.Path.Value,
				imp.Name,
				imp.Comment,
				imp.Doc)
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
		for _, pkg := range pkgs {
			fmt.Println("pkgname:", pkg.Name)
			for fpath, file := range pkg.Files {
				fmt.Println(fpath)
				if error := addPrintf(file); error != nil {
					fmt.Println(error)
					return
				}

				if err := os.Remove(fpath); err != nil {
					panic(err)
				}
				f, err := os.Create(fpath)
				if err != nil {
					panic(err)
				}
				defer f.Close()
				format.Node(f, fset, file)
			}
		}
	}
}
