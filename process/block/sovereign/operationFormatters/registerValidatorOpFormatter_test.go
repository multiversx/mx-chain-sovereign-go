package operationFormatters

import (
	"errors"
	"testing"

	dtoCore "github.com/multiversx/mx-chain-core-go/data/sovereign/dto"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	errMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process/block/sovereign/dto"
	"github.com/multiversx/mx-chain-go/testscommon/sovereign"
	"github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestNewRegisterValidatorOpFormatter(t *testing.T) {
	t.Parallel()

	t.Run("nil peer accounts", func(t *testing.T) {
		opFormatter, err := NewRegisterValidatorOpFormatter(nil, &sovereign.DataCodecMock{}, &sovereign.OutGoingChainNonceMock{})
		require.Nil(t, opFormatter)
		require.Equal(t, errMx.ErrNilPeerAccounts, err)
	})
	t.Run("nil data codec", func(t *testing.T) {
		opFormatter, err := NewRegisterValidatorOpFormatter(&state.AccountsStub{}, nil, &sovereign.OutGoingChainNonceMock{})
		require.Nil(t, opFormatter)
		require.Equal(t, errMx.ErrNilDataCodec, err)
	})
	t.Run("nil chain nonce handler, should return error", func(t *testing.T) {
		opFormatter, err := NewRegisterValidatorOpFormatter(&state.AccountsStub{}, &sovereign.DataCodecMock{}, nil)
		require.Nil(t, opFormatter)
		require.Equal(t, errNilNonceChainHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		opFormatter, err := NewRegisterValidatorOpFormatter(&state.AccountsStub{}, &sovereign.DataCodecMock{}, &sovereign.OutGoingChainNonceMock{})
		require.Nil(t, err)
		require.False(t, opFormatter.IsInterfaceNil())
	})
}

func TestRegisterNewValidatorOpFormatter_CreateOperationData(t *testing.T) {
	t.Parallel()

	mainChainID := []byte{0xfc}
	blsKey := []byte("blsKey")
	ownerAddress := []byte("owner")
	peerAccountsDB := &state.AccountsStub{
		LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
			return &state.PeerAccountHandlerMock{
				MainChainID: mainChainID,
				BLSKey:      blsKey,
			}, nil
		},
	}

	serializedData := []byte("serializedData")
	serializedChainData := map[dtoCore.ChainID][]byte{
		dtoCore.MVX: serializedData,
	}

	nonce := uint64(4)
	dataCodec := &sovereign.DataCodecMock{
		SerializeNewlyRegisteredKeyCalled: func(keyData dto.RegisteredBlsKey) ([]byte, error) {
			require.Equal(t, dto.RegisteredBlsKey{
				ID:    mainChainID,
				Key:   blsKey,
				Owner: ownerAddress,
				Nonce: nonce,
			}, keyData)

			return serializedData, nil
		},
	}

	event := &transaction.Event{
		Address: vm.StakingSCAddress,
		Topics:  [][]byte{blsKey, ownerAddress},
	}
	chainNonceMock := &sovereign.OutGoingChainNonceMock{
		GetAndIncrementNonceCalled: func(chainID dtoCore.ChainID) (uint64, error) {
			require.Equal(t, dtoCore.MVX, chainID)
			nonce++
			return nonce, nil
		},
	}
	opFormatter, _ := NewRegisterValidatorOpFormatter(peerAccountsDB, dataCodec, chainNonceMock)
	res, err := opFormatter.CreateOperationData(event)
	require.Nil(t, err)
	require.Equal(t, serializedChainData, res)
	require.Equal(t, uint64(5), nonce)
}

func TestRegisterNewValidatorOpFormatter_CreateOperationDataErrorCases(t *testing.T) {
	t.Parallel()

	t.Run("invalid num topics", func(t *testing.T) {
		event := &transaction.Event{
			Address: vm.StakingSCAddress,
			Topics:  [][]byte{[]byte("blsKey")},
		}

		opFormatter, _ := NewRegisterValidatorOpFormatter(&state.AccountsStub{}, &sovereign.DataCodecMock{}, &sovereign.OutGoingChainNonceMock{})
		res, err := opFormatter.CreateOperationData(event)
		require.ErrorIs(t, err, errInvalidNumTopicsInRegisterValidator)
		require.Nil(t, res)
	})
	t.Run("invalid event address", func(t *testing.T) {
		event := &transaction.Event{
			Address: vm.ValidatorSCAddress,
			Topics:  [][]byte{[]byte("blsKey"), []byte("owner")},
		}

		opFormatter, _ := NewRegisterValidatorOpFormatter(&state.AccountsStub{}, &sovereign.DataCodecMock{}, &sovereign.OutGoingChainNonceMock{})
		res, err := opFormatter.CreateOperationData(event)
		require.ErrorIs(t, err, vm.ErrInvalidAddress)
		require.Nil(t, res)
	})
	t.Run("cannot load account", func(t *testing.T) {
		event := &transaction.Event{
			Address: vm.StakingSCAddress,
			Topics:  [][]byte{[]byte("blsKey"), []byte("owner")},
		}

		expectedErr := errors.New("load account fails")
		peerAccountsDB := &state.AccountsStub{
			LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
				return nil, expectedErr
			},
		}

		opFormatter, _ := NewRegisterValidatorOpFormatter(peerAccountsDB, &sovereign.DataCodecMock{}, &sovereign.OutGoingChainNonceMock{})
		res, err := opFormatter.CreateOperationData(event)
		require.Equal(t, expectedErr, err)
		require.Nil(t, res)
	})
}
