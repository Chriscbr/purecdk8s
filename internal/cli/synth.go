package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

const defaultOutputDirectory = "dist"

type synthFlags struct {
	app                         string
	appSet                      bool
	output                      string
	outputSet                   bool
	stdout                      bool
	stdoutSet                   bool
	validate                    bool
	validateSet                 bool
	pluginsDirectory            string
	pluginsDirectorySet         bool
	validationReportsOutputFile string
	validationReportsSet        bool
	format                      string
	formatSet                   bool
	chartAPIVersion             string
	chartAPIVersionSet          bool
	chartVersion                string
	chartVersionSet             bool
	help                        bool
	version                     bool
}

type resolvedSynthOptions struct {
	app                         string
	output                      *string
	stdout                      bool
	validate                    bool
	pluginsDirectory            *string
	validationReportsOutputFile *string
	format                      SynthesisFormat
	chartAPIVersion             *HelmChartAPIVersion
	chartVersion                *string
}

type manifestFile struct {
	path        string
	displayPath string
}

func (r *runner) runSynth(args []string) error {
	flags, err := parseSynthFlags(args)
	if err != nil {
		return err
	}
	if flags.help {
		r.writeSynthHelp()
		return nil
	}
	if flags.version {
		fmt.Fprintln(r.stdout, r.version)
		return nil
	}

	config, err := ReadConfig(r.workDir)
	if err != nil {
		return fmt.Errorf("Error: %w", err)
	}
	options, err := r.resolveSynthOptions(config, flags)
	if err != nil {
		return err
	}
	if err := validateSynthOptions(config, options); err != nil {
		return err
	}

	return r.synthesize(options, config)
}

func parseSynthFlags(args []string) (synthFlags, error) {
	var flags synthFlags

	for i := 0; i < len(args); i++ {
		arg := args[i]
		name, inlineValue, hasInlineValue := splitOption(arg)

		switch name {
		case "--help", "-h":
			value, consumed, err := booleanOption(args, i, inlineValue, hasInlineValue, true)
			if err != nil {
				return flags, fmt.Errorf("Invalid value for --help: %s", inlineValue)
			}
			i += consumed
			flags.help = value
		case "--version":
			value, consumed, err := booleanOption(args, i, inlineValue, hasInlineValue, true)
			if err != nil {
				return flags, fmt.Errorf("Invalid value for --version: %s", inlineValue)
			}
			i += consumed
			flags.version = value
		case "--check-upgrade":
			_, consumed, err := booleanOption(args, i, inlineValue, hasInlineValue, true)
			if err != nil {
				return flags, fmt.Errorf("Invalid value for --check-upgrade: %s", inlineValue)
			}
			i += consumed
		case "--no-check-upgrade":
			// Upgrade checks require npm and are intentionally a no-op.
		case "--app", "-a":
			value, consumed, err := stringOption(args, i, name, inlineValue, hasInlineValue)
			if err != nil {
				return flags, err
			}
			i += consumed
			flags.app, flags.appSet = value, true
		case "--output", "-o":
			value, consumed, err := stringOption(args, i, name, inlineValue, hasInlineValue)
			if err != nil {
				return flags, err
			}
			i += consumed
			flags.output, flags.outputSet = value, true
		case "--stdout", "-p":
			value, consumed, err := booleanOption(args, i, inlineValue, hasInlineValue, true)
			if err != nil {
				return flags, fmt.Errorf("Invalid value for --stdout: %s", inlineValue)
			}
			i += consumed
			flags.stdout, flags.stdoutSet = value, true
		case "--no-stdout":
			flags.stdout, flags.stdoutSet = false, true
		case "--validate":
			value, consumed, err := booleanOption(args, i, inlineValue, hasInlineValue, true)
			if err != nil {
				return flags, fmt.Errorf("Invalid value for --validate: %s", inlineValue)
			}
			i += consumed
			flags.validate, flags.validateSet = value, true
		case "--no-validate":
			flags.validate, flags.validateSet = false, true
		case "--plugins-dir":
			value, consumed, err := stringOption(args, i, name, inlineValue, hasInlineValue)
			if err != nil {
				return flags, err
			}
			i += consumed
			flags.pluginsDirectory, flags.pluginsDirectorySet = value, true
		case "--validation-reports-output-file":
			value, consumed, err := stringOption(args, i, name, inlineValue, hasInlineValue)
			if err != nil {
				return flags, err
			}
			i += consumed
			flags.validationReportsOutputFile, flags.validationReportsSet = value, true
		case "--format":
			value, consumed, err := stringOption(args, i, name, inlineValue, hasInlineValue)
			if err != nil {
				return flags, err
			}
			i += consumed
			flags.format, flags.formatSet = value, true
		case "--chart-api-version":
			value, consumed, err := stringOption(args, i, name, inlineValue, hasInlineValue)
			if err != nil {
				return flags, err
			}
			i += consumed
			flags.chartAPIVersion, flags.chartAPIVersionSet = value, true
		case "--chart-version":
			value, consumed, err := stringOption(args, i, name, inlineValue, hasInlineValue)
			if err != nil {
				return flags, err
			}
			i += consumed
			flags.chartVersion, flags.chartVersionSet = value, true
		default:
			switch {
			case strings.HasPrefix(arg, "-a") && len(arg) > 2:
				flags.app, flags.appSet = strings.TrimPrefix(strings.TrimPrefix(arg, "-a"), "="), true
			case strings.HasPrefix(arg, "-o") && len(arg) > 2:
				flags.output, flags.outputSet = strings.TrimPrefix(strings.TrimPrefix(arg, "-o"), "="), true
			case strings.HasPrefix(arg, "-p="):
				value, err := parseBool(strings.TrimPrefix(arg, "-p="))
				if err != nil {
					return flags, fmt.Errorf("Invalid value for --stdout: %s", strings.TrimPrefix(arg, "-p="))
				}
				flags.stdout, flags.stdoutSet = value, true
			case strings.HasPrefix(arg, "-"):
				return flags, fmt.Errorf("Unknown argument: %s", strings.TrimLeft(arg, "-"))
			default:
				return flags, fmt.Errorf("Unknown argument: %s", arg)
			}
		}
	}

	return flags, nil
}

