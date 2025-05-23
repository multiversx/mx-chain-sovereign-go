package mock

import (
	"time"
)

// WatchdogMock -
type WatchdogMock struct {
	SetCalled  func(callback func(alarmID string), duration time.Duration, alarmID string)
	StopCalled func(alarmID string)
}

// Set -
func (w *WatchdogMock) Set(_ func(alarmID string), _ time.Duration, _ string) {
	if w.SetCalled != nil {
		w.SetCalled(callback, duration, alarmID)
	}
}

// SetDefault -
func (w *WatchdogMock) SetDefault(_ time.Duration, _ string) {
}

// Stop -
func (w *WatchdogMock) Stop(_ string) {
	if w.StopCalled != nil {
		w.StopCalled(alarmID)
	}
}

// Reset -
func (w *WatchdogMock) Reset(_ string) {
}

// IsInterfaceNil -
func (w *WatchdogMock) IsInterfaceNil() bool {
	return false
}
