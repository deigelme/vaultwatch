package alert

import (
	"fmt"
	"io"
	"os"
)

// StdoutNotifier writes alerts to an io.Writer (defaults to os.Stdout).
type StdoutNotifier struct {
	Writer io.Writer
}

// NewStdoutNotifier creates a StdoutNotifier that writes to stdout.
func NewStdoutNotifier() *StdoutNotifier {
	return &StdoutNotifier{Writer: os.Stdout}
}

// Name returns the identifier for this notifier.
func (s *StdoutNotifier) Name() string {
	return "stdout"
}

// Send writes the alert as a formatted line to the configured writer.
func (s *StdoutNotifier) Send(a Alert) error {
	_, err := fmt.Fprintln(s.Writer, a.String())
	return err
}
