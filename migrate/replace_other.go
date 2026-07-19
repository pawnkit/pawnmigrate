//go:build !windows

package migrate

import "os"

func atomicReplace(source, target string) error {
	return os.Rename(source, target)
}
