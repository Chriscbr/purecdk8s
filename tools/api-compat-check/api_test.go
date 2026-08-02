package apicheck

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"reflect"
	"testing"
)

func TestPackageDocumentationCollectsExportedAPI(t *testing.T) {
	t.Parallel()

	files := parseFiles(t, `package example

// Mode selects a mode.
type Mode string

const (
	// ModeFast selects the fast mode.
	ModeFast Mode = "fast"
	modeSlow Mode = "slow"
)

// DefaultName is the default.
var DefaultName = "example"

// Options configures a value.
type Options struct {
	// Name names the value.
	Name string
	// Runner supplies embedded behavior.
	Runner
	hidden string
}

// Runner runs values.
type Runner interface {
	// Run runs the value.
	Run()
	hide()
}

type ExtendedRunner interface {
	// Runner supplies the base behavior.
	Runner
}

// NewRunner creates a runner.
func NewRunner() Runner { return nil }

// Execute executes the value.
func (Options) Execute() {}

// ignored is not public documentation.
func ignored() {}

type private struct{}

// Exported is not part of the public API because its receiver is private.
func (private) Exported() {}
`)

	want := map[string]string{
		"const ModeFast":                           "ModeFast selects the fast mode.\n",
		"embedded field Options.Runner":            "Runner supplies embedded behavior.\n",
		"embedded interface ExtendedRunner.Runner": "Runner supplies the base behavior.\n",
		"field Options.Name":                       "Name names the value.\n",
		"func NewRunner":                           "NewRunner creates a runner.\n",
		"interface method Runner.Run":              "Run runs the value.\n",
		"method Options.Execute":                   "Execute executes the value.\n",
		"type Mode":                                "Mode selects a mode.\n",
		"type ExtendedRunner":                      "",
		"type Options":                             "Options configures a value.\n",
		"type Runner":                              "Runner runs values.\n",
		"var DefaultName":                          "DefaultName is the default.\n",
	}
	if got := packageDocumentation(files); !reflect.DeepEqual(got, want) {
		t.Fatalf("packageDocumentation() = %#v, want %#v", got, want)
	}
}

func TestPackageDocumentationHandlesGroupedAndGenericDeclarations(t *testing.T) {
	t.Parallel()

	files := parseFiles(t, `package example

// Status values.
const (
	Ready = true // Ready is available.
	Waiting = false
)

// Pair stores two values.
type Pair[A, B any] struct{}

// Swap swaps the values.
func (*Pair[A, B]) Swap() {}
`)

	want := map[string]string{
		"const Ready":      "Status values. Ready is available.\n",
		"const Waiting":    "Status values.\n",
		"method Pair.Swap": "Swap swaps the values.\n",
		"type Pair":        "Pair stores two values.\n",
	}
	if got := packageDocumentation(files); !reflect.DeepEqual(got, want) {
		t.Fatalf("packageDocumentation() = %#v, want %#v", got, want)
	}
}

func TestCommentTextNormalizesEquivalentExampleFormatting(t *testing.T) {
	t.Parallel()

	oldStyle := &ast.CommentGroup{List: []*ast.Comment{
		{Text: "// Example:"},
		{Text: "//   first()"},
		{Text: "//   second()"},
	}}
	gofmtStyle := &ast.CommentGroup{List: []*ast.Comment{
		{Text: "// Example:"},
		{Text: "//"},
		{Text: "//\tfirst()"},
		{Text: "//\tsecond()"},
	}}
	if old, formatted := commentText(oldStyle, nil), commentText(gofmtStyle, nil); old != formatted {
		t.Fatalf("old-style docs = %q, gofmt-style docs = %q", old, formatted)
	}
}

func TestCommentTextIgnoresTrailingWhitespaceAndExtraBlankLines(t *testing.T) {
	t.Parallel()

	compact := &ast.CommentGroup{List: []*ast.Comment{
		{Text: "// Summary."},
		{Text: "//"},
		{Text: "// Details."},
	}}
	spaced := &ast.CommentGroup{List: []*ast.Comment{
		{Text: "// Summary.   "},
		{Text: "//"},
		{Text: "//"},
		{Text: "// Details.\t"},
		{Text: "//"},
	}}
	if want, got := commentText(compact, nil), commentText(spaced, nil); got != want {
		t.Fatalf("docs with extra whitespace = %q, want %q", got, want)
	}
}

func TestCompareDocumentationReportsUpstreamChangesInOrder(t *testing.T) {
	t.Parallel()

	upstream := map[string]string{
		"type Match":   "same\n",
		"func Missing": "upstream\n",
		"const Added":  "",
	}
	local := map[string]string{
		"type Match":    "same\n",
		"func Missing":  "",
		"const Added":   "local-only docs\n",
		"func LocalAPI": "ignored\n",
	}
	want := []DocumentationChange{
		{Declaration: "const Added", Upstream: "", Local: "local-only docs\n"},
		{Declaration: "func Missing", Upstream: "upstream\n", Local: ""},
	}
	if got := compareDocumentation(upstream, local); !reflect.DeepEqual(got, want) {
		t.Fatalf("compareDocumentation() = %#v, want %#v", got, want)
	}
}

func TestWriteDocumentationChanges(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	err := WriteDocumentationChanges(&output, []DocumentationChange{{
		Declaration: "func Run",
		Upstream:    "Run does work.\n",
	}})
	if err != nil {
		t.Fatal(err)
	}
	want := "doc for func Run differs:\n  upstream: \"Run does work.\\n\"\n  local:    <none>\n"
	if output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}
}

func TestPrepareDoesNotReattachImplementationComments(t *testing.T) {
	t.Parallel()

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "example.go", `package example

func First() {
	// This implementation comment is not documentation.
}

func Between() {}

// Second has documentation after a removed declaration.
func Second() {}

// privateDocs documents a declaration removed by prepare.
func privateDocs() {}
`, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	if err := prepare(file, nil); err != nil {
		t.Fatal(err)
	}

	var overlay bytes.Buffer
	if err := printer.Fprint(&overlay, fileSet, file); err != nil {
		t.Fatal(err)
	}
	prepared := parseFiles(t, overlay.String())
	want := map[string]string{
		"func Between": "",
		"func First":   "",
		"func Second":  "Second has documentation after a removed declaration.\n",
	}
	if got := packageDocumentation(prepared); !reflect.DeepEqual(got, want) {
		t.Fatalf("packageDocumentation() = %#v, want %#v\noverlay:\n%s", got, want, overlay.String())
	}
}

func parseFiles(t *testing.T, source string) []*ast.File {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "example.go", source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	return []*ast.File{file}
}
