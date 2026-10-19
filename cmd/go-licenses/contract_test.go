// The command's contract from the outside: the binary built once, run in
// workspaces of local modules, judged by its exit status, what it prints
// and what it saves.
package main_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/forkcloser/go-licenses/internal/testmods"
)

var contractBinary string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "go-licenses-contract")
	if err != nil {
		panic(err)
	}

	contractBinary = filepath.Join(dir, "go-licenses")

	build := exec.Command("go", "build", "-o", contractBinary, ".")
	if out, err := build.CombinedOutput(); err != nil {
		panic(string(out))
	}

	code := m.Run()

	_ = os.RemoveAll(dir)
	os.Exit(code)
}

func deps() []testmods.Module {
	return []testmods.Module{
		{Path: "example.com/mit", Licenses: map[string]string{"LICENSE": "MIT"}, Imports: []string{"example.com/cc0"}},
		{Path: "example.com/apache", Licenses: map[string]string{"LICENSE": "Apache-2.0"}, Notice: "Apache product notice\n", SubLicense: "MIT"},
		{Path: "example.com/gpl", Licenses: map[string]string{"COPYING": "GPL-3.0"}},
		{Path: "example.com/mpl", Licenses: map[string]string{"LICENSE.md": "MPL-2.0"}},
		{Path: "example.com/cc0", Licenses: map[string]string{"UNLICENSE": "CC0-1.0"}},
	}
}

func badDeps() []testmods.Module {
	return append(deps(),
		testmods.Module{Path: "example.com/agpl", Licenses: map[string]string{"LICENSE": "AGPL-3.0"}},
		testmods.Module{Path: "example.com/nolicense", Licenses: map[string]string{"README.md": "\n# nolicense\n"}},
	)
}

