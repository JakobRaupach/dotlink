package dotlink

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"os"
)

func (app *appEnv) link() error {
	app.printVerb(fmt.Sprintf("walking through the filesystem with root: %q", app.srcroot))
	filepath.WalkDir(app.srcroot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error occured while reading file/dir %v", path)
		}
		relPath, err := filepath.Rel(app.srcroot, path)
		if err != nil { return err }
		if relPath == "." {
			return nil
		}
		symlink := filepath.Join(app.destroot, relPath)

		if app.matcher.Ignored(filepath.ToSlash(relPath), d.IsDir()) {
			if d.IsDir() {
				app.printVerb("ignoring file %v", path)
				return fs.SkipDir
			}
			app.printVerb("ignoring file %v", path)
			return nil
		}

		if d.IsDir() {
			_, err := os.Stat(symlink)
			if err == nil {
				// Directory already exists. Do nothing...
				return nil
			}
			// Directory does not exists so create a symlink
			app.printVerb("creating a symlink for directory %v -> %v", symlink, path)
			os.Symlink(path, symlink)
			// Skip the directory contents as it is already symlinked
			return fs.SkipDir
		}
		app.printVerb(fmt.Sprintf("creating a symlink %v -> %v", symlink, path))
		os.Symlink(path, symlink)
		return nil
	})
	return nil
}

