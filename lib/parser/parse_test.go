package parser

import (
	"strings"
	"testing"
)

func map_eq(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		if w, ok := b[k]; !ok || w != v {
			return false
		}
	}
	return true
}

func array_eq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if w := b[i]; w != v {
			return false
		}
	}
	return true
}

func map_array_eq(a, b map[string][]string) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		if w, ok := b[k]; !ok || !array_eq(w, v) {
			return false
		}
	}
	return true
}

func TestRun_readAndParse(t *testing.T) {
	// parser tester
	tester_parser := func(str string, vars map[string]string, firsts []string, rules map[string][]string) bool {
		ret := true

		r := strings.NewReader(str)
		parser := MakeParser(r)

		if err := parser.readAndParse(); err != nil {
			t.Errorf("error happened: %s", err)
			ret = false
		}

		if !map_eq(parser.varmap, vars) {
			t.Errorf("expected %s to eq %s", parser.varmap, vars)
			ret = false
		}

		if !array_eq(parser.rules.Firsts, firsts) {
			t.Errorf("expected %s to eq %s", parser.rules.Firsts, firsts)
			ret = false
		}

		if !map_array_eq(parser.rules.Rules, rules) {
			t.Errorf("expected %s to eq %s", parser.rules.Rules, rules)
			ret = false
		}

		return ret
	}

	// only varmap tester
	tester_varmap := func(str string, vars map[string]string) bool {
		return tester_parser(str, vars, []string{}, map[string][]string{})
	}

	// only rules tester
	tester_rules := func(str string, firsts []string, rules map[string][]string) bool {
		return tester_parser(str, map[string]string{}, firsts, rules)
	}

	// empty input
	str := ""
	vars := map[string]string{}
	tester_varmap(str, vars)

	// empty input
	str = `
`
	vars = map[string]string{}
	tester_varmap(str, vars)

	// only commnet
	str = `#comment
`
	vars = map[string]string{}
	tester_varmap(str, vars)

	// simple assign
	str = `
VAR1 = var1
VAR2 = var2
VAR3 = var3
`
	vars = map[string]string{
		"VAR1": "var1",
		"VAR2": "var2",
		"VAR3": "var3",
	}
	tester_varmap(str, vars)

	// overwrite assign
	str = `
VAR1 = var1
VAR2 = var2
VAR1 = var3
`
	vars = map[string]string{
		"VAR1": "var3",
		"VAR2": "var2",
	}
	tester_varmap(str, vars)

	// reference variable assign
	str = `
VAR1 = var1
VAR2 = $(VAR1)
VAR3 := $(VAR1)
VAR1 = var2
`
	vars = map[string]string{
		"VAR1": "var2",
		"VAR2": "$(VAR1)",
		"VAR3": "var1",
	}
	tester_varmap(str, vars)

	// undefined variable assign
	str = `
VAR1 = $(VAR0)
VAR2 := $(VAR0)
`
	vars = map[string]string{
		"VAR1": "$(VAR0)",
		"VAR2": "",
	}
	tester_varmap(str, vars)

	// resolve multi reference
	str = `
VAR1 = var1
VAR2 = @$(VAR1)--$(VAR1)++$(VAR1)$
VAR3 := @$(VAR1)--$(VAR1)++$(VAR1)$
VAR1 = var2
`
	vars = map[string]string{
		"VAR1": "var2",
		"VAR2": "@$(VAR1)--$(VAR1)++$(VAR1)$",
		"VAR3": "@var1--var1++var1$",
	}
	tester_varmap(str, vars)

	// empty rule define
	str = `
rule1 : 
`
	rules := map[string][]string{
		"rule1": []string{
			"",
		},
	}
	firsts := []string{"rule1"}
	tester_rules(str, firsts, rules)

	// rule depend other rule
	str = `
rule1 : rule2  rule3
	# comment rule1
rule2 : rule3
	# comment rule2
rule3 :
	# comment rule3
`
	rules = map[string][]string{
		"rule1": []string{
			"rule2  rule3",
		},
		"rule2": []string{
			"rule3",
		},
		"rule3": []string{
			"",
		},
	}
	firsts = []string{"rule1"}
	tester_rules(str, firsts, rules)

	// rule with command
	str = `
rule1 : rule2
	echo rule1
	echo end
rule2 :
	echo rule2
`
	rules = map[string][]string{
		"rule1": []string{
			"rule2",
			"echo rule1",
			"echo end",
		},
		"rule2": []string{
			"",
			"echo rule2",
		},
	}
	firsts = []string{"rule1"}
	tester_rules(str, firsts, rules)

	// rule with reference
	str = `
$(VAR1) : $(VAR2)
	echo $(VAR1)
$(VAR2) :
	echo $(VAR2)
`
	rules = map[string][]string{
		"$(VAR1)": []string{
			"$(VAR2)",
			"echo $(VAR1)",
		},
		"$(VAR2)": []string{
			"",
			"echo $(VAR2)",
		},
	}
	firsts = []string{"$(VAR1)"}
	tester_rules(str, firsts, rules)

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
	vars = map[string]string{
		"VAR":  "rule",
		"VAR2": "$(VAR)2",
		"VAR3": "rule3",
	}
	rules = map[string][]string{
		"rule1": []string{
			"$(VAR2)",
			"echo rule1",
		},
		"rule2": []string{
			"rule3",
			"echo $(VAR2)",
		},
		"$(VAR3)": []string{
			"",
			"echo $(VAR3)",
		},
	}
	firsts = []string{"rule1"}
	tester_parser(str, vars, firsts, rules)
}