// workspace builds the modules with the application under MIT, enters the
// application and sets the offline environment.
func workspace(t *testing.T, mods []testmods.Module, testOnly []testmods.Module) testmods.Workspace {
	t.Helper()

	testmods.Offline(t)

	ws := testmods.Build(t, mods, testOnly)
	if err := os.WriteFile(filepath.Join(ws.App, "LICENSE"), []byte(testmods.Text(t, "MIT")), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Chdir(ws.App)

	return ws
}

// run runs the binary in the current directory: exit status, stdout,
// stderr.
func run(t *testing.T, args ...string) (int, string, string) {
	t.Helper()

	cmd := exec.CommandContext(t.Context(), contractBinary, args...)
	cmd.Env = os.Environ()

	var stdout, stderr strings.Builder

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	var exit *exec.ExitError

	switch {
	case err == nil:
		return 0, stdout.String(), stderr.String()
	case errors.As(err, &exit):
		return exit.ExitCode(), stdout.String(), stderr.String()
	default:
		t.Fatalf("running %s: %v", contractBinary, err)

		return -1, "", ""
	}
}

// findings are stderr's lines that report a library, in order.
func findings(stderr string) []string {
	var out []string

	for line := range strings.SplitSeq(stderr, "\n") {
		if strings.Contains(line, "for library '") {
			out = append(out, line)
		}
	}

	sort.Strings(out)

	return out
}

// TestContractCheckDefaultsToForbiddenAndUnknown: `check` with no flags
// refuses a forbidden license and a library without one, each on a line
// of its own on stderr, exit 1; a tree of notice, restricted, reciprocal
// and unencumbered licenses passes with exit 0 and nothing printed.
func TestContractCheckDefaultsToForbiddenAndUnknown(t *testing.T) {
	workspace(t, badDeps(), nil)

	code, stdout, stderr := run(t, "check", ".")
	if code != 1 || stdout != "" {
		t.Errorf("exit %d stdout %q", code, stdout)
	}

	want := []string{
		"Did not find license for library 'example.com/nolicense'.",
		"License 'AGPL-3.0' of not allowed license type 'Forbidden' found for library 'example.com/agpl'.",
	}
	if got := findings(stderr); !slices.Equal(got, want) {
		t.Errorf("findings %q, want %q", got, want)
	}

	workspace(t, deps(), nil)

	if code, stdout, stderr := run(t, "check", "."); code != 0 || stdout != "" || stderr != "" {
		t.Errorf("a clean tree: exit %d stdout %q stderr %q", code, stdout, stderr)
	}
}

// TestContractCheckDisallowedTypes: --disallowed_types replaces the
// default set: restricted and reciprocal flag GPL and MPL by name and
// type, the forbidden one then passes; a library without a license file
// fails whatever the flags; --allowed_licenses flags every license not
// named; the two flags together are refused.
func TestContractCheckDisallowedTypes(t *testing.T) {
	workspace(t, badDeps(), nil)

	code, _, stderr := run(t, "check", ".", "--disallowed_types=restricted,reciprocal")
	want := []string{
		"Did not find license for library 'example.com/nolicense'.",
		"License 'GPL-3.0' of not allowed license type 'Restricted' found for library 'example.com/gpl'.",
		"License 'MPL-2.0' of not allowed license type 'Reciprocal' found for library 'example.com/mpl'.",
	}

	if code != 1 || !slices.Equal(findings(stderr), want) {
		t.Errorf("--disallowed_types: exit %d, findings %q, want %q", code, findings(stderr), want)
	}

	code, _, stderr = run(t, "check", ".", "--allowed_licenses=MIT,CC0-1.0")
	want = []string{
		"Did not find license for library 'example.com/nolicense'.",
		"Not allowed license 'AGPL-3.0' found for library 'example.com/agpl'.",
		"Not allowed license 'Apache-2.0' found for library 'example.com/apache'.",
		"Not allowed license 'GPL-3.0' found for library 'example.com/gpl'.",
		"Not allowed license 'MPL-2.0' found for library 'example.com/mpl'.",
	}

	if code != 1 || !slices.Equal(findings(stderr), want) {
		t.Errorf("--allowed_licenses: exit %d, findings %q, want %q", code, findings(stderr), want)
	}

	code, _, stderr = run(t, "check", ".", "--allowed_licenses=MIT", "--disallowed_types=forbidden")
	if code != 1 || !strings.Contains(stderr, "can't be used at the same time") {
		t.Errorf("both flags: exit %d stderr %q", code, stderr)
	}
}

// TestContractCheckIgnoreAndIncludeTests: --ignore drops libraries under
// a path prefix from the check, what they import still checked;
// --include_tests brings in a module a test file alone imports.
func TestContractCheckIgnoreAndIncludeTests(t *testing.T) {
	testOnly := testmods.Module{Path: "example.com/testagpl", Licenses: map[string]string{"LICENSE": "AGPL-3.0"}}
	workspace(t, badDeps(), []testmods.Module{testOnly})

	if code, _, stderr := run(t, "check", ".", "--ignore", "example.com/agpl", "--ignore", "example.com/nolicense"); code != 0 || stderr != "" {
		t.Errorf("--ignore: exit %d stderr %q", code, stderr)
	}

	code, _, stderr := run(t, "check", ".", "--ignore", "example.com/agpl", "--ignore", "example.com/nolicense", "--include_tests")
	if code != 1 || !strings.Contains(stderr, "found for library 'example.com/testagpl'") {
		t.Errorf("--include_tests: exit %d stderr %q", code, stderr)
	}
}

// TestContractSaveLaysOutByRestrictiveness: `save` writes one directory
// per library under the save path, named by the library: the license
// file alone for a notice or unencumbered license, with a NOTICE beside
// it when the module has one; the whole package directory for a
// restricted or reciprocal one; the application's own license too. A
// save path that exists is refused; --force replaces it.
func TestContractSaveLaysOutByRestrictiveness(t *testing.T) {
	ws := workspace(t, deps(), nil)
	out := filepath.Join(ws.Root, "out")

	if code, _, stderr := run(t, "save", ".", "--save_path="+out); code != 0 {
		t.Fatalf("save: exit %d stderr %q", code, stderr)
	}

	var saved []string

	if err := filepath.WalkDir(out, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			rel, _ := filepath.Rel(out, path)
			saved = append(saved, filepath.ToSlash(rel))
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}

	sort.Strings(saved)

	want := []string{
		"example.com/apache/LICENSE",
		"example.com/apache/NOTICE",
		"example.com/apache/sub/LICENSE",
		"example.com/app/LICENSE",
		"example.com/cc0/UNLICENSE",
		"example.com/gpl/COPYING",
		"example.com/gpl/go.mod",
		"example.com/gpl/lib.go",
		"example.com/mit/LICENSE",
		"example.com/mpl/LICENSE.md",
		"example.com/mpl/go.mod",
		"example.com/mpl/lib.go",
	}

	if !slices.Equal(saved, want) {
		t.Errorf("saved:\n%s\nwant:\n%s", strings.Join(saved, "\n"), strings.Join(want, "\n"))
	}

	if code, _, stderr := run(t, "save", ".", "--save_path="+out); code != 1 || !strings.Contains(stderr, "already exists") {
		t.Errorf("an existing save path: exit %d stderr %q", code, stderr)
	}

	marker := filepath.Join(out, "stale")
	if err := os.WriteFile(marker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	if code, _, stderr := run(t, "save", ".", "--save_path="+out, "--force"); code != 0 {
		t.Errorf("--force: exit %d stderr %q", code, stderr)
	}

	if _, err := os.Stat(marker); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("--force left the old directory's file: %v", err)
	}
}

// TestContractSaveRefusesWhatItCannotRedistribute: a forbidden license or
// a library without one fails the save, exit 1, the error naming them.
func TestContractSaveRefusesWhatItCannotRedistribute(t *testing.T) {
	ws := workspace(t, badDeps(), nil)
	out := filepath.Join(ws.Root, "out")

	code, _, stderr := run(t, "save", ".", "--save_path="+out)
	if code != 1 || !strings.Contains(stderr, "incompatible/unknown license") || !strings.Contains(stderr, "example.com/agpl") || !strings.Contains(stderr, "example.com/nolicense") {
		t.Errorf("exit %d stderr %q", code, stderr)
	}
}

// TestContractUsage: no command is a usage error, exit 1; --help exits 0
// and names both commands.
func TestContractUsage(t *testing.T) {
	workspace(t, deps(), nil)

	if code, _, _ := run(t); code != 1 {
		t.Errorf("no command: exit %d, want 1", code)
	}

	code, stdout, _ := run(t, "--help")
	if code != 0 || !strings.Contains(stdout, "check") || !strings.Contains(stdout, "save") {
		t.Errorf("--help: exit %d stdout %q", code, stdout)
	}
}
