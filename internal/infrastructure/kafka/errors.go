package kafka

import "fmt"

func errFail(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
