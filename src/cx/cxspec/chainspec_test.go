package cxspec

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/skycoin/skycoin/src/cipher"
)

func TestNew(t *testing.T) {
	pk, sk := cipher.GenerateKeyPair()
	addr := cipher.AddressFromPubKey(pk)
	spec, err := New("skycoin", "SKY", sk, addr, []byte{0, 1, 1})
	require.NoError(t, err)
	b, err := json.MarshalIndent(spec, "", "\t")
	require.NoError(t, err)
	fmt.Println(string(b))
}
