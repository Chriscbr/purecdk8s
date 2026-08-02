//go:build ignore

// check-cdk8splus-api-shape compares the public, non-generated cdk8s+ Go
// interface methods, struct fields, and package functions with upstream.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

const version = "v2.0.46"

type typeShape struct {
	kind      string
	fields    map[string]string
	embedded  []string
	completed bool
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go run scripts/check-cdk8splus-api-shape.go <kubernetes-minor-version>")
		os.Exit(2)
	}
	minor := os.Args[1]
	for _, character := range minor {
		if character < '0' || character > '9' {
			fmt.Fprintln(os.Stderr, "kubernetes-minor-version must be numeric")
			os.Exit(2)
		}
	}
	root, err := os.Getwd()
	if err != nil {
		fail(err)
	}
	packageName := "cdk8splus" + minor
	pureDirectory := filepath.Join(root, packageName, "v2")
	upstreamDirectory := os.Getenv("UPSTREAM_DIR")
	reference := os.Getenv("REFERENCE_LABEL")
	if reference == "" {
		reference = "upstream " + version
	}
	if upstreamDirectory == "" {
		upstreamDirectory, err = downloadDirectory("github.com/cdk8s-team/cdk8s-plus-go/" + packageName + "/v2@" + version)
		if err != nil {
			fail(err)
		}
	}
	upstream, upstreamFunctions, err := collect(upstreamDirectory)
	if err != nil {
		fail(err)
	}
	pure, pureFunctions, err := collect(pureDirectory)
	if err != nil {
		fail(err)
	}

	var differences []string
	for name, expected := range upstream {
		actual, exists := pure[name]
		if !exists {
			differences = append(differences, "missing type "+name)
			continue
		}
		if expected.kind != actual.kind {
			continue // Forwarding packages use exact type aliases.
		}
		if expected.kind == "interface" {
			completeMethods(name, upstream, map[string]bool{})
			completeMethods(name, pure, map[string]bool{})
		}
		for field := range expected.fields {
			// These methods come from constructs.Construct. Upstream JSII emits
			// them directly on every interface, while the native API embeds
			// constructs.Construct instead.
			if field == "Node" || field == "ToString" {
				continue
			}
			actualSignature, found := actual.fields[field]
			if !found {
				differences = append(differences, "missing "+name+"."+field)
				continue
			}
			if expected.fields[field] != actualSignature {
				differences = append(differences, fmt.Sprintf("different %s.%s: upstream %s, purecdk8s %s", name, field, expected.fields[field], actualSignature))
			}
		}
	}
	for name, signature := range upstreamFunctions {
		actual, found := pureFunctions[name]
		if !found {
			differences = append(differences, "missing function "+name)
		} else if signature != actual {
			differences = append(differences, fmt.Sprintf("different function %s: upstream %s, purecdk8s %s", name, signature, actual))
		}
	}
	sort.Strings(differences)
	if len(differences) > 0 {
		fmt.Fprintln(os.Stderr, "different exported cdk8s+ API shapes:")
		for _, difference := range differences {
			fmt.Fprintln(os.Stderr, "  "+difference)
		}
		os.Exit(1)
	}
	fmt.Printf("cdk8splus%s public types, interface methods, struct fields, and functions cover %s (%d types, %d functions).\n", minor, reference, len(upstream), len(upstreamFunctions))
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func downloadDirectory(module string) (string, error) {
	result, err := exec.Command("go", "mod", "download", "-json", module).Output()
	if err != nil {
		return "", err
	}
	var metadata struct{ Dir string }
	if err := json.Unmarshal(result, &metadata); err != nil {
		return "", err
	}
	if metadata.Dir == "" {
		return "", fmt.Errorf("module download did not return a source directory")
	}
	return metadata.Dir, nil
}

func collect(directory string) (map[string]*typeShape, map[string]string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, nil, err
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		files = append(files, filepath.Join(directory, name))
	}
	result := map[string]*typeShape{}
	functions := map[string]string{}
	fset := token.NewFileSet()
	for _, filename := range files {
		file, err := parser.ParseFile(fset, filename, nil, 0)
		if err != nil {
			return nil, nil, err
		}
		for _, declaration := range file.Decls {
			if function, ok := declaration.(*ast.FuncDecl); ok && function.Recv == nil && ast.IsExported(function.Name.Name) {
				functions[function.Name.Name] = functionSignature(function.Type)
				continue
			}
			general, ok := declaration.(*ast.GenDecl)
			if !ok || general.Tok != token.TYPE {
				continue
			}
			for _, specification := range general.Specs {
				typeSpec := specification.(*ast.TypeSpec)
				if !ast.IsExported(typeSpec.Name.Name) {
					continue
				}
				shape := &typeShape{fields: map[string]string{}}
				switch value := typeSpec.Type.(type) {
				case *ast.InterfaceType:
					shape.kind = "interface"
					for _, field := range value.Methods.List {
						if len(field.Names) > 0 {
							for _, name := range field.Names {
								if ast.IsExported(name.Name) {
									shape.fields[name.Name] = functionSignature(field.Type.(*ast.FuncType))
								}
							}
							continue
						}
						if embedded := embeddedName(field.Type); embedded != "" {
							shape.embedded = append(shape.embedded, embedded)
						}
					}
				case *ast.StructType:
					shape.kind = "struct"
					for _, field := range value.Fields.List {
						for _, name := range field.Names {
							if ast.IsExported(name.Name) {
								shape.fields[name.Name] = expressionString(field.Type)
							}
						}
					}
				default:
					shape.kind = "other"
				}
				result[typeSpec.Name.Name] = shape
			}
		}
	}
	return result, functions, nil
}

func embeddedName(expression ast.Expr) string {
	switch value := expression.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.StarExpr:
		return embeddedName(value.X)
	default:
		return ""
	}
}

func completeMethods(name string, shapes map[string]*typeShape, visiting map[string]bool) {
	shape, exists := shapes[name]
	if !exists || shape.kind != "interface" || shape.completed || visiting[name] {
		return
	}
	visiting[name] = true
	for _, embedded := range shape.embedded {
		completeMethods(embedded, shapes, visiting)
		if parent, exists := shapes[embedded]; exists {
			for method := range parent.fields {
				shape.fields[method] = parent.fields[method]
			}
		}
	}
	shape.completed = true
	delete(visiting, name)
}

func expressionString(expression ast.Expr) string {
	var output bytes.Buffer
	if err := printer.Fprint(&output, token.NewFileSet(), expression); err != nil {
		panic(err)
	}
	return output.String()
}

func functionSignature(function *ast.FuncType) string {
	parameters := fieldTypes(function.Params)
	results := fieldTypes(function.Results)
	if len(results) == 0 {
		return "(" + strings.Join(parameters, ",") + ")"
	}
	if len(results) == 1 {
		return "(" + strings.Join(parameters, ",") + ")->" + results[0]
	}
	return "(" + strings.Join(parameters, ",") + ")->(" + strings.Join(results, ",") + ")"
}

func fieldTypes(fields *ast.FieldList) []string {
	if fields == nil {
		return nil
	}
	var result []string
	for _, field := range fields.List {
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		for range count {
			result = append(result, expressionString(field.Type))
		}
	}
	return result
}
