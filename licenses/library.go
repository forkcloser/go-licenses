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
	"context"
	"errors"
	"fmt"
	"go/build"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/sync/errgroup"
	"golang.org/x/tools/go/packages"
)

// Library is a collection of packages covered by the same license file.
type Library struct {
	// LicenseFile is the path of the file containing the library's license.
	LicenseFile string
	// Packages contains import paths for Go packages in this library.
	// It may not be the complete set of all packages in the library.
	Packages []string
	// Parent go module.
	module *Module
	// List of licenses for found at the LicenseFile.
	Licenses []License
}

// PackagesError aggregates all Packages[].Errors into a single error.
type PackagesError struct {
	pkgs []*packages.Package
}

func (e PackagesError) Error() string {
	var str strings.Builder
	fmt.Fprintf(&str, "errors for %q:", e.pkgs)
	packages.Visit(e.pkgs, nil, func(pkg *packages.Package) {
		for _, err := range pkg.Errors {
			fmt.Fprintf(&str, "\n%s: %s", pkg.PkgPath, err)
		}
	})

	return str.String()
}

// A Warning is something Libraries noticed that does not stop it: what it is
// about (a package, a module or a licence file) and what it noticed.
type Warning struct {
	Subject string
	Message string
}

func (w Warning) String() string {
	return w.Subject + ": " + w.Message
}

