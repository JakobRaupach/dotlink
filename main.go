package main

import(
	"os"
	"github.com/JakobRaupach/dotlink/internal/dotlink"
)

func main() {
	os.Exit(dotlink.Run(os.Args[1:]))
}