func splitOption(arg string) (name, value string, hasValue bool) {
	if !strings.HasPrefix(arg, "-") {
		return arg, "", false
	}
	if before, after, found := strings.Cut(arg, "="); found {
		return before, after, true
	}
	return arg, "", false
}

func stringOption(args []string, index int, name, inlineValue string, hasInlineValue bool) (string, int, error) {
	if hasInlineValue {
		return inlineValue, 0, nil
	}
	if index+1 >= len(args) {
		return "", 0, fmt.Errorf("Not enough arguments following: %s", strings.TrimLeft(name, "-"))
	}
	return args[index+1], 1, nil
}

func booleanOption(args []string, index int, inlineValue string, hasInlineValue, defaultValue bool) (bool, int, error) {
	if hasInlineValue {
		value, err := parseBool(inlineValue)
		return value, 0, err
	}
	if index+1 < len(args) {
		if value, err := parseBool(args[index+1]); err == nil {
			return value, 1, nil
		}
	}
	return defaultValue, 0, nil
}

func (r *runner) resolveSynthOptions(config *Config, flags synthFlags) (resolvedSynthOptions, error) {
	var options resolvedSynthOptions

	switch {
	case flags.appSet:
		options.app = flags.app
	case envStringSet(r.env, &options.app, "CDK8S_APP", "CDK8S_A"):
	case config != nil && config.App != nil:
		options.app = *config.App
	default:
		return options, errors.New("Missing required argument: app")
	}

	options.stdout = false
	if value, ok := envValue(r.env, "CDK8S_STDOUT", "CDK8S_P"); ok {
		parsed, err := parseBool(value)
		if err != nil {
			return options, fmt.Errorf("Invalid value for CDK8S_STDOUT: %s", value)
		}
		options.stdout = parsed
	}
	if flags.stdoutSet {
		options.stdout = flags.stdout
	}

	switch {
	case flags.outputSet:
		options.output = stringPointer(flags.output)
	case envPointer(r.env, &options.output, "CDK8S_OUTPUT", "CDK8S_O"):
	case config != nil && config.Output != nil:
		options.output = stringPointer(*config.Output)
	case !options.stdout:
		options.output = stringPointer(defaultOutputDirectory)
	}

	options.validate = true
	if value, ok := envValue(r.env, "CDK8S_VALIDATE"); ok {
		parsed, err := parseBool(value)
		if err != nil {
			return options, fmt.Errorf("Invalid value for CDK8S_VALIDATE: %s", value)
		}
		options.validate = parsed
	}
	if flags.validateSet {
		options.validate = flags.validate
	}

	switch {
	case flags.pluginsDirectorySet:
		options.pluginsDirectory = stringPointer(flags.pluginsDirectory)
	case envPointer(r.env, &options.pluginsDirectory, "CDK8S_PLUGINS_DIR"):
	case config != nil && config.PluginsDirectory != nil:
		options.pluginsDirectory = stringPointer(*config.PluginsDirectory)
	}

	switch {
	case flags.validationReportsSet:
		options.validationReportsOutputFile = stringPointer(flags.validationReportsOutputFile)
	case envPointer(r.env, &options.validationReportsOutputFile, "CDK8S_VALIDATION_REPORTS_OUTPUT_FILE"):
	}

	format := string(SynthesisFormatPlain)
	switch {
	case flags.formatSet:
		format = flags.format
	case envStringSet(r.env, &format, "CDK8S_FORMAT"):
	case config != nil && config.SynthConfig != nil && config.SynthConfig.Format != nil:
		format = string(*config.SynthConfig.Format)
	}
	options.format = SynthesisFormat(format)

	var chartAPIVersion string
	switch {
	case flags.chartAPIVersionSet:
		chartAPIVersion = flags.chartAPIVersion
		options.chartAPIVersion = helmAPIVersionPointer(chartAPIVersion)
	case envStringSet(r.env, &chartAPIVersion, "CDK8S_CHART_API_VERSION"):
		options.chartAPIVersion = helmAPIVersionPointer(chartAPIVersion)
	case config != nil && config.SynthConfig != nil && config.SynthConfig.ChartAPIVersion != nil:
		value := *config.SynthConfig.ChartAPIVersion
		options.chartAPIVersion = &value
	}
	if options.chartAPIVersion == nil && options.format == SynthesisFormatHelm {
		options.chartAPIVersion = helmAPIVersionPointer(string(HelmChartAPIVersionV2))
	}

	var chartVersion string
	switch {
	case flags.chartVersionSet:
		chartVersion = flags.chartVersion
		options.chartVersion = &chartVersion
	case envStringSet(r.env, &chartVersion, "CDK8S_CHART_VERSION"):
		options.chartVersion = &chartVersion
	case config != nil && config.SynthConfig != nil && config.SynthConfig.ChartVersion != nil:
		options.chartVersion = stringPointer(*config.SynthConfig.ChartVersion)
	}

	return options, nil
}

