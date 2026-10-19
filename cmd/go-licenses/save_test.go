package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/forkcloser/go-licenses/licenses"
)

// TestBadLicenses: a library with no license identified is Unknown, one
// with a forbidden license is Forbidden, and the shareable ones are not
// listed; the save judges this before it removes or writes anything.
func TestBadLicenses(t *testing.T) {
	t.Parallel()

	none := &licenses.Library{Packages: []string{"m/none"}}
	forbidden := &licenses.Library{
		Packages: []string{"m/forbidden"},
		Licenses: []licenses.License{{Name: "x", Type: licenses.Forbidden}},
	}
	notice := &licenses.Library{
		Packages: []string{"m/notice"},
		Licenses: []licenses.License{{Name: "MIT", Type: licenses.Notice}},
	}

	got := badLicenses([]*licenses.Library{none, forbidden, notice})

	if len(got) != 2 || len(got[licenses.Unknown]) != 1 || got[licenses.Unknown][0] != none ||
		len(got[licenses.Forbidden]) != 1 || got[licenses.Forbidden][0] != forbidden {
		t.Errorf("badLicenses = %v", got)
	}

	if got := badLicenses([]*licenses.Library{notice}); len(got) != 0 {
		t.Errorf("badLicenses of a notice library = %v, want none", got)
	}
}

// TestSaveJudgesBeforeItRemoves: save judges every library before it
// touches the save path. With --force and a forbidden library, the
// previous save stands, marker and all; with --force and a notice
// library, the previous save is replaced by the notice; without --force an
// existing save path is refused.
func TestSaveJudgesBeforeItRemoves(t *testing.T) {
	t.Parallel()

	src := t.TempDir()
	licenseFile := filepath.Join(src, "LICENSE")

	if err := os.WriteFile(licenseFile, []byte("MIT"), 0o644); err != nil {
		t.Fatal(err)
	}

	forbidden := &licenses.Library{
		Packages:    []string{"m/forbidden"},
		LicenseFile: licenseFile,
		Licenses:    []licenses.License{{Name: "x", Type: licenses.Forbidden}},
	}
	notice := &licenses.Library{
		Packages:    []string{"m/notice"},
		LicenseFile: licenseFile,
		Licenses:    []licenses.License{{Name: "MIT", Type: licenses.Notice}},
	}

	previous := func() (string, string) {
		savePath := filepath.Join(t.TempDir(), "save")
		marker := filepath.Join(savePath, "marker")

		if err := os.MkdirAll(savePath, 0o755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(marker, []byte("previous"), 0o644); err != nil {
			t.Fatal(err)
		}

		return savePath, marker
	}

	savePath, marker := previous()

	failed := save([]*licenses.Library{notice, forbidden}, savePath, true)
	if failed == nil || !strings.Contains(failed.Error(), "incompatible/unknown license") {
		t.Fatalf("save of a forbidden library with --force: %v; want the license error", failed)
	}

	if _, err := os.Stat(marker); err != nil {
		t.Errorf("the previous save did not survive a failing forced save: %v", err)
	}

	savePath, marker = previous()

	if err := save([]*licenses.Library{notice}, savePath, true); err != nil {
		t.Fatalf("forced save of a notice library: %v", err)
	}

	if _, err := os.Stat(marker); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the previous save's marker after a forced save: %v; want gone", err)
	}

	if _, err := os.Stat(filepath.Join(savePath, "m", "notice", "LICENSE")); err != nil {
		t.Errorf("the notice after a forced save: %v", err)
	}

	savePath, _ = previous()

	refused := save([]*licenses.Library{notice}, savePath, false)
	if refused == nil || !strings.Contains(refused.Error(), "already exists") {
		t.Errorf("save over an existing path without --force: %v; want refused", refused)
	}
}
