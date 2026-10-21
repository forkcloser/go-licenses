// The licenses package's contract from the outside, over a workspace of
// local modules whose license texts come from the classifier's corpus.
package licenses_test

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/forkcloser/go-licenses/internal/testmods"
	"github.com/forkcloser/go-licenses/licenses"
)

// standard is the dependency set the contract cases share: a notice
// license, a notice license with a NOTICE file and a sub package under
// MIT, a restricted one, a reciprocal one, an unencumbered one, a
// forbidden one, one with no license file, and a module reached only
// through another.
func standard() []testmods.Module {
	return []testmods.Module{
		{Path: "example.com/mit", Licenses: map[string]string{"LICENSE": "MIT"}, Imports: []string{"example.com/cc0"}},
		{
			Path:       "example.com/apache",
			Licenses:   map[string]string{"LICENSE": "Apache-2.0"},
			Notice:     "Apache product notice\n",
			SubLicense: "MIT",
		},
		{Path: "example.com/gpl", Licenses: map[string]string{"COPYING": "GPL-3.0"}},
		{Path: "example.com/mpl", Licenses: map[string]string{"LICENSE.md": "MPL-2.0"}},
		{Path: "example.com/agpl", Licenses: map[string]string{"LICENSE": "AGPL-3.0"}},
		{Path: "example.com/nolicense", Licenses: map[string]string{"README.md": "\n# nolicense\n\nA library.\n"}},
	}
}

func cc0() testmods.Module {
	return testmods.Module{Path: "example.com/cc0", Licenses: map[string]string{"UNLICENSE": "CC0-1.0"}}
}

type libView struct {
	name, file string
	licenses   []string
	types      []string
	packages   []string
}

func view(ws testmods.Workspace, libs []*licenses.Library) map[string]libView {
	out := map[string]libView{}

	for _, l := range libs {
		v := libView{name: l.Name(), packages: l.Packages}

		if l.LicenseFile != "" {
			rel, _ := filepath.Rel(ws.Root, l.LicenseFile)
			v.file = filepath.ToSlash(rel)
		}

		for _, lic := range l.Licenses {
			v.licenses = append(v.licenses, lic.Name)
			v.types = append(v.types, string(lic.Type))
		}

		out[v.name] = v
	}

	return out
}

// TestContractLibrariesAreTheModulesLicenseFiles: Libraries lists one
// library per license file the dependencies' packages resolve to, named
// by the packages' common ancestor, sorted by name: the module root's
// LICENSE, COPYING, LICENSE.md or UNLICENSE as the classifier names it;
// a package with its own license file is a library of its own; a module
// whose candidates hold no license is a library with no license file;
// a module reached only through another is listed, and so is the
// application's own package, with its own license file or none; the
// standard library is not.
func TestContractLibrariesAreTheModulesLicenseFiles(t *testing.T) {
	testmods.Offline(t)

	ws := testmods.Build(t, append(standard(), cc0()), nil)
	t.Chdir(ws.App)

	classifier, err := licenses.NewClassifier()
	if err != nil {
		t.Fatal(err)
	}

	libs, warnings, err := licenses.Libraries(t.Context(), classifier, false, nil, ".")
	if err != nil {
		t.Fatal(err)
	}

	for _, w := range warnings {
		t.Errorf("warning: %s", w)
	}

	got := view(ws, libs)

	want := map[string]libView{
		"example.com/mit": {
			file:     "mit/LICENSE",
			licenses: []string{"MIT"},
			types:    []string{"notice"},
			packages: []string{"example.com/mit"},
		},
		"example.com/apache": {
			file:     "apache/LICENSE",
			licenses: []string{"Apache-2.0"},
			types:    []string{"notice"},
			packages: []string{"example.com/apache"},
		},
		"example.com/apache/sub": {
			file:     "apache/sub/LICENSE",
			licenses: []string{"MIT"},
			types:    []string{"notice"},
			packages: []string{"example.com/apache/sub"},
		},
		"example.com/gpl": {
			file:     "gpl/COPYING",
			licenses: []string{"GPL-3.0"},
			types:    []string{"restricted"},
			packages: []string{"example.com/gpl"},
		},
		"example.com/mpl": {
			file:     "mpl/LICENSE.md",
			licenses: []string{"MPL-2.0"},
			types:    []string{"reciprocal"},
			packages: []string{"example.com/mpl"},
		},
		"example.com/agpl": {
			file:     "agpl/LICENSE",
			licenses: []string{"AGPL-3.0"},
			types:    []string{"forbidden"},
			packages: []string{"example.com/agpl"},
		},
		"example.com/nolicense": {file: "", packages: []string{"example.com/nolicense"}},
		"example.com/cc0": {
			file:     "cc0/UNLICENSE",
			licenses: []string{"CC0-1.0"},
			types:    []string{"unencumbered"},
			packages: []string{"example.com/cc0"},
		},
		"example.com/app": {file: "", packages: []string{"example.com/app"}},
	}

	for name, w := range want {
		g, ok := got[name]
		if !ok {
			t.Errorf("%s: not listed", name)

			continue
		}

		if g.file != w.file || !slices.Equal(g.licenses, w.licenses) || !slices.Equal(g.types, w.types) ||
			!slices.Equal(g.packages, w.packages) {
			t.Errorf("%s: %+v, want %+v", name, g, w)
		}
	}

	for name := range got {
		if _, ok := want[name]; !ok {
			t.Errorf("%s: listed, not a dependency's library", name)
		}
	}

	names := make([]string, 0, len(libs))
	for _, l := range libs {
		names = append(names, l.Name())
	}

	if !slices.IsSorted(names) {
		t.Errorf("libraries not sorted by name: %v", names)
	}
}

