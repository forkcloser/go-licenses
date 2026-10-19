package main

import (
	"errors"
	"slices"
	"testing"

	"github.com/forkcloser/go-licenses/licenses"
)

// TestDisallowedLicenseTypes: the names are the types, whatever their case
// and spacing; a name that is no type is a usage error, where it was a
// message followed by a check with the default set, exit 0.
func TestDisallowedLicenseTypes(t *testing.T) {
	t.Parallel()

	got, err := getDisallowedLicenseTypes([]string{"Forbidden", " restricted ", "unknown"})
	if err != nil {
		t.Fatal(err)
	}

	if want := []licenses.Type{licenses.Forbidden, licenses.Restricted, licenses.Unknown}; !slices.Equal(got, want) {
		t.Errorf("getDisallowedLicenseTypes = %v, want %v", got, want)
	}

	for _, names := range [][]string{{"typo"}, {"forbidden", "typo"}, {""}} {
		got, err := getDisallowedLicenseTypes(names)
		if !errors.Is(err, errUnknownLicenseType) || got != nil {
			t.Errorf("getDisallowedLicenseTypes(%q) = %v, %v; want errUnknownLicenseType", names, got, err)
		}
	}

	if got, err := getDisallowedLicenseTypes(nil); err != nil || len(got) != 0 {
		t.Errorf("getDisallowedLicenseTypes(nil) = %v, %v; want none", got, err)
	}
}
