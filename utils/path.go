package utils

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func GetRootDir() string {
	wd, _ := os.Getwd()
	vol := filepath.VolumeName(wd)
	if vol != "" {
		return vol + string(os.PathSeparator)
	} else {
		return string(os.PathSeparator)
	}
}

func GetProjectRoot() (string, error) {
	exe, _ := os.Executable()
	// Ensures Project Root paths are found whether built or ran using `go run .`
	if strings.Contains(exe, "/go-build") {
		dir, err := os.Getwd()
		if err != nil {
			return "", err
		}

		for {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				return dir, nil
			}

			if dir == GetRootDir() {
				return "", fmt.Errorf("could not find root project path")
			}

			dir = filepath.Dir(dir)
		}

	} else {
		return filepath.Dir(exe), nil
	}

}

func ProjectPath(relativePath ...string) string {
	base, err := GetProjectRoot()
	if err != nil {
		slog.Error("Error finding project root", "error", err)
	}

	return filepath.Join(base, filepath.Join(relativePath...))
}

func FolderExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
