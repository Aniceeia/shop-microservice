package postgresql

import "fmt"

func parseInput(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}