func TestRun_preprocess(t *testing.T) {
	// parser tester
	tester := func(p *Parser, varmap map[string]string, firsts []string, rules map[string][]string) bool {
		ret := true

		if err := p.preprocess(); err != nil {
			t.Errorf("error happened: %s", err)
			ret = false
		}

		if !map_eq(p.varmap, varmap) {
			t.Errorf("expected %s to eq %s", p.varmap, varmap)
			ret = false
		}

		if !array_eq(p.rules.Firsts, firsts) {
			t.Errorf("expected %s to eq %s", p.rules.Firsts, firsts)
			ret = false
		}

		if !map_array_eq(p.rules.Rules, rules) {
			t.Errorf("expected %s to eq %s", p.rules.Rules, rules)
			ret = false
		}

		return ret
	}

	// reference resolve
	parser := MakeParser(strings.NewReader(""))
	parser.varmap = map[string]string{
		"VAR":  "rule",
		"VAR2": "$(VAR)2",
		"VAR3": "rule3",
	}
	parser.rules.Rules = map[string][]string{
		"rule1": []string{
			"$(VAR2)",
			"echo rule1",
		},
		"rule2": []string{
			"rule3",
			"echo $(VAR2)",
		},
		"$(VAR3)": []string{
			"",
			"echo $(VAR3)",
		},
	}
	parser.rules.Firsts = []string{"$(VAR3)"}

	expected_vars := map[string]string{
		"VAR":  "rule",
		"VAR2": "rule2",
		"VAR3": "rule3",
	}
	expected_rules := map[string][]string{
		"rule1": []string{
			"rule2",
			"echo rule1",
		},
		"rule2": []string{
			"rule3",
			"echo rule2",
		},
		"rule3": []string{
			"",
			"echo rule3",
		},
	}
	expected_firsts := []string{"rule3"}
	tester(parser, expected_vars, expected_firsts, expected_rules)
}

func TestRun_makeRule(t *testing.T) {
	// parser tester
	tester := func(p *Parser, firsts []string, rules map[string][]string) bool {
		ret := true

		if err := p.makeRule(); err != nil {
			t.Errorf("error happened: %s", err)
			ret = false
		}

		if !array_eq(p.rules.Firsts, firsts) {
			t.Errorf("expected %s to eq %s", p.rules.Firsts, firsts)
			ret = false
		}

		if !map_array_eq(p.rules.Rules, rules) {
			t.Errorf("expected %s to eq %s", p.rules.Rules, rules)
			ret = false
		}

		return ret
	}

	// multi target rule
	parser := MakeParser(strings.NewReader(""))
	parser.rules.Rules = map[string][]string{
		"rule1": []string{
			"rule2  rule3",
			"echo rule1",
		},
		"rule2  rule3": []string{
			"",
			"echo rule2",
			"echo rule3",
		},
	}
	parser.rules.Firsts = []string{"rule2  rule3"}
	expected_rules := map[string][]string{
		"rule1": []string{
			"rule2 rule3",
			"echo rule1",
		},
		"rule2": []string{
			"",
			"echo rule2",
			"echo rule3",
		},
		"rule3": []string{
			"",
			"echo rule2",
			"echo rule3",
		},
	}
	expected_firsts := []string{"rule2", "rule3"}
	tester(parser, expected_firsts, expected_rules)
}
