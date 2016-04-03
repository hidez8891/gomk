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

	rules = &Rules{
		o.firsts,
		o.rules,
	}
	return
}

type Rule struct {
	Name     string
	Depends  []string
	Commands []Command
}

type Command struct {
	Exestr   string
	NeedEcho bool
}

type Rules struct {
	first_targets []string
	rule_map      map[string]Rule
}

func (r *Rules) Get(name string) (Rule, bool) {
	rule, ok := r.rule_map[name]
	return rule, ok
}

func (r *Rules) Rules() map[string]Rule {
	return r.rule_map
}

func (r *Rules) Firsts() []string {
	return r.first_targets
}

type Parser struct {
	scanner *bufio.Scanner
	buffer  []string
	varmap  map[string]string
	rules   map[string]Rule
	firsts  []string
}

func MakeParser(r io.Reader) *Parser {
	return &Parser{
		bufio.NewScanner(r),
		[]string{},
		map[string]string{},
		map[string]Rule{},
		[]string{},
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
	rules := map[string]Rule{}
	for name, rule := range o.rules {
		// resolve depends
		depends := []string{}
		for _, depend := range rule.Depends {
			depends = append(depends, o.resolveReference(depend))
		}

		// resolve command
		commands := []Command{}
		for _, cmd := range rule.Commands {
			v := cmd.Exestr
			w := o.resolveReference(v)
			commands = append(commands, Command{w, cmd.NeedEcho})
		}

		// resolve rule name
		name = o.resolveReference(name)

		// update
		if _, exist := rules[name]; exist {
			return errors.New("Error: Duplicate rule define " + name)
		}
		rules[name] = Rule{name, depends, commands}
	}
	o.rules = rules

	// resolve targets reference
	firsts := []string{}
	for _, name := range o.firsts {
		// resolve rule name
		name = o.resolveReference(name)
		firsts = append(firsts, name)
	}
	o.firsts = firsts

	return nil
}

func (o *Parser) makeRule() error {
	// parse & construct rules
	rules := map[string]Rule{}
	for name, rule := range o.rules {
		depends := []string{}
		for _, mix_depend := range rule.Depends {
			for _, d := range strings.Fields(mix_depend) {
				depends = append(depends, d)
			}
		}
		rule.Depends = depends

		for _, t := range strings.Fields(name) {
			if _, exist := rules[t]; exist {
				return errors.New("Error: Duplicate rule define " + t)
			}
			rule.Name = t
			rules[t] = rule
		}
	}
	o.rules = rules

	// parse first targets
	firsts := []string{}
	for _, k := range o.firsts {
		for _, target := range strings.Fields(k) {
			firsts = append(firsts, target)
		}
	}
	o.firsts = firsts

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

	if _, exist := o.rules[lhs]; exist {
		return errors.New("Error: Duplicate rule define " + lhs)
	}
	depends := []string{rhs}
	commands := []Command{}

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

		commands = append(commands, Command{line, true})
	}

	if len(o.firsts) == 0 {
		o.firsts = append(o.firsts, lhs)
	}
	o.rules[lhs] = Rule{lhs, depends, commands}

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
