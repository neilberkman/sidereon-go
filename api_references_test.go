package sidereon

import (
	"encoding/json"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

type catalogGoPackage struct {
	ImportPath, Export             string
	GoFiles, CgoFiles, TestGoFiles []string
}

func TestCatalogExportReferences(t *testing.T) {
	cmd := exec.Command("go", "list", "-deps", "-export", "-test", "-json", ".")
	output, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	packages := map[string]catalogGoPackage{}
	decoder := json.NewDecoder(strings.NewReader(string(output)))
	for {
		var pkg catalogGoPackage
		err = decoder.Decode(&pkg)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if pkg.ImportPath != "" {
			packages[pkg.ImportPath] = pkg
		}
	}
	root, ok := packages["sidereon.dev/go/v3"]
	if !ok {
		t.Fatal("package not listed")
	}
	fset := token.NewFileSet()
	var files []*ast.File
	required := map[string]bool{}
	names := append(append(root.GoFiles, root.CgoFiles...), root.TestGoFiles...)
	for _, name := range names {
		if name == "api_references_test.go" {
			continue
		}
		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, file)
		if name != "celestrak.go" {
			continue
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Name.IsExported() {
					if d.Recv == nil {
						required[d.Name.Name] = true
					} else {
						receiver := d.Recv.List[0].Type
						if star, ok := receiver.(*ast.StarExpr); ok {
							receiver = star.X
						}
						if id, ok := receiver.(*ast.Ident); ok {
							required[id.Name+"."+d.Name.Name] = true
						}
					}
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if s.Name.IsExported() {
							required[s.Name.Name] = true
						}
					case *ast.ValueSpec:
						for _, id := range s.Names {
							if id.IsExported() {
								required[id.Name] = true
							}
						}
					}
				}
			}
		}
	}
	for _, name := range []string{"OMMFromTLE", "OMM.WithCCSDSHeader", "OMM.ToKVNLenient", "OMM.ToXMLLenient", "OMMRecordSpan", "OMMArray.Spans", "OMM.Equal", "OMMArray.Equal", "SGP4FitOMMHeader", "SGP4FitOMMHeaderError", "OMMHeader"} {
		required[name] = true
	}
	lookup := func(path string) (io.ReadCloser, error) { pkg := packages[path]; return os.Open(pkg.Export) }
	info := types.Info{Selections: map[*ast.SelectorExpr]*types.Selection{}}
	conf := types.Config{Importer: importer.ForCompiler(fset, "gc", lookup), Error: func(error) {}, FakeImportC: true}
	_, err = conf.Check("sidereon.dev/go/v3", fset, files, &info)
	if err != nil {
		t.Fatalf("type checking test references: %v", err)
	}
	for _, file := range files {
		if !strings.HasSuffix(fset.Position(file.Pos()).Filename, "_test.go") {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			switch v := n.(type) {
			case *ast.Ident:
				delete(required, v.Name)
			case *ast.CallExpr:
				selector, ok := v.Fun.(*ast.SelectorExpr)
				if !ok {
					break
				}
				if sel := info.Selections[selector]; sel != nil {
					if method, ok := sel.Obj().(*types.Func); ok {
						sig := method.Type().(*types.Signature)
						recv := sig.Recv().Type()
						if ptr, ok := recv.(*types.Pointer); ok {
							recv = ptr.Elem()
						}
						if named, ok := recv.(*types.Named); ok {
							delete(required, named.Obj().Name()+"."+method.Name())
						}
					}
				}
			}
			return true
		})
	}
	for name := range required {
		t.Errorf("exported identifier %s has no test reference", name)
	}
}
