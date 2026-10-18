// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main_test

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var update = flag.Bool("update", false, "update golden files")

// fixtureRoot is the module root, where testdata/modules lives, from this
// package's directory.
const fixtureRoot = "../.."

func TestCheckCommandE2E(t *testing.T) {
	tests := []struct {
		workdir        string
		args           []string // additional arguments to pass to report command.
		goldenFilePath string
		wantExitCode   int
	}{
		{"testdata/modules/hello01", nil, "output-check-forbidden.txt", 0},
		{
			"testdata/modules/hello01",
			[]string{"--disallowed_types=forbidden,notice"},
			"output-check-notice-forbidden.txt",
			1,
		},
		{"testdata/modules/cli02", nil, "output-check-forbidden.txt", 0},
		{
			"testdata/modules/cli02",
			[]string{"--disallowed_types=forbidden,notice"},
			"output-check-notice-forbidden.txt",
			1,
		},
		{"testdata/modules/cli02", []string{"--allowed_licenses=Apache-2.0"}, "output-check-license-names-1.txt", 1},
		{
			"testdata/modules/cli02",
			[]string{"--allowed_licenses=Apache-2.0,MIT"},
			"output-check-license-names-2.txt",
			1,
		},
		{
			"testdata/modules/cli02",
			[]string{"--allowed_licenses= Apache-2.0, MIT"},
			"output-check-license-names-2.txt",
			1,
		},
		{"testdata/modules/nolicense05", nil, "output-check.txt", 1},
		{"testdata/modules/complex", nil, "output-check-complex.txt", 0},
	}

	originalWorkDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	goLicensesPath := buildGoLicenses(t)

	for _, tt := range tests {
		t.Run(tt.workdir, func(t *testing.T) {
			t.Chdir(filepath.Join(originalWorkDir, fixtureRoot, tt.workdir))

			cmd := exec.Command("go", "mod", "download")

			log, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("downloading go modules:\n%s", string(log))
			}
			// Extra flags to disable logging.
			// Logs are not deterministic for this golden test.
			args := append([]string{"check", "."}, tt.args...)
			cmd = exec.Command(goLicensesPath, args...)
			cmd.Env = goldenPlatform()
			// Capture stderr to buffer.
			var stderr bytes.Buffer

			cmd.Stderr = &stderr
			exitCode := 0

			t.Logf("%s $ go-licenses check .", tt.workdir)

			output, err := cmd.Output()
			if err != nil {
				exitCode = -1

				exitError := &exec.ExitError{}
				if errors.As(err, &exitError) {
					exitCode = exitError.ExitCode()
				}
			}

			if len(output) != 0 {
				t.Fatalf("unexpected output running go-licenses check: %s.", string(output))
			}

			if exitCode != tt.wantExitCode {
				t.Fatalf(
					"unexpected exit code running go-licenses check, expected %d but got %d",
					tt.wantExitCode,
					exitCode,
				)
			}

			got := filterOutput(stderr.String())
			if *update {
				err = os.WriteFile(tt.goldenFilePath, []byte(got), 0o600)
				if err != nil {
					t.Fatalf("writing golden file: %s", err)
				}
			}

			goldenBytes, err := os.ReadFile(tt.goldenFilePath)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					t.Fatalf("reading golden file: %s. Create a golden file by running `go test --update .`", err)
				}

				t.Fatalf("reading golden file: %s", err)
			}

			golden := string(goldenBytes)
			if got != golden {
				t.Logf("\n=== start of log ===\n%s=== end of log ===\n\n\n", stderr.String())
				t.Fatalf("result of go-licenses check does not match the golden file.\n"+
					"Diff -golden +got:\n%s\n"+
					"Update the golden by running `go test --update .`",
					cmp.Diff(golden, got))
			}
		})
	}
}

// TestSaveCommandE2E runs save into a fresh path, which it creates, and
// into one that exists, which it refuses without touching.
func TestSaveCommandE2E(t *testing.T) {
	t.Parallel()

	goLicensesPath := buildGoLicenses(t)
	module := filepath.Join(fixtureRoot, "testdata", "modules", "hello01")

	download := exec.Command("go", "mod", "download")
	download.Dir = module

	if log, err := download.CombinedOutput(); err != nil {
		t.Fatalf("downloading go modules:\n%s", log)
	}

	save := func(t *testing.T, savePath string) (string, error) {
		t.Helper()

		cmd := exec.Command(goLicensesPath, "save", ".", "--save_path="+savePath)
		cmd.Dir = module
		cmd.Env = goldenPlatform()

		var stderr bytes.Buffer

		cmd.Stderr = &stderr
		err := cmd.Run()

		return stderr.String(), err
	}

	t.Run("fresh", func(t *testing.T) {
		t.Parallel()

		savePath := filepath.Join(t.TempDir(), "licenses")

		if stderr, err := save(t, savePath); err != nil {
			t.Fatalf("save: %v\n%s", err, stderr)
		}

		saved := filepath.Join(savePath, "github.com", "forkcloser", "go-licenses", "testdata", "modules", "hello01")
		if _, err := os.Stat(filepath.Join(saved, "LICENSE")); err != nil {
			t.Fatalf("the module's licence was not saved: %v", err)
		}
	})

	t.Run("exists", func(t *testing.T) {
		t.Parallel()

		savePath := t.TempDir()

		stderr, err := save(t, savePath)
		if err == nil {
			t.Fatal("save into an existing path succeeded")
		}

		if !strings.Contains(stderr, savePath+" already exists") {
			t.Fatalf("save did not refuse the existing path:\n%s", stderr)
		}

		entries, err := os.ReadDir(savePath)
		if err != nil || len(entries) != 0 {
			t.Fatalf("save wrote into the existing path: %v %v", entries, err)
		}
	})
}

// buildGoLicenses builds the CLI into a temporary directory.
func buildGoLicenses(t *testing.T) string {
	t.Helper()

	goLicensesPath := filepath.Join(t.TempDir(), goLicensesBinary())

	if _, err := exec.Command("go", "build", "-o", goLicensesPath).Output(); err != nil {
		t.Fatal(err)
	}

	t.Logf("Built go-licenses binary in %s.", goLicensesPath)

	return goLicensesPath
}

func filterOutput(output string) string {
	output = regexp.MustCompile(`(?m)W\d+.*\n`).
		ReplaceAllString(output, "")

	output = regexp.MustCompile(`(?m)^/.*\n`).
		ReplaceAllString(output, "")

	return output
}

// goLicensesBinary is the built CLI's file name: Windows runs only a file
// with an executable extension.
func goLicensesBinary() string {
	if runtime.GOOS == "windows" {
		return "go-licenses.exe"
	}

	return "go-licenses"
}

// goldenPlatform is the environment the CLI runs in: the goldens are
// linux/amd64's, and the packages a module loads, so the libraries it
// reports, change with the platform (golang.org/x/sys/cpu is imported on
// amd64 only). LOG_LEVEL=error keeps the warnings off the compared output.
func goldenPlatform() []string {
	return append(os.Environ(), "GOOS=linux", "GOARCH=amd64", "LOG_LEVEL=error")
}
