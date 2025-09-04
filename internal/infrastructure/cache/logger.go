package cache

import "log"

func cacheLog(format string, args ...any) {
	log.Printf(format, args...)
}
