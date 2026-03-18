package sdk

// LicenseStatus represents the current status of a license.
type LicenseStatus = string

const (
	LicenseStatusValid    LicenseStatus = "valid"
	LicenseStatusExpiring LicenseStatus = "expiring"
	LicenseStatusGrace    LicenseStatus = "grace"
	LicenseStatusExpired  LicenseStatus = "expired"
)

// LicenseType represents the type of license (e.g. "commercial").
type LicenseType = string
