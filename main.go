package main

import "github.com/CHTJonas/go-lg/cmd"

// Software version defaults to the value below but is overridden by the compiler in Makefile.
var version = "dev-edge"

func main() {
	cmd.Execute(version)
}
