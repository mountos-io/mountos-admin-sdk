# mountOS Admin SDK for Go

Go SDK for the mountOS vendor API. Zero external dependencies — pure stdlib.

## Install

```bash
go get github.com/mountos-app/mountos-admin-sdk/go
```

## Usage

```go
package main

import (
  "context"
  "fmt"
  "log"

  sdk "github.com/mountos-app/mountos-admin-sdk/go"
)

func main() {
  client, err := sdk.NewClient(sdk.Config{
    BaseURL:    "https://appserv.example.com",
    PrivateKey: "base64-encoded-ed25519-private-key",
  })
  if err != nil {
    log.Fatal(err)
  }

  ctx := context.Background()

  // Accounts
  acct, err := client.Accounts.Create(ctx, &sdk.CreateAccountRequest{Name: "Acme"})
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println("Created account:", acct.ID)

  list, _ := client.Accounts.List(ctx, &sdk.ListOptions{Page: 1, Limit: 10})
  fmt.Println("Total accounts:", list.Pagination.Total)

  acct, _ = client.Accounts.Get(ctx, acct.ID)
  _ = client.Accounts.Lock(ctx, acct.ID)
  _ = client.Accounts.Unlock(ctx, acct.ID)
  _ = client.Accounts.Activate(ctx, acct.ID)

  // Users
  user, _ := client.Users.Add(ctx, &sdk.AddUserRequest{
    AccountID: acct.ID, Email: "a@b.com", Username: "alice",
  })
  users, _ := client.Users.List(ctx, &sdk.UserListOptions{AccountID: acct.ID})
  _ = users
  _, _ = client.Users.Edit(ctx, user.ID, &sdk.EditUserRequest{
    Username: "bob", Email: "b@c.com",
  })

  // Regions
  region, _ := client.Regions.Create(ctx, &sdk.CreateRegionRequest{
    AccountID: acct.ID, Name: "us-east", DNS: "us.example.com",
  })
  _ = region

  // Storages
  storage, _ := client.Storages.Create(ctx, &sdk.CreateStorageRequest{
    AccountID: acct.ID, RegionID: region.ID,
    Name: "prod-s3", StorageType: "object",
    ProviderType: "s3", Endpoint: "https://s3.example.com",
  })
  _ = storage

  // Volumes
  _ = client.Volumes.UpdateQuota(ctx, "volume-uuid", &sdk.UpdateVolumeQuotaRequest{
    QuotaLimit: 1073741824,
  })

  // Audit logs
  logs, _ := client.AuditLogs.List(ctx, &sdk.AuditLogListOptions{
    AccountID: acct.ID, Limit: 20,
  })
  fmt.Println("Audit log entries:", len(logs.Items))
}
```

## Error Handling

```go
acct, err := client.Accounts.Get(ctx, 999)
if err != nil {
  var sdkErr *sdk.Error
  if errors.As(err, &sdkErr) {
    fmt.Println(sdkErr.Message)   // "account not found"
    fmt.Println(sdkErr.Status)    // 404
    fmt.Println(sdkErr.ErrorCode) // 10900
  }
}
```

## Auth

JWT tokens are auto-generated using your ED25519 private key and cached for ~55 minutes (1h TTL with 5min refresh margin). Thread-safe via `sync.Mutex`.

## License

MIT
