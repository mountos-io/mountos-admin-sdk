package sdk

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

const (
	tokenTTL        = 3600
	refreshMargin   = 300
	clockSkewLeeway = 5
)

type tokenCache struct {
	mu     sync.Mutex
	token  string
	expiry int64
	key    ed25519.PrivateKey
	kfp    string
}

func newTokenCache(privateKeyBase64 string) (*tokenCache, error) {
	raw, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("mountos: invalid private key: %w", err)
	}
	if len(raw) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("mountos: private key must be %d bytes, got %d", ed25519.PrivateKeySize, len(raw))
	}
	k := ed25519.PrivateKey(raw)
	h := sha256.Sum256([]byte(k.Public().(ed25519.PublicKey)))
	return &tokenCache{key: k, kfp: hex.EncodeToString(h[:16])}, nil
}

func (tc *tokenCache) getToken() (string, error) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	now := time.Now().Unix()
	if tc.token != "" && now < tc.expiry-refreshMargin {
		return tc.token, nil
	}

	exp := now + tokenTTL
	jti := generateUUID()

	header := fmt.Sprintf(`{"alg":"EdDSA","typ":"JWT","kid":"%s"}`, tc.kfp)
	payload := fmt.Sprintf(
		`{"sub":"mountos:provider","aud":"mountos/appserv","iat":%d,"nbf":%d,"exp":%d,"jti":"%s","scope":"service","kfp":"%s"}`,
		now, now-clockSkewLeeway, exp, jti, tc.kfp,
	)

	signingInput := base64URLEncode([]byte(header)) + "." + base64URLEncode([]byte(payload))
	sig := ed25519.Sign(tc.key, []byte(signingInput))

	tc.token = signingInput + "." + base64URLEncode(sig)
	tc.expiry = exp
	return tc.token, nil
}

func generateUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}
