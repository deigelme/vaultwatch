package alert

import (
	"errors"
	"fmt"
)

// MultiNotifier fans out an Alert to multiple Notifier implementations.
// All notifiers are attempted even if one fails; errors are joined.
type MultiNotifier struct {
	notifiers []Notifier
}

// NewMultiNotifier returns a MultiNotifier that sends to every provided Notifier.
func NewMultiNotifier(notifiers ...Notifier) *MultiNotifier {
	return &MultiNotifier{notifiers: notifiers}
}

// Send delivers the alert to all configured notifiers.
// If one or more notifiers fail, their errors are combined and returned.
func (m *MultiNotifier) Send(a Alert) error {
	var errs []error
	for _, n := range m.notifiers {
		if err := n.Send(a); err != nil {
			errs = append(errs, fmt.Errorf("%T: %w", n, err))
		}
	}
	return errors.Join(errs...)
}
