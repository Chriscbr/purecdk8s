package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestRootHelpAndVersion(t *testing.T) {
	t.Run("no arguments prints help", func(t *testing.T) {
		code, stdout, stderr := runTestCLI(t, t.TempDir(), nil, nil)
		if code != 0 {
			t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr)
		}
		if !strings.Contains(stdout, "purecdk8s [command]") ||
			!strings.Contains(stdout, "purecdk8s init TYPE") ||
			!strings.Contains(stdout, "purecdk8s synth") ||
			!strings.Contains(stdout, "CDK8S_OUTPUT") {
			t.Fatalf("root help did not contain expected sections:\n%s", stdout)
		}
		if stderr != "" {
			t.Fatalf("stderr = %q, want empty", stderr)
		}
	})

	t.Run("version is a bare version string", func(t *testing.T) {
		code, stdout, stderr := runTestCLI(t, t.TempDir(), []string{"--version"}, nil)
		if code != 0 || stdout != "test-version\n" || stderr != "" {
			t.Fatalf("Run(--version) = code %d, stdout %q, stderr %q", code, stdout, stderr)
		}
	})

	t.Run("synth help does not require app", func(t *testing.T) {
		code, stdout, stderr := runTestCLI(t, t.TempDir(), []string{"synthesize", "--help"}, nil)
		if code != 0 {
			t.Fatalf("Run(synthesize --help) code = %d, want 0; stderr = %q", code, stderr)
		}
		if !strings.Contains(stdout, "purecdk8s synth") ||
			!strings.Contains(stdout, "--app, -a") ||
			!strings.Contains(stdout, "--stdout, -p") {
			t.Fatalf("synth help did not contain expected options:\n%s", stdout)
		}
	})
}

func TestSynthUsesConfigAndCleansOutput(t *testing.T) {
	requireUnixShell(t)
	dir := t.TempDir()
	command := helperCommand(t)
	writeTestFile(t, filepath.Join(dir, ConfigFileName), fmt.Sprintf(
		"app: %s\noutput: configured-dist\n",
		strconv.Quote(command),
	))
	stale := filepath.Join(dir, "configured-dist", "stale.yaml")
	if err := os.MkdirAll(filepath.Dir(stale), 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, stale, "stale: true\n")

	code, stdout, stderr := runTestCLI(t, dir, []string{"synth", "--no-check-upgrade"}, helperEnvironment())
	if code != 0 {
		t.Fatalf("Run(synth) code = %d, want 0; stdout = %q; stderr = %q", code, stdout, stderr)
	}
	wantStdout := "Synthesizing application\n" +
		"  - " + filepath.Join("configured-dist", "z.yaml") + "\n" +
		"  - " + filepath.Join("configured-dist", "sub", "a.yaml") + "\n"
	if stdout != wantStdout {
		t.Fatalf("stdout = %q, want %q", stdout, wantStdout)
	}
	if strings.Contains(stdout, "helper stdout") {
		t.Fatalf("application stdout leaked into CLI stdout: %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale output was not removed; stat error = %v", err)
	}
	if got := readTestFile(t, filepath.Join(dir, "configured-dist", "z.yaml")); got != "kind: Root\n" {
		t.Fatalf("root manifest = %q", got)
	}
}

func TestSynthCLIFlagsOverrideEnvironmentAndConfig(t *testing.T) {
	requireUnixShell(t)
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, ConfigFileName), "app: exit 91\noutput: config-dist\n")

	env := append(helperEnvironment(),
		"CDK8S_APP=exit 92",
		"CDK8S_OUTPUT=env-dist",
	)
	args := []string{"synthesize", "-a", helperCommand(t), "-o", "cli-dist"}
	code, stdout, stderr := runTestCLI(t, dir, args, env)
	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stdout = %q; stderr = %q", code, stdout, stderr)
	}
	if !strings.Contains(stdout, filepath.Join("cli-dist", "z.yaml")) {
		t.Fatalf("stdout = %q, want CLI output path", stdout)
	}
	if _, err := os.Stat(filepath.Join(dir, "cli-dist", "z.yaml")); err != nil {
		t.Fatalf("CLI-selected output was not synthesized: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "env-dist")); !os.IsNotExist(err) {
		t.Fatalf("environment output unexpectedly used; stat error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "config-dist")); !os.IsNotExist(err) {
		t.Fatalf("config output unexpectedly used; stat error = %v", err)
	}
}

func TestSynthEnvironmentOverridesConfig(t *testing.T) {
	requireUnixShell(t)
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, ConfigFileName), "app: exit 91\noutput: config-dist\n")

	env := append(helperEnvironment(),
		"CDK8S_APP="+helperCommand(t),
		"CDK8S_OUTPUT=env-dist",
	)
	code, stdout, stderr := runTestCLI(t, dir, []string{"synth"}, env)
	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stdout = %q; stderr = %q", code, stdout, stderr)
	}
	if !strings.Contains(stdout, filepath.Join("env-dist", "z.yaml")) {
		t.Fatalf("stdout = %q, want environment-selected output path", stdout)
	}
	if _, err := os.Stat(filepath.Join(dir, "env-dist", "z.yaml")); err != nil {
		t.Fatalf("environment-selected output was not synthesized: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "config-dist")); !os.IsNotExist(err) {
		t.Fatalf("config output unexpectedly used; stat error = %v", err)
	}
}

