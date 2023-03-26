//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
)

var (
	// Default target to run when none is specified
	// If not set, running mage will list available targets
	Default        = Build
	app     string = "tar-split"
	Stdout         = ourStdout
	Stderr         = ourStderr

	golangcilintVersion = "v1.51.2"
)

// Run all-the-things
func All() error {
	mg.Deps(Vet)
	mg.Deps(Test)
	mg.Deps(Build)
	mg.Deps(Lint)
	return nil
}

// A build step that requires additional params, or platform specific steps for example
func Build() error {
	mg.Deps(InstallDeps)
	fmt.Println("Building...")
	cmd := exec.Command("go", "build", "-v", "-o", app, "./cmd/tar-split")
	cmd.Stdout = Stdout
	cmd.Stderr = Stderr
	return cmd.Run()
}

// Vet the codes
func Vet() error {
	fmt.Println("go vet...")
	cmd := exec.Command("go", "vet", "./...")
	cmd.Stdout = Stdout
	cmd.Stderr = Stderr
	return cmd.Run()
}

// Run the Linters
func Lint() error {
	mg.Deps(InstallToolsLint)
	fmt.Println("Linting...")
	cmd := exec.Command("golangci-lint", "run")
	cmd.Stdout = Stdout
	cmd.Stderr = Stderr
	return cmd.Run()
}

// Run the tests available
func Test() error {
	fmt.Println("Testing...")
	cmd := exec.Command("go", "test", "-v", "./...")
	cmd.Stdout = Stdout
	cmd.Stderr = Stderr
	return cmd.Run()
}

// A custom install step if you need your bin someplace other than go/bin
func Install() error {
	mg.Deps(Build)
	fmt.Println("Installing...")
	return os.Rename(app, "/usr/local/bin/"+app)
}

// Manage your deps, or running package managers.
func InstallDeps() error {
	mg.Deps(Tidy)
	fmt.Println("Installing Deps...")
	cmd := exec.Command("go", "get", "./...")
	cmd.Stdout = Stdout
	cmd.Stderr = Stderr
	return cmd.Run()
}

// Tools used during build/dev/test
func InstallTools() error {
	mg.Deps(InstallToolsLint)
	return nil
}

func InstallToolsLint() error {
	fmt.Println("Installing Deps...")
	cmd := exec.Command("go", "install", "github.com/golangci/golangci-lint/cmd/golangci-lint@"+golangcilintVersion)
	cmd.Stdout = Stdout
	cmd.Stderr = Stderr
	return cmd.Run()
}

// Tidy go modules
func Tidy() error {
	fmt.Println("Tidy up...")
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Stdout = Stdout
	cmd.Stderr = Stderr
	return cmd.Run()
}

// Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll(app)
}
