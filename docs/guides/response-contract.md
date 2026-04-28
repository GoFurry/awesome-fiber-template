# Response Contract Guide

Phase 9 also formalizes the default generated API response shape.

## JSON Envelope

Generated business handlers use:

```json
{
  "code": 1,
  "message": "success",
  "data": {}
}
```

Error responses use the same envelope with:

- `code = 0`
- a meaningful HTTP status
- `message` copied from the routed error

## Important Behavior

- Missing routes should remain `404`, not collapse into `500`.
- `fiber.Error` status codes should be preserved by generated routers.
- `docs`, `ui`, and `metrics` are operational routes and may return non-JSON content.

## Testing Expectation

The generator regression suite now checks missing-route responses for a JSON `404` envelope on the generated services it starts.
