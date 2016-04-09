package parser

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"
)

type MakeRule struct {
	Targets map[string]int
	Rules   []Rule
}

type Rule struct {
	Depends  []string
	Commands []Command
}

type Command struct {
	Exestr   string
	NeedEcho bool
}

type Parser struct {
	scanner *bufio.Scanner
	buffer  []string
	varmap  map[string]string
	targets map[string]int
	rules   []Rule
}

func Parse(r io.Reader) (mr *MakeRule, err error) {
	o := &Parser{
		scanner: bufio.NewScanner(r),
		buffer:  []string{},
		varmap:  map[string]string{},
		targets: map[string]int{},
		rules:   []Rule{},
	}

	if err = o.readAndParse(); err != nil {
		return
	}

	if err = o.preprocess(); err != nil {
		return
	}

	mr = &MakeRule{
		Targets: o.targets,
		Rules:   o.rules,
	}
	return
}

func (o *Parser) readAndParse() error {
	rule_class := regexp.MustCompile(`^(.+?)\s*(:=|=|:)\s*(.*?)$`)

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

func (o *Parser) parseAssign(lhs, rhs string, immediate bool) error {
	lhs = strings.TrimSpace(lhs)
	rhs = strings.TrimSpace(rhs)

	if immediate {
		rhs = o.resolveVariable(rhs)
	}
	o.varmap[lhs] = rhs

	return nil
}

func (o *Parser) parseRule(lhs, rhs string) error {
	lhs = strings.TrimSpace(lhs)
	rhs = strings.TrimSpace(rhs)

	if _, exist := o.targets[lhs]; exist {
		return errors.New("Error: Duplicate rule define " + lhs)
	}
	target := lhs
	depends := []string{rhs}
	commands := o.parseCommands()

	o.targets[target] = len(o.rules)
	o.rules = append(o.rules, Rule{depends, commands})

	return nil
}

func (o *Parser) parseCommands() []Command {
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

	return commands
}

func (o *Parser) preprocess() error {
	// varmap
	varmap := map[string]string{}
	for k, v := range o.varmap {
		varmap[k] = o.resolveVariable(v)
	}
	o.varmap = varmap

	// targets
	targets := map[string]int{}
	for name, id := range o.targets {
		name = o.resolveVariable(name)
		names := strings.Fields(name)

		for _, n := range names {
			if _, ok := targets[n]; ok {
				return errors.New("Error: Duplicate rule define " + n)
			}
			targets[n] = id
		}
	}
	o.targets = targets

	// rules
	rules := []Rule{}
	for _, rule := range o.rules {
		// depends
		depends := []string{}
		for _, depend := range rule.Depends {
			depend := o.resolveVariable(depend)
			ds := strings.Fields(depend)

			depends = append(depends, ds...)
		}

		// commands
		commands := []Command{}
		for _, cmd := range rule.Commands {
			exestr := o.resolveVariable(cmd.Exestr)

			// echo flag
			if strings.HasPrefix(exestr, "@") {
				commands = append(commands, Command{exestr[1:], false})
			} else {
				commands = append(commands, Command{exestr, true})
			}
		}

		rules = append(rules, Rule{depends, commands})
	}
	o.rules = rules

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

func (o *Parser) resolveVariable(str string) string {
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
				res += o.resolveVariable(val)
			}
			lastindex = pos[1]
		}

		if lastindex != len(str) {
			res += str[lastindex:len(str)]
		}

		return res
	}
}
