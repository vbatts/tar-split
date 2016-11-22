package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	flag.Parse()

	failed := 0
	for _, arg := range flag.Args() {
		cmd := exec.Command("bash", arg)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			failed++
		}
	}
	if failed > 0 {
		fmt.Printf("%d FAILED tests\n", failed)
		os.Exit(1)
	}
}