func envStringSet(env []string, target *string, names ...string) bool {
	value, ok := envValue(env, names...)
	if ok {
		*target = value
	}
	return ok
}

func envPointer(env []string, target **string, names ...string) bool {
	value, ok := envValue(env, names...)
	if ok {
		*target = stringPointer(value)
	}
	return ok
}

func stringPointer(value string) *string {
	return &value
}

func helmAPIVersionPointer(value string) *HelmChartAPIVersion {
	converted := HelmChartAPIVersion(value)
	return &converted
}

func validateSynthOptions(config *Config, options resolvedSynthOptions) error {
	if options.stdout && options.output != nil {
		configOutputMatches := config != nil && config.Output != nil && *config.Output == *options.output
		if !configOutputMatches {
			return errors.New("Error: '--output' and '--stdout' are mutually exclusive. Please only use one.")
		}
	}

	if options.format != SynthesisFormatPlain && options.format != SynthesisFormatHelm {
		return fmt.Errorf(
			"Error: You need to specify synthesis format either as %s or %s but received: %s",
			SynthesisFormatPlain,
			SynthesisFormatHelm,
			options.format,
		)
	}
	if options.chartAPIVersion != nil &&
		*options.chartAPIVersion != "" &&
		*options.chartAPIVersion != HelmChartAPIVersionV1 &&
		*options.chartAPIVersion != HelmChartAPIVersionV2 {
		return fmt.Errorf(
			"Error: You need to specify helm chart api version either as %s or %s but received: %s",
			HelmChartAPIVersionV1,
			HelmChartAPIVersionV2,
			*options.chartAPIVersion,
		)
	}
	if options.format == SynthesisFormatHelm &&
		(options.chartVersion == nil || *options.chartVersion == "") {
		return errors.New("Error: You need to specify '--chart-version' when '--format' is set as 'helm'.")
	}
	if options.chartVersion != nil && *options.chartVersion != "" && !isValidSemVer(*options.chartVersion) {
		return fmt.Errorf(
			"Error: The value specified for '--chart-version': %s does not follow SemVer-2(https://semver.org/).",
			*options.chartVersion,
		)
	}
	if options.stdout && options.format == SynthesisFormatHelm {
		return errors.New("Error: Helm format synthesis does not support 'stdout'. Please use 'outdir' instead.")
	}
	if options.format == SynthesisFormatPlain &&
		options.chartAPIVersion != nil &&
		*options.chartAPIVersion != "" {
		return errors.New("Error: You need to specify '--format' as 'helm' when '--chart-api-version' is set.")
	}
	if options.format == SynthesisFormatPlain &&
		options.chartVersion != nil &&
		*options.chartVersion != "" {
		return errors.New("Error: You need to specify '--format' as 'helm' when '--chart-version' is set.")
	}
	if options.chartAPIVersion != nil &&
		*options.chartAPIVersion == HelmChartAPIVersionV1 &&
		crdsArePresent(configImports(config)) {
		return fmt.Errorf(
			"Error: Your application uses CRDs, which are not supported when '--chart-api-version' is set to %s. Please either set '--chart-api-version' to %s, or remove the CRDs from your cdk8s.yaml configuration file",
			HelmChartAPIVersionV1,
			HelmChartAPIVersionV2,
		)
	}
	if options.validate && hasValidations(config) {
		return errors.New("Error: JavaScript validation plugins are not supported by the pure-Go CLI; use '--no-validate'.")
	}

	return nil
}

