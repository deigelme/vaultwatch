package alert

import "fmt"

// Notifier is the interface that all alert backends must implement.
type Notifier interface {
	Send(a Alert) error
}

// MultiNotifier fans out a single alert to multiple Notifier backends.
type MultiNotifier struct {
	notifiers []Notifier
}

// NewMultiNotifier creates a MultiNotifier that dispatches to all provided notifiers.
func NewMultiNotifier(notifiers ...Notifier) *MultiNotifier {
	return &MultiNotifier{notifiers: notifiers}
}

// Send sends the alert to every registered notifier, collecting all errors.
func (m *MultiNotifier) Send(a Alert) error {
	var errs []error
	for _, n := range m.notifiers {
		if err := n.Send(a); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("multi-notifier encountered %d error(s): %v", len(errs), errs)
}
