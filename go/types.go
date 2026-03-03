package sdk

import "encoding/json"

// Config holds client configuration.
type Config struct {
	BaseURL    string
	PrivateKey string // base64-encoded 64-byte ED25519 private key
}

// ListOptions for page-based pagination.
type ListOptions struct {
	Page  int
	Limit int
}

// PaginationMeta from paginated responses.
type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"totalPages"`
}

// PaginatedResponse wraps page-based results.
type PaginatedResponse[T any] struct {
	Items      []T            `json:"items"`
	Pagination PaginationMeta `json:"pagination"`
}

// CursorPaginatedResponse wraps cursor-based results.
type CursorPaginatedResponse[T any] struct {
	Items      []T    `json:"items"`
	NextCursor *int64 `json:"nextCursor"`
}

// IDResponse returned by create/edit/toggle endpoints.
type IDResponse struct {
	ID int64 `json:"id"`
}

// StringIDResponse returned by storage/volume endpoints.
type StringIDResponse struct {
	ID string `json:"id"`
}

// Accounts

type CreateAccountRequest struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	VendorInfo  map[string]any `json:"vendorInfo,omitempty"`
}

type EditAccountRequest struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	VendorInfo  map[string]any `json:"vendorInfo,omitempty"`
}

type Account struct {
	ID          int64          `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	VendorInfo  map[string]any `json:"vendorInfo,omitempty"`
	IsActive    bool           `json:"isActive"`
	Locked      bool           `json:"locked"`
	CreatedAt   string         `json:"createdAt"`
	UpdatedAt   string         `json:"updatedAt"`
}

// Users

type AddUserRequest struct {
	AccountID  int64          `json:"accountId"`
	Username   string         `json:"username"`
	Email      string         `json:"email"`
	Name       string         `json:"name,omitempty"`
	VendorInfo map[string]any `json:"vendorInfo,omitempty"`
}

type EditUserRequest struct {
	Username   string         `json:"username"`
	Email      string         `json:"email"`
	Name       string         `json:"name,omitempty"`
	VendorInfo map[string]any `json:"vendorInfo,omitempty"`
}

type User struct {
	ID        int64  `json:"id"`
	AccountID int64  `json:"accountId"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Name      string `json:"name"`
}

type UserListOptions struct {
	AccountID int64
	Page      int
	Limit     int
}

// Regions

type CreateRegionRequest struct {
	AccountID int64  `json:"accountId"`
	Name      string `json:"name"`
	DNS       string `json:"dns"`
}

type EditRegionRequest struct {
	AccountID int64  `json:"accountId"`
	Name      string `json:"name"`
	DNS       string `json:"dns"`
}

type Region struct {
	ID        int64  `json:"id"`
	AccountID int64  `json:"accountId"`
	Name      string `json:"name"`
	DNS       string `json:"dns"`
	IsActive  bool   `json:"isActive"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// Storages

type CreateStorageRequest struct {
	AccountID    int64  `json:"accountId"`
	RegionID     int64  `json:"regionId"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	StorageType  string `json:"storageType"`
	ProviderType string `json:"providerType"`
	Endpoint     string `json:"endpoint"`
	Region       string `json:"region,omitempty"`
	Bucket       string `json:"bucket,omitempty"`
	Base         string `json:"base,omitempty"`
	BlockRegion  string `json:"blockRegion,omitempty"`
	BlockType    string `json:"blockType,omitempty"`
	BlockSize    int32  `json:"blockSize,omitempty"`
	AccessKey    string `json:"accessKey,omitempty"`
	SecretKey    string `json:"secretKey,omitempty"`
}

type EditStorageRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Endpoint    string `json:"endpoint,omitempty"`
	AccessKey   string `json:"accessKey,omitempty"`
	SecretKey   string `json:"secretKey,omitempty"`
}

type Storage struct {
	ID           string `json:"id"`
	ShardID      int64  `json:"shardId"`
	AccountID    int64  `json:"accountId"`
	RegionID     int64  `json:"regionId"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	StorageType  string `json:"storageType"`
	ProviderType string `json:"providerType"`
	BlockType    string `json:"blockType,omitempty"`
	Endpoint     string `json:"endpoint"`
	Region       string `json:"region,omitempty"`
	Bucket       string `json:"bucket,omitempty"`
	Base         string `json:"base,omitempty"`
	BlockRegion  string `json:"blockRegion,omitempty"`
	BlockSize    int32  `json:"blockSize,omitempty"`
	IsActive     bool   `json:"isActive"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type StorageListOptions struct {
	AccountID int64
	Page      int
	Limit     int
}

type CreateStorageResponse struct {
	ID      string `json:"id"`
	ShardID int64  `json:"shardId"`
}

// Volumes

type UpdateVolumeQuotaRequest struct {
	QuotaLimit int64 `json:"quotaLimit"`
}

// Audit Logs

type AuditLog struct {
	ID          int64           `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description,omitempty"`
	Subject     string          `json:"subject,omitempty"`
	Success     bool            `json:"success"`
	Data        json.RawMessage `json:"data,omitempty"`
	CreatedBy   string          `json:"createdBy,omitempty"`
	AccountID   string          `json:"accountId,omitempty"`
	CreatedAt   string          `json:"createdAt,omitempty"`
	UpdatedAt   string          `json:"updatedAt,omitempty"`
}

type AuditLogListOptions struct {
	AccountID int64
	Cursor    int64
	Limit     int
	Subject   string
}
