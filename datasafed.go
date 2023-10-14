package main

import (
	"github.com/apecloud/datasafed/cmd"
	_ "github.com/apecloud/datasafed/pkg/certs" // set env for bundled certs
)

func main() {
	cmd.Execute()
}
