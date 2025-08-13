package factory

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/process/headerCheck"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	dataRetrieverMocks "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/headerSigVerifier"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
)

func createHeaderSigVerifierArgs() *headerCheck.ArgsHeaderSigVerifier {
	return &headerCheck.ArgsHeaderSigVerifier{
		Marshalizer:                  &mock.MarshalizerMock{},
		Hasher:                       &hashingMocks.HasherMock{},
		NodesCoordinator:             &shardingMocks.NodesCoordinatorMock{},
		MultiSigContainer:            cryptoMocks.NewMultiSignerContainerMock(cryptoMocks.NewMultiSigner()),
		SingleSigVerifier:            &mock.SignerMock{},
		KeyGen:                       &mock.SingleSignKeyGenMock{},
		FallbackHeaderValidator:      &testscommon.FallBackHeaderValidatorStub{},
		EnableEpochsHandler:          enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
		HeadersPool:                  &mock.HeadersCacherStub{},
		ProofsPool:                   &dataRetrieverMocks.ProofsPoolMock{},
		StorageService:               &genericMocks.ChainStorerMock{},
		ExtraHeaderSigVerifierHolder: &headerSigVerifier.ExtraHeaderSigVerifierHolderMock{},
	}
}

func TestHeaderSignatureVerifyFactory_CreateHeaderSignatureVerifier(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	factory := NewHeaderSignatureVerifyFactory()
	require.False(t, factory.IsInterfaceNil())

	valSyncer, err := factory.CreateHeaderSignatureVerifier(args)
	require.Nil(t, err)
	require.Equal(t, "*headerCheck.HeaderSigVerifier", fmt.Sprintf("%T", valSyncer))
}
