package dotlink

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"os"
	"errors"
)

type link struct {
	src, dest	string
	isDir		bool
}


type conflict struct {
	dest, reason	string
}


func (app *appEnv) planLink() ([]link, []conflict, error) {
	var links []link
	var conflicts []conflict

	err := filepath.WalkDir(app.srcroot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			conflicts = append(conflicts, conflict{path, err.Error() })
			return nil
		}
		rel, err := filepath.Rel(app.srcroot, path)
		if rel == "." {
			return nil
		}


		if app.matcher.Ignored(filepath.ToSlash(rel), d.IsDir()) {
			if d.IsDir() {
				app.printVerb("ignoring directory %v", path)
				return fs.SkipDir
			}
			app.printVerb("ignoring file %v", path)
			return nil
		}
		dest := filepath.Join(app.destroot, rel)
		destInfo, err := os.Lstat(dest)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			links = append(links, link{src: path, dest: dest, isDir: d.IsDir()})
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		case err != nil:
			conflicts = append(conflicts, conflict{dest, err.Error()})
		}

		if destInfo.Mode()&fs.ModeSymlink != 0 {
			if sameTarget(dest, path) {
				app.printVerb("already linked: %v", dest)
			} else {
				target, _ := os.Readlink(dest)
				conflicts = append(conflicts, conflict{dest, "is a symlink to " + target})
			}
			if d.IsDir() {
				return fs.SkipDir;
			}
			return nil
		}

		if d.IsDir() && destInfo.IsDir() {
			return nil
		}

		if d.IsDir() {
			conflicts = append(conflicts, conflict{dest, "is a file but src is a directory"})
			return fs.SkipDir;
		}
		conflicts = append(conflicts, conflict{dest, "already exists but is not a symlink"})
		return nil
	})
	return links, conflicts, err
}


func sameTarget(src, dest string) bool {
	a, err := os.Stat(dest)
	if err != nil {
		return false
	}
	b, err := os.Stat(src)
	if err != nil {
		return false
	}
	return os.SameFile(a, b)
}



func (app *appEnv) link() error {
	links, conflicts, err := app.planLink()
	if err != nil {
		return err
	}

	if len(conflicts) > 0 {
		for _, c := range conflicts {
			fmt.Fprintf(os.Stderr, "conflict: %s, %s\n", c.dest, c.reason)
		}
		return fmt.Errorf("%d conflict(s), nothing was linked", len(conflicts))
	}

	for _, l := range links {
		app.printVerb("linking %v -> %v", l.dest, l.src)
		if app.dryRun {
			continue
		}
		if err := os.Symlink(l.src, l.dest); err != nil {
			return err
		}

	}
	return nil
}

