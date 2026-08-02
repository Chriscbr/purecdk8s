//go:build ignore

// sync-api-docs copies public Go documentation from an upstream package's
// source directory onto matching declarations in a local package.
package main

import (
	"fmt"
	"go/ast"
	"go/doc/comment"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type documentation struct {
	raw        string
	normalized string
}

type target struct {
	filename      string
	anchor        int
	leadingStart  int
	trailingStart int
	trailingEnd   int
	hasLeading    bool
	hasTrailing   bool
	inheritsGroup bool
	documentation documentation
}

type edit struct {
	start       int
	end         int
	replacement string
	declaration string
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: go run scripts/sync-api-docs.go <upstream-source-dir> <local-source-dir>")
		os.Exit(2)
	}

	upstream, _, err := scan(os.Args[1])
	if err != nil {
		fail(err)
	}
	_, local, err := scan(os.Args[2])
	if err != nil {
		fail(err)
	}

	editsByFile := map[string][]edit{}
	var unresolved []string
	for declaration, wanted := range upstream {
		current, ok := local[declaration]
		if !ok {
			if wanted.normalized != "" {
				unresolved = append(unresolved, declaration)
			}
			continue
		}
		if wanted.normalized == current.documentation.normalized {
			continue
		}

		contents, err := os.ReadFile(current.filename)
		if err != nil {
			fail(err)
		}
		indent := lineIndent(contents, current.anchor)
		replacement := renderDocumentation(wanted.raw, indent)
		if wanted.normalized == "" && current.inheritsGroup && !current.hasLeading {
			// A non-nil, empty spec comment overrides documentation inherited
			// from a grouped declaration without disturbing sibling specs.
			replacement = "//\n" + indent
		}
		start := current.anchor
		if current.hasLeading {
			start = current.leadingStart
		}
		editsByFile[current.filename] = append(editsByFile[current.filename], edit{
			start:       start,
			end:         current.anchor,
			replacement: replacement,
			declaration: declaration,
		})
		if current.hasTrailing {
			editsByFile[current.filename] = append(editsByFile[current.filename], edit{
				start:       current.trailingStart,
				end:         current.trailingEnd,
				declaration: declaration + " trailing comment",
			})
		}
	}

	if len(unresolved) != 0 {
		fmt.Printf("skipped %d documentation locations provided by aliases or embedded types\n", len(unresolved))
	}

	changed := 0
	for filename, edits := range editsByFile {
		contents, err := os.ReadFile(filename)
		if err != nil {
			fail(err)
		}
		edits, err = mergeEdits(edits)
		if err != nil {
			fail(fmt.Errorf("%s: %w", filename, err))
		}
		sort.Slice(edits, func(i, j int) bool { return edits[i].start > edits[j].start })
		for _, edit := range edits {
			contents = append(contents[:edit.start], append([]byte(edit.replacement), contents[edit.end:]...)...)
			changed++
		}
		if err := os.WriteFile(filename, contents, 0o644); err != nil {
			fail(err)
		}
	}
	fmt.Printf("updated %d documentation locations in %s\n", changed, os.Args[2])
}

func scan(directory string) (map[string]documentation, map[string]target, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, nil, err
	}
	var filenames []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		filenames = append(filenames, filepath.Join(directory, entry.Name()))
	}
	sort.Strings(filenames)

	docs := map[string]documentation{}
	targets := map[string]target{}
	for _, filename := range filenames {
		contents, err := os.ReadFile(filename)
		if err != nil {
			return nil, nil, err
		}
		fileSet := token.NewFileSet()
		file, err := parser.ParseFile(fileSet, filename, contents, parser.ParseComments)
		if err != nil {
			return nil, nil, err
		}
		tokenFile := fileSet.File(file.Pos())
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
				add(docs, targets, filename, tokenFile, name, declaration.Pos(), declaration.Doc, nil, nil)
			case *ast.GenDecl:
				collectGeneralDeclaration(docs, targets, filename, tokenFile, declaration)
			}
		}
	}
	return docs, targets, nil
}

func collectGeneralDeclaration(docs map[string]documentation, targets map[string]target, filename string, tokenFile *token.File, declaration *ast.GenDecl) {
	switch declaration.Tok {
	case token.CONST, token.VAR:
		kind := declaration.Tok.String()
		for _, item := range declaration.Specs {
			spec := item.(*ast.ValueSpec)
			leading, anchor, inherited := specificationDocumentation(declaration, spec.Doc, spec.Pos())
			for _, name := range spec.Names {
				if name.IsExported() {
					add(docs, targets, filename, tokenFile, kind+" "+name.Name, anchor, leading, spec.Comment, inherited)
				}
			}
		}
	case token.TYPE:
		for _, item := range declaration.Specs {
			spec := item.(*ast.TypeSpec)
			if !spec.Name.IsExported() {
				continue
			}
			leading, anchor, inherited := specificationDocumentation(declaration, spec.Doc, spec.Pos())
			add(docs, targets, filename, tokenFile, "type "+spec.Name.Name, anchor, leading, spec.Comment, inherited)
			collectTypeMembers(docs, targets, filename, tokenFile, spec)
		}
	}
}

