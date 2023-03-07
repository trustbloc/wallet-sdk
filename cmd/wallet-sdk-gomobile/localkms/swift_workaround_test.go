package localkms_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/localkms"
)

func TestConstructors_NewKMS(t *testing.T) {
	constructors := &localkms.Constructors{}

	localKMS, err := constructors.NewKMS(localkms.NewMemKMSStore())
	require.NoError(t, err)
	require.NotNil(t, localKMS)
}
