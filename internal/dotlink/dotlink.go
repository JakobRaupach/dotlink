package internal

import (
	"os"
	"io/fs"
	"fmt"
	"flag"
	"path/filepath"
	"strings"
	"github.com/JakobRaupach/dotlink/internal/ignore"
)

func Run(args []string) int {
	var app appEnv
	err := app.fromArgs(args)
	if err != nil {
		return 2;
	}
	if err = app.run(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		return 1
	}
	return 0
}

type appEnv struct {
	srcroot 	string
	destroot	string
	matcher		*ignore.Matcher
	verbose		bool
	quiet		bool
}


func (app *appEnv) fromArgs(args []string) error {
	fl := flag.NewFlagSet("dotlink", flag.ContinueOnError)
	fl.StringVar(&app.destroot, "dest", "", "Destination of the symlinks")
	var ignore string
	fl.StringVar(&ignore, "i", "", "Files to ignore")
	fl.BoolVar(&app.verbose, "v", false, "Verbose output")
	fl.BoolVar(&app.quiet, "q", false, "Output only errors")
	if err := fl.Parse(args); err != nil {
		return err
	}
	if fl.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "usage : dotlink [flags] <src>\n")
		return fmt.Errorf("usage : dotlink [flags] <src>")
	}
	srcroot, err := filepath.Abs(fl.Arg(0))
	if err != nil {
		return err
	}
	app.srcroot = srcroot
	if _, err := os.Stat(app.srcroot); err != nil {
		fmt.Fprintf(os.Stderr, "%q is not a valid path\n", app.srcroot)
		flag.Usage()
		return flag.ErrHelp
	}

	if app.destroot == "" {
		app.destroot = filepath.Dir(app.srcroot)
	} else {
		destroot, err := filepath.Abs(app.destroot)
		if err != nil {
			return err
		}
		app.destroot = destroot
	}

	if err := app.loadIgnoreMatcher(ignore); err != nil {
		fmt.Fprintf(os.Stderr, "Could not read ignored files (either of -i or .ignore)\n%v",err)
		flag.Usage()
		return flag.ErrHelp
	}
	app.srcroot = filepath.Clean(app.srcroot)
	app.destroot = filepath.Clean(app.destroot)

	app.printVerb("src: %v; dest: %v;", app.srcroot, app.destroot)
	return nil
}

func loadIgnoreFlag(ignoreFlag string) []ignore.Pattern {
	var ps []ignore.Pattern
	for _, p := range strings.Split(ignoreFlag, ",") {
		if pattern, ok := ignore.Compile(strings.TrimSpace(p)); ok {
			ps = append(ps, pattern)
		}
	}
	return ps
}


func (app *appEnv) loadIgnoreMatcher(ignoreFlag string) error {
	patterns, err := ignore.CompileFile(filepath.Join(app.srcroot, ".ignore"))
	if err != nil { return err }
	patterns = append(patterns, loadIgnoreFlag(ignoreFlag)...)

	extra := []string{".ignore", ".git", ".gitignore", "README.*", "LICENSE.*", "RCS", "CVS"}
	for _, line := range extra {
		if p, ok := ignore.Compile(line); ok {
			patterns = append(patterns, p)
		}
	}

	app.matcher = ignore.NewMatcher(patterns)
	return nil
}


func (app *appEnv) run() error {
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

func (app *appEnv) printVerb(format string, a ...any) {
	if app.verbose {
		fmt.Fprintf(os.Stdout, "[i] %s\n", fmt.Sprintf(format, a...))
	}
}
