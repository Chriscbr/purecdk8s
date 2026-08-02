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
	"io"
	"os"
	"path/filepath"
	"sort"
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
	hasAPIChanges := hasIncompatibleChanges(report.API)
	if !hasAPIChanges && len(report.Documentation) == 0 {
		fmt.Printf("%s API and documentation match %s.\n", *name, *upstreamPackage)
		return
	}

	fmt.Fprintf(os.Stderr, "%s API or documentation differs from %s:\n", *name, *upstreamPackage)
	if hasAPIChanges {
		if err := report.API.TextIncompatible(os.Stderr, false); err != nil {
			fatal(err)
		}
	}
	if err := WriteDocumentationChanges(os.Stderr, report.Documentation); err != nil {
		fatal(err)
	}
	os.Exit(1)
}

type replacementFlags []string

// Report contains both the type-level API report and documentation changes.
// Documentation is compared for every exported declaration in the upstream
// package; additional declarations in the local package are ignored just as
// compatible additions are ignored by apidiff.
type Report struct {
	API           apidiff.Report
	Documentation []DocumentationChange
}

// DocumentationChange describes one exported declaration whose Go doc comment
// differs between the upstream and local packages. An empty string means that
// the declaration has no doc comment.
type DocumentationChange struct {
	Declaration string
	Upstream    string
	Local       string
}

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

func check(upstreamPackage, localPackage, sourceDir string, replacements replacements) (Report, error) {
	sourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return Report{}, err
	}
	overlay, err := apiOverlay(sourceDir, replacements)
	if err != nil {
		return Report{}, err
	}

	upstream, err := loadPackage(upstreamPackage, nil)
	if err != nil {
		return Report{}, err
	}
	local, err := loadPackage(localPackage, overlay)
	if err != nil {
		return Report{}, err
	}
	return Report{
		API:           apidiff.Changes(upstream.Types, local.Types),
		Documentation: compareDocumentation(packageDocumentation(upstream.Syntax), packageDocumentation(local.Syntax)),
	}, nil
}

func hasIncompatibleChanges(report apidiff.Report) bool {
	for _, change := range report.Changes {
		if !change.Compatible {
			return true
		}
	}
	return false
}

// WriteDocumentationChanges writes documentation changes in a stable,
// human-readable format.
func WriteDocumentationChanges(writer io.Writer, changes []DocumentationChange) error {
	for _, change := range changes {
		if _, err := fmt.Fprintf(writer, "doc for %s differs:\n  upstream: %s\n  local:    %s\n", change.Declaration, formatDocumentation(change.Upstream), formatDocumentation(change.Local)); err != nil {
			return err
		}
	}
	return nil
}

func formatDocumentation(documentation string) string {
	if documentation == "" {
		return "<none>"
	}
	return strconv.Quote(documentation)
}

func compareDocumentation(upstream, local map[string]string) []DocumentationChange {
	declarations := make([]string, 0, len(upstream))
	for declaration := range upstream {
		declarations = append(declarations, declaration)
	}
	sort.Strings(declarations)

	var changes []DocumentationChange
	for _, declaration := range declarations {
		if upstream[declaration] == local[declaration] {
			continue
		}
		changes = append(changes, DocumentationChange{
			Declaration: declaration,
			Upstream:    upstream[declaration],
			Local:       local[declaration],
		})
	}
	return changes
}

func packageDocumentation(files []*ast.File) map[string]string {
	documentation := map[string]string{}
	for _, file := range files {
		for _, declaration := range file.Decls {
			switch declaration := declaration.(type) {
			case *ast.FuncDecl:
				if !declaration.Name.IsExported() || !hasExportedReceiver(declaration) {
					continue
				}
				name := "func " + declaration.Name.Name
				if declaration.Recv != nil {
					name = "method " + receiverName(declaration.Recv.List[0].Type).Name + "." + declaration.Name.Name
				}
				documentation[name] = commentText(declaration.Doc, nil)
			case *ast.GenDecl:
				declarationDocumentation(documentation, declaration)
			}
		}
	}
	return documentation
}

