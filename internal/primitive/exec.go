package primitive

import (
	"bufio"
	"context"
	"os/exec"
	"syscall"
	"time"

	"github.com/turbot/pipe-fittings/modconfig"
	"github.com/turbot/pipe-fittings/perr"
	"github.com/turbot/pipe-fittings/schema"
)

type Exec struct{}

func (e *Exec) ValidateInput(ctx context.Context, i modconfig.Input) error {
	if i["command"] == nil {
		return perr.BadRequestWithMessage("Exec input must define a command")
	}
	return nil
}

func (e *Exec) Run(ctx context.Context, input modconfig.Input) (*modconfig.Output, error) {
	if err := e.ValidateInput(ctx, input); err != nil {
		return nil, err
	}

	// TODO - support arguments per https://www.terraform.io/language/resources/provisioners/local-exec#argument-reference

	//nolint:gosec // TODO G204: Subprocess launched with a potential tainted input or cmd arguments (gosec)
	cmd := exec.Command("sh", "-c", input["command"].(string))

	// Capture stdout in real-time
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stdoutLines := []string{}
	// TODO - by default this has a max line size of 64K, see https://stackoverflow.com/a/16615559
	stdoutScanner := bufio.NewScanner(stdout)
	stdoutScanner.Buffer(make([]byte, bufio.MaxScanTokenSize*40), bufio.MaxScanTokenSize*40)
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
	// TODO - by default this has a max line size of 64K, see https://stackoverflow.com/a/16615559
	stderrScanner := bufio.NewScanner(stderr)
	stderrScanner.Buffer(make([]byte, bufio.MaxScanTokenSize*40), bufio.MaxScanTokenSize*40)
	go func() {
		for stderrScanner.Scan() {
			t := stderrScanner.Text()
			// TODO - send to logs immediately
			stderrLines = append(stderrLines, t)
		}
	}()

	start := time.Now().UTC()
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
	finish := time.Now().UTC()

	o := modconfig.Output{
		Data: map[string]interface{}{},
	}

	o.Data["exit_code"] = exitCode
	o.Data["stdout_lines"] = stdoutLines
	o.Data["stderr_lines"] = stderrLines
	o.Data[schema.AttributeTypeStartedAt] = start
	o.Data[schema.AttributeTypeFinishedAt] = finish

	return &o, nil
}
