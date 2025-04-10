package data

import (
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// SimulationResultsWithVMOutput is the data transfer object which will hold results for simulation a transaction's execution
type SimulationResultsWithVMOutput struct {
	transaction.SimulationResults
	ReturnData    [][]byte            `json:"returnData,omitempty"`
	ReturnCode    vmcommon.ReturnCode `json:"returnCode,omitempty"`
	ReturnMessage string              `json:"returnMessage,omitempty"`
	Sender        string              `json:"sender,omitempty"`
	Receiver      string              `json:"receiver,omitempty"`
	VMOutput      *vmcommon.VMOutput  `json:"-"`
}
