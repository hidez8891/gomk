package runner

import (
	"bytes"
	"testing"
)

func TestRun_Run(t *testing.T) {
	tester := func(cmd, expected_out, expected_err string) bool {
		result := true

		outStream, errStream := new(bytes.Buffer), new(bytes.Buffer)
		runner := New(outStream, errStream)

		if err := runner.Run(cmd); err != nil {
			t.Errorf("error happened: %s", err)
			result = false
		}
		if outStream.String() != expected_out {
			t.Errorf("expected %q to eq %q", outStream.String(), expected_out)
			result = false
		}
		if errStream.String() != expected_err {
			t.Errorf("expected %q to eq %q", errStream.String(), expected_err)
			result = false
		}

		return result
	}

	// empty execute
	cmd := ""
	expected_out := ""
	if !tester(cmd, expected_out, "") {
		t.Skip()
	}

	// simple echo
	cmd = "echo HOGE"
	expected_out = "HOGE\n"
	if !tester(cmd, expected_out, "") {
		t.Skip()
	}
}
