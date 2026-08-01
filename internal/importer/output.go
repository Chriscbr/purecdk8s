package importer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// replacePackageDirectory prepares one generated import package. Import
// directories are generated artifacts, so replacing the exact package
// directory prevents stale upstream JSII proxy files from remaining beside
// purecdk8s output.
func replacePackageDirectory(outputDir, packageName string) (string, error) {
	output, err := filepath.Abs(outputDir)
	if err != nil {
		return "", fmt.Errorf("resolve import output directory: %w", err)
	}
	directory, err := filepath.Abs(filepath.Join(output, packageName))
	if err != nil {
		return "", fmt.Errorf("resolve generated package directory: %w", err)
	}
	relative, err := filepath.Rel(output, directory)
	if err != nil {
		return "", fmt.Errorf("validate generated package directory: %w", err)
	}
	if relative == "." || filepath.IsAbs(relative) || filepath.Dir(relative) != "." ||
		relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("refusing to replace generated package outside one direct child of %s: %s", output, directory)
	}
	if err := os.RemoveAll(directory); err != nil {
		return "", fmt.Errorf("replace generated package %s: %w", directory, err)
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return "", fmt.Errorf("create generated package %s: %w", directory, err)
	}
	return directory, nil
}
