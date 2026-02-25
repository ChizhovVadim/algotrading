package cli

import (
	"os/user"
	"path/filepath"
	"strings"
)

func MapPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		curUser, err := user.Current()
		if err != nil {
			return path
		}
		return filepath.Join(curUser.HomeDir, strings.TrimPrefix(path, "~/"))
	}
	return path
}
