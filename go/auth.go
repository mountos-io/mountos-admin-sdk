package sdk

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"
	"time"
)

const (
	tokenTTL      = 3600
	refreshMargin = 300
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
	jti := strconv.FormatInt(time.Now().UnixNano(), 10)

	header := `{"alg":"EdDSA","typ":"JWT"}`
	payload := fmt.Sprintf(
		`{"sub":"vendor","aud":["mountos/appserv"],"iat":%d,"nbf":%d,"exp":%d,"jti":"%s","scope":"service","kfp":"%s"}`,
		now, now, exp, jti, tc.kfp,
	)

	signingInput := base64URLEncode([]byte(header)) + "." + base64URLEncode([]byte(payload))
	sig := ed25519.Sign(tc.key, []byte(signingInput))

	tc.token = signingInput + "." + base64URLEncode(sig)
	tc.expiry = exp
	return tc.token, nil
}

func base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}
