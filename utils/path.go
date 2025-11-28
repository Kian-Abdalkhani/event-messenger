package utils

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func GetProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		if dir == "/" {
			return "", fmt.Errorf("could not find root project path")
		}

		dir = filepath.Dir(dir)
	}

}

func ProjectPath(relativePath ...string) string {
	root, err := GetProjectRoot()
	if err != nil {
		slog.Error("Could not find Project Path")
	}

	rel := filepath.Join(relativePath...)
	return filepath.Join(root, rel)
}
