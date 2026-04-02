package main

import (
	"os"

	"github.com/GoFurry/awesome-fiber-template/v3/extra-light/cmd"
)

func main() {
	os.Exit(cmd.Execute(os.Args[1:]))
}
