// Command cdk8splus-api-check compares a purecdk8s+ package with the
// corresponding upstream cdk8s-plus-go API.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strconv"

	"golang.org/x/exp/apidiff"
	"golang.org/x/tools/go/packages"
)

const (
	upstreamPackage = "github.com/cdk8s-team/cdk8s-plus-go/cdk8splus34/v2"
	localModule     = "github.com/Chriscbr/purecdk8s"
)

var upstreamImports = map[string]struct {
	path string
	name string
}{
	"github.com/Chriscbr/purecdk8s/cdk8s/v2": {
		path: "github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2",
		name: "cdk8s",
	},
	"github.com/Chriscbr/purecdk8s/constructs/v10": {
		path: "github.com/aws/constructs-go/constructs/v10",
		name: "constructs",
	},
	"github.com/Chriscbr/purecdk8s/jsii": {
		path: "github.com/aws/jsii-runtime-go",
		name: "jsii",
	},
	"github.com/Chriscbr/purecdk8s/cdk8splus34/v2": {
		path: upstreamPackage,
		name: "cdk8splus34",
	},
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: go run . <kubernetes-minor-version> <source-dir>")
		os.Exit(2)
	}

	minorVersion := os.Args[1]
	if _, err := strconv.Atoi(minorVersion); err != nil {
		fmt.Fprintln(os.Stderr, "kubernetes-minor-version must be numeric")
		os.Exit(2)
	}

	sourceDir, err := filepath.Abs(os.Args[2])
	if err != nil {
		fatal(err)
	}
	overlay, err := apiOverlay(sourceDir)
	if err != nil {
		fatal(err)
	}

	localPackage := fmt.Sprintf("%s/cdk8splus%s/v2", localModule, minorVersion)
	oldPackage, err := loadPackage(upstreamPackage, nil)
	if err != nil {
		fatal(err)
	}
	newPackage, err := loadPackage(localPackage, overlay)
	if err != nil {
		fatal(err)
	}

	report := apidiff.Changes(oldPackage.Types, newPackage.Types)
	if !hasIncompatibleChanges(report) {
		fmt.Printf("cdk8splus%s API is compatible with %s.\n", minorVersion, upstreamPackage)
		return
	}

	fmt.Fprintf(os.Stderr, "cdk8splus%s API differs from %s:\n", minorVersion, upstreamPackage)
	if err := report.TextIncompatible(os.Stderr, false); err != nil {
		fatal(err)
	}
	os.Exit(1)
}

func hasIncompatibleChanges(report apidiff.Report) bool {
	for _, change := range report.Changes {
		if !change.Compatible {
			return true
		}
	}
	return false
}

func loadPackage(path string, overlay map[string][]byte) (*packages.Package, error) {
	loaded, err := packages.Load(&packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedSyntax |
			packages.NeedTypes,
		Overlay: overlay,
	}, path)
	if err != nil {
		return nil, err
	}
	if len(loaded) != 1 {
		return nil, fmt.Errorf("expected one package for %s, found %d", path, len(loaded))
	}
	if len(loaded[0].Errors) != 0 {
		return nil, packageErrors(path, loaded[0].Errors)
	}
	return loaded[0], nil
}

func packageErrors(path string, errors []packages.Error) error {
	message := fmt.Sprintf("loading %s", path)
	for _, err := range errors {
		message += "\n" + err.Error()
	}
	return fmt.Errorf("%s", message)
}

func apiOverlay(sourceDir string) (map[string][]byte, error) {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, err
	}

	overlay := map[string][]byte{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" || hasTestSuffix(entry.Name()) {
			continue
		}

		filename := filepath.Join(sourceDir, entry.Name())
		fileSet := token.NewFileSet()
		file, err := parser.ParseFile(fileSet, filename, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		prepare(file)

		var output bytes.Buffer
		if err := (&printer.Config{Mode: printer.TabIndent | printer.UseSpaces, Tabwidth: 8}).Fprint(&output, fileSet, file); err != nil {
			return nil, err
		}
		overlay[filename] = output.Bytes()
	}
	return overlay, nil
}

