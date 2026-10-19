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
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFindCandidates(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Cannot get working directory: %v", err)
	}

	for _, test := range []struct {
		desc                      string
		dir                       string
		rootDir                   string
		wantLicensePathCandidates []string
	}{
		{
			desc: "licenSe",
			dir:  "testdata",
			wantLicensePathCandidates: []string{
				filepath.Join(wd, "testdata/LICENSE"),
			},
		},
		{
			desc: "licenCe",
			dir:  "testdata/licence",
			wantLicensePathCandidates: []string{
				filepath.Join(wd, "testdata/licence/LICENCE"),
				filepath.Join(wd, "testdata/LICENSE"),
			},
		},
		{
			desc: "LICENSE.MIT",
			dir:  "testdata/MIT",
			wantLicensePathCandidates: []string{
				filepath.Join(wd, "testdata/MIT/LICENSE.MIT"),
				filepath.Join(wd, "testdata/LICENSE"),
			},
		},
		{
			desc: "COPYING",
			dir:  "testdata/copying",
			wantLicensePathCandidates: []string{
				filepath.Join(wd, "testdata/copying/COPYING"),
				filepath.Join(wd, "testdata/LICENSE"),
			},
		},
		{
			desc: "NOTICE",
			dir:  "testdata/notice",
			wantLicensePathCandidates: []string{
				filepath.Join(wd, "testdata/notice/NOTICE.txt"),
				filepath.Join(wd, "testdata/LICENSE"),
			},
		},
		{
			desc: "README",
			dir:  "testdata/readme",
			wantLicensePathCandidates: []string{
				filepath.Join(wd, "testdata/readme/README.md"),
				filepath.Join(wd, "testdata/LICENSE"),
			},
		},
		{
			desc: "parent dir",
			dir:  "testdata/internal",
			wantLicensePathCandidates: []string{
				filepath.Join(wd, "testdata/LICENSE"),
			},
		},
		{
			desc: "lowercase",
			dir:  "testdata/lowercase",
			wantLicensePathCandidates: []string{
				filepath.Join(wd, "testdata/lowercase/license"),
				filepath.Join(wd, "testdata/LICENSE"),
			},
		},
		{
			desc: "license-apache-2.0.txt",
			dir:  "testdata/license-apache-2.0",
			wantLicensePathCandidates: []string{
				filepath.Join(wd, "testdata/license-apache-2.0/LICENSE-APACHE-2.0.txt"),
				filepath.Join(wd, "testdata/LICENSE"),
			},
		},
		{
			desc:    "proprietary-license",
			dir:     "testdata/proprietary-license",
			rootDir: "testdata/proprietary-license",
			wantLicensePathCandidates: []string{
				filepath.Join(wd, "testdata/proprietary-license/LICENSE"),
			},
		},
		{
			desc: "UNLICENSE",
			dir:  "testdata/unlicense",
			wantLicensePathCandidates: []string{
				filepath.Join(wd, "testdata/unlicense/UNLICENSE"),
				filepath.Join(wd, "testdata/LICENSE"),
			},
		},
	} {
		t.Run(test.desc, func(t *testing.T) {
			if test.rootDir == "" {
				test.rootDir = "./testdata"
			}

			licensePathCandidates, err := FindCandidates(test.dir, test.rootDir)
			if err != nil {
				t.Fatalf(
					"FindCandidates(%q) = (%#v, %q), want (%q, nil)",
					test.dir,
					licensePathCandidates,
					err,
					test.wantLicensePathCandidates,
				)
			}

			if diff := cmp.Diff(test.wantLicensePathCandidates, licensePathCandidates); diff != "" {
				t.Fatalf(
					"FindCandidates(%q) = %q, %q, want (%q, nil); diff (-want +got): %s",
					test.dir,
					licensePathCandidates,
					err,
					test.wantLicensePathCandidates,
					diff,
				)
			}
		})
	}
}

// TestFindCandidatesBound: the search stays in rootDir as a path, not as a
// string prefix. A directory in a sibling whose name starts with the root's
// (m@v1.2.0-rc1 beside m@v1.2.0 in the module cache) is outside it and
// refused; a directory in the root finds the root's license and not one
// above it.
func TestFindCandidatesBound(t *testing.T) {
	base := t.TempDir()

	for _, dir := range []string{"root/x", "root2/x"} {
		if err := os.MkdirAll(filepath.Join(base, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	for _, file := range []string{"LICENSE", "root/LICENSE", "root2/LICENSE"} {
		if err := os.WriteFile(filepath.Join(base, file), []byte("license"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	root := filepath.Join(base, "root")

	if got, err := FindCandidates(filepath.Join(base, "root2", "x"), root); err == nil {
		t.Errorf("FindCandidates of a sibling sharing the root's name prefix = %q, want an error", got)
	}

	got, err := FindCandidates(filepath.Join(root, "x"), root)
	if err != nil {
		t.Fatal(err)
	}

	if want := []string{filepath.Join(root, "LICENSE")}; !cmp.Equal(got, want) {
		t.Errorf("FindCandidates in the root = %q, want %q (the root's license, not the one above it)", got, want)
	}
}
