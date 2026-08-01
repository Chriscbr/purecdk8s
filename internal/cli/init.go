package cli

import (
	"errors"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
)

const (
	initTemplateGo               = "go"
	initTemplateGoApp            = "go-app"
	initTemplateGoLibrary        = "go-library"
	initTemplateGoLibraryPublic  = "go-library-public"
	initTemplateGoLibraryPrivate = "go-library-private"
	purecdk8sModule              = "github.com/purecdk8s/purecdk8s"
	purecdk8sInitialVersion      = "v0.1.0"
)

var initTemplateNames = []string{
	initTemplateGo,
	initTemplateGoApp,
	initTemplateGoLibrary,
	initTemplateGoLibraryPublic,
	initTemplateGoLibraryPrivate,
}

type initTemplateKind int

const (
	initKindApp initTemplateKind = iota
	initKindPublicLibrary
	initKindPrivateLibrary
)

type initArguments struct {
	template string
	help     bool
	version  bool
}

func (r *runner) runInit(args []string) error {
	parsed, err := parseInitArguments(args)
	if err != nil {
		return err
	}
	if parsed.help {
		r.writeInitHelp()
		return nil
	}
	if parsed.version {
		fmt.Fprintln(r.stdout, r.version)
		return nil
	}

	kind, ok := resolveInitTemplate(parsed.template)
	if !ok {
		return invalidInitTemplateError(parsed.template)
	}
	if err := ensureInitDirectoryEmpty(r.workDir); err != nil {
		return err
	}

	fmt.Fprintf(r.stderr, "Initializing a project from the %s template\n", parsed.template)
	return generateGoProject(r.workDir, kind)
}

func parseInitArguments(args []string) (initArguments, error) {
	var parsed initArguments

	for i := 0; i < len(args); i++ {
		arg := args[i]
		name, inlineValue, hasInlineValue := splitOption(arg)
		switch name {
		case "--help", "-h":
			value, consumed, err := booleanOption(args, i, inlineValue, hasInlineValue, true)
			if err != nil {
				return parsed, fmt.Errorf("Invalid value for --help: %s", inlineValue)
			}
			i += consumed
			parsed.help = value
		case "--version":
			value, consumed, err := booleanOption(args, i, inlineValue, hasInlineValue, true)
			if err != nil {
				return parsed, fmt.Errorf("Invalid value for --version: %s", inlineValue)
			}
			i += consumed
			parsed.version = value
		case "--check-upgrade":
			_, consumed, err := booleanOption(args, i, inlineValue, hasInlineValue, true)
			if err != nil {
				return parsed, fmt.Errorf("Invalid value for --check-upgrade: %s", inlineValue)
			}
			i += consumed
		case "--no-check-upgrade":
			// The pure-Go CLI has no npm upgrade check.
		default:
			if strings.HasPrefix(arg, "-") {
				return parsed, fmt.Errorf("Unknown argument: %s", strings.TrimLeft(arg, "-"))
			}
			if parsed.template != "" {
				return parsed, fmt.Errorf("Unknown argument: %s", arg)
			}
			parsed.template = arg
		}
	}

	if parsed.help || parsed.version {
		return parsed, nil
	}
	if parsed.template == "" {
		return parsed, errors.New("Not enough non-option arguments: got 0, need at least 1")
	}
	return parsed, nil
}

func resolveInitTemplate(name string) (initTemplateKind, bool) {
	switch name {
	case initTemplateGo, initTemplateGoApp:
		return initKindApp, true
	case initTemplateGoLibrary, initTemplateGoLibraryPublic:
		return initKindPublicLibrary, true
	case initTemplateGoLibraryPrivate:
		return initKindPrivateLibrary, true
	default:
		return 0, false
	}
}

