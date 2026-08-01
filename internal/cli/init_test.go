package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

func TestInitHelpAndVersion(t *testing.T) {
	t.Run("help lists native Go templates", func(t *testing.T) {
		code, stdout, stderr := runTestCLI(t, t.TempDir(), []string{"init", "--help"}, nil)
		if code != 0 || stderr != "" {
			t.Fatalf("Run(init --help) = code %d, stderr %q; want success", code, stderr)
		}
		for _, template := range initTemplateNames {
			if !strings.Contains(stdout, `"`+template+`"`) {
				t.Fatalf("init help does not list %q:\n%s", template, stdout)
			}
		}
	})

	t.Run("version works without a template", func(t *testing.T) {
		code, stdout, stderr := runTestCLI(t, t.TempDir(), []string{"init", "--version"}, nil)
		if code != 0 || stdout != "test-version\n" || stderr != "" {
			t.Fatalf("Run(init --version) = code %d, stdout %q, stderr %q", code, stdout, stderr)
		}
	})
}

func TestInitGoTemplates(t *testing.T) {
	tests := []struct {
		name            string
		template        string
		app             bool
		private         bool
		expectedGoFiles []string
	}{
		{
			name:            "go alias",
			template:        initTemplateGo,
			app:             true,
			expectedGoFiles: []string{"main.go"},
		},
		{
			name:            "canonical go app",
			template:        initTemplateGoApp,
			app:             true,
			expectedGoFiles: []string{"main.go"},
		},
		{
			name:            "go library alias",
			template:        initTemplateGoLibrary,
			expectedGoFiles: []string{"chart.go", "chart_test.go"},
		},
		{
			name:            "public go library",
			template:        initTemplateGoLibraryPublic,
			expectedGoFiles: []string{"chart.go", "chart_test.go"},
		},
		{
			name:            "private go library",
			template:        initTemplateGoLibraryPrivate,
			private:         true,
			expectedGoFiles: []string{"chart.go", "chart_test.go"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			projectDir := filepath.Join(t.TempDir(), "my-project")
			if err := os.Mkdir(projectDir, 0o755); err != nil {
				t.Fatal(err)
			}

			code, stdout, stderr := runTestCLI(
				t,
				projectDir,
				[]string{"init", test.template, "--no-check-upgrade"},
				nil,
			)
			if code != 0 {
				t.Fatalf("Run(init %s) code = %d, want 0; stdout = %q; stderr = %q", test.template, code, stdout, stderr)
			}
			if stdout != "" {
				t.Fatalf("stdout = %q, want empty", stdout)
			}
			wantStatus := "Initializing a project from the " + test.template + " template\n"
			if stderr != wantStatus {
				t.Fatalf("stderr = %q, want %q", stderr, wantStatus)
			}

			expectedFiles := append([]string{ConfigFileName, "README.md", "go.mod", "help"}, test.expectedGoFiles...)
			sort.Strings(expectedFiles)
			if got := directoryEntryNames(t, projectDir); strings.Join(got, ",") != strings.Join(expectedFiles, ",") {
				t.Fatalf("generated files = %v, want %v", got, expectedFiles)
			}

			config, err := ReadConfig(projectDir)
			if err != nil {
				t.Fatalf("ReadConfig() error = %v", err)
			}
			assertStringPointer(t, "Language", config.Language, "go")
			if got := strings.Join(config.Imports, ","); got != "k8s" {
				t.Fatalf("Imports = %q, want k8s", got)
			}
			if test.app {
				assertStringPointer(t, "App", config.App, "go run .")
			} else if config.App != nil {
				t.Fatalf("library App = %q, want omitted", *config.App)
			}

			goMod := readTestFile(t, filepath.Join(projectDir, "go.mod"))
			if !strings.Contains(goMod, "module example.com/my-project") ||
				!strings.Contains(goMod, "require "+purecdk8sModule+" "+purecdk8sInitialVersion) {
				t.Fatalf("go.mod does not contain expected native module requirement:\n%s", goMod)
			}
			assertNoLegacyRuntimeReferences(t, projectDir)

			readme := readTestFile(t, filepath.Join(projectDir, "README.md"))
			if !strings.Contains(readme, "replace "+purecdk8sModule+" => /path/to/purecdk8s") {
				t.Fatalf("README is missing source-checkout replace guidance:\n%s", readme)
			}
			if !test.app {
				wantVisibility := "Visibility: Public"
				wantHelpVisibility := "Visibility: public"
				if test.private {
					wantVisibility = "Visibility: Private"
					wantHelpVisibility = "Visibility: private"
				}
				if !strings.Contains(readme, wantVisibility) {
					t.Fatalf("README does not contain %q:\n%s", wantVisibility, readme)
				}
				help := readTestFile(t, filepath.Join(projectDir, "help"))
				if !strings.Contains(help, wantHelpVisibility) {
					t.Fatalf("help does not contain %q:\n%s", wantHelpVisibility, help)
				}
				if test.private && !strings.Contains(readme, "GOPRIVATE=example.com/my-project") {
					t.Fatalf("private README is missing GOPRIVATE guidance:\n%s", readme)
				}
			}

			compileGeneratedProject(t, projectDir)
		})
	}
}

