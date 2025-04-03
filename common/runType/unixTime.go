package runType

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"

	"github.com/multiversx/mx-chain-go/errors"
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

// TimeUnit enum for round configuration
type TimeUnit int

const (
	// Seconds specifies round configuration in seconds
	Seconds TimeUnit = iota
	// Milliseconds specifies round configuration in milliseconds
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

// TimeToUnix returns the time to unix based on current configuration
func TimeToUnix(t time.Time) int64 {
	return timeToUnix(t)
}

// UnixToTime converts int64 to time based on current configuration
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

// TimeDurationToUnix converts duration time to unix based on current configuration
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

// CheckRoundDuration checks round duration based on current  configuration
func CheckRoundDuration(roundDuration uint64) error {
	switch currentUnit {
	case Seconds:
		return checkRoundDurationSec(roundDuration)
	case Milliseconds:
		return checkRoundDurationMilliSec(roundDuration)
	default:
		return checkRoundDurationSec(roundDuration)
	}
}

func checkRoundDurationSec(roundDuration uint64) error {
	roundDurationSec := roundDuration / 1000
	if roundDurationSec < 1 {
		return errors.ErrInvalidRoundDuration
	}

	return nil
}

func checkRoundDurationMilliSec(roundDuration uint64) error {
	if roundDuration < core.MinRoundDurationMS {
		return errors.ErrInvalidRoundDuration
	}

	return nil
}
