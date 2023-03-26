//go:build mage
// +build mage

package main

import (
	"io"
	"os"

	"github.com/fatih/color"
)

var (
	ourStdout = cw{c: color.New(color.FgGreen), o: os.Stdout}
	ourStderr = cw{c: color.New(color.FgRed), o: os.Stderr}
)

// hack around color.Color not implementing Write()
type cw struct {
	c *color.Color
	o io.Writer
}

func (cw cw) Write(p []byte) (int, error) {
	i := len(p)
	_, err := cw.c.Fprint(cw.o, string(p)) // discarding the number of bytes written for now...
	return i, err
}
