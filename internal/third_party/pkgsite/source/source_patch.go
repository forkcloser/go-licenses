// Copyright 2022 Google LLC
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

package source

import "strings"

// This file includes all local additions to source package for google/go-licenses use-cases.

// SetCommit overrides commit to a specified commit. Usually, you should pass your version to
// ModuleInfo(). However, when you do not know the version and just wants a link that points to
// a known commit/branch/tag. You can use this method to directly override the commit like
// info.SetCommit("master").
//
// Note this is different from directly passing "master" as version to ModuleInfo(), because for
// modules not at the root of a repo, there are conventions that add a module's relative dir in
// front of the version as the actual git tag. For example, for a sub module at ./submod whose
// version is v1.0.1, the actual git tag should be submod/v1.0.1.
func (i *Info) SetCommit(commit string) {
	if i == nil {
		return
	}
	i.commit = commit
}

// gopkgInInfo resolves a gopkg.in module without asking gopkg.in, whose
// mapping is fixed (https://labix.org/gopkg.in): gopkg.in/pkg.vN is
// github.com/go-pkg/pkg, gopkg.in/user/pkg.vN is github.com/user/pkg, and the
// module's versions are that repository's tags. The live lookup it replaces
// went through gopkg.in's proxy, which on CI answered 502 after ten seconds or
// not at all (forkcloser/go-licenses#8). It reports false for any other path,
// and for a gopkg.in path outside that scheme, which then takes the lookup.
func gopkgInInfo(modulePath, version string) (*Info, bool) {
	rest, ok := strings.CutPrefix(modulePath, "gopkg.in/")
	if !ok {
		return nil, false
	}

	user, pkg, hasUser := strings.Cut(rest, "/")
	if !hasUser {
		pkg = user
	}

	dot := strings.LastIndex(pkg, ".v")
	if dot <= 0 || strings.Contains(pkg, "/") || !isDigits(pkg[dot+2:]) {
		return nil, false
	}

	name := pkg[:dot]
	if !hasUser {
		user = "go-" + name
	}

	commit, _ := commitFromVersion(version, "")

	return NewGitHubInfo("https://github.com/"+user+"/"+name, "", commit), true
}

func isDigits(s string) bool {
	return s != "" && !strings.ContainsFunc(s, func(r rune) bool { return r < '0' || r > '9' })
}
