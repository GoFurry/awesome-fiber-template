# s3 addon

`addons/s3` is a reusable S3-compatible object storage addon.

It is designed to be copied into any template tier when a project needs object storage, without coupling the `v3/*` templates to a specific provider.

## Features

- AWS SDK for Go v2 based thin wrapper
- default bucket with per-call `WithBucket(...)` override
- upload helpers for `[]byte`, `io.Reader`, and local files
- download as `[]byte` or streaming `io.ReadCloser`
- object metadata lookup via `HeadObject`
- idempotent `DeleteObject`
- pre-signed `GET`, `PUT`, and `DELETE` URLs
- supports custom endpoint and path-style mode for MinIO, R2, Ceph, and other S3-compatible services

## Install

```bash
cd addons/s3
go mod tidy
```

## Config

```go
cfg := s3.Config{
    Region:         "ap-southeast-1",
    Endpoint:       "http://127.0.0.1:9000",
    AccessKey:      "minioadmin",
    SecretKey:      "minioadmin",
    Bucket:         "assets",
    UseSSL:         false,
    ForcePathStyle: true,
    PresignExpire:  15 * time.Minute,
}
```

Fields:

- `Region`: required
- `Endpoint`: optional custom endpoint for S3-compatible providers
- `AccessKey` / `SecretKey` / `SessionToken`: optional explicit credentials
- `Bucket`: optional default bucket
- `UseSSL`: used when `Endpoint` does not already contain a scheme
- `ForcePathStyle`: recommended for MinIO and many self-hosted S3-compatible services
- `PresignExpire`: default pre-signed URL expiration, defaults to `15m`

## Usage

```go
package storage

import (
    "context"
    "time"

    appstorage "your/project/internal/storage"
)

func newStorage(ctx context.Context) (*appstorage.Service, error) {
    return appstorage.New(ctx, appstorage.Config{
        Region:         "ap-southeast-1",
        Endpoint:       "http://127.0.0.1:9000",
        AccessKey:      "minioadmin",
        SecretKey:      "minioadmin",
        Bucket:         "assets",
        UseSSL:         false,
        ForcePathStyle: true,
        PresignExpire:  10 * time.Minute,
    })
}
```

Upload:

```go
result, err := svc.UploadBytes(
    ctx,
    "avatars/u1.png",
    fileBytes,
    s3.WithContentType("image/png"),
    s3.WithCacheControl("public, max-age=31536000"),
    s3.WithMetadata(map[string]string{"source": "avatar"}),
)
```

Download:

```go
data, err := svc.DownloadBytes(ctx, "avatars/u1.png")

stream, info, err := svc.DownloadStream(ctx, "avatars/u1.png")
defer stream.Close()
_ = info
```

Bucket override:

```go
info, err := svc.HeadObject(ctx, "backups/full.tar.gz", s3.WithBucket("backup-bucket"))

backupBucket := svc.Bucket("backup-bucket")
url, headers, err := backupBucket.PresignPut(ctx, "daily/report.json")
```

Pre-signed URLs:

```go
getURL, getHeaders, err := svc.PresignGet(ctx, "avatars/u1.png")

putURL, putHeaders, err := svc.PresignPut(
    ctx,
    "uploads/raw.bin",
    s3.WithContentType("application/octet-stream"),
    s3.WithExpire(5*time.Minute),
)

deleteURL, deleteHeaders, err := svc.PresignDelete(ctx, "uploads/raw.bin")
```

## Exposed API

- `New(ctx, cfg)`
- `(*Service).Client()`
- `(*Service).PresignClient()`
- `(*Service).Bucket(name ...string)`
- `(*Service).HeadObject(...)`
- `(*Service).DeleteObject(...)`
- `(*Service).DownloadBytes(...)`
- `(*Service).DownloadStream(...)`
- `(*Service).UploadBytes(...)`
- `(*Service).UploadReader(...)`
- `(*Service).UploadFile(...)`
- `(*Service).PresignGet(...)`
- `(*Service).PresignPut(...)`
- `(*Service).PresignDelete(...)`

## Notes

- This addon is intentionally thin. Listing, multipart upload managers, copy/move, ACL, tagging, and bucket management are out of scope for the first version.
- `ErrObjectNotFound` is returned for stable not-found handling on read-style operations.
- `DeleteObject` treats not found as success.
- If you need advanced S3 features, use `svc.Client()` or `svc.PresignClient()` directly.
