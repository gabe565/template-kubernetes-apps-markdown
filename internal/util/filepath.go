package util

import "path/filepath"

func IsYAMLPath(s string) bool {
	ext := filepath.Ext(s)
	return ext == ".yaml" || ext == ".yml"
}