func hasValidations(config *Config) bool {
	if config == nil || config.Validations == nil {
		return false
	}
	value := reflect.ValueOf(config.Validations)
	switch value.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return value.Len() != 0
	default:
		return true
	}
}

func (r *runner) synthesize(options resolvedSynthOptions, config *Config) error {
	if options.output != nil && *options.output != "" {
		outputPath := r.resolvePath(*options.output)
		if err := r.ensureSafeCleanTarget(outputPath); err != nil {
			return fmt.Errorf("Error: %w", err)
		}
		if err := os.RemoveAll(outputPath); err != nil {
			return fmt.Errorf("Error: remove output directory %q: %w", *options.output, err)
		}
	}

	if options.stdout {
		tempDir, err := os.MkdirTemp("", "cdk8s-")
		if err != nil {
			return fmt.Errorf("Error: create temporary synthesis directory: %w", err)
		}
		defer os.RemoveAll(tempDir)

		manifests, err := r.synthApp(options.app, tempDir, tempDir, true)
		if err != nil {
			return err
		}
		for _, manifest := range manifests {
			if err := copyFile(r.stdout, manifest.path); err != nil {
				return fmt.Errorf("Error: write synthesized manifest %q to stdout: %w", manifest.path, err)
			}
		}
		return nil
	}

	output := ""
	if options.output != nil {
		output = *options.output
	}
	actualOutput := r.resolvePath(output)
	if options.format == SynthesisFormatHelm {
		if actualOutput == "" {
			actualOutput = r.workDir
		}
		if err := r.createHelmScaffolding(actualOutput, options, config); err != nil {
			return fmt.Errorf("Error: create Helm chart scaffolding: %w", err)
		}
		output = filepath.Join(output, "templates")
		actualOutput = filepath.Join(actualOutput, "templates")
	}
	_, err := r.synthApp(options.app, output, actualOutput, false)
	return err
}

func (r *runner) synthApp(command, environmentOutput, actualOutput string, stdout bool) ([]manifestFile, error) {
	if !stdout {
		fmt.Fprintln(r.stdout, "Synthesizing application")
	}
	if err := r.runApp(command, environmentOutput); err != nil {
		return nil, err
	}

	if _, err := os.Stat(actualOutput); err != nil {
		if errors.Is(err, os.ErrNotExist) || actualOutput == "" {
			return nil, fmt.Errorf(`ERROR: synthesis failed, app expected to create "%s"`, environmentOutput)
		}
		return nil, fmt.Errorf("Error: inspect synthesis output %q: %w", environmentOutput, err)
	}

	manifests, err := findManifests(actualOutput, environmentOutput)
	if err != nil {
		return nil, fmt.Errorf("Error: find synthesized manifests in %q: %w", environmentOutput, err)
	}
	if len(manifests) == 0 {
		fmt.Fprintln(r.stderr, "No manifests synthesized")
		return manifests, nil
	}
	if !stdout {
		for _, manifest := range manifests {
			fmt.Fprintf(r.stdout, "  - %s\n", manifest.displayPath)
		}
	}

	return manifests, nil
}

