package dto

import (
	"errors"
)

// ErrInvalidIncomingTopicIdentifier signals that we received invalid/unknown incoming topic identifier
var ErrInvalidIncomingTopicIdentifier = errors.New("received invalid/unknown incoming topic identifier")

// ErrInvalidNumTopicsInEvent signals that we received invalid number of topics in incoming/outgoing event
var ErrInvalidNumTopicsInEvent = errors.New("received invalid number of topics in event")

// ErrInvalidIncomingEventIdentifier signals that we received invalid/unknown incoming event identifier
var ErrInvalidIncomingEventIdentifier = errors.New("received invalid/unknown incoming event identifier")
