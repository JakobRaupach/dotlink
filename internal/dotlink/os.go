package dotlink

import (
	"os"
	"fmt"
)

func (app *appEnv) Remove(path string) error {
	if app.dryRun {
		fmt.Fprintf(os.Stdout, "Deleting %v\n", path)
		return nil
	}
	err := os.Remove(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while deleting %v\nError: %v", path, err)
	}
	return err
}
