package factory

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSovereignHeaderSignatureVerifyFactory_CreateHeaderSignatureVerifier(t *testing.T) {
	t.Parallel()

	args := createHeaderSigVerifierArgs()
	factory := NewSovereignHeaderSignatureVerifyFactory()
	require.False(t, factory.IsInterfaceNil())

	valSyncer, err := factory.CreateHeaderSignatureVerifier(args)
	require.Nil(t, err)
	require.Equal(t, "*headerCheck.sovereignHeaderSignatureVerify", fmt.Sprintf("%T", valSyncer))
}
