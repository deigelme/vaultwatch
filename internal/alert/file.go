package alert

import (
	"fmt"
	"os"
	"time"
)

// FileNotifier appends alert messages to a local file.
type FileNotifier struct {
	path string
}

// NewFileNotifier creates a FileNotifier that writes alerts to the given
// file path. The file is created if it does not exist and is appended to
// on each alert.
func NewFileNotifier(path string) (*FileNotifier, error) {
	if path == "" {
		return nil, fmt.Errorf("file notifier: path must not be empty")
	}
	return &FileNotifier{path: path}, nil
}

// Send appends a formatted alert line to the configured file.
func (f *FileNotifier) Send(a Alert) error {
	file, err := os.OpenFile(f.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("file notifier: open %q: %w", f.path, err)
	}
	defer file.Close()

	line := fmt.Sprintf("%s [%s] %s\n",
		time.Now().UTC().Format(time.RFC3339),
		a.Level,
		a.String(),
	)
	if _, err := file.WriteString(line); err != nil {
		return fmt.Errorf("file notifier: write: %w", err)
	}
	return nil
}
