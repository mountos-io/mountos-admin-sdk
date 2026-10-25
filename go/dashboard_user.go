package sdk

import (
	"crypto/hmac"
	"crypto/rand"
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

// signedDashboardUser is the signed payload: the identity plus the request it is
// bound to. appserv rejects a header whose method or requestUri differs from the
// request it arrives with, or whose nonce was already used.
type signedDashboardUser struct {
	DashboardUser
	Method     string `json:"method"`
	RequestURI string `json:"requestUri"`
	Nonce      string `json:"nonce"`
}

// SignDashboardUser produces the signed header value for X-MountOS-Dashboard-User
// for exactly one request. method is the HTTP method and requestURI is the request
// target as sent on the wire (path and query). Each call sets exp to now + 10
// minutes and a fresh nonce, so the header cannot be reused for another request.
// hmacSecret is the dedicated dashboard-user HMAC secret shared with appserv
// (DASHBOARD_USER_HMAC_KEY).
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
func SignDashboardUser(user *DashboardUser, hmacSecret, method, requestURI string) (string, error) {
	if hmacSecret == "" {
		return "", fmt.Errorf("mountos: dashboard HMAC secret is required to sign a dashboard user header")
	}

	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("mountos: dashboard nonce: %w", err)
	}

	u := *user
	exp := time.Now().Add(defaultDashboardUserTTL).Unix()
	u.Exp = &exp

	payload, err := json.Marshal(signedDashboardUser{
		DashboardUser: u,
		Method:        method,
		RequestURI:    requestURI,
		Nonce:         base64.RawURLEncoding.EncodeToString(nonce),
	})
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