// TestContractIncludeTestsAndIgnore: a module imported by a test file
// alone is listed only with includeTests; an ignored path prefix drops
// the libraries under it and keeps what they import.
func TestContractIncludeTestsAndIgnore(t *testing.T) {
	testmods.Offline(t)

	testOnly := testmods.Module{Path: "example.com/testonly", Licenses: map[string]string{"LICENSE": "MIT"}}
	ws := testmods.Build(t, append(standard(), cc0()), []testmods.Module{testOnly})
	t.Chdir(ws.App)

	classifier, err := licenses.NewClassifier()
	if err != nil {
		t.Fatal(err)
	}

	names := func(includeTests bool, ignore []string) []string {
		libs, _, err := licenses.Libraries(t.Context(), classifier, includeTests, ignore, ".")
		if err != nil {
			t.Fatal(err)
		}

		out := make([]string, 0, len(libs))
		for _, l := range libs {
			out = append(out, l.Name())
		}

		return out
	}

	if n := names(false, nil); slices.Contains(n, "example.com/testonly") {
		t.Errorf("without includeTests, the test-only module is listed: %v", n)
	}

	if n := names(true, nil); !slices.Contains(n, "example.com/testonly") {
		t.Errorf("with includeTests, the test-only module is not listed: %v", n)
	}

	n := names(false, []string{"example.com/mit", "example.com/apache"})
	for _, dropped := range []string{"example.com/mit", "example.com/apache", "example.com/apache/sub"} {
		if slices.Contains(n, dropped) {
			t.Errorf("ignored prefix still listed: %s in %v", dropped, n)
		}
	}

	if !slices.Contains(n, "example.com/cc0") {
		t.Errorf("what an ignored module imports is not listed: %v", n)
	}
}

