package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"

	"golang.org/x/tools/go/ast/astutil"
)

func main() {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", os.Stdin, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	branchId := 0
	numberOfCases := 0
	astutil.Apply(file, nil, func(c *astutil.Cursor) bool {
		n := c.Node()
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Name.Name == os.Args[1] {
				c.Replace(astutil.Apply(n, nil, func(c *astutil.Cursor) bool {
					n := c.Node()
					switch x := n.(type) {
					case *ast.IfStmt:
						cond := x.Cond
						c.Replace(&ast.IfStmt{
							Init: x.Init,
							Cond: &ast.CallExpr{
								Fun: &ast.Ident{Name: "branchCoverage"},
								Args: []ast.Expr{
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
									Fun: &ast.Ident{Name: "branchCoverage"},
									Args: []ast.Expr{
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
						body := x.Body
						newBody := []ast.Stmt{}
						hasDefault := false
						for caseId, c := range body.List {
							newCase := *c.(*ast.CaseClause)
							hasDefault = hasDefault || newCase.List == nil
							newCase.Body = append([]ast.Stmt{&ast.ExprStmt{X: &ast.CallExpr{
								Fun: &ast.Ident{Name: "outputCoverage"},
								Args: []ast.Expr{
									&ast.BasicLit{Kind: token.STRING, Value: "\"" + fmt.Sprint(branchId, caseId) + "\\n\""},
								},
							}}}, newCase.Body...)
							newBody = append(newBody, &newCase)
						}
						if !hasDefault {
							newCase := ast.CaseClause{Body: []ast.Stmt{&ast.ExprStmt{X: &ast.CallExpr{
								Fun: &ast.Ident{Name: "outputCoverage"},
								Args: []ast.Expr{
									&ast.BasicLit{Kind: token.STRING, Value: "\"" + fmt.Sprint(branchId, "default") + "\\n\""},
								},
							}}}}
							newBody = append(newBody, &newCase)
						}
						c.Replace(&ast.SwitchStmt{
							Init: x.Init,
							Tag:  x.Tag,
							Body: &ast.BlockStmt{List: newBody},
						})
						branchId++
						numberOfCases += len(newBody)
					case *ast.TypeSwitchStmt:
						body := x.Body
						newBody := []ast.Stmt{}
						hasDefault := false
						for caseId, c := range body.List {
							newCase := *c.(*ast.CaseClause)
							hasDefault = hasDefault || newCase.List == nil
							newCase.Body = append([]ast.Stmt{&ast.ExprStmt{X: &ast.CallExpr{
								Fun: &ast.Ident{Name: "outputCoverage"},
								Args: []ast.Expr{
									&ast.BasicLit{Kind: token.STRING, Value: "\"" + fmt.Sprint(branchId, caseId) + "\\n\""},
								},
							}}}, newCase.Body...)
							newBody = append(newBody, &newCase)
						}
						if !hasDefault {
							newCase := ast.CaseClause{Body: []ast.Stmt{&ast.ExprStmt{X: &ast.CallExpr{
								Fun: &ast.Ident{Name: "outputCoverage"},
								Args: []ast.Expr{
									&ast.BasicLit{Kind: token.STRING, Value: "\"" + fmt.Sprint(branchId, "default") + "\\n\""},
								},
							}}}}
							newBody = append(newBody, &newCase)
						}
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
							Fun: &ast.Ident{Name: "outputCoverage"},
							Args: []ast.Expr{
								&ast.BasicLit{Kind: token.STRING, Value: "\"" + fmt.Sprint(branchId, "enter") + "\\n\""},
							},
						}}}, newRange.Body.List...)
						c.Replace(&ast.BlockStmt{
							List: []ast.Stmt{&newRange,
								&ast.ExprStmt{X: &ast.CallExpr{
									Fun: &ast.Ident{Name: "outputCoverage"},
									Args: []ast.Expr{
										&ast.BasicLit{Kind: token.STRING, Value: "\"" + fmt.Sprint(branchId, "exit") + "\\n\""},
									},
								}}},
						})
						branchId++
						numberOfCases += 2
					case *ast.ForStmt:
						newFor := *x
						newFor.Body.List = append([]ast.Stmt{&ast.ExprStmt{X: &ast.CallExpr{
							Fun: &ast.Ident{Name: "outputCoverage"},
							Args: []ast.Expr{
								&ast.BasicLit{Kind: token.STRING, Value: "\"" + fmt.Sprint(branchId, "enter") + "\\n\""},
							},
						}}}, newFor.Body.List...)
						c.Replace(&ast.BlockStmt{
							List: []ast.Stmt{&newFor,
								&ast.ExprStmt{X: &ast.CallExpr{
									Fun: &ast.Ident{Name: "outputCoverage"},
									Args: []ast.Expr{
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
	printer.Fprint(os.Stdout, fset, file)
	fmt.Println("func branchCoverage(id int, cond bool) bool {\n" +
		"\toutputCoverage(fmt.Sprintln(id, cond))\n" +
		"\treturn cond\n" +
		"}\n" +
		"func outputCoverage(id string) {\n" +
		"\tcoverageFile, err := os.OpenFile(\"coverage_" + os.Args[1] +
		"\", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)\n" +
		"\tif err != nil {\n" +
		"\t\tpanic(err)\n" +
		"\t}\n" +
		"\tcoverageFile.WriteString(id)\n" +
		"\tcoverageFile.Close()\n" +
		"}")
	log.Println(numberOfCases)
}
