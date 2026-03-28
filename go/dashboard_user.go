package sdk

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

const defaultDashboardUserTTL = 10 * time.Minute

// deriveVerificationKey derives the Ed25519 public key from a base64-encoded
// signing key (32-byte seed) and returns it as a standard base64 string.
func deriveVerificationKey(signingKeyBase64 string) (string, error) {
	seed, err := base64.StdEncoding.DecodeString(signingKeyBase64)
	if err != nil {
		return "", fmt.Errorf("mountos: decode signing key: %w", err)
	}
	if len(seed) == 64 {
		seed = seed[:32]
	}
	if len(seed) != 32 {
		return "", fmt.Errorf("mountos: signing key must be 32 bytes, got %d", len(seed))
	}
	pub := ed25519.NewKeyFromSeed(seed).Public().(ed25519.PublicKey)
	return base64.StdEncoding.EncodeToString(pub), nil
}

// deriveDashboardHMACKey derives a dedicated HMAC key from the verification key
// using domain separation.
func deriveDashboardHMACKey(verificationKey string) []byte {
	mac := hmac.New(sha256.New, []byte(verificationKey))
	mac.Write([]byte("dashboard-user"))
	return mac.Sum(nil)
}

// SignDashboardUser produces the signed header value for X-MountOS-Dashboard-User.
// Sets exp to now + 10 minutes. The HMAC key is derived from the Ed25519
// verification key (derived from privateKey) with domain separation.
// Format: base64url(json).base64url(hmac-sha256(base64url(json), derived_key))
func SignDashboardUser(user *DashboardUser, privateKey string) (string, error) {
	verificationKey, err := deriveVerificationKey(privateKey)
	if err != nil {
		return "", err
	}

	u := *user
	u.Exp = time.Now().Add(defaultDashboardUserTTL).Unix()

	payload, err := json.Marshal(u)
	if err != nil {
		return "", fmt.Errorf("mountos: marshal dashboard user: %w", err)
	}

	key := deriveDashboardHMACKey(verificationKey)
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(encoded))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return encoded + "." + sig, nil
}
