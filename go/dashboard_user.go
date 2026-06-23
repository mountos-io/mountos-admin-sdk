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

// deriveDashboardHMACKey derives the dashboard-user signing key from the shared
// secret using domain separation. The secret must match the appserv vault
// DASHBOARD_USER_HMAC_KEY; it is a dedicated secret, NOT the public provider
// verification key.
func deriveDashboardHMACKey(secret string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("dashboard-user"))
	return mac.Sum(nil)
}

// SignDashboardUser produces the signed header value for X-MountOS-Dashboard-User.
// Sets exp to now + 10 minutes. hmacSecret is the dedicated dashboard-user HMAC
// secret shared with appserv (DASHBOARD_USER_HMAC_KEY).
// Format: base64url(json).base64url(hmac-sha256(base64url(json), derived_key))
func SignDashboardUser(user *DashboardUser, hmacSecret string) (string, error) {
	if hmacSecret == "" {
		return "", fmt.Errorf("mountos: dashboard HMAC secret is required to sign a dashboard user header")
	}

	u := *user
	u.Exp = time.Now().Add(defaultDashboardUserTTL).Unix()

	payload, err := json.Marshal(u)
	if err != nil {
		return "", fmt.Errorf("mountos: marshal dashboard user: %w", err)
	}

	key := deriveDashboardHMACKey(hmacSecret)
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(encoded))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return encoded + "." + sig, nil
}
