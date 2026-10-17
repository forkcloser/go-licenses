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
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/mycophonic/primordium/app/shutdown"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/forkcloser/go-licenses/licenses"
)

// errDisallowed is a check that found licences it does not allow. The
// findings are printed already; main exits 1 without another message.
var errDisallowed = shutdown.ExitError(1)

type checkCmd struct {
	AllowedLicenses []string `help:"list of allowed license names, can't be used in combination with disallowed_types"                                  name:"allowed_licenses"`
	DisallowedTypes []string `help:"list of disallowed license types, can't be used in combination with allowed_licenses (default: forbidden, unknown)" name:"disallowed_types"`

	Packages []string `arg:"" help:"Go packages to check." name:"package"`
}

// Help is the long help, after the flags.
func (*checkCmd) Help() string {
	return packageHelp
}

func (c *checkCmd) Run(ctx context.Context, root *cli) error {
	var disallowedLicenseTypes []licenses.Type

	allowedLicenseNames := getAllowedLicenseNames(c.AllowedLicenses)
	disallowedLicenseTypes = getDisallowedLicenseTypes(c.DisallowedTypes)

	hasLicenseNames := len(allowedLicenseNames) > 0
	hasLicenseType := len(disallowedLicenseTypes) > 0

	if hasLicenseNames && hasLicenseType {
		return errors.New("allowed_licenses && disallowed_types can't be used at the same time")
	}

	if !hasLicenseNames && !hasLicenseType {
		// fallback to original behaviour to avoid breaking changes
		disallowedLicenseTypes = []licenses.Type{licenses.Forbidden, licenses.Unknown}
		hasLicenseType = true
	}

	classifier, err := licenses.NewClassifier()
	if err != nil {
		return err
	}

	libs, warnings, err := licenses.Libraries(ctx, classifier, root.IncludeTests, root.Ignore, c.Packages...)
	if err != nil {
		return err
	}

	logWarnings(ctx, warnings)

	// indicate that a forbidden license was found
	found := false

	for _, lib := range libs {
		if lib.LicenseFile == "" {
			fmt.Fprintf(os.Stderr, "Did not find license for library '%v'.\n", lib)

			found = true

			continue
		}

		for _, license := range lib.Licenses {
			if hasLicenseNames && !isAllowedLicenseName(license.Name, allowedLicenseNames) {
				fmt.Fprintf(os.Stderr, "Not allowed license '%s' found for library '%v'.\n", license.Name, lib)

				found = true
			} else if hasLicenseType && isDisallowedLicenseType(license.Type, disallowedLicenseTypes) {
				fmt.Fprintf(
					os.Stderr,
					"License '%s' of not allowed license type '%s' found for library '%v'.\n",
					license.Name,
					cases.Title(language.English).String(license.Type.String()),
					lib,
				)

				found = true
			}
		}
	}

	if found {
		return errDisallowed
	}

	return nil
}

func getDisallowedLicenseTypes(disallowedTypes []string) []licenses.Type {
	if len(disallowedTypes) == 0 {
		return []licenses.Type{}
	}

	excludedLicenseTypes := make([]licenses.Type, 0)

	for _, v := range disallowedTypes {
		switch strings.TrimSpace(strings.ToLower(v)) {
		case "forbidden":
			excludedLicenseTypes = append(excludedLicenseTypes, licenses.Forbidden)
		case "notice":
			excludedLicenseTypes = append(excludedLicenseTypes, licenses.Notice)
		case "permissive":
			excludedLicenseTypes = append(excludedLicenseTypes, licenses.Permissive)
		case "reciprocal":
			excludedLicenseTypes = append(excludedLicenseTypes, licenses.Reciprocal)
		case "restricted":
			excludedLicenseTypes = append(excludedLicenseTypes, licenses.Restricted)
		case "unencumbered":
			excludedLicenseTypes = append(excludedLicenseTypes, licenses.Unencumbered)
		case "unknown":
			excludedLicenseTypes = append(excludedLicenseTypes, licenses.Unknown)
		default:
			fmt.Fprintf(
				os.Stderr,
				"Unknown license type '%s' provided.\n"+
					"Allowed types: forbidden, notice, permissive, reciprocal, restricted, unencumbered, unknown\n",
				v,
			)
		}
	}

	return excludedLicenseTypes
}

func isDisallowedLicenseType(licenseType licenses.Type, excludedLicenseTypes []licenses.Type) bool {
	return slices.Contains(excludedLicenseTypes, licenseType)
}

func getAllowedLicenseNames(allowedLicenses []string) []string {
	if len(allowedLicenses) == 0 {
		return []string{}
	}

	var allowed []string

	for _, licenseName := range allowedLicenses {
		allowed = append(allowed, strings.TrimSpace(licenseName))
	}

	return allowed
}

func isAllowedLicenseName(licenseName string, allowedLicenseNames []string) bool {
	return slices.Contains(allowedLicenseNames, licenseName)
}

// logWarnings logs what Libraries noticed and went past.
func logWarnings(ctx context.Context, warnings []licenses.Warning) {
	for _, warning := range warnings {
		slog.WarnContext(ctx, "dependency warning", "subject", warning.Subject, "warning", warning.Message)
	}
}
