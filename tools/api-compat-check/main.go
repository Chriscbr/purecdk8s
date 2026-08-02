// Package apicheck compares a purecdk8s package with its upstream API.
package apicheck

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/exp/apidiff"
	"golang.org/x/tools/go/packages"
)

type replacement struct {
	path string
	name string
}

type replacements map[string]replacement

// Main runs the API compatibility checker command.
func Main() {
	var (
		name            = flag.String("name", "package", "name used in result messages")
		upstreamPackage = flag.String("upstream", "", "upstream package import path")
		localPackage    = flag.String("local", "", "local package import path")
		sourceDir       = flag.String("source", "", "local package source directory")
		replacementArgs replacementFlags
	)
	flag.Var(&replacementArgs, "replace", "local=upstream,package-name import replacement (repeatable)")
	flag.Parse()

	if *upstreamPackage == "" || *localPackage == "" || *sourceDir == "" || flag.NArg() != 0 {
		flag.Usage()
		os.Exit(2)
	}

	replacements, err := replacementArgs.parse()
	if err != nil {
		fatal(err)
	}
	report, err := check(*upstreamPackage, *localPackage, *sourceDir, replacements)
	if err != nil {
		fatal(err)
	}
	if !hasIncompatibleChanges(report) {
		fmt.Printf("%s API is compatible with %s.\n", *name, *upstreamPackage)
		return
	}

	fmt.Fprintf(os.Stderr, "%s API differs from %s:\n", *name, *upstreamPackage)
	if err := report.TextIncompatible(os.Stderr, false); err != nil {
		fatal(err)
	}
	os.Exit(1)
}

type replacementFlags []string

func (values *replacementFlags) String() string {
	return strings.Join(*values, " ")
}

func (values *replacementFlags) Set(value string) error {
	*values = append(*values, value)
	return nil
}

func (values replacementFlags) parse() (replacements, error) {
	result := make(replacements, len(values))
	for _, value := range values {
		local, upstreamAndName, found := strings.Cut(value, "=")
		upstream, name, foundName := strings.Cut(upstreamAndName, ",")
		if !found || !foundName || local == "" || upstream == "" || name == "" {
			return nil, fmt.Errorf("invalid replacement %q; expected local=upstream,package-name", value)
		}
		if _, exists := result[local]; exists {
			return nil, fmt.Errorf("duplicate replacement for %s", local)
		}
		result[local] = replacement{path: upstream, name: name}
	}
	return result, nil
}

func check(upstreamPackage, localPackage, sourceDir string, replacements replacements) (apidiff.Report, error) {
	sourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return apidiff.Report{}, err
	}
	overlay, err := apiOverlay(sourceDir, replacements)
	if err != nil {
		return apidiff.Report{}, err
	}

	upstream, err := loadPackage(upstreamPackage, nil)
	if err != nil {
		return apidiff.Report{}, err
	}
	local, err := loadPackage(localPackage, overlay)
	if err != nil {
		return apidiff.Report{}, err
	}
	return apidiff.Changes(upstream.Types, local.Types), nil
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

func apiOverlay(sourceDir string, replacements replacements) (map[string][]byte, error) {
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

func prepare(file *ast.File, replacements replacements) error {
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
				importSpec.Path.Value = strconv.Quote(replacement.path)
				if importSpec.Name == nil {
					importSpec.Name = ast.NewIdent(replacement.name)
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

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