func declarationDocumentation(documentation map[string]string, declaration *ast.GenDecl) {
	switch declaration.Tok {
	case token.CONST, token.VAR:
		kind := declaration.Tok.String()
		for _, item := range declaration.Specs {
			spec := item.(*ast.ValueSpec)
			doc := spec.Doc
			if doc == nil {
				doc = declaration.Doc
			}
			for _, name := range spec.Names {
				if name.IsExported() {
					documentation[kind+" "+name.Name] = commentText(doc, spec.Comment)
				}
			}
		}
	case token.TYPE:
		for _, item := range declaration.Specs {
			spec := item.(*ast.TypeSpec)
			if !spec.Name.IsExported() {
				continue
			}
			doc := spec.Doc
			if doc == nil {
				doc = declaration.Doc
			}
			documentation["type "+spec.Name.Name] = commentText(doc, spec.Comment)
			typeMemberDocumentation(documentation, spec)
		}
	}
}

func typeMemberDocumentation(documentation map[string]string, spec *ast.TypeSpec) {
	var (
		fields       *ast.FieldList
		kind         string
		embeddedKind string
	)
	switch definition := spec.Type.(type) {
	case *ast.StructType:
		fields = definition.Fields
		kind = "field "
		embeddedKind = "embedded field "
	case *ast.InterfaceType:
		fields = definition.Methods
		kind = "interface method "
		embeddedKind = "embedded interface "
	default:
		return
	}

	for _, field := range fields.List {
		if len(field.Names) == 0 {
			name := receiverName(field.Type)
			if name.IsExported() {
				documentation[embeddedKind+spec.Name.Name+"."+name.Name] = commentText(field.Doc, field.Comment)
			}
			continue
		}
		for _, name := range field.Names {
			if name.IsExported() {
				documentation[kind+spec.Name.Name+"."+name.Name] = commentText(field.Doc, field.Comment)
			}
		}
	}
}

func commentText(leading, trailing *ast.CommentGroup) string {
	var result string
	if leading != nil {
		result = leading.Text()
	}
	if trailing != nil {
		result += trailing.Text()
	}
	return result
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
	retainDocumentationComments(file)
	return nil
}

// prepare removes function bodies and private declarations before the local
// package is loaded through an overlay. Keep only comments attached to the
// remaining public API so comments from a removed body or declaration cannot
// be reattached to a neighboring declaration when the overlay is printed and
// parsed again.
func retainDocumentationComments(file *ast.File) {
	kept := map[*ast.CommentGroup]bool{}
	keep := func(group *ast.CommentGroup) {
		if group != nil {
			kept[group] = true
		}
	}

	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			if declaration.Name.IsExported() && hasExportedReceiver(declaration) {
				keep(declaration.Doc)
			}
		case *ast.GenDecl:
			for _, item := range declaration.Specs {
				switch spec := item.(type) {
				case *ast.ValueSpec:
					for _, name := range spec.Names {
						if name.IsExported() {
							keep(declaration.Doc)
							keep(spec.Doc)
							keep(spec.Comment)
							break
						}
					}
				case *ast.TypeSpec:
					if !spec.Name.IsExported() {
						continue
					}
					keep(declaration.Doc)
					keep(spec.Doc)
					keep(spec.Comment)
					var fields *ast.FieldList
					switch definition := spec.Type.(type) {
					case *ast.StructType:
						fields = definition.Fields
					case *ast.InterfaceType:
						fields = definition.Methods
					}
					if fields != nil {
						for _, field := range fields.List {
							keep(field.Doc)
							keep(field.Comment)
						}
					}
				}
			}
		}
	}

	comments := file.Comments[:0]
	for _, group := range file.Comments {
		if kept[group] {
			comments = append(comments, group)
		}
	}
	file.Comments = comments
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
	case *ast.SelectorExpr:
		return expression.Sel
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
