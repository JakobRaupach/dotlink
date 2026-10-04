package dotlink

import (
	"os"
	"fmt"
	"github.com/JakobRaupach/dotlink/internal/ignore"
)

type appEnv struct {
	srcroot 	string
	destroot	string
	matcher		*ignore.Matcher
	verbose		bool
	quiet		bool
	dryRun		bool
	modeDelete	bool
	modeReload	bool
}


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

func (app *appEnv) run() error {
	if app.modeDelete {
		return app.delete()
	}
	if app.modeReload {
		return app.reload()
	}
	return app.link()
}


func (app *appEnv) printVerb(format string, a ...any) {
	if app.verbose {
		fmt.Fprintf(os.Stdout, "[i] %s\n", fmt.Sprintf(format, a...))
	}
}
