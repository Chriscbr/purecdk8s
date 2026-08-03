# purecdk8s

`purecdk8s` is a native-Go implementation of cdk8s.
It aims to be a drop-in replacement for the original set of cdk8s Go packages like `github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2` and `github.com/cdk8s-team/cdk8s-plus-go/cdk8splus34/v2`, but without the JSII runtime and Node.js.
The `purecdk8s` CLI is a drop-in replacement for the original `cdk8s` CLI, and all of the Go packages expose the same APIs as the original packages.

The compatibility guarantee is kept in check by a multi-layered testing approach:

- Unit tests cover the constructs API, cdk8s core API, cdk8s+ API, and purecdk8s CLI, based on the original cdk8s tests.
- Integration tests that use Docker to build and synthesize an application with both the original cdk8s Go implementation and the purecdk8s Go implementation, and compare the output byte-for-byte.
- Automated API checks that compare the exported types and Go doc comments in constructs, cdk8s, and cdk8s+ against their upstream packages.


The module has no dependency on `github.com/aws/jsii-runtime-go`, npm, or Node.js. Its only third-party dependency is `gopkg.in/yaml.v3`.

## Install

```console
go install github.com/Chriscbr/purecdk8s/cmd/purecdk8s@latest
```

## Import compatibility

These are the main import-path changes between the original cdk8s and the purecdk8s implementation.

| Original import | Replacement import |
| --- | --- |
| `github.com/cdk8s-team/cdk8s-core-go/cdk8s/v2` | `github.com/Chriscbr/purecdk8s/cdk8s/v2` |
| `github.com/aws/constructs-go/constructs/v10` | `github.com/Chriscbr/purecdk8s/constructs/v10` |
| `github.com/aws/jsii-runtime-go` | `github.com/Chriscbr/purecdk8s/jsii` |
| `github.com/cdk8s-team/cdk8s-plus-go/cdk8splus34/v2` | `github.com/Chriscbr/purecdk8s/cdk8splus34/v2` |
| `github.com/cdk8s-team/cdk8s-plus-go/cdk8splus35/v2` | `github.com/Chriscbr/purecdk8s/cdk8splus35/v2` |
| `github.com/cdk8s-team/cdk8s-plus-go/cdk8splus36/v2` | `github.com/Chriscbr/purecdk8s/cdk8splus36/v2` |

Note that even though there is a `jsii` package, there is no actual JSII runtime or Node.js used.
The `jsii` package is only there to help you migrate your code, so calls like `jsii.String("default")` and `jsii.Number(3)` do not need to change.

## Start a new project

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

We recommend consulting the original cdk8s documentation for more information on how to use the library and its CLI.

## Migrate an existing cdk8s Go app

1. Replace the imports shown in
   [Import compatibility](#import-compatibility). Do this in handwritten
   source, not in the existing generated import tree.
2. Regenerate the imported APIs. The native importer safely replaces each
   generated target package:

   ```console
   purecdk8s import --output imports
   ```

3. Add the purecdk8s module:

   ```console
   go get github.com/Chriscbr/purecdk8s@latest
   go mod tidy
   ```

4. Verify source compatibility and synthesis:

   ```console
   go test ./...
   purecdk8s synth
   ```

After `go mod tidy`, the upstream core, constructs, and JSII runtime modules
should disappear from the application's module graph.

## Development

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
make vet
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
  feature dependency - `purecdk8s` doesn't automatically download it.
- Go maps do not retain insertion order. Default deterministic synthesis
  matches cdk8s ordering, but the `CDK8S_DISABLE_SORT` escape hatch cannot
  reproduce JavaScript object insertion order.
