// Copyright 2019 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/mycophonic/primordium/app/logger"
	"github.com/mycophonic/primordium/app/shutdown"
)

// The help texts are upstream's, laid out in paragraphs: kong reflows the
// lines within one.
const (
	description = `go-licenses helps you work with licenses of your go project's dependencies.

Prerequisites: Go v1.16 or later; run it in your go project, after "go mod download".`

	packageHelp = `Typically, specify the Go package that builds your Go binary. go-licenses expects the same package argument format as "go build". For example:

- A rooted import path like "github.com/forkcloser/go-licenses/cmd/go-licenses" or "github.com/forkcloser/go-licenses/licenses".

- A relative path that denotes the package in that directory, like "." or "./cmd/some-command".

To learn more about Go package argument, run "go help packages".`
)

// cli is the command line: the flags check and save share, and the two
// commands. The flag names are upstream's, underscores included.
type cli struct {
	IncludeTests bool     `help:"Include packages only imported by testing code."                                                                                 name:"include_tests"`
	Ignore       []string `help:"Package path prefixes to be ignored. Dependencies from the ignored packages are still checked. Can be specified multiple times."`

	Check checkCmd `cmd:"" help:"Checks whether licenses for a package are not allowed."`
	Save  saveCmd  `cmd:"" help:"Saves licenses, copyright notices and source code, as required by a Go package's dependencies, to a directory."`
}

func main() {
	// slog through primordium (LOG_LEVEL; tint on a terminal, JSON otherwise),
	// and a context the first shutdown signal cancels. Run exits with the status
	// an ExitError asks for, silently: a failed check has printed its findings.
	ctx := context.Background()
	logger.SetDefaultsForLogger(ctx)

	shutdown.Run(shutdown.SetDefaults(ctx), func(ctx context.Context) error {
		return run(ctx, os.Args[1:])
	})
}

// run parses args and runs the command. Parse errors are kong's, printed with
// the usage; every failure exits 1, as it always has, and --help exits 0.
func run(ctx context.Context, args []string) error {
	var root cli

	parser, err := kong.New(&root,
		kong.Name("go-licenses"),
		kong.Description(description),
		kong.UsageOnError(),
		kong.Exit(func(code int) {
			if code != 0 {
				code = 1
			}

			shutdown.Exit(code)
		}),
		kong.BindTo(ctx, (*context.Context)(nil)),
	)
	if err != nil {
		return fmt.Errorf("building the command line: %w", err)
	}

	kctx, err := parser.Parse(args)
	parser.FatalIfErrorf(err)

	switch err := kctx.Run(&root); {
	case err == nil:
		return nil
	case errors.Is(err, errDisallowed):
		return errDisallowed
	default:
		// A failure is a message to the user, as the findings are, not a log record.
		fmt.Fprintln(os.Stderr, "go-licenses:", err)

		return shutdown.ExitError(1)
	}
}

// Unvendor removes the "*/vendor/" prefix from the given import path, if present.
func unvendor(importPath string) string {
	if vendorerAndVendoree := strings.SplitN(importPath, "/vendor/", 2); len(vendorerAndVendoree) == 2 {
		return vendorerAndVendoree[1]
	}

	return importPath
}
