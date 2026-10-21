// Package testmods builds a workspace of local Go modules for the contract
// tests: an application module whose dependencies are modules beside it,
// reached through replace directives, so `go list` resolves them offline.
// Each dependency carries the license text the test names, taken from the
// classifier's own corpus, so the classifier's verdict on it is known.
package testmods

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

// Module describes one dependency: its path, the license files it carries
// (file name to the corpus text to copy, or literal content when the
// value starts with a newline), the packages it imports, and a
// sub-package with its own license file when SubLicense is set.
type Module struct {
	Path       string
	Licenses   map[string]string
	Imports    []string
	Notice     string // content of a NOTICE file beside the license, when set
	SubLicense string // corpus text for a LICENSE in a "sub" package of the module
}

// The modes of what Build writes: directories the test alone enters, files
// the test alone reads.
const (
	dirMode  = 0o750
	fileMode = 0o600
)

// Workspace is the built tree: Root holds one directory per module; App is
// the application module's directory, the one to run the tool in.
type Workspace struct {
	Root string
	App  string
}

// corpus is the classifier's license corpus, where the texts come from.
func corpus(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("no caller")
	}

	return filepath.Join(filepath.Dir(file), "..", "third_party", "licenseclassifier", "assets", "License")
}

// Text is the pristine text of the named license from the corpus.
func Text(t *testing.T, name string) string {
	t.Helper()

	dir := filepath.Join(corpus(t), name)

	for _, candidate := range []string{"pristine.txt", "license.txt", "a.txt"} {
		b, err := os.ReadFile(filepath.Join(dir, candidate)) // #nosec G304 -- a corpus file inside the module
		if err == nil {
			return string(b)
		}
	}

	t.Fatalf("no text for %s in %s", name, dir)

	return ""
}

// Build writes the workspace: the application module importing every
// module in deps (and testOnly from a test file alone), each dependency a
// module of its own with its files.
func Build(t *testing.T, deps, testOnly []Module) Workspace {
	t.Helper()

	root := t.TempDir()
	workspace := Workspace{Root: root, App: filepath.Join(root, "app")}

	all := slices.Concat(deps, testOnly)

	var gomod strings.Builder

	gomod.WriteString("module example.com/app\n\ngo 1.26\n\n")

	for _, m := range all {
		gomod.WriteString("require " + m.Path + " v0.0.0\n")
		gomod.WriteString("replace " + m.Path + " => ../" + dirOf(m.Path) + "\n")
	}

	write(t, filepath.Join(workspace.App, "go.mod"), gomod.String())
	write(t, filepath.Join(workspace.App, "main.go"), program("main", deps, "func main() { use() }\n"))

	if len(testOnly) > 0 {
		write(t, filepath.Join(workspace.App, "main_test.go"), program("main", testOnly, "func init() { use() }\n"))
	}

	for _, module := range all {
		dir := filepath.Join(root, dirOf(module.Path))

		var mod strings.Builder

		mod.WriteString("module " + module.Path + "\n\ngo 1.26\n\n")

		for _, imp := range module.Imports {
			mod.WriteString("require " + imp + " v0.0.0\n")
			mod.WriteString("replace " + imp + " => ../" + dirOf(imp) + "\n")
		}

		write(t, filepath.Join(dir, "go.mod"), mod.String())
		write(t, filepath.Join(dir, "lib.go"), program(dirOf(module.Path), modulesOf(all, module.Imports), ""))

		for name, text := range module.Licenses {
			if strings.HasPrefix(text, "\n") {
				write(t, filepath.Join(dir, name), text[1:])
			} else {
				write(t, filepath.Join(dir, name), Text(t, text))
			}
		}

		if module.Notice != "" {
			write(t, filepath.Join(dir, "NOTICE"), module.Notice)
		}

		if module.SubLicense != "" {
			write(
				t,
				filepath.Join(dir, "sub", "sub.go"),
				"package sub\n\n// Use is the sub package's use.\nfunc Use() {}\n",
			)
			write(t, filepath.Join(dir, "sub", "LICENSE"), Text(t, module.SubLicense))
			write(t, filepath.Join(dir, "lib.go"), program(dirOf(module.Path), nil, "")+"\n")
		}
	}

	return workspace
}

func dirOf(path string) string {
	return path[strings.LastIndex(path, "/")+1:]
}

func modulesOf(all []Module, paths []string) []Module {
	var out []Module

	for _, m := range all {
		for _, p := range paths {
			if m.Path == p {
				out = append(out, m)
			}
		}
	}

	return out
}

// program is a Go source file of the package importing every module given
// (its root package, and its sub package when it has one), with a use()
// that references each, plus trailer.
func program(pkg string, imports []Module, trailer string) string {
	var text strings.Builder

	text.WriteString("package " + pkg + "\n\n")

	if len(imports) > 0 {
		text.WriteString("import (\n")

		for _, m := range imports {
			text.WriteString("\t" + dirOf(m.Path) + " \"" + m.Path + "\"\n")

			if m.SubLicense != "" {
				text.WriteString("\t" + dirOf(m.Path) + "sub \"" + m.Path + "/sub\"\n")
			}
		}

		text.WriteString(")\n\n")
	}

	text.WriteString("func use() {\n")

	for _, m := range imports {
		text.WriteString("\t" + dirOf(m.Path) + ".Use()\n")

		if m.SubLicense != "" {
			text.WriteString("\t" + dirOf(m.Path) + "sub.Use()\n")
		}
	}

	text.WriteString("}\n\n// Use is the package's use.\nfunc Use() { use() }\n")
	text.WriteString(trailer)

	return text.String()
}

func write(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), dirMode); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(content), fileMode); err != nil {
		t.Fatal(err)
	}
}

// Offline sets the Go environment the workspace needs: modules resolved
// by their replace directives alone, no proxy, no workspace file, no
// toolchain switch.
func Offline(t *testing.T) {
	t.Helper()

	t.Setenv("GOFLAGS", "-mod=mod")
	t.Setenv("GOPROXY", "off")
	t.Setenv("GOWORK", "off")
	t.Setenv("GOTOOLCHAIN", "local")
	t.Setenv("GONOSUMDB", "example.com")
	t.Setenv("GONOSUMCHECK", "1")
	t.Setenv("GOFLAGS", "-mod=mod")
}
