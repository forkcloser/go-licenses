package licenses_test

import (
	"os"
	"path/filepath"
)

func writeFile(path string) error {
	return writeContent(path, "x\n")
}

func writeContent(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(content), 0o600)
}
