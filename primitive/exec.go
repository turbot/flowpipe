package primitive

import (
	"bufio"
	"context"
	"errors"
	"os/exec"
	"syscall"

	"github.com/turbot/steampipe-pipelines/pipeline"
)

type Exec struct{}

func (e *Exec) ValidateInput(ctx context.Context, i pipeline.StepInput) error {
	if i["command"] == nil {
		return errors.New("Exec input must define a command")
	}
	return nil
}

func (e *Exec) Run(ctx context.Context, input pipeline.StepInput) (pipeline.StepOutput, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	// TODO - support arguments per https://www.terraform.io/language/resources/provisioners/local-exec#argument-reference

	cmd := exec.Command("sh", "-c", input["command"].(string))

	// Capture stdout in real-time
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stdoutLines := []string{}
	stdoutScanner := bufio.NewScanner(stdout)
	go func() {
		for stdoutScanner.Scan() {
			t := stdoutScanner.Text()
			// TODO - send to logs immediately
			stdoutLines = append(stdoutLines, t)
		}
	}()

	// Capture stderr in real-time
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	stderrLines := []string{}
	stderrScanner := bufio.NewScanner(stderr)
	go func() {
		for stderrScanner.Scan() {
			t := stderrScanner.Text()
			// TODO - send to logs immediately
			stderrLines = append(stderrLines, t)
		}
	}()

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	exitCode := 0

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0

			// This works on both Unix and Windows. Although package
			// syscall is generally platform dependent, WaitStatus is
			// defined for both Unix and Windows and in both cases has
			// an ExitStatus() method with the same signature.
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		} else {
			// Unexpected error type, set exit_code to -1 (because I don't have a better idea)
			// TODO - log a warning
			exitCode = -1
		}
	}

	output := pipeline.StepOutput{
		"exit_code":    exitCode,
		"stdout_lines": stdoutLines,
		"stderr_lines": stderrLines,
	}

	return output, nil
}
