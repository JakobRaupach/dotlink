package main

import(
	"os"
	"github.com/JakobRaupach/dotlink/internal/dotlink"
)

func main() {
	os.Exit(internal.Run(os.Args[1:]))
}