func invalidInitTemplateError(name string) error {
	quoted := make([]string, len(initTemplateNames))
	for index, template := range initTemplateNames {
		quoted[index] = strconv.Quote(template)
	}
	return fmt.Errorf(
		"Invalid values:\n  Argument: TYPE, Given: %q, Choices: %s",
		name,
		strings.Join(quoted, ", "),
	)
}

func ensureInitDirectoryEmpty(directory string) error {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("Error: inspect project directory: %w", err)
	}
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), ".") {
			return errors.New("Cannot initialize a project in a non-empty directory")
		}
	}
	return nil
}

func generateGoProject(directory string, kind initTemplateKind) error {
	base := filepath.Base(directory)
	modulePath := "example.com/" + base
	packageName := goPackageName(base)

	files := map[string]string{
		ConfigFileName: goLibraryConfig,
		"go.mod":       goModuleFile(modulePath),
	}

	switch kind {
	case initKindApp:
		files[ConfigFileName] = goAppConfig
		files["main.go"] = goAppSource(base)
		files["help"] = goAppHelp
		files["README.md"] = goAppReadme(base, modulePath)
	case initKindPublicLibrary:
		files["chart.go"] = goLibrarySource(packageName)
		files["chart_test.go"] = goLibraryTestSource(packageName)
		files["help"] = goLibraryHelp(false)
		files["README.md"] = goLibraryReadme(base, modulePath, false)
	case initKindPrivateLibrary:
		files["chart.go"] = goLibrarySource(packageName)
		files["chart_test.go"] = goLibraryTestSource(packageName)
		files["help"] = goLibraryHelp(true)
		files["README.md"] = goLibraryReadme(base, modulePath, true)
	default:
		return fmt.Errorf("Error: unknown internal Go template kind %d", kind)
	}

	for name, contents := range files {
		if strings.HasSuffix(name, ".go") {
			formatted, err := format.Source([]byte(contents))
			if err != nil {
				return fmt.Errorf("Error: format generated %s: %w", name, err)
			}
			contents = string(formatted)
		}
		path := filepath.Join(directory, name)
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			return fmt.Errorf("Error: write generated %s: %w", name, err)
		}
	}
	return nil
}

func goPackageName(base string) string {
	var builder strings.Builder
	for _, character := range strings.ToLower(base) {
		if unicode.IsLetter(character) || unicode.IsDigit(character) || character == '_' {
			builder.WriteRune(character)
		}
	}
	name := builder.String()
	if name == "" {
		name = "charts"
	}
	first, _ := utf8FirstRune(name)
	if unicode.IsDigit(first) {
		name = "charts" + name
	}
	if goKeywords[name] {
		name += "charts"
	}
	return name
}

func utf8FirstRune(value string) (rune, int) {
	for _, character := range value {
		return character, len(string(character))
	}
	return 0, 0
}

var goKeywords = map[string]bool{
	"break": true, "default": true, "func": true, "interface": true, "select": true,
	"case": true, "defer": true, "go": true, "map": true, "struct": true,
	"chan": true, "else": true, "goto": true, "package": true, "switch": true,
	"const": true, "fallthrough": true, "if": true, "range": true, "type": true,
	"continue": true, "for": true, "import": true, "return": true, "var": true,
}

func goModuleFile(modulePath string) string {
	return fmt.Sprintf(`module %s

go 1.23.0

require %s %s
`, modulePath, purecdk8sModule, purecdk8sInitialVersion)
}

const goAppConfig = `language: go
app: go run .
imports:
  - k8s
`

const goLibraryConfig = `language: go
imports:
  - k8s
`

func goAppSource(base string) string {
	return fmt.Sprintf(`package main

import (
	"%s/cdk8s/v2"
	"%s/constructs/v10"
)

type MyChartProps struct {
	cdk8s.ChartProps
}

func NewMyChart(scope constructs.Construct, id string, props *MyChartProps) cdk8s.Chart {
	var chartProps cdk8s.ChartProps
	if props != nil {
		chartProps = props.ChartProps
	}
	chart := cdk8s.NewChart(scope, &id, &chartProps)

	// Define resources here.

	return chart
}

func main() {
	app := cdk8s.NewApp(nil)
	NewMyChart(app, %q, nil)
	app.Synth()
}
`, purecdk8sModule, purecdk8sModule, base)
}

