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

package licenses

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLibraries(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Cannot get working directory: %v", err)
	}

	classifier := classifierStub{
		licenses: map[string][]License{
			"testdata/LICENSE": {
				{
					Name: "foo",
					Type: Notice,
				},
			},
			"testdata/direct/LICENSE": {
				{
					Name: "foo",
					Type: Notice,
				},
			},
			"testdata/indirect/LICENSE": {
				{
					Name: "foo",
					Type: Notice,
				},
			},
		},
	}

	type wantedLibInfo struct {
		Name        string
		LicenseFile string
		Licenses    []License
	}

	testdataLibInfo := wantedLibInfo{
		Name:        "github.com/forkcloser/go-licenses/licenses/testdata",
		LicenseFile: filepath.Join(wd, "testdata", "LICENSE"),
		Licenses:    classifier.licenses["testdata/LICENSE"],
	}
	directLibInfo := wantedLibInfo{
		Name:        "github.com/forkcloser/go-licenses/licenses/testdata/direct",
		LicenseFile: filepath.Join(wd, "testdata", "direct", "LICENSE"),
		Licenses:    classifier.licenses["testdata/direct/LICENSE"],
	}
	indirectLibInfo := wantedLibInfo{
		Name:        "github.com/forkcloser/go-licenses/licenses/testdata/indirect",
		LicenseFile: filepath.Join(wd, "testdata", "indirect", "LICENSE"),
		Licenses:    classifier.licenses["testdata/indirect/LICENSE"],
	}

	for _, test := range []struct {
		desc         string
		importPath   string
		goflags      string
		includeTests bool
		ignore       []string
		wantLibs     []wantedLibInfo
	}{
		{
			desc:       "Detects direct dependency",
			importPath: "github.com/forkcloser/go-licenses/licenses/testdata/direct",
			wantLibs:   []wantedLibInfo{directLibInfo, indirectLibInfo},
		},
		{
			desc:       "Detects transitive dependency",
			importPath: "github.com/forkcloser/go-licenses/licenses/testdata",
			wantLibs:   []wantedLibInfo{testdataLibInfo, directLibInfo, indirectLibInfo},
		},
		{
			desc:       "Ignores a package path",
			importPath: "github.com/forkcloser/go-licenses/licenses/testdata",
			ignore: []string{
				"github.com/forkcloser/go-licenses/licenses/testdata/direct",
			},
			wantLibs: []wantedLibInfo{testdataLibInfo, indirectLibInfo},
		},
		{
			desc:         "Detects the dependencies only imported in testing code",
			importPath:   "github.com/forkcloser/go-licenses/licenses/testdata/testlib",
			includeTests: true,
			wantLibs: []wantedLibInfo{
				indirectLibInfo,
				{
					Name:        "github.com/forkcloser/go-licenses/licenses/testdata/testlib",
					LicenseFile: filepath.Join(wd, "testdata", "LICENSE"),
					Licenses:    classifier.licenses["testdata/LICENSE"],
				},
			},
		},
		{
			desc:         "Should not detect the dependencies only imported in testing code",
			importPath:   "github.com/forkcloser/go-licenses/licenses/testdata/testlib",
			includeTests: false,
			wantLibs: []wantedLibInfo{
				{
					Name:        "github.com/forkcloser/go-licenses/licenses/testdata/testlib",
					LicenseFile: filepath.Join(wd, "testdata", "LICENSE"),
					Licenses:    classifier.licenses["testdata/LICENSE"],
				},
			},
		},
		{
			desc:       "Build tagged package",
			importPath: "github.com/forkcloser/go-licenses/licenses/testdata/tags",
			goflags:    "-tags=tags",
			wantLibs: []wantedLibInfo{
				indirectLibInfo,
				{
					Name:        "github.com/forkcloser/go-licenses/licenses/testdata/tags",
					LicenseFile: filepath.Join(wd, "testdata", "LICENSE"),
					Licenses:    classifier.licenses["testdata/LICENSE"],
				},
			},
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			if test.goflags != "" {
				t.Setenv("GOFLAGS", test.goflags)
			}

			gotLibs, _, err := Libraries(t.Context(), classifier, test.includeTests, test.ignore, test.importPath)
			if err != nil {
				t.Fatalf("Libraries(_, %q) = (_, %q), want (_, nil)", test.importPath, err)
			}

			for i, gotLib := range gotLibs {
				if diff := cmp.Diff(test.wantLibs[i].Name, gotLib.Name()); diff != "" {
					t.Errorf("Libraries(_, %q)[%d].Name() diff (-want +got): %s", test.importPath, i, diff)
				}

				if diff := cmp.Diff(test.wantLibs[i].LicenseFile, gotLib.LicenseFile); diff != "" {
					t.Errorf("Libraries(_, %q)[%d].LicenseFile diff (-want +got): %s", test.importPath, i, diff)
				}

				if diff := cmp.Diff(test.wantLibs[i].Licenses, gotLib.Licenses); diff != "" {
					t.Errorf("Libraries(_, %q)[%d].Licenses diff (-want +got): %s", test.importPath, i, diff)
				}
			}

			if diff := cmp.Diff(len(test.wantLibs), len(gotLibs)); diff != "" {
				t.Errorf("len(Libraries(_, %q)) diff (-want +got): %s", test.importPath, diff)
			}
		})
	}
}

