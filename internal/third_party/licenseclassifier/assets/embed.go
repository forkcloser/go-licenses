// Copyright 2022 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package assets holds the licence texts the classifier matches against.
package assets

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"

	"github.com/forkcloser/go-licenses/internal/third_party/licenseclassifier"
)

//go:embed */*/*
var licenseFS embed.FS

// threshold is the confidence a match needs to be reported.
const threshold = 0.8

// DefaultClassifier returns a classifier loaded with the contents of the
// assets directory.
func DefaultClassifier() (*licenseclassifier.Classifier, error) {
	c := licenseclassifier.NewClassifier(threshold)

	err := fs.WalkDir(licenseFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		b, err := licenseFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		splits := strings.Split(path, "/")
		category, name, variant := splits[0], splits[1], splits[2]

		return c.AddContent(category, name, variant, b)
	})
	if err != nil {
		return nil, fmt.Errorf("loading the licence texts: %w", err)
	}

	return c, nil
}
