package runType

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/errors"
)

func TestConfigureUnixTimeSeconds(t *testing.T) {
	ConfigureUnixTime(Seconds)

	now := time.Now().Truncate(time.Second)
	unix := TimeToUnix(now)
	require.Equal(t, now.Unix(), unix)

	recovered := UnixToTime(unix)
	require.Equal(t, now, recovered)

	dur := 30 * time.Second
	unixDur := TimeDurationToUnix(dur)
	require.Equal(t, int64(30), unixDur)

	err := CheckRoundDuration(1000)
	require.NoError(t, err)

	err = CheckRoundDuration(999)
	require.ErrorIs(t, err, errors.ErrInvalidRoundDuration)

	rounds := ComputeRoundsPerDay(time.Second)
	require.Equal(t, uint64(86400), rounds)
}

func TestConfigureUnixTimeMilliseconds(t *testing.T) {
	ConfigureUnixTime(Milliseconds)
	defer ConfigureUnixTime(Seconds)

	now := time.Now().Truncate(time.Millisecond)
	unix := TimeToUnix(now)
	require.Equal(t, now.UnixMilli(), unix)

	recovered := UnixToTime(unix)
	require.Equal(t, now, recovered)

	dur := 600 * time.Millisecond
	unixDur := TimeDurationToUnix(dur)
	require.Equal(t, int64(600), unixDur)

	err := CheckRoundDuration(core.MinRoundDurationMS)
	require.NoError(t, err)

	err = CheckRoundDuration(core.MinRoundDurationMS - 1)
	require.ErrorIs(t, err, errors.ErrInvalidRoundDuration)

	rounds := ComputeRoundsPerDay(600 * time.Millisecond)
	expected := uint64(NumberOfMillisecondsInDay) / 600
	require.Equal(t, expected, rounds)
}
