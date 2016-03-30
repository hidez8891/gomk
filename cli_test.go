package main

import (
	"bytes"
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
	tester := func(exe_str, expected string) bool {
		outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
		cli := &CLI{outStream: outStream, errStream: errStream}
		result := true

		args := strings.Split(exe_str, " ")
		status := cli.Run(args)

		if status != ExitCodeOK {
			t.Errorf("expected %d to eq %d", status, ExitCodeOK)
			result = false
		}

		if outStream.String() != expected {
			t.Errorf("expected %q to eq %q", outStream.String(), expected)
			result = false
		}

		if errStream.String() != "" {
			t.Errorf("expected %q to eq %q", errStream.String(), "")
			result = false
		}

		return result
	}

	// default rule
	exe_str := "./gomk -f test/test002.mk"
	expected := "echo echo1\necho1\necho echo2\necho2\n"
	if ok := tester(exe_str, expected); !ok {
		t.Skip()
	}

	// specify
	exe_str = "./gomk -f test/test002.mk echo2"
	expected = "echo echo2\necho2\n"
	if ok := tester(exe_str, expected); !ok {
		t.Skip()
	}

	// set rules
	exe_str = "./gomk -f test/test002.mk echo2 echo1"
	expected = "echo echo2\necho2\necho echo1\necho1\n"
	if ok := tester(exe_str, expected); !ok {
		t.Skip()
	}
}
