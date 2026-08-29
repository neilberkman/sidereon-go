package native

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func callbackAssignsName(function *ast.FuncLit, name string) bool {
	assigned := false
	ast.Inspect(function.Body, func(node ast.Node) bool {
		assignment, ok := node.(*ast.AssignStmt)
		if !ok || assignment.Tok != token.ASSIGN {
			return true
		}
		for _, expression := range assignment.Lhs {
			identifier, ok := expression.(*ast.Ident)
			if ok && identifier.Name == name {
				assigned = true
				return false
			}
		}
		return true
	})
	return assigned
}

func callbackDeclaresName(function *ast.FuncLit, name string) bool {
	declared := false
	ast.Inspect(function.Body, func(node ast.Node) bool {
		if nested, ok := node.(*ast.FuncLit); ok && nested != function {
			return false
		}
		switch value := node.(type) {
		case *ast.AssignStmt:
			if value.Tok != token.DEFINE {
				return true
			}
			for _, expression := range value.Lhs {
				identifier, ok := expression.(*ast.Ident)
				if ok && identifier.Name == name {
					declared = true
					return false
				}
			}
		case *ast.DeclStmt:
			declaration, ok := value.Decl.(*ast.GenDecl)
			if !ok {
				return true
			}
			for _, specification := range declaration.Specs {
				value, ok := specification.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, identifier := range value.Names {
					if identifier.Name == name {
						declared = true
						return false
					}
				}
			}
		}
		return true
	})
	return declared
}

func literalErrorReturn(statement *ast.ReturnStmt) bool {
	if len(statement.Results) != 1 {
		return false
	}
	switch value := statement.Results[0].(type) {
	case *ast.BasicLit, *ast.CompositeLit:
		return true
	case *ast.Ident:
		return value.Name == "nil" || value.Name == "true" || value.Name == "false"
	default:
		return false
	}
}

func callbackReturnsLiteral(function *ast.FuncLit) bool {
	literal := false
	ast.Inspect(function.Body, func(node ast.Node) bool {
		if nested, ok := node.(*ast.FuncLit); ok && nested != function {
			return false
		}
		if statement, ok := node.(*ast.ReturnStmt); ok && literalErrorReturn(statement) {
			literal = true
		}
		return true
	})
	return literal
}

func TestHandleCallbacksDoNotEraseCapturedErrors(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	var findings []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, entry.Name(), nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", entry.Name(), err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			assignment, ok := node.(*ast.AssignStmt)
			if !ok || len(assignment.Lhs) != len(assignment.Rhs) {
				return true
			}
			for index, lhs := range assignment.Lhs {
				identifier, ok := lhs.(*ast.Ident)
				if !ok || !strings.HasSuffix(strings.ToLower(identifier.Name), "err") {
					continue
				}
				ast.Inspect(assignment.Rhs[index], func(candidate ast.Node) bool {
					function, ok := candidate.(*ast.FuncLit)
					if !ok {
						return true
					}
					if !callbackDeclaresName(function, identifier.Name) && callbackAssignsName(function, identifier.Name) && callbackReturnsLiteral(function) {
						position := fset.Position(function.Pos())
						findings = append(findings, fmt.Sprintf("%s:%d", filepath.Base(position.Filename), position.Line))
					}
					return false
				})
			}
			return true
		})
	}
	if len(findings) != 0 {
		sort.Strings(findings)
		t.Fatalf("callbacks overwrite their receiving error and return a literal: %s", strings.Join(findings, ", "))
	}
}
