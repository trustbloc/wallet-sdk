package did_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/did"
)

func TestConstructors_NewCreatorWithKeyReader(t *testing.T) {
	localKMS := createTestKMS(t)

	constructors := &did.Constructors{}

	didCreator, err := constructors.NewCreatorWithKeyWriter(localKMS)
	require.NoError(t, err)
	require.NotNil(t, didCreator)
}

func TestConstructors_NewCreatorWithKeyWriter(t *testing.T) {
	localKMS := createTestKMS(t)

	constructors := &did.Constructors{}

	didCreator, err := constructors.NewCreatorWithKeyReader(localKMS)
	require.NoError(t, err)
	require.NotNil(t, didCreator)
}

func TestConstructors_NewResolver(t *testing.T) {
	constructors := &did.Constructors{}

	didResolver, err := constructors.NewResolver("")
	require.NoError(t, err)
	require.NotNil(t, didResolver)
}
