package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
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
	file, err := makefilePath(file)
	if err != nil {
		fmt.Fprintf(cli.errStream, "Not found %s\n", file)
		return ExitCodeError
	}

	// Get targets
	targets := flags.Args()

	// Parse makefile
	rules, err := cli.parseMakefile(file)
	if err != nil {
		fmt.Fprintf(cli.errStream, "%s\n", err)
		return ExitCodeError
	}

	// if not defined target, set default target
	if len(targets) == 0 {
		targets = searchMapKeys(rules.Targets, 0)
	}

	// Run targets
	for _, target := range targets {
		if err := cli.runRules(rules, target); err != nil {
			fmt.Fprintf(cli.errStream, "%s\n", err)
			return ExitCodeError
		}
	}

	return ExitCodeOK
}

func (cli *CLI) parseMakefile(path string) (*parser.MakeRule, error) {
	reader, err := openMakefile(path)
	if err != nil {
		return nil, err
	}
	defer closeMakefile(reader)

	rules, err := parser.Parse(reader)
	if err != nil {
		return nil, err
	}

	if len(rules.Rules) == 0 {
		return nil, errors.New("Not defined make rule")
	}

	return rules, nil
}

func (cli *CLI) runRules(rules *parser.MakeRule, root string) error {
	pre_time := int64(0)
	pre_target := ""
	at_least_one_running := false

	schedule := cli.makeExecuteSchedule(rules, root)
	if len(schedule) == 0 {
		return nil
	}

	for _, target := range schedule {
		do_execute := true

		target_t, err := modTime(target)
		if err != nil {
			target_t = 0
		}

		id, ok := rules.Targets[target]
		if target_t == 0 && !ok {
			return errors.New("Not found make rule " + target)
		}
		if !ok {
			do_execute = false
		}

		rule := rules.Rules[id]
		if inArray(rule.Depends, pre_target) && pre_time < target_t {
			do_execute = false
		}

		pre_target = target
		pre_time = target_t

		if !do_execute {
			continue
		}

		runner := runner.New(cli.outStream, cli.errStream)
		for _, cmd := range rule.Commands {
			if cmd.NeedEcho {
				fmt.Fprintf(cli.outStream, "%s\n", cmd.Exestr)
			}

			if err := runner.Run(cmd.Exestr); err != nil {
				return err
			}
		}

		at_least_one_running = true
	}

	if !at_least_one_running {
		fmt.Fprintf(cli.outStream, "'%s' is up to date\n", root)
	}

	return nil
}

func (cli *CLI) makeExecuteSchedule(rules *parser.MakeRule, target string) []string {
	return cli.makeExecuteScheduleImpl(rules, target, []string{}, []string{})
}

func (cli *CLI) makeExecuteScheduleImpl(rules *parser.MakeRule, target string, parent, schedule []string) []string {
	inSameRule := func(keys []string, t string) bool {
		id, ok := rules.Targets[t]
		if !ok {
			return false
		}

		for _, key := range keys {
			key_id, _ := rules.Targets[key]
			if key_id == id {
				return true
			}
		}
		return false
	}

	if inSameRule(schedule, target) {
		return schedule
	}

	id, ok := rules.Targets[target]
	if !ok {
		return append(schedule, target)
	}
	rule := rules.Rules[id]

	parent = append(parent, target)
	for _, depend := range rule.Depends {
		if inSameRule(parent, depend) {
			fmt.Fprintf(cli.errStream, "Circular %s <- %s dependency dropped\n", target, depend)
			continue
		}

		schedule = cli.makeExecuteScheduleImpl(rules, depend, parent, schedule)
	}

	return append(schedule, target)
}
