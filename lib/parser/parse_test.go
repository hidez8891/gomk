package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
)

func make_parser(r io.Reader) *Parser {
	return &Parser{
		scanner: bufio.NewScanner(r),
		buffer:  []string{},
		varmap:  map[string]string{},
		targets: map[string]int{},
		rules:   []Rule{},
	}
}

func make_rule(depends []string, commands []string) Rule {
	cmds := []Command{}
	for _, cmd := range commands {
		cmds = append(cmds, Command{cmd, true})
	}

	return Rule{depends, cmds}
}

func make_rule2(depends []string, commands []Command) Rule {
	return Rule{depends, commands}
}

func TestRun_readAndParse(t *testing.T) {
	// parser varmap tester
	tester_varmap := func(str string, varmap map[string]string) error {
		r := strings.NewReader(str)
		parser := make_parser(r)

		if err := parser.readAndParse(); err != nil {
			return errors.New(fmt.Sprintf("error happened: %q", err))
		}

		if !reflect.DeepEqual(parser.varmap, varmap) {
			return errors.New(fmt.Sprintf("expected %q to eq %q", parser.varmap, varmap))
		}

		return nil
	}

	// parser rules tester
	tester_rules := func(str string, targets map[string]int, rules []Rule) error {
		r := strings.NewReader(str)
		parser := make_parser(r)

		if err := parser.readAndParse(); err != nil {
			return errors.New(fmt.Sprintf("error happened: %q", err))
		}

		if !reflect.DeepEqual(parser.targets, targets) {
			return errors.New(fmt.Sprintf("expected %q to eq %q", parser.targets, targets))
		}

		if !reflect.DeepEqual(parser.rules, rules) {
			return errors.New(fmt.Sprintf("expected %q to eq %q", parser.rules, rules))
		}

		return nil
	}

	// parser tester
	tester_parser := func(str string, varmap map[string]string, targets map[string]int, rules []Rule) error {
		r := strings.NewReader(str)
		parser := make_parser(r)

		if err := parser.readAndParse(); err != nil {
			return errors.New(fmt.Sprintf("error happened: %q", err))
		}

		if !reflect.DeepEqual(parser.varmap, varmap) {
			return errors.New(fmt.Sprintf("expected %q to eq %q", parser.varmap, varmap))
		}

		if !reflect.DeepEqual(parser.targets, targets) {
			return errors.New(fmt.Sprintf("expected %q to eq %q", parser.targets, targets))
		}

		if !reflect.DeepEqual(parser.rules, rules) {
			return errors.New(fmt.Sprintf("expected %q to eq %q", parser.rules, rules))
		}

		return nil
	}

	// empty input
	str := ""
	varmap := map[string]string{}
	if err := tester_varmap(str, varmap); err != nil {
		t.Error(err)
	}

	// empty input
	str = `
`
	varmap = map[string]string{}
	if err := tester_varmap(str, varmap); err != nil {
		t.Error(err)
	}

	// only commnet
	str = `#comment
`
	varmap = map[string]string{}
	if err := tester_varmap(str, varmap); err != nil {
		t.Error(err)
	}

	// simple assign
	str = `
VAR1 = var1
VAR2 = var2
VAR3 = var3
`
	varmap = map[string]string{
		"VAR1": "var1",
		"VAR2": "var2",
		"VAR3": "var3",
	}
	if err := tester_varmap(str, varmap); err != nil {
		t.Error(err)
	}

	// overwrite assign
	str = `
VAR1 = var1
VAR2 = var2
VAR1 = var3
`
	varmap = map[string]string{
		"VAR1": "var3",
		"VAR2": "var2",
	}
	if err := tester_varmap(str, varmap); err != nil {
		t.Error(err)
	}

	// reference variable assign
	str = `
VAR1 = var1
VAR2 = $(VAR1)
VAR3 := $(VAR1)
VAR1 = var2
`
	varmap = map[string]string{
		"VAR1": "var2",
		"VAR2": "$(VAR1)",
		"VAR3": "var1",
	}
	if err := tester_varmap(str, varmap); err != nil {
		t.Error(err)
	}

	// undefined variable assign
	str = `
VAR1 = $(VAR0)
VAR2 := $(VAR0)
`
	varmap = map[string]string{
		"VAR1": "$(VAR0)",
		"VAR2": "",
	}
	if err := tester_varmap(str, varmap); err != nil {
		t.Error(err)
	}

	// resolve multi reference
	str = `
VAR1 = var1
VAR2 = @$(VAR1)--$(VAR1)++$(VAR1)$
VAR3 := @$(VAR1)--$(VAR1)++$(VAR1)$
VAR1 = var2
`
	varmap = map[string]string{
		"VAR1": "var2",
		"VAR2": "@$(VAR1)--$(VAR1)++$(VAR1)$",
		"VAR3": "@var1--var1++var1$",
	}
	if err := tester_varmap(str, varmap); err != nil {
		t.Error(err)
	}

	// empty rule define
	str = `
rule1 : 
`
	rules := []Rule{
		make_rule(
			[]string{""},
			[]string{},
		),
	}
	targets := map[string]int{
		"rule1": 0,
	}
	if err := tester_rules(str, targets, rules); err != nil {
		t.Error(err)
	}

	// rule depend other rule
	str = `
rule1 : rule2  rule3
	# comment rule1
rule2 : rule3
	# comment rule2
rule3 :
	# comment rule3
`
	rules = []Rule{
		make_rule(
			[]string{"rule2  rule3"},
			[]string{},
		),
		make_rule(
			[]string{"rule3"},
			[]string{},
		),
		make_rule(
			[]string{""},
			[]string{},
		),
	}
	targets = map[string]int{
		"rule1": 0,
		"rule2": 1,
		"rule3": 2,
	}
	if err := tester_rules(str, targets, rules); err != nil {
		t.Error(err)
	}

	// rule with command
	str = `
rule1 : rule2
	echo rule1
	echo end
rule2 :
	echo rule2
`
	rules = []Rule{
		make_rule(
			[]string{"rule2"},
			[]string{
				"echo rule1",
				"echo end",
			},
		),
		make_rule(
			[]string{""},
			[]string{
				"echo rule2",
			},
		),
	}
	targets = map[string]int{
		"rule1": 0,
		"rule2": 1,
	}
	if err := tester_rules(str, targets, rules); err != nil {
		t.Error(err)
	}

	// rule with reference
	str = `
$(VAR1) : $(VAR2)
	echo $(VAR1)
$(VAR2) :
	echo $(VAR2)
`
	rules = []Rule{
		make_rule(
			[]string{"$(VAR2)"},
			[]string{"echo $(VAR1)"},
		),
		make_rule(
			[]string{""},
			[]string{"echo $(VAR2)"},
		),
	}
	targets = map[string]int{
		"$(VAR1)": 0,
		"$(VAR2)": 1,
	}
	if err := tester_rules(str, targets, rules); err != nil {
		t.Error(err)
	}

	// rule have no-echo commands
	str = `
rule1 :
	@echo echo1
	echo echo2
`
	rules = []Rule{
		make_rule2(
			[]string{""},
			[]Command{
				Command{"@echo echo1", true},
				Command{"echo echo2", true},
			},
		),
	}
	targets = map[string]int{
		"rule1": 0,
	}
	if err := tester_rules(str, targets, rules); err != nil {
		t.Error(err)
	}

	// simple rule
	str = `
VAR  = rule
VAR2 = $(VAR)2 
VAR3 := $(VAR)3

rule1 : $(VAR2)
	echo rule1
rule2 : rule3
	echo $(VAR2)
$(VAR3) :
	echo $(VAR3)
`
	varmap = map[string]string{
		"VAR":  "rule",
		"VAR2": "$(VAR)2",
		"VAR3": "rule3",
	}
	rules = []Rule{
		make_rule(
			[]string{"$(VAR2)"},
			[]string{"echo rule1"},
		),
		make_rule(
			[]string{"rule3"},
			[]string{"echo $(VAR2)"},
		),
		make_rule(
			[]string{""},
			[]string{"echo $(VAR3)"},
		),
	}
	targets = map[string]int{
		"rule1":   0,
		"rule2":   1,
		"$(VAR3)": 2,
	}
	if err := tester_parser(str, varmap, targets, rules); err != nil {
		t.Error(err)
	}

	// rule have variable no-echo command
	str = `
ECHO = echo
CMD1 = $(ECHO) rule1
CMD2 = @$(ECHO) rule2

rule1 :
	$(CMD1)
rule2 :
	$(CMD2)
`
	varmap = map[string]string{
		"ECHO": "echo",
		"CMD1": "$(ECHO) rule1",
		"CMD2": "@$(ECHO) rule2",
	}
	rules = []Rule{
		make_rule2(
			[]string{""},
			[]Command{
				Command{"$(CMD1)", true},
			},
		),
		make_rule2(
			[]string{""},
			[]Command{
				Command{"$(CMD2)", true},
			},
		),
	}
	targets = map[string]int{
		"rule1": 0,
		"rule2": 1,
	}
	if err := tester_parser(str, varmap, targets, rules); err != nil {
		t.Error(err)
	}
}

