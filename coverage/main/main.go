package main

import (
	"bufio"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

func main() {
	if len(os.Args)%2 != 1 {
		log.Fatal("Arguments should be a list of pairs file path - function name.")
	}
	tasks := make(map[string][]string)
	for i := 1; i != len(os.Args); i += 2 {
		if functions, found := tasks[os.Args[i]]; found {
			functions = append(functions, os.Args[i+1])
			tasks[os.Args[i]] = functions
		} else {
			tasks[os.Args[i]] = []string{os.Args[i+1]}
		}
	}
	numberOfCases := make(map[string]map[string]int)
	for filePath, functions := range tasks {
		numberOfCases[filePath] = make(map[string]int)
		if err := os.MkdirAll(filepath.Dir("tmp/"+filePath), 0770); err != nil {
			panic(err)
		}
		os.Rename(filePath, "tmp/"+filePath)
		old, err := os.Open("tmp/" + filePath)
		if err != nil {
			panic(err)
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, filePath, old, parser.ParseComments)
		if err != nil {
			log.Fatal(err)
		}
		for _, function := range functions {
			file, numberOfCases[filePath][function] = updateFunction(file, filePath, function)
		}
		file.Decls = append([]ast.Decl{&ast.GenDecl{
			Tok:   token.IMPORT,
			Specs: []ast.Spec{&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: "\"github.com/hajimehoshi/ebiten/v2/coverage\""}}},
		}}, file.Decls...)
		new, err := os.Create(filePath)
		if err != nil {
			panic(err)
		}
		printer.Fprint(new, fset, file)
		new.Close()
		old.Close()
	}
	exec.Command("go", "test").Run()
	coverages := make(map[string]map[string]float64)
	for filePath, functions := range tasks {
		coverages[filePath] = make(map[string]float64)
		err := os.Remove(filePath)
		if err != nil {
			panic(err)
		}
		err = os.Rename("tmp/"+filePath, filePath)
		if err != nil {
			panic(err)
		}
		for _, function := range functions {
			file, err := os.Open(coverageId(filePath, function))
			if errors.Is(err, os.ErrNotExist) {
				coverages[filePath][function] = 0
			} else if err != nil {
				panic(err)
			} else {
				fileScanner := bufio.NewScanner(file)
				fileScanner.Split(bufio.ScanLines)
				fileLines := make(map[string]bool)
				for fileScanner.Scan() {
					fileLines[fileScanner.Text()] = true
				}
				coverages[filePath][function] = float64(len(fileLines)) /
					float64(numberOfCases[filePath][function])
			}
			file.Close()
			os.Remove(coverageId(filePath, function))
			fmt.Println(filePath, function, coverages[filePath][function])
		}
	}
	os.RemoveAll("tmp")
}

