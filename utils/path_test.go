package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetRootDir(t *testing.T) {
	root := GetRootDir()
	if runtime.GOOS == "windows" {
		// Should be something like "C:\"
		if !strings.HasSuffix(root, string(os.PathSeparator)) || !strings.Contains(root, ":") {
			t.Errorf("Windows root dir invalid: %s", root)
		}
	} else {
		// Should be "/"
		if root != string(os.PathSeparator) {
			t.Errorf("Unix root dir invalid: %s", root)
		}
	}
}

func TestProjectRoot(t *testing.T) {
	cwd, _ := os.Getwd()
	tests := []struct {
		dir     string
		wantErr bool
	}{
		{"/", true},
		{cwd, false},
	}

	for _, tt := range tests {
		t.Run(tt.dir, func(t *testing.T) {
			_ = os.Chdir(tt.dir)
			_, err := GetProjectRoot()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProjectRoot(%s) error = %v, wantErr %v", tt.dir, err, tt.wantErr)
			}
		})
	}
}

func TestProjectPath(t *testing.T) {
	cwd, _ := os.Getwd()
	tests := []struct {
		relativePath []string
		expectedPath string
	}{
		{[]string{"data", "uploads"}, filepath.Join(filepath.Dir(cwd), "data", "uploads")},
		{[]string{"data", "app.db"}, filepath.Join(filepath.Dir(cwd), "data", "app.db")},
	}

	for _, tt := range tests {
		t.Run(filepath.Join(tt.relativePath...), func(t *testing.T) {
			path := ProjectPath(tt.relativePath...)
			if path != tt.expectedPath {
				t.Errorf("ProjectPath(%s) Path=%s Expected Path=%s", strings.Join(tt.relativePath, string(os.PathSeparator)), path, tt.expectedPath)
			}
		})
	}
}
