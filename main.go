package main

import (
	"fmt"
	"myblogx/core"
	"myblogx/flags"
)

func main() {
	flags.Parse()
	fmt.Printf("flags: %+v\n", flags.FlagOptions)
	core.ReadCfg()
}
