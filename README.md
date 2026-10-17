# Licenses tool

> This is not an officially supported Google product.

`go-licenses` analyzes the dependency tree of a Go package/binary. It can check
that the libraries used are under licenses a policy allows, and collect all of
the license documents, copyright notices and source code into a directory in
order to comply with license terms on redistribution.

## Before you start

To use this tool, make sure:

* [You have Go v1.16 or later installed](https://golang.org/dl/).
* Change directory to your go project, **for example**:

  ```shell
  git clone git@github.com:forkcloser/go-licenses.git
  cd go-licenses
  ```

* Download required modules:

  ```shell
  go mod download
  ```

## Installation

Use the following command to download and install this tool:

```shell
go install github.com/forkcloser/go-licenses/cmd/go-licenses@latest
```

If you were using `go get` to install this tool, note that
[starting in Go 1.17, go get is deprecated for installing binaries](https://go.dev/doc/go-get-install-deprecation).

## Save licenses, copyright notices and source code (depending on license type)

```shell
go-licenses save "github.com/forkcloser/go-licenses/cmd/go-licenses" --save_path="/tmp/go-licenses-cli"
```

This command analyzes a binary/package's dependencies and determines what needs
to be redistributed alongside that binary/package in order to comply with the
license terms. This typically includes the license itself and a copyright
notice, but may also include the dependency's source code. All of the required
artifacts will be saved in the directory indicated by `--save_path`.

## Checking for forbidden licenses

```shell
$ go-licenses check github.com/logrusorgru/aurora
Forbidden license type WTFPL for library github.com/logrusorgru/auroraexit status 1
```

This command analyzes a package's dependencies and determines if any are
considered forbidden by the license classifer. See
[github.com/google/licenseclassifier](https://github.com/google/licenseclassifier/blob/842c0d70d7027215932deb13801890992c9ba364/license_type.go#L323)
for licenses considered forbidden.

## Usages

### Global
Typically, specify the Go package that builds your Go binary.
go-licenses expects the same package argument format as `go build`.  For examples:

* A rooted import path like `github.com/forkcloser/go-licenses/cmd/go-licenses` or `github.com/forkcloser/go-licenses/licenses`.
* A relative path that denotes the package in that directory, like `.` or `./cmd/some-command`.

To learn more about package argument, run `go help packages`.

To learn more about go-licenses usages, run `go-licenses --help`, or `go-licenses <command> --help`.

### Save

Save licenses, copyright notices and source code (depending on license type):

```shell
go-licenses save <package> [package...] --save_path=<save_path>
```

### Check

Checking for forbidden and unknown licenses usage:

```shell
go-licenses check <package> [package...]
```

**Tip**: Usually you'll want to

* append `/...` to the end of an import path prefix (e.g., your repo path) to include all packages matching that pattern
* add `--include_tests` to also check packages only imported by testing code (e.g., testing libraries/frameworks)

```shell
go-licenses check --include_tests github.com/forkcloser/go-licenses/...
```

Checking for disallowed license types:

```shell
go-licenses check <package> [package...] --disallowed_types=<comma separated license types> 
```

Supported license types:

* See `forbidden` list: [github.com/google/licenseclassifier](https://github.com/google/licenseclassifier/blob/e6a9bb99b5a6f71d5a34336b8245e305f5430f99/license_type.go#L341)
* See `notice` list:  [github.com/google/licenseclassifier](https://github.com/google/licenseclassifier/blob/e6a9bb99b5a6f71d5a34336b8245e305f5430f99/license_type.go#L249)
* See `permissive` list:  [github.com/google/licenseclassifier](https://github.com/google/licenseclassifier/blob/e6a9bb99b5a6f71d5a34336b8245e305f5430f99/license_type.go#L321)
* See `reciprocal` list:  [github.com/google/licenseclassifier](https://github.com/google/licenseclassifier/blob/e6a9bb99b5a6f71d5a34336b8245e305f5430f99/license_type.go#L225)
* See `restricted` list:  [github.com/google/licenseclassifier](https://github.com/google/licenseclassifier/blob/e6a9bb99b5a6f71d5a34336b8245e305f5430f99/license_type.go#L185)
* See `unencumbered` list:  [github.com/google/licenseclassifier](https://github.com/google/licenseclassifier/blob/e6a9bb99b5a6f71d5a34336b8245e305f5430f99/license_type.go#L324)
* `unknown`

Allow only specific license names:

```shell
go-licenses check <package> [package...] --allowed_licenses=<comma separated license names> 
```

* See supported license names: [github.com/google/licenseclassifier](https://github.com/google/licenseclassifier/blob/e6a9bb99b5a6f71d5a34336b8245e305f5430f99/license_type.go#L28)

### Build tags

To read dependencies from packages with
[build tags](https://golang.org/pkg/go/build/#hdr-Build_Constraints). Use the
`$GOFLAGS` environment variable.

```shell
$ GOFLAGS="-tags=tools" go-licenses check google.golang.org/grpc/test/tools
```

### Ignoring packages

Use the `--ignore` global flag to specify package path prefixes to be ignored.
For example, to ignore your organization's internal packages under `github.com/example-corporation`:

```shell
$ go-licenses check \
    github.com/example-corporation/example-product \
    --ignore github.com/example-corporation
```

Note that dependencies from the ignored packages are still resolved and checked.
This flag makes effect to the `check` and `save` commands.

### Include testing packages

Use the `--include_tests` global flag to include packages only imported by testing code (e.g., testing libraries/frameworks).
Example command:

```shell
go-licenses check --include_tests "github.com/forkcloser/go-licenses/..."
```

This flag makes effect to the `check` and `save` commands.

## Warnings and errors

The tool will log warnings and errors in some scenarios. This section provides
guidance on addressing them.

Warnings are logged to stderr, colored on a terminal and as JSON otherwise;
`LOG_LEVEL` (`debug`, `info`, `warn`, `error`) sets how much is shown.

### Dependency contains non-Go code

A warning will be logged when a dependency contains non-Go code. This is because
it is not possible to check the non-Go code for further dependencies, which may
conceal additional license requirements. You should investigate this code to
determine whether it has dependencies and take action to comply with their
license terms.
