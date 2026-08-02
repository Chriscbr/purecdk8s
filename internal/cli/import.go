package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Chriscbr/purecdk8s/internal/importer"
	"gopkg.in/yaml.v3"
)

type importFlags struct {
	specs              []string
	output             string
	language           string
	classPrefix        string
	classPrefixSet     bool
	disableClassPrefix bool
	save               bool
	excludes           []string
}

func (r *runner) runImport(args []string) error {
	flags := importFlags{save: true}
	for index := 0; index < len(args); index++ {
		arg := args[index]
		next := func(name string) (string, error) {
			if index+1 >= len(args) {
				return "", fmt.Errorf("Not enough arguments following: %s", name)
			}
			index++
			return args[index], nil
		}
		switch {
		case arg == "--help" || arg == "-h":
			r.writeImportHelp()
			return nil
		case arg == "--version":
			fmt.Fprintln(r.stdout, r.version)
			return nil
		case arg == "--save" || arg == "-s":
			flags.save = true
		case arg == "--no-save":
			flags.save = false
		case arg == "--output" || arg == "-o":
			value, err := next(arg)
			if err != nil {
				return err
			}
			flags.output = value
		case strings.HasPrefix(arg, "--output="):
			flags.output = strings.TrimPrefix(arg, "--output=")
		case arg == "--language" || arg == "-l":
			value, err := next(arg)
			if err != nil {
				return err
			}
			flags.language = value
		case strings.HasPrefix(arg, "--language="):
			flags.language = strings.TrimPrefix(arg, "--language=")
		case arg == "--class-prefix":
			value, err := next(arg)
			if err != nil {
				return err
			}
			flags.classPrefix, flags.classPrefixSet = value, true
		case strings.HasPrefix(arg, "--class-prefix="):
			flags.classPrefix, flags.classPrefixSet = strings.TrimPrefix(arg, "--class-prefix="), true
		case arg == "--no-class-prefix":
			flags.disableClassPrefix = true
		case arg == "--exclude":
			start := index + 1
			for index+1 < len(args) && !strings.HasPrefix(args[index+1], "-") {
				index++
				flags.excludes = append(flags.excludes, args[index])
			}
			if index < start {
				return fmt.Errorf("Not enough arguments following: %s", arg)
			}
		case strings.HasPrefix(arg, "--exclude="):
			flags.excludes = append(flags.excludes, strings.TrimPrefix(arg, "--exclude="))
		case arg == "--check-upgrade":
			if index+1 < len(args) {
				if _, err := parseBool(args[index+1]); err == nil {
					index++
				}
			}
		case strings.HasPrefix(arg, "--check-upgrade="):
			value := strings.TrimPrefix(arg, "--check-upgrade=")
			if _, err := parseBool(value); err != nil {
				return fmt.Errorf("Invalid value for --check-upgrade: %s", value)
			}
		case arg == "--no-check-upgrade":
			// Upgrade checks are intentionally a no-op in the standalone binary.
		case strings.HasPrefix(arg, "-"):
			return fmt.Errorf("Unknown argument: %s", strings.TrimLeft(arg, "-"))
		default:
			flags.specs = append(flags.specs, arg)
		}
	}

	config, err := ReadConfig(r.workDir)
	if err != nil {
		return err
	}
	if flags.output == "" {
		if config != nil && config.ImportDirectory != nil {
			flags.output = *config.ImportDirectory
		} else {
			flags.output = "imports"
		}
	}
	if flags.language == "" && config != nil && config.Language != nil {
		flags.language = *config.Language
	}
	if flags.language == "" {
		flags.language = "go"
	}
	if flags.language != "go" {
		return fmt.Errorf("Error: purecdk8s generates Go imports; unsupported language: %s", flags.language)
	}
	if len(flags.specs) == 0 && config != nil && len(config.Imports) > 0 {
		flags.specs = append(flags.specs, config.Imports...)
	}
	if len(flags.specs) == 0 {
		flags.specs = []string{"k8s"}
	}

	output := flags.output
	if !filepath.IsAbs(output) {
		output = filepath.Join(r.workDir, output)
	}
	for _, rawSpec := range flags.specs {
		packageName, source := parseNamedImport(rawSpec)
		resolvedSource := resolveImportSource(r.workDir, source)
		options := importer.Options{
			Source:      resolvedSource,
			OutputDir:   output,
			PackageName: packageName,
			Excludes:    append([]string(nil), flags.excludes...),
			HTTPClient:  r.httpClient,
			WorkDir:     r.workDir,
		}
		if flags.classPrefixSet {
			options.ClassNamePrefix = flags.classPrefix
		}
		if flags.disableClassPrefix {
			options.DisableClassNamePrefix = true
		}
		result, err := importer.Import(context.Background(), options)
		if err != nil {
			return fmt.Errorf("Error: import %s: %w", rawSpec, err)
		}
		for _, generated := range result.Packages {
			relative, relativeErr := filepath.Rel(r.workDir, generated.File)
			if relativeErr != nil {
				relative = generated.File
			}
			fmt.Fprintf(r.stdout, "generated %s (%d resources)\n", relative, len(generated.Resources))
		}
	}

	if flags.save {
		if err := saveImportsConfig(r.workDir, config, flags.specs, flags.output); err != nil {
			return err
		}
	}
	return nil
}

