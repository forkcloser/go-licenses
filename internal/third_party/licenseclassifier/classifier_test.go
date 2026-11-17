package licenseclassifier_test

import (
	"strings"
	"testing"

	"github.com/forkcloser/go-licenses/internal/third_party/licenseclassifier/assets"
)

// TestMatchKeepsNoCopyrightLineALicenseHolds: a copyright line inside a
// license's text is the license's, and is not reported beside it, whether
// the license starts the text or follows other words.
func TestMatchKeepsNoCopyrightLineALicenseHolds(t *testing.T) {
	t.Parallel()

	classifier, err := assets.DefaultClassifier()
	if err != nil {
		t.Fatal(err)
	}

	const license = "Redistribution and use in source and binary forms, with or without\n" +
		"modification, are permitted provided that the following conditions are met:\n" +
		"Copyright (c) 2020 The Authors\n" +
		"1. Redistributions of source code must retain the above copyright notice, this\n" +
		"list of conditions and the following disclaimer.\n" +
		"2. Redistributions in binary form must reproduce the above copyright notice,\n" +
		"this list of conditions and the following disclaimer in the documentation\n" +
		"and/or other materials provided with the distribution.\n" +
		"THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS \"AS IS\" AND\n" +
		"ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED\n" +
		"WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE\n" +
		"DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR\n" +
		"ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES\n" +
		"(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;\n" +
		"LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND\n" +
		"ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT\n" +
		"(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS\n" +
		"SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.\n"

	for name, text := range map[string]string{
		"the license first":       license,
		"the license after words": "Words of a project that come before its license.\n" + license,
	} {
		results, err := classifier.MatchFrom(strings.NewReader(text))
		if err != nil {
			t.Fatal(err)
		}

		var got []string
		for _, match := range results.Matches {
			got = append(got, match.MatchType+" "+match.Name)
		}

		if len(got) != 1 || got[0] != "License BSD-2-Clause" {
			t.Errorf("%s: %v, want the license alone", name, got)
		}
	}
}
