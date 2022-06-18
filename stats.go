package gorequest

import "time"

type Stats struct {
	RequestBytes  int64
	ResponseBytes int64

	RequestDuration time.Duration
}

func copyStats(old Stats) Stats {
	newStats := old
	return newStats
}
