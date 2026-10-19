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
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/otiai10/copy"

	"github.com/forkcloser/go-licenses/licenses"
)

var noticeRegexp = regexp.MustCompile(`^NOTICE(\.(txt|md))?$`)

type saveCmd struct {
	// SavePath is where the output of the command is written to.
	SavePath string `help:"Directory into which files should be saved that are required by license terms" name:"save_path" required:""`
	// Force replaces the directory SavePath names when it already exists;
	// without it, the command fails.
	Force bool `help:"Delete the destination directory if it already exists."`

	Packages []string `arg:"" help:"Go packages whose dependencies' files to save." name:"package"`
}

// Help is the long help, after the flags.
func (*saveCmd) Help() string {
	return packageHelp
}

func (c *saveCmd) Run(ctx context.Context, root *cli) error {
	savePath := c.SavePath

	classifier, err := licenses.NewClassifier()
	if err != nil {
		return err
	}

	libs, warnings, err := licenses.Libraries(ctx, classifier, root.IncludeTests, root.Ignore, c.Packages...)
	if err != nil {
		return err
	}

	logWarnings(ctx, warnings)

	return save(libs, savePath, c.Force)
}

// save writes the libraries' licenses, or sources, under savePath. Every
// library is judged before anything is removed or written, so a run that
// fails on a license leaves the previous save as it was; only then is the
// previous save removed when forced, and refused when not.
func save(libs []*licenses.Library, savePath string, force bool) error {
	if libsWithBadLicenses := badLicenses(libs); len(libsWithBadLicenses) > 0 {
		return fmt.Errorf("one or more libraries have an incompatible/unknown license: %q", libsWithBadLicenses)
	}

	if force {
		if err := os.RemoveAll(savePath); err != nil {
			return err
		}
	}

	// Check that the save path doesn't exist, otherwise it'd end up with a mix of
	// existing files and the output of this command.
	if _, err := os.Stat(savePath); err == nil {
		return fmt.Errorf("%s already exists", savePath)
	} else if !os.IsNotExist(err) {
		return err
	}

	for _, lib := range libs {
		libSaveDir := filepath.Join(savePath, unvendor(lib.Name()))

		switch restrictiveness(lib) {
		case licenses.RestrictionsShareCode:
			// Copy the entire source directory for the library.
			libDir := filepath.Dir(lib.LicenseFile)
			if err := copySrc(libDir, libSaveDir); err != nil {
				return err
			}
		case licenses.RestrictionsShareLicense:
			// Just copy the license and copyright notice.
			if err := copyNotices(lib.LicenseFile, libSaveDir); err != nil {
				return err
			}
		}
	}

	return nil
}

// restrictiveness is what a library's licenses together ask of a save.
func restrictiveness(lib *licenses.Library) licenses.LicenseRestrictiveness {
	licenseTypes := make([]licenses.Type, 0, len(lib.Licenses))
	for _, license := range lib.Licenses {
		licenseTypes = append(licenseTypes, license.Type)
	}

	return licenses.LicenseTypeRestrictiveness(licenseTypes...)
}

// badLicenses is every library a save cannot fulfil, by license type: one
// with no license identified (Unknown), and one whose licenses include a
// type that is neither shared as code nor as a notice.
func badLicenses(libs []*licenses.Library) map[licenses.Type][]*licenses.Library {
	libsWithBadLicenses := make(map[licenses.Type][]*licenses.Library)

	for _, lib := range libs {
		switch restrictiveness(lib) {
		case licenses.RestrictionsShareCode, licenses.RestrictionsShareLicense:
			continue
		}

		if len(lib.Licenses) == 0 {
			// If we can't identify the license, we can't fulfill its requirements.
			libsWithBadLicenses[licenses.Unknown] = append(libsWithBadLicenses[licenses.Unknown], lib)

			continue
		}

		// Register all bad licenses, so we can print them out at the end.
		for _, license := range lib.Licenses {
			switch license.Type {
			case licenses.Notice, licenses.Permissive, licenses.Unencumbered, licenses.Restricted, licenses.Reciprocal:
				// these are allowed
				continue
			}

			libsWithBadLicenses[license.Type] = append(libsWithBadLicenses[license.Type], lib)
		}
	}

	return libsWithBadLicenses
}

func copySrc(src, dest string) error {
	// Skip the .git directory for copying, if it exists, since we don't want to save the user's
	// local Git config along with the source code.
	opt := copy.Options{
		Skip: func(_ os.FileInfo, src, dest string) (bool, error) {
			return strings.HasSuffix(src, ".git"), nil
		},
		AddPermission: 0o600,
	}
	if err := copy.Copy(src, dest, opt); err != nil {
		return err
	}

	return nil
}

func copyNotices(licensePath, dest string) error {
	if err := copy.Copy(licensePath, filepath.Join(dest, filepath.Base(licensePath))); err != nil {
		return err
	}

	src := filepath.Dir(licensePath)

	files, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, f := range files {
		if fName := f.Name(); !f.IsDir() && noticeRegexp.MatchString(fName) {
			if err := copy.Copy(filepath.Join(src, fName), filepath.Join(dest, fName)); err != nil {
				return err
			}
		}
	}

	return nil
}
