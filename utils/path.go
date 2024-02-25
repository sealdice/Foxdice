package utils

import (
	"path"
)

func DataDir(dirs ...string) string {
	p := []string{"data"}
	p = append(p, dirs...)
	return path.Join(p...)
}
