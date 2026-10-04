package dotlink

import (
	"flag"
	"fmt"
	"path/filepath"
	"os"
	"strings"
	"github.com/JakobRaupach/dotlink/internal/ignore"
)

func (app *appEnv) fromArgs(args []string) error {
	fl := flag.NewFlagSet("dotlink", flag.ContinueOnError)
	fl.StringVar(&app.destroot, "dest", "", "Destination of the symlinks")
	var ignore string
	fl.StringVar(&ignore, "i", "", "Files to ignore")
	fl.BoolVar(&app.verbose, "v", false, "Verbose output")
	fl.BoolVar(&app.quiet, "q", false, "Output only errors")
	fl.BoolVar(&app.modeDelete, "D", false, "Delete symlinks")
	fl.BoolVar(&app.modeReload, "R", false, "Reload symlinks (removes and recreates links)")
	fl.BoolVar(&app.dryRun, "n", false, "Simulating the result. Not touching the file system")
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
	// error can be safely ignored, as '.ignore' is not required to be present
	patterns, _ := ignore.CompileFile(filepath.Join(app.srcroot, ".ignore"))
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