// Libraries returns the collection of libraries used by this package, directly or transitively.
// A library is a collection of one or more packages covered by the same license file.
// Packages not covered by a license will be returned as individual libraries.
// Standard library packages will be ignored.
func Libraries(
	ctx context.Context,
	classifier Classifier,
	includeTests bool,
	ignoredPaths []string,
	importPaths ...string,
) ([]*Library, []Warning, error) {
	// These are the steps we take to find libraries:
	// 1. we list all modules and all packages
	// 2. for each package, we find a list of candidates
	// 3. we deduplicate all candidates
	// 4. for each candidate, we classify if the candidate is a license file
	// 5. for each package, we select the first candidates that is a license
	//    file & add the package to a list of packages for that license file
	// 6. we return an array of libraries (which are the license files, the
	//    found licenses in that file, all the packages that had that file as
	//    its first candidate and the module in which those packages live)
	cfg := &packages.Config{
		Context: ctx,
		Mode:    packages.NeedImports | packages.NeedDeps | packages.NeedFiles | packages.NeedName | packages.NeedModule,
		Tests:   includeTests,
	}

	rootPkgs, err := packages.Load(cfg, importPaths...)
	if err != nil {
		return nil, nil, err
	}

	vendoredSearch := []*Module{}

	for _, parentPkg := range rootPkgs {
		if parentPkg.Module == nil {
			continue
		}

		module := newModule(parentPkg.Module)
		if module.Dir == "" {
			continue
		}

		vendoredSearch = append(vendoredSearch, module)
	}

	type pkgInfo struct {
		// pkgPath is the import path of the package.
		pkgPath string
		// modulePath is the module path of the package.
		modulePath string

		// pkgDir is the directory containing the package's source code.
		pkgDir string
		// moduleDir is the directory containing the module's source code.
		moduleDir string
	}

	var warnings []Warning

	allModules := map[string]*Module{}
	allPackages := []pkgInfo{}
	{
		pkgErrorOccurred := false

		var loadErrs []error

		packages.Visit(rootPkgs, func(p *packages.Package) bool {
			if len(p.Errors) > 0 {
				pkgErrorOccurred = true
				return false
			}

			if isStdLib(p) {
				// No license requirements for the Go standard library.
				return false
			}

			if includeTests && isTestBinary(p) {
				// A test binary only imports the standard library, so we do not need to check its license.
				// Moreover, Find below will return an error because pkgDir is not under p.Module.Dir
				// as pkgDir is under GOCACHE instead.
				return false
			}

			for _, i := range ignoredPaths {
				if strings.HasPrefix(p.PkgPath, i) {
					// Marked to be ignored.
					return true
				}
			}

			if len(p.OtherFiles) > 0 {
				warnings = append(warnings, Warning{
					Subject: p.PkgPath,
					Message: "contains non-Go code that can't be inspected for further dependencies: " +
						strings.Join(p.OtherFiles, ", "),
				})
			}

			var pkgDir string

			switch {
			case len(p.GoFiles) > 0:
				pkgDir = filepath.Dir(p.GoFiles[0])
			case len(p.CompiledGoFiles) > 0:
				pkgDir = filepath.Dir(p.CompiledGoFiles[0])
			case len(p.OtherFiles) > 0:
				pkgDir = filepath.Dir(p.OtherFiles[0])
			default:
				// This package is empty - nothing to do.
				return true
			}

			if p.Module == nil {
				loadErrs = append(loadErrs, fmt.Errorf(
					"package %s does not have module info: projects outside Go modules are not supported",
					p.PkgPath,
				))

				return false
			}

			module := newModule(p.Module)

			if module.Dir == "" {
				// A known cause is that the module is vendored, so some information is lost.
				isVendored := strings.Contains(pkgDir, string(filepath.Separator)+"vendor"+string(filepath.Separator))
				if !isVendored {
					warnings = append(warnings, Warning{
						Subject: module.Path,
						Message: "module has no directory and is not vendored: its licence search is not bounded by it",
					})
				} else {
					// This is vendored. Handle this known special case.

					// Extra note why we identify a vendored package like this.
					//
					// For a normal package:
					// * if it's not in a module, lib.module == nil
					// * if it's in a module, lib.module.Dir != ""
					// Only vendored modules will have lib.module != nil && lib.module.Path != "" && lib.module.Dir ==
					// "" as far as I know.
					// So the if condition above is already very strict for vendored packages.
					// On top of it, we checked the lib.LicensePath contains a vendor folder in it.
					// So it's rare to have a false positive for both conditions at the same time, although it may
					// happen in theory.
					//
					// These assumptions may change in the future,
					// so we need to keep this updated with go tooling changes.
					for _, parentModule := range vendoredSearch {
						if strings.HasPrefix(pkgDir, parentModule.Dir) {
							module = parentModule
							break
						}
					}

					if module.Dir == "" {
						warnings = append(warnings, Warning{
							Subject: module.Path,
							Message: "vendored, but no module vendors it",
						})
					}
				}
			}

			allPackages = append(allPackages, pkgInfo{
				pkgPath:    p.PkgPath,
				modulePath: module.Path,
				pkgDir:     pkgDir,
				moduleDir:  module.Dir,
			})
			allModules[module.Path] = module

			return true
		}, nil)

		if pkgErrorOccurred {
			return nil, nil, PackagesError{
				pkgs: rootPkgs,
			}
		}

		if len(loadErrs) > 0 {
			return nil, nil, fmt.Errorf("loading direct and transitive dependency packages: %w", errors.Join(loadErrs...))
		}
	}

	pkgCandidates := map[string][]string{}
	allCandidates := map[string]struct{}{}

	for _, pkg := range allPackages {
		candidates, err := FindCandidates(pkg.pkgDir, pkg.moduleDir)
		if err != nil {
			return nil, nil, err
		}

		pkgCandidates[pkg.pkgDir] = candidates
		for _, candidate := range candidates {
			allCandidates[candidate] = struct{}{}
		}
	}

	group, _ := errgroup.WithContext(ctx)
	foundLicenseSlice := make([]struct {
		candidate string
		licenes   []License
		err       error
	}, len(allCandidates))

	counter := 0
	for candidate := range allCandidates {
		idx := counter
		counter++

		group.Go(func() error {
			// A file that fails to parse is reported, and the others still count.
			licenses, err := classifier.Identify(candidate)

			foundLicenseSlice[idx] = struct {
				candidate string
				licenes   []License
				err       error
			}{
				candidate: candidate,
				licenes:   licenses,
				err:       err,
			}

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, nil, err
	}

	foundLicenses := map[string][]License{}

	for _, found := range foundLicenseSlice {
		if found.err != nil {
			warnings = append(warnings, Warning{
				Subject: found.candidate,
				Message: "failed to parse: " + found.err.Error(),
			})

			continue
		}

		if len(found.licenes) == 0 {
			continue
		}

		foundLicenses[found.candidate] = found.licenes
	}

	pkgsByLicense := make(map[string][]pkgInfo)

	for _, pkg := range allPackages {
		candidates := pkgCandidates[pkg.pkgDir]

		bestCandidate := ""

		for _, candidate := range candidates {
			if _, ok := foundLicenses[candidate]; ok {
				bestCandidate = candidate
				break
			}
		}

		pkgsByLicense[bestCandidate] = append(pkgsByLicense[bestCandidate], pkg)
	}

	var libraries []*Library

	for licenseFile, pkgs := range pkgsByLicense {
		if licenseFile == "" {
			// No license for these packages - return each one as a separate library.
			for _, p := range pkgs {
				libraries = append(libraries, &Library{
					Packages: []string{p.pkgPath},
					module:   allModules[p.modulePath],
				})
			}

			continue
		}

		lib := &Library{
			LicenseFile: licenseFile,
			Licenses:    foundLicenses[licenseFile],
			Packages:    make([]string, len(pkgs)),
			module:      allModules[pkgs[0].modulePath],
		}

		for i, p := range pkgs {
			lib.Packages[i] = p.pkgPath
		}

		libraries = append(libraries, lib)
	}

	// Sort libraries to produce a stable result for snapshot diffing.
	sort.Slice(libraries, func(i, j int) bool {
		return libraries[i].Name() < libraries[j].Name()
	})

	return libraries, warnings, nil
}

// Name is the common prefix of the import paths for all of the packages in this library.
func (l *Library) Name() string {
	return commonAncestor(l.Packages)
}

func commonAncestor(paths []string) string {
	if len(paths) == 0 {
		return ""
	}

	if len(paths) == 1 {
		return paths[0]
	}

	sort.Strings(paths)
	min, max := paths[0], paths[len(paths)-1]
	lastSlashIndex := 0

	for i := 0; i < len(min) && i < len(max); i++ {
		if min[i] != max[i] {
			return min[:lastSlashIndex]
		}

		if min[i] == '/' {
			lastSlashIndex = i
		}
	}

	return min
}

func (l *Library) String() string {
	return l.Name()
}

func (l *Library) Version() string {
	if l.module != nil {
		return l.module.Version
	}

	return ""
}

// isStdLib returns true if this package is part of the Go standard library.
func isStdLib(pkg *packages.Package) bool {
	if pkg.Name == "unsafe" {
		// Special case unsafe stdlib, because it does not contain go files.
		return true
	}

	if len(pkg.GoFiles) == 0 {
		return false
	}

	prefix := build.Default.GOROOT

	sep := string(filepath.Separator)
	if !strings.HasSuffix(prefix, sep) {
		prefix += sep
	}

	return strings.HasPrefix(pkg.GoFiles[0], prefix)
}

// isTestBinary returns true iff pkg is a test binary.
func isTestBinary(pkg *packages.Package) bool {
	return strings.HasSuffix(pkg.PkgPath, ".test")
}
