package runner

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
)

type Runner struct {
	outStream, errStream io.Writer
}

func New(out, err io.Writer) *Runner {
	return &Runner{out, err}
}

func (r *Runner) Run(command string) error {
	args := []string{"cmd", "/C", command}
	cmd := exec.Command(args[0], args[1:]...)

	out_reader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	err_reader, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	go r.echoStdout(out_reader)
	go r.echoStderr(err_reader)

	err = cmd.Wait()
	return err
}

func (r *Runner) echoStdout(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Fprintf(r.outStream, "%s\n", scanner.Text())
	}
}

func (r *Runner) echoStderr(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Fprintf(r.errStream, "%s\n", scanner.Text())
	}
}
