package sdk

import (
	"crypto/ed25519"
	"encoding/base64"
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
}

func newTokenCache(privateKeyBase64 string) (*tokenCache, error) {
	raw, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("mountos: invalid private key: %w", err)
	}
	if len(raw) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("mountos: private key must be %d bytes, got %d", ed25519.PrivateKeySize, len(raw))
	}
	return &tokenCache{key: ed25519.PrivateKey(raw)}, nil
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
		`{"sub":"vendor","aud":["mountos/appserv"],"iat":%d,"nbf":%d,"exp":%d,"jti":"%s","scope":"service"}`,
		now, now, exp, jti,
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
