# purecdk8s

`purecdk8s` is a native-Go implementation of cdk8s. It provides the cdk8s
construct model, synthesis APIs, generated Kubernetes and CRD bindings, and a
`cdk8s`-compatible command-line workflow without starting the JSII kernel or a
Node.js process.

The compatibility target is
[`cdk8s-core-go/cdk8s/v2` v2.70.85](https://pkg.go.dev/github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2@v2.70.85).
The package preserves its pointer-oriented Go API so an existing application
can normally migrate by changing imports and regenerating imported APIs.

## What is native

- The construct tree, charts, API objects, dependency ordering, name
  generation, lazy values, JSON patches, YAML, durations, sizes, and cron
  helpers are ordinary Go.
- Kubernetes and CustomResourceDefinition bindings are generated as ordinary
  Go structs and constructors.
- `synth`, `init`, and `import` are implemented in Go.
- `purecdk8s/jsii` contains source-compatible pointer helpers such as
  `jsii.String` and `jsii.Number`; it is not the JSII runtime.

The module has no dependency on `github.com/aws/jsii-runtime-go`, the JSII
kernel, npm, or Node.js. Its only third-party runtime module is
`gopkg.in/yaml.v3`.

## Install

From a source checkout:

```console
cd /path/to/purecdk8s
go test ./...
go install ./cmd/purecdk8s
```

`purecdk8s` is the native command name for the compatible CLI workflow.

To build standalone binaries instead:

```console
go build -o ./bin/purecdk8s ./cmd/purecdk8s
```

Once a tagged module version is available, the equivalent remote installation
is:

```console
go install github.com/Chriscbr/purecdk8s/cmd/purecdk8s@latest
```

The library requires Go 1.23 or newer.

## Import compatibility

These are the core import-path changes for application code:

| Upstream import | Native replacement |
| --- | --- |
| `github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2` | `github.com/Chriscbr/purecdk8s/cdk8s/v2` |
| `github.com/aws/constructs-go/constructs/v10` | `github.com/Chriscbr/purecdk8s/constructs/v10` |
| `github.com/aws/jsii-runtime-go` | `github.com/Chriscbr/purecdk8s/jsii` |

Native cdk8s+ packages use the same versioned import convention:

| Upstream import | Native replacement |
| --- | --- |
| `github.com/cdk8s-team/cdk8s-plus-go/cdk8splus34/v2` | `github.com/Chriscbr/purecdk8s/cdk8splus34/v2` |
| `github.com/cdk8s-team/cdk8s-plus-go/cdk8splus35/v2` | `github.com/Chriscbr/purecdk8s/cdk8splus35/v2` |
| `github.com/cdk8s-team/cdk8s-plus-go/cdk8splus36/v2` | `github.com/Chriscbr/purecdk8s/cdk8splus36/v2` |

The replacement helper package is still named `jsii`, so calls such as
`jsii.String("default")` and `jsii.Number(3)` do not need to change.

## cdk8s+

The native cdk8s+ implementation includes Kubernetes-specific low-level
bindings under each of `cdk8splus34/v2/k8s`, `cdk8splus35/v2/k8s`, and
`cdk8splus36/v2/k8s`. The high-level layer currently provides native
ConfigMap, Secret, Namespace, ServiceAccount, Volume, Container, Pod, Job,
CronJob, Deployment, Service, and Ingress constructs, including deployment
exposure and lazy synthesis. Kubernetes 1.35 and 1.36
are forward ports of the 1.34 high-level behavior because upstream cdk8s+ has
not yet published corresponding reference packages.

A typical native app looks like this:

```go
package main

import (
	"example.com/my-app/imports/k8s"
	"github.com/Chriscbr/purecdk8s/cdk8s/v2"
	"github.com/Chriscbr/purecdk8s/constructs/v10"
	"github.com/Chriscbr/purecdk8s/jsii"
)

func newChart(scope constructs.Construct, id string) cdk8s.Chart {
	chart := cdk8s.NewChart(scope, jsii.String(id), &cdk8s.ChartProps{
		Namespace: jsii.String("default"),
	})

	k8s.NewKubeConfigMap(chart, jsii.String("settings"), &k8s.KubeConfigMapProps{
		Data: &map[string]*string{
			"mode": jsii.String("native-go"),
		},
	})
	return chart
}

func main() {
	app := cdk8s.NewApp(nil)
	newChart(app, "example")
	app.Synth()
}
```

## Start a project

Create an empty directory and initialize a native Go application:

```console
mkdir hello-cdk8s
cd hello-cdk8s
purecdk8s init go-app
purecdk8s import
go mod tidy
purecdk8s synth
```

Supported Go templates are `go`, `go-app`, `go-library`,
`go-library-public`, and `go-library-private`. `go` is an alias for
`go-app`, while `go-library` is an alias for `go-library-public`.

An app project uses the familiar `cdk8s.yaml` format:

```yaml
language: go
app: go run .
output: dist
importDirectory: imports
imports:
  - k8s
```

If the module has not been published at the version used by the generated
`go.mod`, point the project at a local checkout:

```console
go mod edit -replace github.com/Chriscbr/purecdk8s=/absolute/path/to/purecdk8s
go mod tidy
```

## Import Kubernetes and CRD APIs

With no positional arguments, `import` processes the entries in
`cdk8s.yaml`:

```console
purecdk8s import
```

`k8s` imports the default Kubernetes schema, v1.25.0. Pin a schema explicitly
when reproducibility matters:

```console
purecdk8s import k8s@1.25.0
```

Local files, HTTP(S) documents, and `github:` CustomResourceDefinition
shorthand are also supported:

```console
purecdk8s import ./crds/widgets.yaml
purecdk8s import https://example.com/crds/widgets.yaml
purecdk8s import github:crossplane/crossplane@0.14.0
```

Use upstream-compatible `NAME:=SPEC` syntax to name an import:

```console
purecdk8s import widgets:=./crds/widgets.yaml
```

Useful import options include:

- `-o, --output DIR` to select the import directory;
- `-l, --language go` to select Go generation;
- `--class-prefix PREFIX` or `--no-class-prefix` for generated class naming;
- `--exclude REGEXP` to replace matching Kubernetes `$ref` types with
  `interface{}`; this option is repeatable; and
- `--save` or `--no-save` to control whether the import is recorded in
  `cdk8s.yaml`.

Kubernetes imports generate the familiar resource constructors
(`NewKubeDeployment`, `NewKubeService`, and so on), props structs, nested API
types, `*_Manifest`, `*_GVK`, `*_Of`, and type-guard functions. Scalar unions
such as Kubernetes `IntOrString` values retain their `FromNumber` and
`FromString` constructors.

CRDs are generated from `spec.versions[*].schema.openAPIV3Schema`, with a
resource constructor for each served version and ordinary Go packages based
on the API group. Both `apiextensions.k8s.io/v1` and legacy v1beta1-style CRD
documents are accepted.

Generated imports from the upstream cdk8s CLI cannot simply be retained:
those files contain JSII proxies and imports. Regenerate them with
`purecdk8s import` so the generated package is native Go. The importer removes
and recreates only each target package directory, so stale JSII proxy files do
not survive while sibling import packages remain untouched.

Helm imports support local charts, HTTP chart repositories, and OCI
registries:

```console
purecdk8s import helm:./charts/webapp
purecdk8s import helm:https://charts.bitnami.com/bitnami/mysql@9.10.10
purecdk8s import helm:oci://registry-1.docker.io/bitnamicharts/wordpress@17.1.17
```

HTTP and OCI imports use `helm pull` to acquire the schema at generation time.
The generated Go package retains the original runtime chart reference and does
not embed the downloaded archive.

## Synthesize

The CLI reads `cdk8s.yaml`, runs the configured app with `CDK8S_OUTDIR`, and
writes manifests to `output` (default: `dist`):

```console
purecdk8s synth
purecdk8s synth --output manifests
purecdk8s synth --stdout
purecdk8s synth --app "go run ./cmd/app"
purecdk8s synth --format helm --chart-version 1.0.0
```

The `synthesize` alias and the upstream short options `-a`, `-o`, and `-p`
are supported. CLI options take precedence over `CDK8S_*` environment
variables, which take precedence over `cdk8s.yaml`.

Helm-format synthesis writes `Chart.yaml`, `README.md`, and synthesized
manifests under `templates/`. It uses Helm chart API v2 by default; select v1
or v2 with `--chart-api-version`. A SemVer-compatible `--chart-version` is
required. With API v2, imported CRD documents are copied into `crds/`.

## Migrate an existing cdk8s Go app

1. Replace the three imports shown in
   [Import compatibility](#import-compatibility). Do this in handwritten
   source, not in the existing generated import tree.
2. Regenerate the imported APIs. The native importer safely replaces each
   generated target package, including its old JSII proxy files:

   ```console
   purecdk8s import --output imports
   ```

   If you want a backup, copy it outside the Go module first so
   `go test ./...` does not try to compile the old proxy packages.

3. Add the native module. For a local checkout:

   ```console
   go mod edit -replace github.com/Chriscbr/purecdk8s=/absolute/path/to/purecdk8s
   go get github.com/Chriscbr/purecdk8s@v0.0.0
   go mod tidy
   ```

4. Verify source compatibility and synthesis:

   ```console
   go test ./...
   purecdk8s synth
   ```

After `go mod tidy`, the upstream core, constructs, and JSII runtime modules
should disappear from the application's module graph.

## Include and Helm

`Include` reads a local or HTTP(S) multi-document YAML stream and adds every
manifest to a chart:

```go
cdk8s.NewInclude(chart, jsii.String("base"), &cdk8s.IncludeProps{
	Url: jsii.String("./manifests/base.yaml"),
})
```

`Helm` has the cdk8s v2.70.85 API, including values, repository, version,
namespace, release-name, custom executable, and additional flags:

```go
cdk8s.NewHelm(chart, jsii.String("nginx"), &cdk8s.HelmProps{
	Chart:   jsii.String("bitnami/nginx"),
	Version: jsii.String("18.2.4"),
	Values: &map[string]interface{}{
		"service": map[string]interface{}{"type": "ClusterIP"},
	},
})
```

This library feature shells out to `helm template`, just as its contract
requires, so the Helm executable must be installed when `NewHelm` is used.
It does not require Node.js.

## Test the implementation

Run unit tests with:

```console
make test
```

The unit tests cover the constructs API, cdk8s core API, cdk8s+ API, and purecdk8s CLI.

Run an integration test with:

```console
make integration-example EXAMPLE=helm
```

Run every integration test with:

```console
make integration
```

Use `VERBOSE=1` for more detailed output.

Add example cdk8s Go code to `integration/examples` to add another test case.
The integration test uses Docker to run the test case with both the original cdk8s Go
implementation and the pure cdk8s Go implementation, and compares the output byte-for-byte.

To inspect output manually, you can generate the outputs for either implementation (or both):

```console
./integration/run.sh helm upstream
./integration/run.sh helm pure
./integration/run.sh helm both
```

The last command also compares the two trees and prints the temporary output
directory instead of removing it.

The remaining local checks are available through `make`:

```console
make format
make format-check
make build
make api
make test
```

`make test` runs unit tests, API checks, and integration tests.
`make api` checks the exported types and Go doc comments in constructs, cdk8s,
and cdk8s+ against their upstream packages.

## Compatibility notes

- JavaScript validation plugins declared under `validations` in
  `cdk8s.yaml` cannot run in a no-JavaScript runtime. The native CLI reports
  this explicitly; use `purecdk8s synth --no-validate` or replace the plugin
  with native validation.
- Direct calls to JSII kernel APIs such as `Create`, `Invoke`, `UnsafeCast`,
  proxy registration, or cross-language dispatch are intentionally absent.
  Ordinary cdk8s Go applications using the normal constructs APIs should not
  be affected.
- `cdk8s.NewHelm` runs by shelling out to the system's `helm` executable. That
  executable is also used for HTTP and OCI Helm imports. It's an optional
  feature dependency, so `purecdk8s` doesn't automatically download it.
- Kubernetes schema import downloads the selected schema, and HTTP(S) CRD
  and remote Helm imports download the requested inputs. These operations
  require network access; non-Helm synthesis itself does not.
- Go maps do not retain insertion order. Default deterministic synthesis
  matches cdk8s ordering, but the `CDK8S_DISABLE_SORT` escape hatch cannot
  reproduce JavaScript object insertion order.
- There is still a helper package named `jsii` for source code compatibility
  so you can continue to use `jsii.String` and `jsii.Number` in your code.
  To be clear, there isn't an actual JSII or Node.js runtime used.
- The implementation only takes a single dependency on `gopkg.in/yaml.v3`.
