package mock

import (
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-sovereign-go/vm"
)

// SystemSCStub -
type SystemSCStub struct {
	ExecuteCalled       func(args *vmcommon.ContractCallInput) vmcommon.ReturnCode
	SetNewGasCostCalled func(gasCost vm.GasCost)
}

// SetNewGasCost -
func (s *SystemSCStub) SetNewGasCost(gasCost vm.GasCost) {
	if s.SetNewGasCostCalled != nil {
		s.SetNewGasCostCalled(gasCost)
	}
}

// CanUseContract -
func (s *SystemSCStub) CanUseContract() bool {
	return true
}

// Execute -
func (s *SystemSCStub) Execute(args *vmcommon.ContractCallInput) vmcommon.ReturnCode {
	if s.ExecuteCalled != nil {
		return s.ExecuteCalled(args)
	}
	return 0
}

// IsInterfaceNil -
func (s *SystemSCStub) IsInterfaceNil() bool {
	return s == nil
}