// TestContractFindCandidatesWalksUpToTheRoot: the candidates of a
// directory are its files named like a license (LICENSE, LICENCE,
// UNLICENSE, COPYING, README, NOTICE, any case, any suffix), then its
// parent's, up to the root directory and no further; a directory outside
// the root is refused.
func TestContractFindCandidatesWalksUpToTheRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	pkg := filepath.Join(root, "a", "b")

	for _, f := range []string{"LICENSE", "README.md", "NOTICE.txt", "main.go", "a/COPYING", "a/licence.txt", "a/b/UNLICENSE", "a/b/Readme", "a/b/notes.txt"} {
		path := filepath.Join(root, filepath.FromSlash(f))
		if err := writeFile(path); err != nil {
			t.Fatal(err)
		}
	}

	got, err := licenses.FindCandidates(pkg, root)
	if err != nil {
		t.Fatal(err)
	}

	rel := make([]string, 0, len(got))
	for _, g := range got {
		r, _ := filepath.Rel(root, g)
		rel = append(rel, filepath.ToSlash(r))
	}

	want := []string{"a/b/Readme", "a/b/UNLICENSE", "a/COPYING", "a/licence.txt", "LICENSE", "NOTICE.txt", "README.md"}
	if !slices.Equal(rel, want) {
		t.Errorf("candidates %v, want %v (closest first, each directory's in name order)", rel, want)
	}

	if _, err := licenses.FindCandidates(t.TempDir(), root); err == nil {
		t.Error("a directory outside the root: no error")
	}
}

// TestContractIdentifyNamesTheCorpusText: the classifier names a pristine
// corpus text by its license, once, with the type the table gives it; a
// file holding two texts names both; a README with no license text names
// none; an empty path names none.
//
//nolint:paralleltest // Offline sets the Go environment with t.Setenv
func TestContractIdentifyNamesTheCorpusText(t *testing.T) {
	testmods.Offline(t)

	classifier, err := licenses.NewClassifier()
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()

	cases := map[string]struct {
		content string
		want    []licenses.License
	}{
		"MIT": {testmods.Text(t, "MIT"), []licenses.License{{Name: "MIT", Type: licenses.Notice}}},
		"GPL": {testmods.Text(t, "GPL-3.0"), []licenses.License{{Name: "GPL-3.0", Type: licenses.Restricted}}},
		"both": {
			testmods.Text(t, "MIT") + "\n\n" + testmods.Text(t, "Apache-2.0"),
			[]licenses.License{{Name: "MIT", Type: licenses.Notice}, {Name: "Apache-2.0", Type: licenses.Notice}},
		},
		"twice": {
			testmods.Text(t, "MIT") + "\n\n" + testmods.Text(t, "MIT"),
			[]licenses.License{{Name: "MIT", Type: licenses.Notice}},
		},
		"readme": {"# a library\n\nSee the docs.\n", []licenses.License{}},
	}

	for name, c := range cases {
		path := filepath.Join(dir, name)
		if err := writeContent(path, c.content); err != nil {
			t.Fatal(err)
		}

		got, err := classifier.Identify(path)
		if err != nil {
			t.Errorf("%s: %v", name, err)

			continue
		}

		slices.SortFunc(got, func(a, b licenses.License) int { return strings.Compare(a.Name, b.Name) })
		slices.SortFunc(c.want, func(a, b licenses.License) int { return strings.Compare(a.Name, b.Name) })

		if !slices.Equal(got, c.want) {
			t.Errorf("%s: %v, want %v", name, got, c.want)
		}
	}

	if got, err := classifier.Identify(""); err != nil || got != nil {
		t.Errorf("an empty path: %v, %v; want nothing and no error", got, err)
	}

	if _, err := classifier.Identify(filepath.Join(dir, "missing")); err == nil {
		t.Error("a missing file: no error")
	}
}

