package sdk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

const defaultDashboardUserTTL = 10 * time.Minute

// deriveDashboardHMACKey derives a dedicated HMAC key from the ED25519 signing key
// using domain separation.
func deriveDashboardHMACKey(privateKey string) []byte {
	mac := hmac.New(sha256.New, []byte(privateKey))
	mac.Write([]byte("dashboard-user"))
	return mac.Sum(nil)
}

// SignDashboardUser produces the signed header value for X-MountOS-Dashboard-User.
// Sets exp to now + 10 minutes. The HMAC key is derived from privateKey with domain separation.
// Format: base64url(json).base64url(hmac-sha256(base64url(json), derived_key))
func SignDashboardUser(user *DashboardUser, privateKey string) (string, error) {
	u := *user
	u.Exp = time.Now().Add(defaultDashboardUserTTL).Unix()

	payload, err := json.Marshal(u)
	if err != nil {
		return "", fmt.Errorf("mountos: marshal dashboard user: %w", err)
	}

	key := deriveDashboardHMACKey(privateKey)
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(encoded))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return encoded + "." + sig, nil
}