func TestRun_preprocess(t *testing.T) {
	// parser tester
	tester_parser := func(parser *Parser, varmap map[string]string, targets map[string]int, rules []Rule) error {
		if err := parser.preprocess(); err != nil {
			return errors.New(fmt.Sprintf("error happened: %q", err))
		}

		if !reflect.DeepEqual(parser.varmap, varmap) {
			return errors.New(fmt.Sprintf("expected %q to eq %q", parser.varmap, varmap))
		}

		if !reflect.DeepEqual(parser.targets, targets) {
			return errors.New(fmt.Sprintf("expected %q to eq %q", parser.targets, targets))
		}

		if !reflect.DeepEqual(parser.rules, rules) {
			return errors.New(fmt.Sprintf("expected %q to eq %q", parser.rules, rules))
		}

		return nil
	}

	// reference resolve
	parser := make_parser(strings.NewReader(""))
	parser.varmap = map[string]string{
		"VAR":  "rule",
		"VAR2": "$(VAR)2",
		"VAR3": "rule3",
	}
	parser.rules = []Rule{
		make_rule(
			[]string{"$(VAR2)"},
			[]string{"echo rule1"},
		),
		make_rule(
			[]string{"rule3"},
			[]string{"echo $(VAR2)"},
		),
		make_rule(
			[]string{""},
			[]string{"echo $(VAR3)"},
		),
	}
	parser.targets = map[string]int{
		"rule1":   0,
		"rule2":   1,
		"$(VAR3)": 2,
	}

	expected_varmap := map[string]string{
		"VAR":  "rule",
		"VAR2": "rule2",
		"VAR3": "rule3",
	}
	expected_rules := []Rule{
		make_rule(
			[]string{"rule2"},
			[]string{"echo rule1"},
		),
		make_rule(
			[]string{"rule3"},
			[]string{"echo rule2"},
		),
		make_rule(
			[]string{},
			[]string{"echo rule3"},
		),
	}
	expected_targets := map[string]int{
		"rule1": 0,
		"rule2": 1,
		"rule3": 2,
	}
	if err := tester_parser(parser, expected_varmap, expected_targets, expected_rules); err != nil {
		t.Error(err)
	}

	// rule have variable no-echo command
	parser = make_parser(strings.NewReader(""))
	parser.varmap = map[string]string{
		"ECHO": "echo",
		"CMD1": "$(ECHO) rule1",
		"CMD2": "@$(ECHO) rule2",
	}
	parser.rules = []Rule{
		make_rule2(
			[]string{""},
			[]Command{
				Command{"$(CMD1)", true},
			},
		),
		make_rule2(
			[]string{""},
			[]Command{
				Command{"$(CMD2)", true},
			},
		),
	}
	parser.targets = map[string]int{
		"rule1": 0,
		"rule2": 1,
	}

	expected_varmap = map[string]string{
		"ECHO": "echo",
		"CMD1": "echo rule1",
		"CMD2": "@echo rule2",
	}
	expected_rules = []Rule{
		make_rule2(
			[]string{},
			[]Command{
				Command{"echo rule1", true},
			},
		),
		make_rule2(
			[]string{},
			[]Command{
				Command{"echo rule2", false},
			},
		),
	}
	expected_targets = map[string]int{
		"rule1": 0,
		"rule2": 1,
	}
	if err := tester_parser(parser, expected_varmap, expected_targets, expected_rules); err != nil {
		t.Error(err)
	}

	// multi target rule
	parser = make_parser(strings.NewReader(""))
	parser.varmap = map[string]string{}
	parser.rules = []Rule{
		make_rule(
			[]string{"rule2  rule3"},
			[]string{"echo rule1"},
		),
		make_rule(
			[]string{""},
			[]string{
				"echo rule2",
				"echo rule3",
			},
		),
	}
	parser.targets = map[string]int{
		"rule1":        0,
		"rule2  rule3": 1,
	}

	expected_varmap = map[string]string{}
	expected_rules = []Rule{
		make_rule(
			[]string{"rule2", "rule3"},
			[]string{"echo rule1"},
		),
		make_rule(
			[]string{},
			[]string{
				"echo rule2",
				"echo rule3",
			},
		),
	}
	expected_targets = map[string]int{
		"rule1": 0,
		"rule2": 1,
		"rule3": 1,
	}
	if err := tester_parser(parser, expected_varmap, expected_targets, expected_rules); err != nil {
		t.Error(err)
	}
}
