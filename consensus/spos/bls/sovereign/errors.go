package sovereign

import "errors"

// ErrNilSubRoundsFactoryInSovereign signals that a nil sub round factory handler provided in sovereign
var ErrNilSubRoundsFactoryInSovereign = errors.New("nil sub round factory handler provided in sovereign")
