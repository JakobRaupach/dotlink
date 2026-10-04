package dotlink

import (
	"os"
	"fmt"
)

func (app *appEnv) Symlink(target, symlink string) error {
	if app.dryRun {
		fmt.Fprintf(os.Stdout, "Linking %v -> %v\n", symlink, target)
		return nil
	}
	err := os.Symlink(target, symlink)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while creating a symlink\n%v -> %v\nError: %v", symlink, target, err)
	}
	return err
}


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
