// Command checkdoc verifies conventional documentation for the public package.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

type finding struct {
	file string
	line int
	kind string
	name string
	text string
}

func main() {
	root, err := os.Getwd()
	if err != nil {
		fatal(err)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		fatal(err)
	}

	fset := token.NewFileSet()
	var findings []finding
	total := 0
	packageOverview := ""
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(root, name), nil, parser.ParseComments)
		if err != nil {
			fatal(err)
		}
		if name == "doc.go" && file.Doc != nil {
			packageOverview = file.Doc.Text()
		}
		for _, declaration := range file.Decls {
			switch value := declaration.(type) {
			case *ast.FuncDecl:
				if value.Name.IsExported() {
					total++
					findings = checkComment(findings, fset, name, value.Pos(), "function or method", value.Name.Name, value.Doc)
				}
			case *ast.GenDecl:
				for _, rawSpec := range value.Specs {
					switch spec := rawSpec.(type) {
					case *ast.TypeSpec:
						if !spec.Name.IsExported() {
							continue
						}
						total++
						doc := spec.Doc
						if doc == nil {
							doc = value.Doc
						}
						findings = checkComment(findings, fset, name, spec.Pos(), "type", spec.Name.Name, doc)
						total, findings = checkInterfaceMethods(total, findings, fset, name, spec)
					case *ast.ValueSpec:
						for _, identifier := range spec.Names {
							if !identifier.IsExported() {
								continue
							}
							total++
							doc := spec.Doc
							if doc == nil {
								doc = value.Doc
							}
							findings = checkComment(findings, fset, name, identifier.Pos(), value.Tok.String(), identifier.Name, doc)
						}
					}
				}
			}
		}
	}

	overviewTerms := []string{"Package sidereon", "cgo", "CGO_ENABLED", "Close", "concurrent"}
	for _, term := range overviewTerms {
		if !strings.Contains(packageOverview, term) {
			findings = append(findings, finding{file: "doc.go", line: 1, kind: "package overview", name: term, text: "missing required overview term"})
		}
	}

	sort.Slice(findings, func(i, j int) bool {
		if findings[i].file != findings[j].file {
			return findings[i].file < findings[j].file
		}
		if findings[i].line != findings[j].line {
			return findings[i].line < findings[j].line
		}
		return findings[i].name < findings[j].name
	})
	if len(findings) != 0 {
		for _, item := range findings {
			fmt.Fprintf(os.Stderr, "%s:%d: %s %s: %s\n", item.file, item.line, item.kind, item.name, item.text)
		}
		fmt.Fprintf(os.Stderr, "documentation: %d failures across %d exported declarations and fields\n", len(findings), total)
		os.Exit(1)
	}
	fmt.Printf("documentation: %d exported declarations and fields have conventional comments\n", total)
}

func checkInterfaceMethods(total int, findings []finding, fset *token.FileSet, filename string, spec *ast.TypeSpec) (int, []finding) {
	value, ok := spec.Type.(*ast.InterfaceType)
	if !ok {
		return total, findings
	}
	for _, field := range value.Methods.List {
		for _, identifier := range field.Names {
			if !identifier.IsExported() {
				continue
			}
			total++
			findings = checkComment(findings, fset, filename, identifier.Pos(), "field or interface method", identifier.Name, field.Doc)
		}
	}
	return total, findings
}

func checkComment(findings []finding, fset *token.FileSet, filename string, pos token.Pos, kind, name string, doc *ast.CommentGroup) []finding {
	line := fset.Position(pos).Line
	if doc == nil || strings.TrimSpace(doc.Text()) == "" {
		return append(findings, finding{file: filename, line: line, kind: kind, name: name, text: "missing doc comment"})
	}
	text := strings.TrimSpace(doc.Text())
	startsWithName := strings.HasPrefix(text, name)
	if startsWithName && len(text) > len(name) {
		next, _ := utf8.DecodeRuneInString(text[len(name):])
		startsWithName = next != '_' && !unicode.IsLetter(next) && !unicode.IsDigit(next)
	}
	if !startsWithName {
		return append(findings, finding{file: filename, line: line, kind: kind, name: name, text: fmt.Sprintf("comment must start with %q", name)})
	}
	return findings
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
