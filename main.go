package main

import(
	"os"
	"github.com/JakobRaupach/dotfilesmgr/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
