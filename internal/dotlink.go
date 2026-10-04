package internal

import (
	"os"
	"io/fs"
	"fmt"
	"flag"
	"bufio"
	"path/filepath"
	"strings"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
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
	matcher		gitignore.Matcher
	verbose		bool
	quiet		bool
}


func (app *appEnv) fromArgs(args []string) error {
	fl := flag.NewFlagSet("dotfilesmgr", flag.ContinueOnError)
	fl.StringVar(&app.destroot, "dest", "", "Destination of the symlinks")
	var ignore string
	fl.StringVar(&ignore, "i", "", "Files to ignore")
	fl.BoolVar(&app.verbose, "v", true, "Verbose")
	fl.BoolVar(&app.quiet, "q", false, "Output only errors")
	if err := fl.Parse(args); err != nil {
		return err
	}
	if fl.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "usage : dotfilesmgr [flags] <src>\n")
		return fmt.Errorf("usage : dotfilesmgr [flags] <src>")
	}
	app.srcroot = filepath.Clean(fl.Arg(0))
	if _, err := os.Stat(app.srcroot); err != nil {
		fmt.Fprintf(os.Stderr, "%q is not a valid path\n", app.srcroot)
		flag.Usage()
		return flag.ErrHelp
	}

	if app.destroot == "" {
		app.destroot = filepath.Dir(filepath.Clean(app.srcroot))
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


func loadIgnoreFile(path string, domain []string) ([]gitignore.Pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var ps []gitignore.Pattern
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ps = append(ps, gitignore.ParsePattern(line, domain))
	}
	return ps, sc.Err()
}


func loadIgnoreFlag(ignoreFlag string) []gitignore.Pattern {
	var ps []gitignore.Pattern
	for _, p := range strings.Split(ignoreFlag, ",") {
		ps = append(ps, gitignore.ParsePattern(strings.TrimSpace(p),nil))
	}
	return ps
}


func (app *appEnv) loadIgnoreMatcher(ignoreFlag string) error {
	domain := strings.Split(filepath.ToSlash(app.srcroot), "/")
	patterns, err := loadIgnoreFile(filepath.Join(app.srcroot, ".ignore"), domain)
	if err != nil { return err }
	patterns = append(patterns, loadIgnoreFlag(ignoreFlag)...)

	extra := []string{".ignore", ".git", ".gitignore", "README.*", "LICENSE.*", "RCS", "CVS"}
	for _, line := range extra {
		patterns = append(patterns, gitignore.ParsePattern(line, nil))
	}

	app.matcher = gitignore.NewMatcher(patterns)
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
		symlink := filepath.Join(app.destroot, relPath)

		if app.matcher.Match(strings.Split(filepath.ToSlash(relPath), "/"), false) {
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