// TestContractTypesAndRestrictiveness: the type table names each class
// the README lists, Unknown printing as "unknown"; the restrictiveness of
// a set is NotAllowed with any forbidden or by-exception type, Unknown
// with none or any unknown, ShareCode with any restricted or reciprocal,
// and ShareLicense otherwise.
func TestContractTypesAndRestrictiveness(t *testing.T) {
	t.Parallel()

	for name, want := range map[string]licenses.Type{
		"MIT": licenses.Notice, "Apache-2.0": licenses.Notice, "BSD-3-Clause": licenses.Notice,
		"GPL-3.0": licenses.Restricted, "LGPL-3.0": licenses.Restricted, "MPL-2.0": licenses.Reciprocal,
		"CC0-1.0": licenses.Unencumbered, "Unlicense": licenses.Unencumbered, "AGPL-3.0": licenses.Forbidden,
		"WTFPL": licenses.Forbidden, "no-such-license": licenses.Unknown,
	} {
		if got := licenses.LicenseType(name); got != want {
			t.Errorf("%s: %q, want %q", name, got, want)
		}
	}

	if licenses.Unknown.String() != "unknown" || licenses.Notice.String() != "notice" {
		t.Errorf("Unknown prints %q, Notice %q", licenses.Unknown.String(), licenses.Notice.String())
	}

	for _, c := range []struct {
		types []licenses.Type
		want  licenses.LicenseRestrictiveness
	}{
		{nil, licenses.RestrictionsUnknown},
		{[]licenses.Type{licenses.Notice}, licenses.RestrictionsShareLicense},
		{[]licenses.Type{licenses.Permissive, licenses.Unencumbered}, licenses.RestrictionsShareLicense},
		{[]licenses.Type{licenses.Notice, licenses.Restricted}, licenses.RestrictionsShareCode},
		{[]licenses.Type{licenses.Reciprocal}, licenses.RestrictionsShareCode},
		{[]licenses.Type{licenses.Notice, licenses.Unknown}, licenses.RestrictionsUnknown},
		{[]licenses.Type{licenses.Restricted, licenses.Unknown}, licenses.RestrictionsUnknown},
		{[]licenses.Type{licenses.Notice, licenses.Forbidden}, licenses.RestrictionsNotAllowed},
		{[]licenses.Type{licenses.ByExceptionOnly}, licenses.RestrictionsNotAllowed},
		{[]licenses.Type{licenses.Forbidden, licenses.Unknown}, licenses.RestrictionsNotAllowed},
	} {
		if got := licenses.LicenseTypeRestrictiveness(c.types...); got != c.want {
			t.Errorf("%v: %q, want %q", c.types, got, c.want)
		}
	}
}

// TestContractLibraryNameIsTheCommonAncestor: a library is named by the
// directory its packages share: sibling packages sharing a prefix name
// their parent, never the shorter package; one package names itself;
// packages that share no directory name nothing.
func TestContractLibraryNameIsTheCommonAncestor(t *testing.T) {
	t.Parallel()

	for _, c := range []struct {
		packages []string
		want     string
	}{
		{[]string{"m/pkg", "m/pkgextra"}, "m"},
		{[]string{"m/pkg", "m/pkgextra", "m/pkg/sub"}, "m"},
		{[]string{"m/pkg", "m/pkg/sub"}, "m/pkg"},
		{[]string{"m"}, "m"},
		{[]string{"m/a/x", "m/b/y"}, "m"},
		{[]string{"m", "mextra"}, ""},
	} {
		if got := (&licenses.Library{Packages: c.packages}).Name(); got != c.want {
			t.Errorf("%v: %q, want %q", c.packages, got, c.want)
		}
	}
}

// TestContractFindCandidatesStaysUnderTheRoot: a directory that is not
// under the root, even one whose path starts with the root's, is refused,
// and the walk up from a package never leaves the root.
func TestContractFindCandidatesStaysUnderTheRoot(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	root := filepath.Join(base, "root")
	sibling := filepath.Join(base, "root2", "x")

	for _, d := range []string{filepath.Join(root, "pkg"), sibling} {
		if err := os.MkdirAll(d, 0o750); err != nil {
			t.Fatal(err)
		}
	}

	for _, f := range []string{filepath.Join(base, "root2", "LICENSE"), filepath.Join(base, "LICENSE")} {
		if err := writeFile(f); err != nil {
			t.Fatal(err)
		}
	}

	if got, err := licenses.FindCandidates(sibling, root); err == nil {
		t.Errorf("a sibling of the root sharing its prefix: %v and no error, want a refusal", got)
	}

	got, err := licenses.FindCandidates(filepath.Join(root, "pkg"), root)
	if err != nil || len(got) != 0 {
		t.Errorf(
			"a package under a root with no license: %v, %v; want none (the LICENSE above the root is not the root's)",
			got,
			err,
		)
	}
}