func specificationDocumentation(declaration *ast.GenDecl, specDoc *ast.CommentGroup, specPosition token.Pos) (*ast.CommentGroup, token.Pos, *ast.CommentGroup) {
	if specDoc != nil {
		return specDoc, specPosition, nil
	}
	if len(declaration.Specs) == 1 {
		return declaration.Doc, declaration.Pos(), nil
	}
	return nil, specPosition, declaration.Doc
}

func collectTypeMembers(docs map[string]documentation, targets map[string]target, filename string, tokenFile *token.File, spec *ast.TypeSpec) {
	var fields *ast.FieldList
	kind := ""
	embeddedKind := ""
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
				add(docs, targets, filename, tokenFile, embeddedKind+spec.Name.Name+"."+name.Name, field.Pos(), field.Doc, field.Comment, nil)
			}
			continue
		}
		for _, name := range field.Names {
			if name.IsExported() {
				add(docs, targets, filename, tokenFile, kind+spec.Name.Name+"."+name.Name, field.Pos(), field.Doc, field.Comment, nil)
			}
		}
	}
}

func add(docs map[string]documentation, targets map[string]target, filename string, tokenFile *token.File, declaration string, anchor token.Pos, leading, trailing, inherited *ast.CommentGroup) {
	raw := rawDocumentation(leading, trailing)
	if leading == nil && inherited != nil {
		raw = rawDocumentation(inherited, trailing)
	}
	doc := documentation{raw: raw, normalized: normalizeDocumentation(raw)}
	docs[declaration] = doc
	value := target{
		filename:      filename,
		anchor:        tokenFile.Offset(anchor),
		hasLeading:    leading != nil,
		hasTrailing:   trailing != nil,
		inheritsGroup: leading == nil && inherited != nil,
		documentation: doc,
	}
	if leading != nil {
		value.leadingStart = tokenFile.Offset(leading.Pos())
	}
	if trailing != nil {
		value.trailingStart = tokenFile.Offset(trailing.Pos())
		value.trailingEnd = tokenFile.Offset(trailing.End())
	}
	targets[declaration] = value
}

func rawDocumentation(leading, trailing *ast.CommentGroup) string {
	var result string
	if leading != nil {
		result = leading.Text()
	}
	if trailing != nil {
		result += trailing.Text()
	}
	return result
}

func normalizeDocumentation(value string) string {
	if value == "" {
		return ""
	}
	parser := comment.Parser{}
	printer := comment.Printer{TextWidth: -1}
	return string(printer.Text(parser.Parse(value)))
}

func renderDocumentation(value, indent string) string {
	if value == "" {
		return ""
	}
	parser := comment.Parser{}
	printer := comment.Printer{TextWidth: 80}
	rendered := strings.TrimSuffix(string(printer.Text(parser.Parse(value))), "\n")
	lines := strings.Split(rendered, "\n")
	for index, line := range lines {
		switch {
		case line == "":
			lines[index] = "//"
		case strings.HasPrefix(line, "\t"):
			lines[index] = "//" + line
		default:
			lines[index] = "// " + line
		}
	}
	return strings.Join(lines, "\n"+indent) + "\n" + indent
}

func lineIndent(contents []byte, offset int) string {
	lineStart := offset
	for lineStart > 0 && contents[lineStart-1] != '\n' {
		lineStart--
	}
	return string(contents[lineStart:offset])
}

func mergeEdits(edits []edit) ([]edit, error) {
	byRange := map[string]edit{}
	for _, candidate := range edits {
		key := fmt.Sprintf("%d:%d", candidate.start, candidate.end)
		if existing, ok := byRange[key]; ok {
			if existing.replacement != candidate.replacement {
				return nil, fmt.Errorf("conflicting edits for %s and %s", existing.declaration, candidate.declaration)
			}
			continue
		}
		byRange[key] = candidate
	}
	result := make([]edit, 0, len(byRange))
	for _, candidate := range byRange {
		result = append(result, candidate)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].start < result[j].start })
	for index := 1; index < len(result); index++ {
		if result[index].start < result[index-1].end {
			return nil, fmt.Errorf("overlapping edits for %s and %s", result[index-1].declaration, result[index].declaration)
		}
	}
	return result, nil
}

func hasExportedReceiver(function *ast.FuncDecl) bool {
	if function.Recv == nil {
		return true
	}
	return len(function.Recv.List) == 1 && receiverName(function.Recv.List[0].Type).IsExported()
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

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
