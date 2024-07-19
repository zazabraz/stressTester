package utils

import "fmt"

func NewErrorWrapper(field string) func(err error) error {
	return func(err error) error {
		return fmt.Errorf("%s: %w", field, err)
	}
}