func updateFunction(file *ast.File, filePath string, functionName string) (*ast.File, int) {
	branchId := 0
	numberOfCases := 0
	astutil.Apply(file, nil, func(c *astutil.Cursor) bool {
		n := c.Node()
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Name.Name == functionName {
				c.Replace(astutil.Apply(n, nil, func(c *astutil.Cursor) bool {
					n := c.Node()
					switch x := n.(type) {
					case *ast.IfStmt:
						cond := x.Cond
						c.Replace(&ast.IfStmt{
							Init: x.Init,
							Cond: &ast.CallExpr{
								Fun: &ast.Ident{Name: "coverage.BranchCoverage"},
								Args: []ast.Expr{
									&ast.BasicLit{Kind: token.STRING, Value: "\"" +
										coverageId(filePath, functionName) + "\""},
									&ast.BasicLit{Kind: token.INT, Value: fmt.Sprint(branchId)},
									cond,
								},
							},
							Body: x.Body,
							Else: x.Else,
						})
						branchId++
						numberOfCases += 2
					case *ast.BinaryExpr:
						if x.Op == token.LAND || x.Op == token.LOR {
							c.Replace(&ast.BinaryExpr{
								Op: x.Op,
								X: &ast.CallExpr{
									Fun: &ast.Ident{Name: "coverage.BranchCoverage"},
									Args: []ast.Expr{
										&ast.BasicLit{Kind: token.STRING, Value: "\"" +
											coverageId(filePath, functionName) + "\""},
										&ast.BasicLit{Kind: token.INT, Value: fmt.Sprint(branchId)},
										x.X,
									},
								},
								Y: x.Y,
							})
							branchId++
							numberOfCases += 2
						}
					case *ast.SwitchStmt:
						newBody := switchBody(x.Body.List, branchId, filePath, functionName)
						c.Replace(&ast.SwitchStmt{
							Init: x.Init,
							Tag:  x.Tag,
							Body: &ast.BlockStmt{List: newBody},
						})
						branchId++
						numberOfCases += len(newBody)
					case *ast.TypeSwitchStmt:
						newBody := switchBody(x.Body.List, branchId, filePath, functionName)
						c.Replace(&ast.TypeSwitchStmt{
							Init:   x.Init,
							Assign: x.Assign,
							Body:   &ast.BlockStmt{List: newBody},
						})
						branchId++
						numberOfCases += len(newBody)
					case *ast.RangeStmt:
						newRange := *x
						newRange.Body.List = append([]ast.Stmt{&ast.ExprStmt{X: &ast.CallExpr{
							Fun: &ast.Ident{Name: "coverage.OutputCoverage"},
							Args: []ast.Expr{
								&ast.BasicLit{Kind: token.STRING, Value: "\"" +
									coverageId(filePath, functionName) + "\""},
								&ast.BasicLit{Kind: token.STRING, Value: "\"" + fmt.Sprint(branchId, "enter") + "\\n\""},
							},
						}}}, newRange.Body.List...)
						c.Replace(&ast.BlockStmt{
							List: []ast.Stmt{&newRange,
								&ast.ExprStmt{X: &ast.CallExpr{
									Fun: &ast.Ident{Name: "coverage.OutputCoverage"},
									Args: []ast.Expr{
										&ast.BasicLit{Kind: token.STRING, Value: "\"" +
											coverageId(filePath, functionName) + "\""},
										&ast.BasicLit{Kind: token.STRING, Value: "\"" + fmt.Sprint(branchId, "exit") + "\\n\""},
									},
								}}},
						})
						branchId++
						numberOfCases += 2
					case *ast.ForStmt:
						newFor := *x
						newFor.Body.List = append([]ast.Stmt{&ast.ExprStmt{X: &ast.CallExpr{
							Fun: &ast.Ident{Name: "coverage.OutputCoverage"},
							Args: []ast.Expr{
								&ast.BasicLit{Kind: token.STRING, Value: "\"" +
									coverageId(filePath, functionName) + "\""},
								&ast.BasicLit{Kind: token.STRING, Value: "\"" + fmt.Sprint(branchId, "enter") + "\\n\""},
							},
						}}}, newFor.Body.List...)
						c.Replace(&ast.BlockStmt{
							List: []ast.Stmt{&newFor,
								&ast.ExprStmt{X: &ast.CallExpr{
									Fun: &ast.Ident{Name: "coverage.OutputCoverage"},
									Args: []ast.Expr{
										&ast.BasicLit{Kind: token.STRING, Value: "\"" +
											coverageId(filePath, functionName) + "\""},
										&ast.BasicLit{Kind: token.STRING, Value: "\"" + fmt.Sprint(branchId, "exit") + "\\n\""},
									},
								}}},
						})
						branchId++
						numberOfCases += 2
					}
					return true
				}))
			}
		}
		return true
	})
	return file, numberOfCases
}

func switchBody(body []ast.Stmt, branchId int, filePath, functionName string) []ast.Stmt {
	newBody := []ast.Stmt{}
	hasDefault := false
	for caseId, c := range body {
		newCase := *c.(*ast.CaseClause)
		hasDefault = hasDefault || newCase.List == nil
		newCase.Body = append([]ast.Stmt{&ast.ExprStmt{X: &ast.CallExpr{
			Fun: &ast.Ident{Name: "coverage.OutputCoverage"},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: "\"" +
					coverageId(filePath, functionName) + "\""},
				&ast.BasicLit{Kind: token.STRING, Value: "\"" + fmt.Sprint(branchId, caseId) + "\\n\""},
			},
		}}}, newCase.Body...)
		newBody = append(newBody, &newCase)
	}
	if !hasDefault {
		newCase := ast.CaseClause{Body: []ast.Stmt{&ast.ExprStmt{X: &ast.CallExpr{
			Fun: &ast.Ident{Name: "coverage.OutputCoverage"},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.STRING, Value: "\"" +
					coverageId(filePath, functionName) + "\""},
				&ast.BasicLit{Kind: token.STRING, Value: "\"" + fmt.Sprint(branchId, "default") + "\\n\""},
			},
		}}}}
		newBody = append(newBody, &newCase)
	}
	return newBody
}

func coverageId(filePath, functionName string) string {
	return "coverage_" + strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(filePath,
		"\\", "_"), "/", "_"), ".", "_") + "_" + functionName
}
