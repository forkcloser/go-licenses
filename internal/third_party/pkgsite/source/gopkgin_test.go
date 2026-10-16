// Copyright 2026 Farcloser
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package source //nolint:testpackage // gopkgInInfo is unexported

import "testing"

func TestGopkgInInfo(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		modulePath, version  string
		wantRepo, wantCommit string // empty wantRepo: not resolved statically
	}{
		{"gopkg.in/yaml.v2", "v2.4.0", "https://github.com/go-yaml/yaml", "v2.4.0"},
		{"gopkg.in/ini.v1", "v1.62.0", "https://github.com/go-ini/ini", "v1.62.0"},
		{"gopkg.in/boltdb/bolt.v1", "v1.3.0", "https://github.com/boltdb/bolt", "v1.3.0"},
		{"gopkg.in/check.v1", "v1.0.0-20201130134442-10cb98267c6c", "https://github.com/go-check/check", "10cb98267c6c"},
		{"gopkg.in/src-d/go-git.v4", "v4.13.1+incompatible", "https://github.com/src-d/go-git", "v4.13.1"},
		{"gopkg.in/yaml", "v1.0.0", "", ""},
		{"gopkg.in/yaml.vx", "v1.0.0", "", ""},
		{"gopkg.in/a/b/c.v1", "v1.0.0", "", ""},
		{"gopkg.in/.v1", "v1.0.0", "", ""},
		{"github.com/go-yaml/yaml", "v2.4.0", "", ""},
	} {
		t.Run(tc.modulePath, func(t *testing.T) {
			t.Parallel()

			info, ok := gopkgInInfo(tc.modulePath, tc.version)
			if tc.wantRepo == "" {
				if ok {
					t.Fatalf("gopkgInInfo(%q) = %+v, want no static resolution", tc.modulePath, info)
				}

				return
			}

			if !ok || info.repoURL != tc.wantRepo || info.commit != tc.wantCommit || info.moduleDir != "" {
				t.Fatalf(
					"gopkgInInfo(%q, %q) = %+v, %v, want repo %s at %s",
					tc.modulePath,
					tc.version,
					info,
					ok,
					tc.wantRepo,
					tc.wantCommit,
				)
			}
		})
	}
}
