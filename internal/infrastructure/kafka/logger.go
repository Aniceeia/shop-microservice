package kafka

import "log"

func kafkaLog(format string, args ...any) {
	log.Printf(format, args...)
}
