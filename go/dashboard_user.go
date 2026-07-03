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
//
// Access-control contract (what appserv does with this header):
//   - The header is OPT-IN. If you never send it, appserv treats the caller as
//     an unrestricted admin.
//   - appserv/hub understands exactly ONE role string: "user". For role="user"
//     it confines the caller to their own AccountID / VolumeID and uses UserID
//     for audit + volume file-operation attribution. It does NOT enforce which
//     operations or fields a role may use. Any other role value is admin.
//   - ALL access-level policy (role->capability, forbidden fields, per-endpoint
//     rules) is YOUR responsibility as the caller/integrator. The reference
//     admin dashboard implements it in its backend proxy (open-source); you may
//     fork or replace that policy. appserv only guarantees account/volume
//     scoping + attribution for role="user".
//
// If you DO send the header it must be well-formed, or appserv rejects the
// request (403) rather than falling back to admin: Role must be non-empty, and
// role="user" must set both AccountID and UserID (the fields it is scoped and
// attributed by). A signed-but-malformed identity is treated as an attempted
// bypass. Any other non-empty Role is accepted as an admin.
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