func goLibrarySource(packageName string) string {
	return fmt.Sprintf(`package %s

import (
	"%s/cdk8s/v2"
	"%s/constructs/v10"
)

// MyChartProps configures MyChart.
type MyChartProps struct {
	cdk8s.ChartProps
}

// NewMyChart creates a reusable cdk8s chart.
func NewMyChart(scope constructs.Construct, id string, props *MyChartProps) cdk8s.Chart {
	var chartProps cdk8s.ChartProps
	if props != nil {
		chartProps = props.ChartProps
	}
	chart := cdk8s.NewChart(scope, &id, &chartProps)

	// Define reusable resources here.

	return chart
}
`, packageName, purecdk8sModule, purecdk8sModule)
}

func goLibraryTestSource(packageName string) string {
	return fmt.Sprintf(`package %s

import (
	"testing"

	"%s/cdk8s/v2"
)

func TestNewMyChart(t *testing.T) {
	app := cdk8s.Testing_App(nil)
	if chart := NewMyChart(app, "test", nil); chart == nil {
		t.Fatal("NewMyChart returned nil")
	}
}
`, packageName, purecdk8sModule)
}

func goAppReadme(base, modulePath string) string {
	return fmt.Sprintf(`# %s

A pure-Go cdk8s application.

## Commands

    purecdk8s import
    purecdk8s synth
    kubectl apply -f dist/

The module path is %s. When developing against a source checkout, add this to
go.mod and replace the example path:

    replace %s => /path/to/purecdk8s

Then run `+"`go mod tidy`"+`.
`, base, modulePath, purecdk8sModule)
}

func goLibraryReadme(base, modulePath string, private bool) string {
	visibility := "Public"
	visibilityText := "This scaffold is intended for a publicly importable Go module."
	privateNote := ""
	if private {
		visibility = "Private"
		visibilityText = "This scaffold is intended for a private Go module."
		privateNote = fmt.Sprintf(`
Configure Go for your private module path before resolving dependencies:

    go env -w GOPRIVATE=%s
`, modulePath)
	}

	return fmt.Sprintf(`# %s

A reusable pure-Go cdk8s chart library.

## Visibility: %s

%s
%s
## Development

    purecdk8s import
    go test ./...

When developing against a source checkout, add this to go.mod and replace the
example path:

    replace %s => /path/to/purecdk8s

Then run `+"`go mod tidy`"+`.
`, base, visibility, visibilityText, privateNote, purecdk8sModule)
}

const goAppHelp = `========================================================================================================

 Your purecdk8s Go project is ready!

   cat help          Prints this message
   purecdk8s synth   Synthesize k8s manifests to dist/
   purecdk8s import  Imports k8s API objects to "imports/k8s"

  Deploy:
   kubectl apply -f dist/

========================================================================================================
`

func goLibraryHelp(private bool) string {
	visibility := "public"
	if private {
		visibility = "private"
	}
	return fmt.Sprintf(`========================================================================================================

 Your purecdk8s Go library is ready!
 Visibility: %s

   cat help          Prints this message
   purecdk8s import  Imports k8s API objects to "imports/k8s"
   go test ./...     Tests the reusable chart library

========================================================================================================
`, visibility)
}

func (r *runner) writeInitHelp() {
	fmt.Fprintf(r.stdout, `%s init TYPE

Create a new cdk8s project from a template.

Positionals:
  TYPE  Project type  [required] [choices: "go", "go-app", "go-library", "go-library-public", "go-library-private"]

Options:
  --version        Show version number  [boolean]
  --check-upgrade  Check for cdk8s-cli upgrade  [boolean] [default: true]
  --help           Show help  [boolean]
`, r.name)
}