func (r *runner) runApp(command, outdir string) error {
	shell := "/bin/sh"
	args := []string{"-c", command}
	if runtime.GOOS == "windows" {
		shell = "cmd.exe"
		args = []string{"/d", "/s", "/c", command}
	}

	env := setEnv(r.env, "CDK8S_OUTDIR", outdir)
	if _, ok := envValue(env, "CDK8S_RECORD_CONSTRUCT_METADATA"); !ok {
		env = setEnv(env, "CDK8S_RECORD_CONSTRUCT_METADATA", "false")
	}

	cmd := exec.Command(shell, args...)
	cmd.Dir = r.workDir
	cmd.Env = env
	cmd.Stdin = r.stdin
	cmd.Stdout = io.Discard
	cmd.Stderr = r.stderr

	err := cmd.Run()
	if err == nil {
		return nil
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return fmt.Errorf(
			"Error: command %q at %s returned a non-zero exit code %d",
			command,
			r.workDir,
			exitError.ExitCode(),
		)
	}
	return fmt.Errorf("Error: command %q at %s failed: %w", command, r.workDir, err)
}

func (r *runner) resolvePath(path string) string {
	if path == "" || filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(r.workDir, path)
}

func (r *runner) ensureSafeCleanTarget(path string) error {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve output directory %q: %w", path, err)
	}
	absolute = filepath.Clean(absolute)

	root := filepath.VolumeName(absolute) + string(filepath.Separator)
	if absolute == root {
		return fmt.Errorf("refusing to remove filesystem root as the output directory")
	}

	relativeWorkDir, err := filepath.Rel(absolute, r.workDir)
	if err != nil {
		return fmt.Errorf("compare output directory %q with working directory: %w", path, err)
	}
	if relativeWorkDir == "." ||
		(relativeWorkDir != ".." && !strings.HasPrefix(relativeWorkDir, ".."+string(filepath.Separator))) {
		return fmt.Errorf("refusing to remove output directory %q because it contains the working directory", path)
	}
	return nil
}

func findManifests(directory, displayDirectory string) ([]manifestFile, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	var manifests []manifestFile
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			manifests = append(manifests, manifestFile{
				path:        filepath.Join(directory, entry.Name()),
				displayPath: filepath.Join(displayDirectory, entry.Name()),
			})
		}
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		nested, err := findManifests(
			filepath.Join(directory, entry.Name()),
			filepath.Join(displayDirectory, entry.Name()),
		)
		if err != nil {
			return nil, err
		}
		manifests = append(manifests, nested...)
	}

	return manifests, nil
}

func copyFile(destination io.Writer, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(destination, file)
	return err
}

func (r *runner) writeSynthHelp() {
	fmt.Fprintf(r.stdout, `%s synth

Synthesizes Kubernetes manifests for all charts in your app.

Options:
  --version                         Show version number  [boolean]
  --check-upgrade                   Check for cdk8s-cli upgrade  [boolean] [default: true]
  --help                            Show help  [boolean]
  --app, -a                         Command to use in order to execute cdk8s app  [required]
  --output, -o                      Output directory
  --stdout, -p                      Write synthesized manifests to STDOUT instead of the output directory  [boolean]
  --plugins-dir                     Directory to store cdk8s plugins.
  --validate                        Apply validation plugins on the resulting manifests (use --no-validate to disable)  [boolean]
  --validation-reports-output-file  File to write a JSON representation of the validation reports to
  --format                          Synthesis format for Kubernetes manifests. The default synthesis format is plain kubernetes manifests.  [string]
  --chart-api-version               Chart API version of helm chart. The default value would be 'v2' api version when synthesis format is helm. There is no default set when synthesis format is plain.  [string]
  --chart-version                   Chart version of helm chart. This is required if synthesis format is helm.
`, r.name)
}
