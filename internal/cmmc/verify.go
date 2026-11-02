package cmmc

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// VerifySignedBundle only validates detached signature authenticity for an exact
// bundle byte sequence. Trust in the out-of-band Ed25519 public key is assumed.
// It does not prove compliance, software safety, or assessor qualification.
func VerifySignedBundle(bundlePath, signaturePath, publicKeyPath string) (string, error) {
	sigText, err := os.ReadFile(signaturePath)
	if err != nil {
		return "", err
	}
	keyText, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return "", err
	}
	if len(sigText) > 1024 || len(keyText) > 1024 {
		return "", errors.New("key or signature file too long")
	}
	sig, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(sigText)))
	if err != nil {
		return "", fmt.Errorf("signature: %w", err)
	}
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(keyText)))
	if err != nil {
		return "", fmt.Errorf("public key: %w", err)
	}
	if len(sig) != ed25519.SignatureSize || len(key) != ed25519.PublicKeySize {
		return "", errors.New("invalid Ed25519 signature or public key size")
	}
	f, err := os.Open(bundlePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return "", err
	}
	if !st.Mode().IsRegular() || st.Size() > 32<<20 {
		return "", errors.New("bundle must be a regular file <=32 MiB")
	}
	b, err := io.ReadAll(io.LimitReader(f, (32<<20)+1))
	if err != nil {
		return "", err
	}
	if len(b) > 32<<20 {
		return "", errors.New("bundle too large")
	}
	if !ed25519.Verify(ed25519.PublicKey(key), b, sig) {
		return "", errors.New("bundle signature verification failed")
	}
	digest := sha256.Sum256(b)
	return hex.EncodeToString(digest[:]), nil
}
