package azuread

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
)

// verifyRS256 verifies an RS256 JWT signature.
// signingInput is "header.payload", sig is the base64url-encoded signature.
func verifyRS256(signingInput, sig string, key *rsa.PublicKey) error {
	sigBytes, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil {
		return err
	}

	hash := sha256.Sum256([]byte(signingInput))
	return rsa.VerifyPKCS1v15(key, crypto.SHA256, hash[:], sigBytes)
}
