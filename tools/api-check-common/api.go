// Package apicheck loads a normalized API-only view of a native purecdk8s
// package for comparison with its upstream JSII-backed counterpart.
package apicheck

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

// Replacement identifies an equivalent upstream import and its package name.
type Replacement struct {
	Path string
	Name string
}

// Options identifies the two packages to compare and imports that should be
// normalized before type-checking the pure Go implementation.
type Options struct {
	UpstreamPackage string
	LocalPackage    string
	SourceDir       string
	Replacements    map[string]Replacement
}

// Check compares the upstream package with the public API of the local source.
func Check(options Options) (apidiff.Report, error) {
	if options.UpstreamPackage == "" || options.LocalPackage == "" || options.SourceDir == "" {
		return apidiff.Report{}, fmt.Errorf("upstream package, local package, and source directory are required")
	}

	sourceDir, err := filepath.Abs(options.SourceDir)
	if err != nil {
		return apidiff.Report{}, err
	}
	overlay, err := apiOverlay(sourceDir, options.Replacements)
	if err != nil {
		return apidiff.Report{}, err
	}

	upstream, err := loadPackage(options.UpstreamPackage, nil)
	if err != nil {
		return apidiff.Report{}, err
	}
	local, err := loadPackage(options.LocalPackage, overlay)
	if err != nil {
		return apidiff.Report{}, err
	}
	return apidiff.Changes(upstream.Types, local.Types), nil
}

// HasIncompatibleChanges reports whether a compatibility-breaking change was
// found. Additive local APIs are permitted.
func HasIncompatibleChanges(report apidiff.Report) bool {
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

func apiOverlay(sourceDir string, replacements map[string]Replacement) (map[string][]byte, error) {
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
		if err := prepare(file, replacements); err != nil {
			return nil, err
		}

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

func prepare(file *ast.File, replacements map[string]Replacement) error {
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
				return err
			}
			if replacement, ok := replacements[path]; ok {
				importSpec.Path.Value = strconv.Quote(replacement.Path)
				if importSpec.Name == nil {
					importSpec.Name = ast.NewIdent(replacement.Name)
				}
			}

			name, err := importName(importSpec)
			if err != nil {
				return err
			}
			if name == "." || name == "_" || usedImports[name] {
				imports = append(imports, importSpec)
			}
		}
		importDeclaration.Specs = imports
	}
	return nil
}

func prepareType(spec *ast.TypeSpec) {
	if !spec.Name.IsExported() {
		return
	}

	switch definition := spec.Type.(type) {
	case *ast.InterfaceType:
		definition.Methods.List = exportedFields(definition.Methods.List)
	case *ast.StructType:
		definition.Fields.List = exportedFields(definition.Fields.List)
	}
}

func exportedFields(fields []*ast.Field) []*ast.Field {
	result := make([]*ast.Field, 0, len(fields))
	for _, field := range fields {
		if len(field.Names) == 0 || field.Names[0].IsExported() {
			result = append(result, field)
		}
	}
	return result
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

func importName(spec *ast.ImportSpec) (string, error) {
	if spec.Name != nil {
		return spec.Name.Name, nil
	}
	path, err := strconv.Unquote(spec.Path.Value)
	if err != nil {
		return "", err
	}
	return filepath.Base(path), nil
}
