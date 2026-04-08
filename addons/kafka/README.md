# kafka addon

Status: placeholder only.

## Purpose

Provide a reusable Kafka producer / consumer wrapper.

## Intended Shape

- `Config` for brokers, topic, group, and auth
- `NewProducer(...)`
- `NewConsumer(...)`

## Notes

- Keep producer and consumer concerns separated.
- Let the application layer decide message contracts.
