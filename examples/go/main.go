package main

import (
  "context"
  "errors"
  "fmt"
  "log"
  "os"

  sdk "github.com/mountos-app/mountos-admin-sdk/go"
)

func main() {
  client, err := sdk.NewClient(sdk.Config{
    BaseURL:    envOrDefault("MOUNTOS_BASE_URL", "https://appserv.example.com"),
    PrivateKey: os.Getenv("MOUNTOS_PRIVATE_KEY"),
  })
  if err != nil {
    log.Fatal("init client:", err)
  }

  ctx := context.Background()

  // --- Accounts ---
  acct, err := client.Accounts.Create(ctx, &sdk.CreateAccountRequest{
    Name:        "Acme Corp",
    Description: "Demo account",
    ProviderInfo: map[string]any{"tier": "enterprise"},
  })
  if err != nil {
    log.Fatal("create account:", err)
  }
  fmt.Println("Created account ID:", acct.ID)

  account, err := client.Accounts.Get(ctx, acct.ID)
  if err != nil {
    log.Fatal("get account:", err)
  }
  fmt.Printf("Account: %s (active=%v, locked=%v)\n", account.Name, account.IsActive, account.Locked)

  list, err := client.Accounts.List(ctx, &sdk.ListOptions{Page: 1, Limit: 10})
  if err != nil {
    log.Fatal("list accounts:", err)
  }
  fmt.Printf("Accounts: %d total, page %d of %d\n", list.Pagination.Total, list.Pagination.Page, list.Pagination.TotalPages)

  _, _ = client.Accounts.Edit(ctx, acct.ID, &sdk.EditAccountRequest{
    Name:        "Acme Corp Updated",
    Description: "Updated description",
  })

  // --- Users ---
  user, err := client.Users.Add(ctx, &sdk.AddUserRequest{
    AccountID: acct.ID,
    Username:  "alice",
    Email:     "alice@acme.com",
    Name:      "Alice Smith",
  })
  if err != nil {
    log.Fatal("add user:", err)
  }
  fmt.Println("Added user ID:", user.ID)

  users, err := client.Users.List(ctx, &sdk.UserListOptions{AccountID: acct.ID, Page: 1, Limit: 20})
  if err != nil {
    log.Fatal("list users:", err)
  }
  for _, u := range users.Items {
    fmt.Printf("  User: %s <%s> (id=%d)\n", u.Username, u.Email, u.ID)
  }

  _, _ = client.Users.Edit(ctx, user.ID, &sdk.EditUserRequest{
    Username: "alice-updated",
    Email:    "alice-new@acme.com",
  })

  // --- Regions ---
  regionResp, err := client.Regions.Create(ctx, &sdk.CreateRegionRequest{
    AccountID: acct.ID,
    Name:      "us-east-1",
  })
  if err != nil {
    log.Fatal("create region:", err)
  }
  fmt.Println("Created region ID:", regionResp.ID)

  region, _ := client.Regions.Get(ctx, regionResp.ID)
  fmt.Printf("Region: %s (exportId=%s)\n", region.Name, region.ExportID)

  // --- Storages ---
  storage, err := client.Storages.Create(ctx, &sdk.CreateStorageRequest{
    AccountID:    acct.ID,
    RegionID:     regionResp.ID,
    Name:         "prod-s3-bucket",
    StorageType:  "object",
    ProviderType: "s3",
    Endpoint:     "https://s3.us-east-1.amazonaws.com",
    Bucket:       "my-mountos-bucket",
    Region:       "us-east-1",
  })
  if err != nil {
    log.Fatal("create storage:", err)
  }
  fmt.Println("Created storage ID:", storage.ID)

  storages, _ := client.Storages.List(ctx, &sdk.StorageListOptions{AccountID: acct.ID})
  for _, s := range storages.Items {
    fmt.Printf("  Storage: %s (type=%s, active=%v)\n", s.Name, s.StorageType, s.IsActive)
  }

  // --- Volumes ---
  _, err = client.Volumes.UpdateQuota(ctx, 1, &sdk.UpdateVolumeQuotaRequest{
    QuotaLimit: 10 * 1024 * 1024 * 1024, // 10 GiB
  })
  if err != nil {
    // Expected to fail if volume doesn't exist
    var sdkErr *sdk.Error
    if errors.As(err, &sdkErr) {
      fmt.Printf("UpdateQuota error: %s (status=%d)\n", sdkErr.Message, sdkErr.Status)
    }
  }

  // --- Audit Logs ---
  logs, err := client.AuditLogs.List(ctx, &sdk.AuditLogListOptions{
    AccountID: &acct.ID,
    Limit:     5,
  })
  if err != nil {
    log.Fatal("list audit logs:", err)
  }
  fmt.Printf("Audit logs: %d entries\n", len(logs.Items))
  for _, entry := range logs.Items {
    fmt.Printf("  [%d] %s (success=%v)\n", entry.ID, entry.Title, entry.Success)
  }

  // --- Service Nodes ---
  nodes, err := client.ServiceNodes.List(ctx, regionResp.ID, "", "", 0)
  if err != nil {
    fmt.Println("List nodes:", err)
  } else {
    fmt.Printf("Service nodes: %d\n", len(nodes))
    for _, n := range nodes {
      fmt.Printf("  Node: %s (type=%s, status=%s)\n", n.NodeID, n.ServiceType, n.Status)
    }
  }

  // --- Vault ---
  if err := client.Vault.Resync(ctx); err != nil {
    fmt.Println("Vault resync:", err)
  } else {
    fmt.Println("Vault resynced")
  }

  fmt.Println("Done.")
}

func envOrDefault(key, fallback string) string {
  if v := os.Getenv(key); v != "" {
    return v
  }
  return fallback
}
