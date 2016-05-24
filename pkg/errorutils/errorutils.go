// Package errorutils provides helper functions for dealing with errors.
package errorutils

// First returns the first non-nil error in a set of errors.
func First(errors ...error) error {
	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}
