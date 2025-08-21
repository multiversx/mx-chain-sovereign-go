package runType

var createPersisterNextEpoch bool

func init() {
	SetShouldCreatePersisterForNextEpoch(false)
}

// ShouldCreatePersister return true if it should create persister
func ShouldCreatePersister() bool {
	return createPersisterNextEpoch
}

// SetShouldCreatePersisterForNextEpoch will set value if it should create persister
func SetShouldCreatePersisterForNextEpoch(value bool) {
	createPersisterNextEpoch = value
}