func hasTestSuffix(name string) bool {
	return len(name) > len("_test.go") && name[len(name)-len("_test.go"):] == "_test.go"
}

func prepare(file *ast.File) {
	declarations := make([]ast.Decl, 0, len(file.Decls))
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok {
			if !function.Name.IsExported() || !hasExportedReceiver(function) {
				continue
			}
			function.Body = panicBody()
		}
		if typeDeclaration, ok := declaration.(*ast.GenDecl); ok && typeDeclaration.Tok == token.TYPE {
			for _, spec := range typeDeclaration.Specs {
				prepareType(spec.(*ast.TypeSpec))
			}
		}
		declarations = append(declarations, declaration)
	}
	file.Decls = declarations

	usedImports := importsUsedByDeclarations(file)
	for _, declaration := range file.Decls {
		importDeclaration, ok := declaration.(*ast.GenDecl)
		if !ok || importDeclaration.Tok != token.IMPORT {
			continue
		}

		imports := make([]ast.Spec, 0, len(importDeclaration.Specs))
		for _, spec := range importDeclaration.Specs {
			importSpec := spec.(*ast.ImportSpec)
			path, err := strconv.Unquote(importSpec.Path.Value)
			if err != nil {
				panic(err)
			}
			if replacement, ok := upstreamImports[path]; ok {
				importSpec.Path.Value = strconv.Quote(replacement.path)
				if importSpec.Name == nil {
					importSpec.Name = ast.NewIdent(replacement.name)
				}
			}

			name := importName(importSpec)
			if name == "." || name == "_" || usedImports[name] {
				imports = append(imports, importSpec)
			}
		}
		importDeclaration.Specs = imports
	}
}

func prepareType(spec *ast.TypeSpec) {
	if !spec.Name.IsExported() {
		return
	}

	switch definition := spec.Type.(type) {
	case *ast.InterfaceType:
		fields := make([]*ast.Field, 0, len(definition.Methods.List))
		for _, field := range definition.Methods.List {
			if len(field.Names) == 0 || field.Names[0].IsExported() {
				fields = append(fields, field)
			}
		}
		definition.Methods.List = fields
	case *ast.StructType:
		fields := make([]*ast.Field, 0, len(definition.Fields.List))
		for _, field := range definition.Fields.List {
			if len(field.Names) == 0 || field.Names[0].IsExported() {
				fields = append(fields, field)
			}
		}
		definition.Fields.List = fields
	}
}

func hasExportedReceiver(function *ast.FuncDecl) bool {
	if function.Recv == nil {
		return true
	}
	if len(function.Recv.List) != 1 {
		return false
	}
	return receiverName(function.Recv.List[0].Type).IsExported()
}

func receiverName(expression ast.Expr) *ast.Ident {
	switch expression := expression.(type) {
	case *ast.Ident:
		return expression
	case *ast.StarExpr:
		return receiverName(expression.X)
	case *ast.IndexExpr:
		return receiverName(expression.X)
	case *ast.IndexListExpr:
		return receiverName(expression.X)
	default:
		return ast.NewIdent("")
	}
}

func panicBody() *ast.BlockStmt {
	return &ast.BlockStmt{List: []ast.Stmt{
		&ast.ExprStmt{X: &ast.CallExpr{
			Fun:  ast.NewIdent("panic"),
			Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote("API-only package")}},
		}},
	}}
}

func importsUsedByDeclarations(file *ast.File) map[string]bool {
	used := map[string]bool{}
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			inspectImportUses(declaration.Type, used)
		case *ast.GenDecl:
			if declaration.Tok != token.IMPORT {
				inspectImportUses(declaration, used)
			}
		}
	}
	return used
}

func inspectImportUses(node ast.Node, used map[string]bool) {
	ast.Inspect(node, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if identifier, ok := selector.X.(*ast.Ident); ok {
			used[identifier.Name] = true
		}
		return true
	})
}

func importName(spec *ast.ImportSpec) string {
	if spec.Name != nil {
		return spec.Name.Name
	}
	path, err := strconv.Unquote(spec.Path.Value)
	if err != nil {
		panic(err)
	}
	return filepath.Base(path)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
