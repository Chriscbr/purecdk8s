package cli

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const DefaultVersion = "dev"

// Options supplies the process-facing dependencies used by Run. The zero value
// uses the current process streams, environment, and working directory.
type Options struct {
	Stdin   io.Reader
	Stdout  io.Writer
	Stderr  io.Writer
	Env     []string
	WorkDir string
	Version string
	Name    string
	// HTTPClient overrides network access for import operations.
	HTTPClient *http.Client
}

type runner struct {
	stdin      io.Reader
	stdout     io.Writer
	stderr     io.Writer
	env        []string
	workDir    string
	version    string
	name       string
	httpClient *http.Client
}

// Run executes the purecdk8s command and returns the process exit code.
func Run(args []string, options Options) int {
	r, err := newRunner(options)
	if err != nil {
		stderr := options.Stderr
		if stderr == nil {
			stderr = os.Stderr
		}
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	if err := r.run(args); err != nil {
		fmt.Fprintln(r.stderr, err)
		return 1
	}
	return 0
}

func newRunner(options Options) (*runner, error) {
	if options.Stdin == nil {
		options.Stdin = os.Stdin
	}
	if options.Stdout == nil {
		options.Stdout = os.Stdout
	}
	if options.Stderr == nil {
		options.Stderr = os.Stderr
	}
	if options.Env == nil {
		options.Env = os.Environ()
	}
	if options.WorkDir == "" {
		workDir, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("determine working directory: %w", err)
		}
		options.WorkDir = workDir
	}
	absoluteWorkDir, err := filepath.Abs(options.WorkDir)
	if err != nil {
		return nil, fmt.Errorf("resolve working directory %q: %w", options.WorkDir, err)
	}
	if options.Version == "" {
		options.Version = DefaultVersion
	}
	if options.Name == "" {
		options.Name = "purecdk8s"
	}

	return &runner{
		stdin:      options.Stdin,
		stdout:     options.Stdout,
		stderr:     options.Stderr,
		env:        append([]string(nil), options.Env...),
		workDir:    absoluteWorkDir,
		version:    options.Version,
		name:       options.Name,
		httpClient: options.HTTPClient,
	}, nil
}

func (r *runner) run(args []string) error {
	if len(args) == 0 {
		r.writeRootHelp()
		return nil
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--help" || arg == "-h":
			r.writeRootHelp()
			return nil
		case strings.HasPrefix(arg, "--help="):
			value, err := parseBool(strings.TrimPrefix(arg, "--help="))
			if err != nil {
				return fmt.Errorf("Invalid value for --help: %s", strings.TrimPrefix(arg, "--help="))
			}
			if value {
				r.writeRootHelp()
				return nil
			}
		case arg == "--version":
			fmt.Fprintln(r.stdout, r.version)
			return nil
		case strings.HasPrefix(arg, "--version="):
			value, err := parseBool(strings.TrimPrefix(arg, "--version="))
			if err != nil {
				return fmt.Errorf("Invalid value for --version: %s", strings.TrimPrefix(arg, "--version="))
			}
			if value {
				fmt.Fprintln(r.stdout, r.version)
				return nil
			}
		case arg == "--check-upgrade" || arg == "--no-check-upgrade":
			if arg == "--check-upgrade" && i+1 < len(args) {
				if _, err := parseBool(args[i+1]); err == nil {
					i++
				}
			}
		case strings.HasPrefix(arg, "--check-upgrade="):
			if _, err := parseBool(strings.TrimPrefix(arg, "--check-upgrade=")); err != nil {
				return fmt.Errorf("Invalid value for --check-upgrade: %s", strings.TrimPrefix(arg, "--check-upgrade="))
			}
		case arg == "synth" || arg == "synthesize":
			return r.runSynth(args[i+1:])
		case arg == "import" || arg == "generate" || arg == "gen":
			return r.runImport(args[i+1:])
		case arg == "init":
			return r.runInit(args[i+1:])
		case strings.HasPrefix(arg, "-"):
			return fmt.Errorf("Unknown argument: %s", strings.TrimLeft(arg, "-"))
		default:
			return fmt.Errorf("Unknown command: %s", arg)
		}
	}

	r.writeRootHelp()
	return nil
}

func (r *runner) writeRootHelp() {
	fmt.Fprintf(r.stdout, `%s [command]

Commands:
  %s init TYPE  Create a new cdk8s project from a template.
  %s import      Imports API objects by generating native Go constructs.  [aliases: generate, gen]
  %s synth      Synthesizes Kubernetes manifests for all charts in your app.  [aliases: synthesize]

Options:
  --version        Show version number  [boolean]
  --check-upgrade  Check for cdk8s-cli upgrade  [boolean] [default: true]
  --help           Show help  [boolean]

Options can be specified via environment variables with the "CDK8S_" prefix (e.g. "CDK8S_OUTPUT")
`, r.name, r.name, r.name, r.name)
}

func parseBool(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "t", "true", "y", "yes", "on":
		return true, nil
	case "0", "f", "false", "n", "no", "off":
		return false, nil
	default:
		return false, strconv.ErrSyntax
	}
}

func envValue(env []string, names ...string) (string, bool) {
	for i := len(env) - 1; i >= 0; i-- {
		key, value, found := strings.Cut(env[i], "=")
		if !found {
			continue
		}
		for _, name := range names {
			if key == name {
				return value, true
			}
		}
	}
	return "", false
}

func setEnv(env []string, key, value string) []string {
	result := make([]string, 0, len(env)+1)
	for _, entry := range env {
		entryKey, _, found := strings.Cut(entry, "=")
		if found && entryKey == key {
			continue
		}
		result = append(result, entry)
	}
	return append(result, key+"="+value)
}
