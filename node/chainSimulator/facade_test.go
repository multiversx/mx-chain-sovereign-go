package chainSimulator

import (
	"errors"
	"testing"

	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-sovereign-go/common"
	"github.com/multiversx/mx-chain-sovereign-go/factory"
	factoryMock "github.com/multiversx/mx-chain-sovereign-go/factory/mock"
	"github.com/multiversx/mx-chain-sovereign-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-sovereign-go/node/chainSimulator/process"
	"github.com/multiversx/mx-chain-sovereign-go/sharding"
	"github.com/multiversx/mx-chain-sovereign-go/state"
	"github.com/multiversx/mx-chain-sovereign-go/testscommon"
	"github.com/multiversx/mx-chain-sovereign-go/testscommon/chainSimulator"
	stateMock "github.com/multiversx/mx-chain-sovereign-go/testscommon/state"
	"github.com/multiversx/mx-chain-sovereign-go/testscommon/vmcommonMocks"
)

var expectedErr = errors.New("expected error")

func TestNewChainSimulatorFacade(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		facade, err := NewChainSimulatorFacade(&chainSimulator.ChainSimulatorMock{
			GetNodeHandlerCalled: func(shardID uint32) process.NodeHandler {
				return &chainSimulator.NodeHandlerMock{}
			},
		})
		require.NoError(t, err)
		require.NotNil(t, facade)
	})
	t.Run("nil chain simulator should error", func(t *testing.T) {
		t.Parallel()

		facade, err := NewChainSimulatorFacade(nil)
		require.Equal(t, errNilChainSimulator, err)
		require.Nil(t, facade)
	})
	t.Run("nil node handler returned by chain simulator should error", func(t *testing.T) {
		t.Parallel()

		facade, err := NewChainSimulatorFacade(&chainSimulator.ChainSimulatorMock{
			GetNodeHandlerCalled: func(shardID uint32) process.NodeHandler {
				return nil
			},
		})
		require.Equal(t, errNilMetachainNode, err)
		require.Nil(t, facade)
	})
}

func TestChainSimulatorFacade_GetExistingAccountFromBech32AddressString(t *testing.T) {
	t.Parallel()

	t.Run("address decode failure should error", func(t *testing.T) {
		t.Parallel()

		facade, err := NewChainSimulatorFacade(&chainSimulator.ChainSimulatorMock{
			GetNodeHandlerCalled: func(shardID uint32) process.NodeHandler {
				return &chainSimulator.NodeHandlerMock{
					GetCoreComponentsCalled: func() factory.CoreComponentsHolder {
						return &mock.CoreComponentsStub{
							AddressPubKeyConverterField: &testscommon.PubkeyConverterStub{
								DecodeCalled: func(humanReadable string) ([]byte, error) {
									return nil, expectedErr
								},
							},
						}
					},
				}
			},
		})
		require.NoError(t, err)

		handler, err := facade.GetExistingAccountFromBech32AddressString("address")
		require.Equal(t, expectedErr, err)
		require.Nil(t, handler)
	})
	t.Run("nil shard node should error", func(t *testing.T) {
		t.Parallel()

		facade, err := NewChainSimulatorFacade(&chainSimulator.ChainSimulatorMock{
			GetNodeHandlerCalled: func(shardID uint32) process.NodeHandler {
				if shardID != common.MetachainShardId {
					return nil
				}

				return &chainSimulator.NodeHandlerMock{
					GetCoreComponentsCalled: func() factory.CoreComponentsHolder {
						return &mock.CoreComponentsStub{
							AddressPubKeyConverterField: &testscommon.PubkeyConverterStub{},
						}
					},
					GetShardCoordinatorCalled: func() sharding.Coordinator {
						return &testscommon.ShardsCoordinatorMock{
							ComputeIdCalled: func(address []byte) uint32 {
								return 0
							},
						}
					},
				}
			},
		})
		require.NoError(t, err)

		handler, err := facade.GetExistingAccountFromBech32AddressString("address")
		require.True(t, errors.Is(err, errShardSetupError))
		require.Nil(t, handler)
	})
	t.Run("shard node GetExistingAccount should error", func(t *testing.T) {
		t.Parallel()

		facade, err := NewChainSimulatorFacade(&chainSimulator.ChainSimulatorMock{
			GetNodeHandlerCalled: func(shardID uint32) process.NodeHandler {
				return &chainSimulator.NodeHandlerMock{
					GetCoreComponentsCalled: func() factory.CoreComponentsHolder {
						return &mock.CoreComponentsStub{
							AddressPubKeyConverterField: &testscommon.PubkeyConverterStub{},
						}
					},
					GetShardCoordinatorCalled: func() sharding.Coordinator {
						return &testscommon.ShardsCoordinatorMock{
							ComputeIdCalled: func(address []byte) uint32 {
								return 0
							},
						}
					},
					GetStateComponentsCalled: func() factory.StateComponentsHolder {
						return &factoryMock.StateComponentsHolderStub{
							AccountsAdapterCalled: func() state.AccountsAdapter {
								return &stateMock.AccountsStub{
									GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
										return nil, expectedErr
									},
								}
							},
						}
					},
				}
			},
		})
		require.NoError(t, err)

		handler, err := facade.GetExistingAccountFromBech32AddressString("address")
		require.Equal(t, expectedErr, err)
		require.Nil(t, handler)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedAccount := &vmcommonMocks.UserAccountStub{}
		facade, err := NewChainSimulatorFacade(&chainSimulator.ChainSimulatorMock{
			GetNodeHandlerCalled: func(shardID uint32) process.NodeHandler {
				return &chainSimulator.NodeHandlerMock{
					GetCoreComponentsCalled: func() factory.CoreComponentsHolder {
						return &mock.CoreComponentsStub{
							AddressPubKeyConverterField: &testscommon.PubkeyConverterStub{},
						}
					},
					GetShardCoordinatorCalled: func() sharding.Coordinator {
						return &testscommon.ShardsCoordinatorMock{
							ComputeIdCalled: func(address []byte) uint32 {
								return 0
							},
						}
					},
					GetStateComponentsCalled: func() factory.StateComponentsHolder {
						return &factoryMock.StateComponentsHolderStub{
							AccountsAdapterCalled: func() state.AccountsAdapter {
								return &stateMock.AccountsStub{
									GetExistingAccountCalled: func(addressContainer []byte) (vmcommon.AccountHandler, error) {
										return providedAccount, nil
									},
								}
							},
						}
					},
				}
			},
		})
		require.NoError(t, err)

		handler, err := facade.GetExistingAccountFromBech32AddressString("address")
		require.NoError(t, err)
		require.True(t, handler == providedAccount) // pointer testing
	})
}