func TestLibraryName(t *testing.T) {
	for _, test := range []struct {
		desc     string
		lib      *Library
		wantName string
	}{
		{
			desc:     "Library with no packages",
			lib:      &Library{},
			wantName: "",
		},
		{
			desc: "Library with 1 package",
			lib: &Library{
				Packages: []string{
					"github.com/google/trillian/crypto",
				},
			},
			wantName: "github.com/google/trillian/crypto",
		},
		{
			desc: "Library with 2 packages",
			lib: &Library{
				Packages: []string{
					"github.com/google/trillian/crypto",
					"github.com/google/trillian/server",
				},
			},
			wantName: "github.com/google/trillian",
		},
		{
			desc: "Vendored library",
			lib: &Library{
				Packages: []string{
					"github.com/google/trillian/vendor/coreos/etcd",
				},
			},
			wantName: "github.com/google/trillian/vendor/coreos/etcd",
		},
		// A package whose import path is a prefix of a sibling's: the
		// ancestor is the directory before them, not the shorter path.
		{
			desc:     "Sibling packages sharing a name prefix",
			lib:      &Library{Packages: []string{"m/pkg", "m/pkgextra"}},
			wantName: "m",
		},
		{
			desc:     "Sibling packages sharing a name prefix, with a sub-package",
			lib:      &Library{Packages: []string{"m/pkg", "m/pkg/sub", "m/pkgextra"}},
			wantName: "m",
		},
		{
			desc:     "A package and its sub-package",
			lib:      &Library{Packages: []string{"m/pkg", "m/pkg/sub"}},
			wantName: "m/pkg",
		},
		{
			desc:     "Top-level names sharing a prefix",
			lib:      &Library{Packages: []string{"m", "mextra"}},
			wantName: "",
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			if got, want := test.lib.Name(), test.wantName; got != want {
				t.Fatalf("Name() = %q, want %q", got, want)
			}
		})
	}
}

// TestLibrariesVendored runs Libraries over a module that vendors its
// dependency: the vendored package's module carries no directory of its own,
// and Libraries finds it its parent module's, which bounds the licence search.
func TestLibrariesVendored(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "testdata", "modules", "vendored03"))
	if err != nil {
		t.Fatal(err)
	}

	t.Chdir(root)
	// The fixture is about vendor/: a GOFLAGS=-mod=mod in the environment
	// would read go-homedir from the module cache instead.
	t.Setenv("GOFLAGS", "-mod=vendor")

	classifier, err := NewClassifier()
	if err != nil {
		t.Fatal(err)
	}

	libs, _, err := Libraries(t.Context(), classifier, false, nil, ".")
	if err != nil {
		t.Fatalf("Libraries(vendored03) = %v", err)
	}

	for _, lib := range libs {
		if lib.Name() != "github.com/mitchellh/go-homedir" {
			continue
		}

		if lib.module == nil || lib.module.Dir != root {
			t.Fatalf("go-homedir's module = %+v, want the vendoring module at %s", lib.module, root)
		}

		want := filepath.Join(root, "vendor", "github.com", "mitchellh", "go-homedir", "LICENSE")
		if lib.LicenseFile != want {
			t.Errorf("go-homedir's licence = %q, want %q", lib.LicenseFile, want)
		}

		return
	}

	t.Fatalf("Libraries(vendored03) has no go-homedir library: %v", libs)
}

// TestLibrariesWarnings pins the warnings Libraries returns: a package with a
// non-Go file it cannot read for further imports is named, with the files.
func TestLibrariesWarnings(t *testing.T) {
	t.Parallel()

	classifier, err := NewClassifier()
	if err != nil {
		t.Fatal(err)
	}

	importPath := "github.com/forkcloser/go-licenses/licenses/testdata/nongo"

	_, warnings, err := Libraries(t.Context(), classifier, false, nil, importPath)
	if err != nil {
		t.Fatalf("Libraries(%q) = %v", importPath, err)
	}

	for _, warning := range warnings {
		if warning.Subject == importPath &&
			strings.HasPrefix(warning.Message, "contains non-Go code that can't be inspected") &&
			strings.HasSuffix(warning.Message, "nongo.s") {
			return
		}
	}

	t.Errorf("Libraries(%q) warnings = %v, want one for its non-Go file", importPath, warnings)
}
