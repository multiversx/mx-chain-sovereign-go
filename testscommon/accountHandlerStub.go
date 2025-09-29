package testscommon

import (
	"fmt"

	storageCommon "github.com/multiversx/mx-chain-storage-go/common"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// AccountHandlerStub is a stub for the AccountDataHandler interface
type AccountHandlerStub struct {
	AddressBytesCalled        func() []byte
	RetrieveValueCalled       func(key []byte) ([]byte, uint32, error)
	SaveKeyValueCalled        func(key, value []byte) error
	RemoveValueCalled         func(key []byte) error
	IsInterfaceNilCalled      func() bool
	MigrateDataTrieLeavesCalled func(args vmcommon.ArgsMigrateDataTrieLeaves) error
	data                      map[string][]byte
}

// NewAccountHandlerStub returns a new instance of AccountHandlerStub
func NewAccountHandlerStub() *AccountHandlerStub {
	return &AccountHandlerStub{
		data: make(map[string][]byte),
	}
}

// AddressBytes returns the address bytes
func (ahs *AccountHandlerStub) AddressBytes() []byte {
	if ahs.AddressBytesCalled != nil {
		return ahs.AddressBytesCalled()
	}
	return nil
}

// RetrieveValue retrieves a value from the stub's data map
func (ahs *AccountHandlerStub) RetrieveValue(key []byte) ([]byte, uint32, error) {
	if ahs.RetrieveValueCalled != nil {
		return ahs.RetrieveValueCalled(key)
	}
	if val, ok := ahs.data[string(key)]; ok {
		return val, 0, nil
	}
	return nil, 0, storageCommon.ErrKeyNotFound
}

// SaveKeyValue saves a key-value pair to the stub's data map
func (ahs *AccountHandlerStub) SaveKeyValue(key, value []byte) error {
	if ahs.SaveKeyValueCalled != nil {
		return ahs.SaveKeyValueCalled(key, value)
	}
	ahs.data[string(key)] = value
	return nil
}

// RemoveValue removes a value from the stub's data map
func (ahs *AccountHandlerStub) RemoveValue(key []byte) error {
	if ahs.RemoveValueCalled != nil {
		return ahs.RemoveValueCalled(key)
	}
	if _, ok := ahs.data[string(key)]; !ok {
		return fmt.Errorf("key not found")
	}
	delete(ahs.data, string(key))
	return nil
}

// IsInterfaceNil returns true if the interface is nil
func (ahs *AccountHandlerStub) IsInterfaceNil() bool {
	if ahs.IsInterfaceNilCalled != nil {
		return ahs.IsInterfaceNilCalled()
	}
	return ahs == nil
}

// MigrateDataTrieLeaves is a mock implementation
func (ahs *AccountHandlerStub) MigrateDataTrieLeaves(args vmcommon.ArgsMigrateDataTrieLeaves) error {
	if ahs.MigrateDataTrieLeavesCalled != nil {
		return ahs.MigrateDataTrieLeavesCalled(args)
	}
	// The handler logic is complex and not needed for these tests.
	// A nil return satisfies the interface.
	return nil
}

var _ vmcommon.AccountDataHandler = &AccountHandlerStub{}