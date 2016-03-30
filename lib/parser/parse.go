package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

func Parse(r io.Reader) (rules *Rules, err error) {
	o := MakeParser(r)
	if err = o.readAndParse(); err != nil {
		return
	}
	if err = o.preprocess(); err != nil {
		return
	}
	if err = o.makeRule(); err != nil {
		return
	}

	rules = o.rules
	return
}

type Rules struct {
	Firsts []string
	Rules  map[string][]string
}

type Parser struct {
	scanner *bufio.Scanner
	buffer  []string
	varmap  map[string]string
	rules   *Rules
}

func MakeParser(r io.Reader) *Parser {
	return &Parser{
		bufio.NewScanner(r),
		[]string{},
		map[string]string{},
		&Rules{
			[]string{},
			map[string][]string{},
		},
	}
}

func (o *Parser) readAndParse() error {
	regstr_sym := `[a-zA-Z0-9_\.\$\(\){}]+`
	regstr_lhs := fmt.Sprintf(`%[1]s(?:\s+%[1]s)?`, regstr_sym)
	regstr_rhs := `.*`
	rule_class := regexp.MustCompile(fmt.Sprintf(
		`^(%[1]s)\s*(:=|=|:)\s*(%[2]s)$`, regstr_lhs, regstr_rhs))

	for o.inputHasNext() {
		line := o.inputText()

		// skip comment line
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		// rule parsing
		m := rule_class.FindStringSubmatch(line)
		if len(m) == 0 {
			return errors.New("Parse error: " + line)
		}

		lhs := m[1]
		ope := m[2]
		rhs := m[3]

		switch ope {
		case ":=":
			// immediate value assign
			if err := o.parseAssign(lhs, rhs, true); err != nil {
				return err
			}
		case "=":
			// lazy value assign
			if err := o.parseAssign(lhs, rhs, false); err != nil {
				return err
			}
		case ":":
			// rule description
			if err := o.parseRule(lhs, rhs); err != nil {
				return err
			}
		}
	}

	return nil
}

func (o *Parser) inputHasNext() bool {
	return len(o.buffer) > 0 || o.scanner.Scan()
}

func (o *Parser) inputText() string {
	if len(o.buffer) > 0 {
		str := o.buffer[len(o.buffer)-1]
		o.buffer = o.buffer[:len(o.buffer)-1]
		return str
	} else {
		return o.scanner.Text()
	}
}

func (o *Parser) inputUnget(str string) {
	o.buffer = append(o.buffer, str)
}

func (o *Parser) preprocess() error {
	// resolve value reference
	varmap := map[string]string{}
	for k, v := range o.varmap {
		varmap[k] = o.resolveReference(v)
	}
	o.varmap = varmap

	// resolve rule reference
	rules := map[string][]string{}
	for k, ary := range o.rules.Rules {
		// resolve exec
		exec := []string{}
		for _, v := range ary {
			w := o.resolveReference(v)
			exec = append(exec, w)
		}

		// resolve rule name
		k = o.resolveReference(k)

		// update
		if _, exist := rules[k]; exist {
			return errors.New("Error: Duplicate rule define " + k)
		}
		rules[k] = exec
	}
	o.rules.Rules = rules

	// resolve targets reference
	firsts := []string{}
	for _, k := range o.rules.Firsts {
		// resolve rule name
		k = o.resolveReference(k)
		firsts = append(firsts, k)
	}
	o.rules.Firsts = firsts

	return nil
}

func (o *Parser) makeRule() error {
	rules := map[string][]string{}

	// parse & construct rules
	for k, ary := range o.rules.Rules {
		depends := strings.Fields(ary[0])
		ary[0] = strings.Join(depends, " ")

		targets := strings.Fields(k)
		for _, t := range targets {
			if _, exist := rules[t]; exist {
				return errors.New("Error: Duplicate rule define " + t)
			}
			rules[t] = ary
		}
	}
	o.rules.Rules = rules

	// parse first targets
	firsts := []string{}
	for _, k := range o.rules.Firsts {
		for _, target := range strings.Fields(k) {
			firsts = append(firsts, target)
		}
	}
	o.rules.Firsts = firsts

	return nil
}

func (o *Parser) parseAssign(lhs, rhs string, immediate bool) error {
	lhs = strings.TrimSpace(lhs)
	rhs = strings.TrimSpace(rhs)

	// allow overwrite
	if immediate {
		rhs = o.resolveReference(rhs)
	}
	o.varmap[lhs] = rhs

	return nil
}

func (o *Parser) parseRule(lhs, rhs string) error {
	lhs = strings.TrimSpace(lhs)
	rhs = strings.TrimSpace(rhs)

	if _, exist := o.rules.Rules[lhs]; exist {
		return errors.New("Error: Duplicate rule define " + lhs)
	}
	exec := []string{rhs}

	for o.inputHasNext() {
		line := o.inputText()

		// end un-tab line
		if len(line) == 0 || line[0] != '\t' {
			o.inputUnget(line)
			break
		}

		line = strings.TrimSpace(line)

		// skip comment line
		if len(line) == 0 || line[0] == '#' {
			continue
		}

		exec = append(exec, line)
	}

	if len(o.rules.Firsts) == 0 {
		o.rules.Firsts = append(o.rules.Firsts, lhs)
	}
	o.rules.Rules[lhs] = exec
	return nil
}

func (o *Parser) resolveReference(str string) string {
	regstr_name := `\$(?:{\w+}|\(\w+\))`
	pickup_name := regexp.MustCompile(regstr_name)

	expand_pos := pickup_name.FindAllStringIndex(str, -1)
	if len(expand_pos) == 0 {
		return str
	} else {
		res := ""
		lastindex := 0

		for _, pos := range expand_pos {
			res += str[lastindex:pos[0]]
			name := str[pos[0]+2 : pos[1]-1]

			if _, ok := o.varmap[name]; !ok {
				// deal with empty string
			} else {
				val, _ := o.varmap[name]
				res += o.resolveReference(val)
			}
			lastindex = pos[1]
		}

		if lastindex != len(str) {
			res += str[lastindex:len(str)]
		}

		return res
	}
}