func TestSynthStdoutAndAliases(t *testing.T) {
	requireUnixShell(t)
	dir := t.TempDir()

	code, stdout, stderr := runTestCLI(
		t,
		dir,
		[]string{"synthesize", "-a", helperCommand(t), "-p"},
		helperEnvironment(),
	)
	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr)
	}
	if want := "kind: Root\nkind: Nested\n"; stdout != want {
		t.Fatalf("stdout = %q, want concatenated manifests %q", stdout, want)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if strings.Contains(stdout, "Synthesizing application") || strings.Contains(stdout, "helper stdout") {
		t.Fatalf("stdout mode contained non-manifest output: %q", stdout)
	}
}

func TestSynthStdoutFromEnvironment(t *testing.T) {
	requireUnixShell(t)
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, ConfigFileName), fmt.Sprintf("app: %s\n", strconv.Quote(helperCommand(t))))

	env := append(helperEnvironment(), "CDK8S_STDOUT=true")
	code, stdout, stderr := runTestCLI(t, dir, []string{"synth"}, env)
	if code != 0 || stderr != "" {
		t.Fatalf("Run() = code %d, stderr %q; want success", code, stderr)
	}
	if want := "kind: Root\nkind: Nested\n"; stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
}

func TestSynthConfigOutputRemainsCompatibleWithStdout(t *testing.T) {
	requireUnixShell(t)
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, ConfigFileName), fmt.Sprintf(
		"app: %s\noutput: configured-dist\n",
		strconv.Quote(helperCommand(t)),
	))
	marker := filepath.Join(dir, "configured-dist", "old")
	if err := os.MkdirAll(filepath.Dir(marker), 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, marker, "old")

	code, stdout, stderr := runTestCLI(t, dir, []string{"synth", "--stdout"}, helperEnvironment())
	if code != 0 || stderr != "" {
		t.Fatalf("Run() = code %d, stderr %q; want success", code, stderr)
	}
	if want := "kind: Root\nkind: Nested\n"; stdout != want {
		t.Fatalf("stdout = %q, want %q", stdout, want)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("configured output was not cleaned before stdout synthesis; stat error = %v", err)
	}
}

func TestSynthRejectsExplicitOutputWithStdout(t *testing.T) {
	dir := t.TempDir()
	code, _, stderr := runTestCLI(
		t,
		dir,
		[]string{"synth", "--app", "exit 99", "--stdout", "--output", "other"},
		nil,
	)
	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if !strings.Contains(stderr, "'--output' and '--stdout' are mutually exclusive") {
		t.Fatalf("stderr = %q, want mutual-exclusion error", stderr)
	}
}

func TestSynthRequiresApp(t *testing.T) {
	code, _, stderr := runTestCLI(t, t.TempDir(), []string{"synth"}, nil)
	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stderr != "Missing required argument: app\n" {
		t.Fatalf("stderr = %q", stderr)
	}
}

func TestCLIHelperProcess(t *testing.T) {
	if os.Getenv("PURECDK8S_HELPER_PROCESS") != "1" {
		return
	}
	if os.Getenv("CDK8S_RECORD_CONSTRUCT_METADATA") != "false" {
		fmt.Fprintln(os.Stderr, "unexpected CDK8S_RECORD_CONSTRUCT_METADATA")
		os.Exit(93)
	}
	outdir := os.Getenv("CDK8S_OUTDIR")
	if outdir == "" {
		fmt.Fprintln(os.Stderr, "CDK8S_OUTDIR is empty")
		os.Exit(94)
	}
	if err := os.MkdirAll(filepath.Join(outdir, "sub"), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(95)
	}
	if err := os.WriteFile(filepath.Join(outdir, "z.yaml"), []byte("kind: Root\n"), 0o600); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(96)
	}
	if err := os.WriteFile(filepath.Join(outdir, "sub", "a.yaml"), []byte("kind: Nested\n"), 0o600); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(97)
	}
	if err := os.WriteFile(filepath.Join(outdir, "ignored.yml"), []byte("ignored: true\n"), 0o600); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(98)
	}
	fmt.Fprintln(os.Stdout, "helper stdout")
	os.Exit(0)
}

func runTestCLI(t *testing.T, workDir string, args, extraEnv []string) (code int, stdout, stderr string) {
	t.Helper()
	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer

	env := testBaseEnvironment()
	env = append(env, extraEnv...)
	code = Run(args, Options{
		Stdout:  &stdoutBuffer,
		Stderr:  &stderrBuffer,
		Env:     env,
		WorkDir: workDir,
		Version: "test-version",
		Name:    "purecdk8s",
	})
	return code, stdoutBuffer.String(), stderrBuffer.String()
}

func testBaseEnvironment() []string {
	var env []string
	for _, entry := range os.Environ() {
		key, _, found := strings.Cut(entry, "=")
		if found && (strings.HasPrefix(key, "CDK8S_") || key == "PURECDK8S_HELPER_PROCESS") {
			continue
		}
		env = append(env, entry)
	}
	return env
}

func helperEnvironment() []string {
	return []string{"PURECDK8S_HELPER_PROCESS=1"}
}

func helperCommand(t *testing.T) string {
	t.Helper()
	absolute, err := filepath.Abs(os.Args[0])
	if err != nil {
		t.Fatalf("resolve test binary: %v", err)
	}
	return fmt.Sprintf("%q -test.run=^TestCLIHelperProcess$", absolute)
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func requireUnixShell(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("helper command quoting in this test uses a POSIX shell")
	}
}
