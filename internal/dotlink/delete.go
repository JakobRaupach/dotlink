package dotlink

import (
	"os"
	"io/fs"
	"fmt"
	"path/filepath"
)

func (app *appEnv) delete() error {
	filepath.WalkDir(app.srcroot, func (path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error occoured while reading the file %v\n", path)
			return err
		}
		rel, err := filepath.Rel(app.srcroot, path)
		if err != nil { return err }
		if rel == "." { return nil }

		dest := filepath.Join(app.destroot, rel)
		fi, err := os.Lstat(dest)
		if err != nil { return nil }
		// not a symlink
		if fi.Mode()&os.ModeSymlink == 0 {
			return nil
		}
		target, _ := os.Readlink(dest)
		if !filepath.IsAbs(target) { 
			target = filepath.Join(filepath.Dir(dest), target)
		}
		if target == path {
			app.Remove(dest)
		}
		return nil
	})
	return nil

}
