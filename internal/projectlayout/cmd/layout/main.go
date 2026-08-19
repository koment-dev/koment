package main

import (
	"fmt"
	"os"

	"github.com/koment-dev/koment/internal/projectlayout"
)

func main() {
	if len(os.Args) != 2 {
		fail("usage: layout check|render")
	}
	root, err := projectlayout.RepositoryRoot(".")
	if err != nil {
		fail(err.Error())
	}
	switch os.Args[1] {
	case "check":
		if err := projectlayout.Check(root); err != nil {
			fail(err.Error())
		}
		fmt.Println("repository layout: ok")
	case "render":
		if err := projectlayout.WriteReference(root); err != nil {
			fail(err.Error())
		}
		fmt.Println(projectlayout.ReferencePath)
	default:
		fail("usage: layout check|render")
	}
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}
