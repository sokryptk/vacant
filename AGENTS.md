# vacant — Domain Availability Agent API

`vacant` checks whether domain names are available for registration. It uses a DNS-first strategy: if a domain has NS records it is assumed registered (fast). If no NS records exist, it falls back to WHOIS for confirmation.

## Endpoints

### `GET /api/health`

Returns service status.

Response:
```json
{"status":"ok"}
```

### `POST /api/check`

Check one or many domains in a single request.

Body (single):
```json
{"domain":"example.com"}
```

Body (batch):
```json
{"domains":["example.com","example.io"]}
```

Max 50 domains per request.

Response:
```json
{
  "results": [
    {
      "domain": "example.com",
      "available": false,
      "method": "dns",
      "reason": "NS records found"
    },
    {
      "domain": "example.io",
      "available": true,
      "method": "whois",
      "reason": "matched \"not found\" on whois.nic.io"
    }
  ]
}
```

Field meanings:
- `method`: `"dns"` means the decision was made via NS lookup. `"whois"` means a WHOIS server was queried because no NS records existed.
- `reason`: Human-readable explanation of the result.
- `error`: Non-empty if the check failed entirely (e.g., WHOIS server unreachable). When `error` is present, `available` reflects the best-effort signal (usually DNS-only).

### `GET /agents.md`

Returns this document as `text/markdown`.

## curl Examples

Single check:
```bash
curl -X POST {{BASE_URL}}/api/check \
  -H "Content-Type: application/json" \
  -d '{"domain":"example.com"}'
```

Batch check:
```bash
curl -X POST {{BASE_URL}}/api/check \
  -H "Content-Type: application/json" \
  -d '{"domains":["foo.com","bar.io","baz.dev"]}'
```

## Limits & Caveats

- 50 domains max per batch.
- Batch concurrency is capped at 5 WHOIS queries in parallel.
- WHOIS responses are not standardized. Rare TLDs may return ambiguous wording.
- Unknown TLDs are resolved via `whois.iana.org` referral.
- WHOIS servers rate-limit aggressively.
- DNS-first results are strong but not absolute.

## Build & Run

```bash
cd web && npm install && npm run build
cd .. && go build
./vacant serve
```
