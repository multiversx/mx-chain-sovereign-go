package runType

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"

	"github.com/multiversx/mx-chain-go/errors"
)

var (
	timeToUnix         func(time.Time) int64
	unixToTime         func(int64) time.Time
	timeDurationToUnix func(time.Duration) int64
	checkRoundDuration func(uint64) error
	unitsInDay         int
)

func init() {
	ConfigureUnixTime(Seconds)
}

// TimeUnit enum for round configuration
type TimeUnit int

const (
	// Seconds specifies round configuration in seconds
	Seconds TimeUnit = iota
	// Milliseconds specifies round configuration in milliseconds
	Milliseconds
)

// NumberOfSecondsInDay represents the number of seconds in a day
const NumberOfSecondsInDay = 86400

// NumberOfMillisecondsInDay represents the number of milliseconds in a day
const NumberOfMillisecondsInDay = NumberOfSecondsInDay * 1000

// ConfigureUnixTime configures timeToUnix singleton to work with a specific time unit
func ConfigureUnixTime(unit TimeUnit) {
	switch unit {
	case Milliseconds:
		timeToUnix = func(t time.Time) int64 { return t.UnixMilli() }
		unixToTime = func(unixTime int64) time.Time { return time.UnixMilli(unixTime) }
		timeDurationToUnix = func(duration time.Duration) int64 { return duration.Milliseconds() }
		checkRoundDuration = checkRoundDurationMilliSec
		unitsInDay = NumberOfMillisecondsInDay
	default:
		timeToUnix = func(t time.Time) int64 { return t.Unix() }
		unixToTime = func(unixTime int64) time.Time { return time.Unix(unixTime, 0) }
		timeDurationToUnix = func(duration time.Duration) int64 { return int64(duration.Seconds()) }
		checkRoundDuration = checkRoundDurationSec
		unitsInDay = NumberOfSecondsInDay
	}
}

// TimeToUnix returns the time to unix based on current configuration
func TimeToUnix(t time.Time) int64 {
	return timeToUnix(t)
}

// UnixToTime converts int64 to time based on current configuration
func UnixToTime(unixTime int64) time.Time {
	return unixToTime(unixTime)
}

// TimeDurationToUnix converts duration time to unix based on current configuration
func TimeDurationToUnix(duration time.Duration) int64 {
	return timeDurationToUnix(duration)
}

// CheckRoundDuration checks round duration based on current configuration
func CheckRoundDuration(roundDuration uint64) error {
	return checkRoundDuration(roundDuration)
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

// ComputeRoundsPerDay computes the rounds per day based on current configuration
func ComputeRoundsPerDay(roundTime time.Duration) uint64 {
	return uint64(unitsInDay) / uint64(timeDurationToUnix(roundTime))
}
