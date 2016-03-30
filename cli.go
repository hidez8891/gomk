package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

import (
	"github.com/hidez8891/gomk/lib/parser"
	"github.com/hidez8891/gomk/lib/runner"
)

// Exit codes are int values that represent an exit code for a particular error.
const (
	ExitCodeOK    int = 0
	ExitCodeError int = 1 + iota
)

// CLI is the command line object
type CLI struct {
	outStream, errStream io.Writer
}

// Run invokes the CLI with the given arguments.
func (cli *CLI) Run(args []string) int {
	var (
		file    string
		version bool
	)

	// Define option flag parse
	flags := flag.NewFlagSet(Name, flag.ContinueOnError)
	flags.SetOutput(cli.errStream)

	flags.StringVar(&file, "f", "", "input makefile")
	flags.BoolVar(&version, "version", false, "Print version information and quit.")

	// Parse commandline flag
	if err := flags.Parse(args[1:]); err != nil {
		fmt.Fprintf(cli.errStream, "%s\n", err)
		return ExitCodeError
	}

	// Show version
	if version {
		fmt.Fprintf(cli.errStream, "%s version %s\n", Name, Version)
		return ExitCodeOK
	}

	// Get makefile path
	file, err := getMakefilePath(file)
	if err != nil {
		fmt.Fprintf(cli.errStream, "Not found %s\n", file)
		return ExitCodeError
	}

	// Open makefile
	reader, err := openMakefile(file)
	if err != nil {
		fmt.Fprintf(cli.errStream, "%s\n", err)
		return ExitCodeError
	}
	defer closeMakefile(reader)

	// Parse makefile
	rules, err := parser.Parse(reader)
	if err != nil {
		fmt.Fprintf(cli.errStream, "%s\n", err)
		return ExitCodeError
	}
	if len(rules.Rules) == 0 {
		fmt.Fprintf(cli.errStream, "Not defined make rule\n")
		return ExitCodeError
	}

	// Get targets
	targets := flags.Args()
	if len(targets) == 0 {
		targets = rules.Firsts
	}

	// Run targets
	for _, target := range targets {
		if err := cli.runRules(rules.Rules, target); err != nil {
			fmt.Fprintf(cli.errStream, "%s\n", err)
			return ExitCodeError
		}
	}

	return ExitCodeOK
}

func getMakefilePath(file string) (path string, err error) {
	if file != "" {
		path = file
	} else {
		path = "Makefile"
	}

	_, err = os.Stat(path)
	return
}

func openMakefile(path string) (*os.File, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return fd, nil
}

func closeMakefile(fd *os.File) {
	fd.Close()
}

func (cli *CLI) runRules(rules map[string][]string, target string) error {
	cmds, ok := rules[target]
	if !ok {
		return errors.New("Not found make rule " + target)
	}

	depends := strings.Fields(cmds[0])
	cmds = cmds[1:]

	for _, depend := range depends {
		if err := cli.runRules(rules, depend); err != nil {
			return err
		}
	}

	runner := runner.New(cli.outStream, cli.errStream)
	for _, cmd := range cmds {
		fmt.Fprintf(cli.outStream, "%s\n", cmd)
		if err := runner.Run(cmd); err != nil {
			return err
		}
	}

	return nil
}