func TestInitDirectoryRules(t *testing.T) {
	t.Run("visible entries are rejected", func(t *testing.T) {
		projectDir := t.TempDir()
		writeTestFile(t, filepath.Join(projectDir, "existing.txt"), "keep me")

		code, stdout, stderr := runTestCLI(t, projectDir, []string{"init", initTemplateGoApp}, nil)
		if code != 1 || stdout != "" {
			t.Fatalf("Run() = code %d, stdout %q; want code 1 and empty stdout", code, stdout)
		}
		if stderr != "Cannot initialize a project in a non-empty directory\n" {
			t.Fatalf("stderr = %q", stderr)
		}
		if got := directoryEntryNames(t, projectDir); strings.Join(got, ",") != "existing.txt" {
			t.Fatalf("directory changed after rejected init: %v", got)
		}
	})

	t.Run("hidden entries are preserved and ignored", func(t *testing.T) {
		projectDir := t.TempDir()
		if err := os.Mkdir(filepath.Join(projectDir, ".git"), 0o755); err != nil {
			t.Fatal(err)
		}
		writeTestFile(t, filepath.Join(projectDir, ".keep"), "keep me")

		code, _, stderr := runTestCLI(t, projectDir, []string{"init", initTemplateGo}, nil)
		if code != 0 {
			t.Fatalf("Run() code = %d, stderr = %q; want success", code, stderr)
		}
		if got := readTestFile(t, filepath.Join(projectDir, ".keep")); got != "keep me" {
			t.Fatalf("hidden file contents = %q, want preserved", got)
		}
		if _, err := os.Stat(filepath.Join(projectDir, ".git")); err != nil {
			t.Fatalf("hidden directory was not preserved: %v", err)
		}
	})
}

func TestInitArgumentErrors(t *testing.T) {
	t.Run("missing template", func(t *testing.T) {
		code, _, stderr := runTestCLI(t, t.TempDir(), []string{"init"}, nil)
		if code != 1 || !strings.Contains(stderr, "need at least 1") {
			t.Fatalf("Run(init) = code %d, stderr %q", code, stderr)
		}
	})

	t.Run("unknown template includes choices", func(t *testing.T) {
		code, _, stderr := runTestCLI(t, t.TempDir(), []string{"init", "rust"}, nil)
		if code != 1 {
			t.Fatalf("Run(init rust) code = %d, want 1", code)
		}
		if !strings.Contains(stderr, `Given: "rust"`) {
			t.Fatalf("stderr = %q, want invalid template", stderr)
		}
		for _, template := range initTemplateNames {
			if !strings.Contains(stderr, `"`+template+`"`) {
				t.Fatalf("stderr does not list %q: %s", template, stderr)
			}
		}
	})
}

func directoryEntryNames(t *testing.T, directory string) []string {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("read directory %s: %v", directory, err)
	}
	names := make([]string, len(entries))
	for index, entry := range entries {
		names[index] = entry.Name()
	}
	sort.Strings(names)
	return names
}

func assertNoLegacyRuntimeReferences(t *testing.T, directory string) {
	t.Helper()
	for _, name := range directoryEntryNames(t, directory) {
		path := filepath.Join(directory, name)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.IsDir() {
			continue
		}
		contents := readTestFile(t, path)
		for _, forbidden := range []string{
			"github.com/aws/constructs-go",
			"github.com/aws/jsii-runtime-go",
			"github.com/cdk8s-team/",
		} {
			if strings.Contains(contents, forbidden) {
				t.Fatalf("%s contains legacy dependency %q:\n%s", name, forbidden, contents)
			}
		}
	}
}

func compileGeneratedProject(t *testing.T, projectDir string) {
	t.Helper()
	moduleRoot := purecdk8sTestModuleRoot(t)
	goModPath := filepath.Join(projectDir, "go.mod")
	goMod := readTestFile(t, goModPath)
	goMod += "\nreplace " + purecdk8sModule + " => " + moduleRoot + "\n"
	writeTestFile(t, goModPath, goMod)

	command := exec.Command("go", "test", "./...")
	command.Dir = projectDir
	command.Env = append(filteredEnvironment("GOFLAGS", "GOWORK"), "GOFLAGS=-mod=mod", "GOWORK=off")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("generated project did not compile: %v\n%s", err, output)
	}
}

func purecdk8sTestModuleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller could not locate init_test.go")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func filteredEnvironment(keys ...string) []string {
	excluded := make(map[string]bool, len(keys))
	for _, key := range keys {
		excluded[key] = true
	}
	var result []string
	for _, entry := range os.Environ() {
		key, _, found := strings.Cut(entry, "=")
		if !found || !excluded[key] {
			result = append(result, entry)
		}
	}
	return result
}
