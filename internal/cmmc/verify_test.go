package cmmc

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

func TestDetachedBundleVerification(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	payload := []byte("vendor catalog, verified out of band")
	paths := []string{filepath.Join(dir, "bundle"), filepath.Join(dir, "sig"), filepath.Join(dir, "pub")}
	for i, data := range [][]byte{payload, []byte(base64.StdEncoding.EncodeToString(ed25519.Sign(priv, payload))), []byte(base64.StdEncoding.EncodeToString(pub))} {
		if err := os.WriteFile(paths[i], data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	digest, err := VerifySignedBundle(paths[0], paths[1], paths[2])
	if err != nil || len(digest) != 64 {
		t.Fatalf("good signature failed %v", err)
	}
	if err := os.WriteFile(paths[0], []byte("tampered"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySignedBundle(paths[0], paths[1], paths[2]); err == nil {
		t.Fatal("tampered bundle accepted")
	}
}
