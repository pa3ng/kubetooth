package keys

import (
	"encoding/hex"

	ellcurv "github.com/btcsuite/btcd/btcec"
)

var cachedCurve = ellcurv.S256()

func NewKeyPair() (string, string) {
	priv, _ := ellcurv.NewPrivateKey(cachedCurve)
	privBytes := priv.Serialize()

	_, pub := ellcurv.PrivKeyFromBytes(cachedCurve, privBytes)

	return hex.EncodeToString(pub.SerializeCompressed()), hex.EncodeToString(privBytes)
}
