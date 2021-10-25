package groupTypes

import "github.com/ElrondNetwork/elrond-go/epochStart"

type TriggerStage int32

const (
	Prepare TriggerStage = 0
	Action  TriggerStage = 1
)

// TriggerData represents the trigger data reported to subscribers
type TriggerData struct {
	TriggerID string
	Data interface{}
}

// GroupActionHandler allows defining the trigger handle and adding members to the group
type GroupActionHandler interface {
	AddToGroup(member ActionHandler) error
	GroupID() string
	ActionHandler
}

// ActionHandler defines the Action for a triggered event
type ActionHandler interface {
	HandleAction(triggerData interface{}, stage TriggerStage) error
	IsInterfaceNil() bool
}

// RegistrationHandler allows registering/unregistering of user defined actions to a predefined trigger
type RegistrationHandler interface {
	RegisterHandler(handler epochStart.ActionHandler)
	UnregisterHandler(handler epochStart.ActionHandler)
	GetName() string
	IsInterfaceNil() bool
}

// GroupActionNotifier allows the registration of user defined group actions to preregistered triggers
type GroupActionNotifier interface {
	RegisterGroupAction(groupAction GroupActionHandler, triggerType string) error
	RegisterTrigger(trigger RegistrationHandler) error
	Close() error
}
