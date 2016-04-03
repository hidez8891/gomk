package parser

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func make_rule(name string, depends []string, commands []string) Rule {
	cmds := []Command{}
	for _, cmd := range commands {
		cmds = append(cmds, Command{cmd, true})
	}

	return Rule{name, depends, cmds}
}

func make_rule_map(rules ...Rule) map[string]Rule {
	rule_map := map[string]Rule{}
	for _, rule := range rules {
		rule_map[rule.Name] = rule
	}
	return rule_map
}

func TestRun_readAndParse(t *testing.T) {
	// parser varmap tester
	tester_varmap := func(str string, varmap map[string]string) error {
		r := strings.NewReader(str)
		parser := MakeParser(r)

		if err := parser.readAndParse(); err != nil {
			return errors.New(fmt.Sprintf("error happened: %q", err))
		}

		if !reflect.DeepEqual(parser.varmap, varmap) {
			return errors.New(fmt.Sprintf("expected %q to eq %q", parser.varmap, varmap))
		}

		return nil
	}

	// parser rules tester
	tester_rules := func(str string, firsts []string, rules map[string]Rule) error {
		r := strings.NewReader(str)
		parser := MakeParser(r)

		if err := parser.readAndParse(); err != nil {
			return errors.New(fmt.Sprintf("error happened: %q", err))
		}

		if !reflect.DeepEqual(parser.firsts, firsts) {
			return errors.New(fmt.Sprintf("expected %q to eq %q", parser.firsts, firsts))
		}

		if !reflect.DeepEqual(parser.rules, rules) {
			return errors.New(fmt.Sprintf("expected %q to eq %q", parser.rules, rules))
		}

		return nil
	}

	// parser tester
	tester_parser := func(str string, varmap map[string]string, firsts []string, rules map[string]Rule) error {
		r := strings.NewReader(str)
		parser := MakeParser(r)

		if err := parser.readAndParse(); err != nil {
			return errors.New(fmt.Sprintf("error happened: %q", err))
		}

		if !reflect.DeepEqual(parser.varmap, varmap) {
			return errors.New(fmt.Sprintf("expected %q to eq %q", parser.varmap, varmap))
		}

		if !reflect.DeepEqual(parser.firsts, firsts) {
			return errors.New(fmt.Sprintf("expected %q to eq %q", parser.firsts, firsts))
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
	rules := make_rule_map(
		make_rule(
			"rule1",
			[]string{""},
			[]string{},
		),
	)
	firsts := []string{"rule1"}
	if err := tester_rules(str, firsts, rules); err != nil {
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
	rules = make_rule_map(
		make_rule(
			"rule1",
			[]string{"rule2  rule3"},
			[]string{},
		),
		make_rule(
			"rule2",
			[]string{"rule3"},
			[]string{},
		),
		make_rule(
			"rule3",
			[]string{""},
			[]string{},
		),
	)
	firsts = []string{"rule1"}
	if err := tester_rules(str, firsts, rules); err != nil {
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
	rules = make_rule_map(
		make_rule(
			"rule1",
			[]string{"rule2"},
			[]string{
				"echo rule1",
				"echo end",
			},
		),
		make_rule(
			"rule2",
			[]string{""},
			[]string{
				"echo rule2",
			},
		),
	)
	firsts = []string{"rule1"}
	if err := tester_rules(str, firsts, rules); err != nil {
		t.Error(err)
	}

	// rule with reference
	str = `
$(VAR1) : $(VAR2)
	echo $(VAR1)
$(VAR2) :
	echo $(VAR2)
`
	rules = make_rule_map(
		make_rule(
			"$(VAR1)",
			[]string{"$(VAR2)"},
			[]string{"echo $(VAR1)"},
		),
		make_rule(
			"$(VAR2)",
			[]string{""},
			[]string{"echo $(VAR2)"},
		),
	)
	firsts = []string{"$(VAR1)"}
	if err := tester_rules(str, firsts, rules); err != nil {
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
	rules = make_rule_map(
		make_rule(
			"rule1",
			[]string{"$(VAR2)"},
			[]string{"echo rule1"},
		),
		make_rule(
			"rule2",
			[]string{"rule3"},
			[]string{"echo $(VAR2)"},
		),
		make_rule(
			"$(VAR3)",
			[]string{""},
			[]string{"echo $(VAR3)"},
		),
	)
	firsts = []string{"rule1"}
	if err := tester_parser(str, varmap, firsts, rules); err != nil {
		t.Error(err)
	}
}

func TestRun_preprocess(t *testing.T) {
	// parser tester
	tester_parser := func(parser *Parser, varmap map[string]string, firsts []string, rules map[string]Rule) error {
		if err := parser.preprocess(); err != nil {
			return errors.New(fmt.Sprintf("error happened: %q", err))
		}

		if !reflect.DeepEqual(parser.varmap, varmap) {
			return errors.New(fmt.Sprintf("expected %q to eq %q", parser.varmap, varmap))
		}

		if !reflect.DeepEqual(parser.firsts, firsts) {
			return errors.New(fmt.Sprintf("expected %q to eq %q", parser.firsts, firsts))
		}

		if !reflect.DeepEqual(parser.rules, rules) {
			return errors.New(fmt.Sprintf("expected %q to eq %q", parser.rules, rules))
		}

		return nil
	}

	// reference resolve
	parser := MakeParser(strings.NewReader(""))
	parser.varmap = map[string]string{
		"VAR":  "rule",
		"VAR2": "$(VAR)2",
		"VAR3": "rule3",
	}
	parser.rules = make_rule_map(
		make_rule(
			"rule1",
			[]string{"$(VAR2)"},
			[]string{"echo rule1"},
		),
		make_rule(
			"rule2",
			[]string{"rule3"},
			[]string{"echo $(VAR2)"},
		),
		make_rule(
			"$(VAR3)",
			[]string{""},
			[]string{"echo $(VAR3)"},
		),
	)
	parser.firsts = []string{"$(VAR3)"}

	expected_varmap := map[string]string{
		"VAR":  "rule",
		"VAR2": "rule2",
		"VAR3": "rule3",
	}
	expected_rules := make_rule_map(
		make_rule(
			"rule1",
			[]string{"rule2"},
			[]string{"echo rule1"},
		),
		make_rule(
			"rule2",
			[]string{"rule3"},
			[]string{"echo rule2"},
		),
		make_rule(
			"rule3",
			[]string{""},
			[]string{"echo rule3"},
		),
	)
	expected_firsts := []string{"rule3"}
	if err := tester_parser(parser, expected_varmap, expected_firsts, expected_rules); err != nil {
		t.Error(err)
	}
}

func TestRun_makeRule(t *testing.T) {
	// parser tester
	tester_parser := func(parser *Parser, firsts []string, rules map[string]Rule) error {
		if err := parser.makeRule(); err != nil {
			return errors.New(fmt.Sprintf("error happened: %q", err))
		}

		if !reflect.DeepEqual(parser.firsts, firsts) {
			return errors.New(fmt.Sprintf("expected %q to eq %q", parser.firsts, firsts))
		}

		if !reflect.DeepEqual(parser.rules, rules) {
			return errors.New(fmt.Sprintf("expected %q to eq %q", parser.rules, rules))
		}

		return nil
	}

	// multi target rule
	parser := MakeParser(strings.NewReader(""))
	parser.rules = make_rule_map(
		make_rule(
			"rule1",
			[]string{"rule2  rule3"},
			[]string{"echo rule1"},
		),
		make_rule(
			"rule2  rule3",
			[]string{""},
			[]string{
				"echo rule2",
				"echo rule3",
			},
		),
	)
	parser.firsts = []string{"rule2  rule3"}
	expected_rules := make_rule_map(
		make_rule(
			"rule1",
			[]string{"rule2", "rule3"},
			[]string{"echo rule1"},
		),
		make_rule(
			"rule2",
			[]string{},
			[]string{
				"echo rule2",
				"echo rule3",
			},
		),
		make_rule(
			"rule3",
			[]string{},
			[]string{
				"echo rule2",
				"echo rule3",
			},
		),
	)
	expected_firsts := []string{"rule2", "rule3"}
	if err := tester_parser(parser, expected_firsts, expected_rules); err != nil {
		t.Error(err)
	}
}
