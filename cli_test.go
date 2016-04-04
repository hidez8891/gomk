package main

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestRun_versionFlag(t *testing.T) {
	outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
	cli := &CLI{outStream: outStream, errStream: errStream}
	args := strings.Split("./gomk -version", " ")

	status := cli.Run(args)
	if status != ExitCodeOK {
		t.Errorf("expected %d to eq %d", status, ExitCodeOK)
	}

	expected := fmt.Sprintf("gomk version %s", Version)
	if !strings.Contains(errStream.String(), expected) {
		t.Errorf("expected %q to eq %q", errStream.String(), expected)
	}
}

func TestRun_fileFlag(t *testing.T) {
	var outStream, errStream *bytes.Buffer
	var cli *CLI

	newCLI := func() {
		outStream, errStream = new(bytes.Buffer), new(bytes.Buffer)
		cli = &CLI{outStream: outStream, errStream: errStream}
	}

	// empty parameter
	newCLI()
	args := strings.Split("./gomk -f", " ")
	status := cli.Run(args)

	if status != ExitCodeError {
		t.Errorf("expected %d to eq %d", status, ExitCodeError)
	}

	expected := "flag needs an argument: -f"
	result := strings.Split(errStream.String(), "\n")[0]
	if !strings.HasPrefix(result, expected) {
		t.Errorf("expected %q to eq %q", result, expected)
	}

	// set parameter
	newCLI()
	args = strings.Split("./gomk -f test/test001.mk", " ")
	status = cli.Run(args)

	if status != ExitCodeOK {
		t.Errorf("expected %d to eq %d", status, ExitCodeOK)
	}
	if outStream.String() != "" {
		t.Errorf("expected %q to empty", outStream.String())
	}
	if errStream.String() != "" {
		t.Errorf("expected %q to empty", errStream.String())
	}
}

func TestRun_targetRules(t *testing.T) {
	tester := func(exe_str, expected_out, expected_err string) error {
		outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
		cli := &CLI{outStream: outStream, errStream: errStream}

		args := strings.Split(exe_str, " ")
		status := cli.Run(args)

		if status != ExitCodeOK {
			return errors.New(fmt.Sprintf("expected %d to eq %d", status, ExitCodeOK))
		}

		if outStream.String() != expected_out {
			return errors.New(fmt.Sprintf("expected %q to eq %q", outStream.String(), expected_out))
		}

		if errStream.String() != expected_err {
			return errors.New(fmt.Sprintf("expected %q to eq %q", errStream.String(), expected_err))
		}

		return nil
	}

	// default rule
	exe_str := "./gomk -f test/test002.mk"
	expected_out := "echo echo1\necho1\necho echo2\necho2\n"
	expected_err := ""
	if err := tester(exe_str, expected_out, expected_err); err != nil {
		t.Error(err)
	}

	// set target
	exe_str = "./gomk -f test/test002.mk echo2"
	expected_out = "echo echo2\necho2\n"
	expected_err = ""
	if err := tester(exe_str, expected_out, expected_err); err != nil {
		t.Error(err)
	}

	// set multi-target
	exe_str = "./gomk -f test/test002.mk echo2 echo1"
	expected_out = "echo echo2\necho2\necho echo1\necho1\n"
	expected_err = ""
	if err := tester(exe_str, expected_out, expected_err); err != nil {
		t.Error(err)
	}

	// suppress echo
	exe_str = "./gomk -f test/test002.mk echo3 echo1"
	expected_out = "echo3\necho echo1\necho1\n"
	expected_err = ""
	if err := tester(exe_str, expected_out, expected_err); err != nil {
		t.Error(err)
	}

	// loop rules
	exe_str = "./gomk -f test/test003.mk"
	expected_out = "\"rule3\"\n\"rule2\"\n\"rule1\"\n"
	expected_err = "Circular rule3 <- rule1 dependency dropped\n"
	if err := tester(exe_str, expected_out, expected_err); err != nil {
		t.Error(err)
	}
}
