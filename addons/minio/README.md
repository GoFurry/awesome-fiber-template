# minio addon

Status: placeholder only.

## Purpose

Provide a small MinIO / S3-compatible storage client wrapper.

## Intended Shape

- `Config` for endpoint, access key, secret key, bucket, and SSL options
- `New(...)` for client creation
- helper methods for upload, download, and presigned URLs

## Notes

- Keep it storage-only.
- Do not mix it with application file-management logic.
