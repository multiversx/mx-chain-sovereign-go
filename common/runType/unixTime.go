package runType

import (
	"time"
)

var (
	secondsConverter = func(t time.Time) int64 {
		return t.Unix()
	}
	millisecondsConverter = func(t time.Time) int64 {
		return t.UnixMilli()
	}

	// timeToUnix singleton defaults to seconds
	timeToUnix  = secondsConverter
	currentUnit = Seconds
)

type TimeUnit int

const (
	Seconds TimeUnit = iota
	Milliseconds
)

// ConfigureUnixTime configures timeToUnix singleton to work with a specific time unit
func ConfigureUnixTime(unit TimeUnit) {
	currentUnit = unit
	switch unit {
	case Seconds:
		timeToUnix = secondsConverter
	case Milliseconds:
		timeToUnix = millisecondsConverter
	default:
		timeToUnix = secondsConverter
	}
}

// TimeToUnix returns the time to unix depending on current configuration
func TimeToUnix(t time.Time) int64 {
	return timeToUnix(t)
}

// UnixToTime converts int64 to time depending on current configuration
func UnixToTime(unixTime int64) time.Time {
	switch currentUnit {
	case Seconds:
		return time.Unix(unixTime, 0)
	case Milliseconds:
		return time.UnixMilli(unixTime)
	default:
		return time.Unix(unixTime, 0)
	}
}

// TimeDurationToUnix converts duration time to unix depending on current configuration
func TimeDurationToUnix(duration time.Duration) int64 {
	switch currentUnit {
	case Seconds:
		return int64(duration.Seconds())
	case Milliseconds:
		return duration.Milliseconds()
	default:
		return int64(duration.Seconds())
	}
}
