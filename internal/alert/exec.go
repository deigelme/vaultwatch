package alert

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ExecNotifier runs a local command when an alert fires.
// The alert message is passed as the last argument to the command.
type ExecNotifier struct {
	command string
	args    []string
	timeout time.Duration
}

// NewExecNotifier creates an ExecNotifier that runs the given command.
// command is the executable path; args are optional static arguments prepended
// before the alert message. timeout controls how long to wait for the process.
func NewExecNotifier(command string, args []string, timeout time.Duration) (*ExecNotifier, error) {
	if strings.TrimSpace(command) == "" {
		return nil, fmt.Errorf("exec notifier: command must not be empty")
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &ExecNotifier{
		command: command,
		args:    args,
		timeout: timeout,
	}, nil
}

// Send executes the configured command with the alert message appended as the
// final argument. It returns an error if the process exits non-zero or times out.
func (e *ExecNotifier) Send(a Alert) error {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	cmdArgs := make([]string, len(e.args), len(e.args)+1)
	copy(cmdArgs, e.args)
	cmdArgs = append(cmdArgs, a.String())

	//nolint:gosec // command is operator-supplied configuration
	cmd := exec.CommandContext(ctx, e.command, cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("exec notifier: command %q failed: %w (output: %s)",
			e.command, err, strings.TrimSpace(string(output)))
	}
	return nil
}
