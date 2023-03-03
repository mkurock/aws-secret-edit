package main

import (
	"os"
	"github.com/mkurock/aws-secret-edit/pkg"
)

func main() {
	var secretName string = ""
	args := os.Args[1:]
	if len(args) == 1 {
		secretName = args[0]
	}
  pkg.Run(secretName)
}