func parseNamedImport(spec string) (name string, source string) {
	name, source, found := strings.Cut(spec, ":=")
	if !found {
		return "", spec
	}
	return strings.TrimSpace(name), strings.TrimSpace(source)
}

func resolveImportSource(workDir, source string) string {
	if strings.HasPrefix(source, "helm:") {
		return source
	}
	if source == "k8s" || strings.HasPrefix(source, "k8s@") ||
		strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") ||
		strings.HasPrefix(source, "file://") || strings.HasPrefix(source, "github:") {
		return source
	}
	if filepath.IsAbs(source) {
		return source
	}
	return filepath.Join(workDir, source)
}

func saveImportsConfig(workDir string, config *Config, specs []string, output string) error {
	if config == nil {
		config = &Config{}
	}
	if config.Language == nil {
		language := "go"
		config.Language = &language
	}
	if output != "" && output != "imports" {
		config.ImportDirectory = &output
	}
	seen := make(map[string]bool, len(config.Imports)+len(specs))
	imports := make([]string, 0, len(config.Imports)+len(specs))
	for _, item := range append(append([]string(nil), config.Imports...), specs...) {
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		imports = append(imports, item)
	}
	config.Imports = imports
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("serialize %s: %w", ConfigFileName, err)
	}
	if err := os.WriteFile(filepath.Join(workDir, ConfigFileName), data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", ConfigFileName, err)
	}
	return nil
}

func (r *runner) writeImportHelp() {
	fmt.Fprintf(r.stdout, `%s import [SPEC]

Imports API objects to your app by generating native Go constructs.

Positionals:
  SPEC  [NAME:=]SPEC; supported specs are k8s, k8s@X.Y.Z, local/HTTP/OCI Helm charts, local CRD YAML, and HTTP(S) CRD YAML.  [default: cdk8s.yaml imports or "k8s"]

Options:
  --version              Show version number  [boolean]
  --check-upgrade        Check for cdk8s-cli upgrade  [boolean] [default: true]
  --help                 Show help  [boolean]
  --save, -s             Save import specs in cdk8s.yaml  [boolean] [default: true]
  --output, -o           Output directory  [string] [default: "imports"]
  --class-prefix         Prefix generated resource class names
  --no-class-prefix      Disable the default "Kube" prefix for Kubernetes imports
  --exclude              Exclude Kubernetes definitions whose raw $ref matches this regular expression (repeatable)
  --language, -l         Output language  [choices: "go"] [default: "go"]

Examples:
  %s import k8s
  %s import k8s@1.25.0 --no-class-prefix
  %s import helm:./charts/webapp
  %s import helm:https://charts.example.test/stable/webapp@1.2.3
  %s import helm:oci://registry.example.test/team/webapp@1.2.3
  %s import widgets:=widgets.example.com_crd.yaml
`, r.name, r.name, r.name, r.name, r.name, r.name, r.name)
}
