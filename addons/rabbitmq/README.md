# rabbitmq addon

Status: placeholder only.

## Purpose

Provide a reusable RabbitMQ connection and channel wrapper.

## Intended Shape

- `Config` for URI, exchange, queue, and routing settings
- `New(...)` for connection setup
- `Close()` for cleanup

## Notes

- Keep the API thin enough to be copied into small services.
