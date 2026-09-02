# Agent Access

Add **vacant** as a skill to your LLM agent.

## Quick Setup

Paste this into your agent session:

```
Install and configure vacant by following the instructions here: {{BASE_URL}}/agents.md
```

## Skill Config

```json
{
  "name": "vacant",
  "description": "Check domain name availability via DNS and WHOIS",
  "endpoint": "{{BASE_URL}}/api/check",
  "method": "POST",
  "headers": { "Content-Type": "application/json" },
  "body": { "domains": ["example.com"] },
  "response": {
    "domain": "string",
    "available": "boolean",
    "method": "dns | whois",
    "reason": "string",
    "error": "string"
  }
}
```

## API Reference

### `POST /api/check`

Check one or many domains.

**Single domain:**
```bash
curl -X POST {{BASE_URL}}/api/check \
  -H "Content-Type: application/json" \
  -d '{"domain":"example.com"}'
```

**Batch (max 1000):**
```bash
curl -X POST {{BASE_URL}}/api/check \
  -H "Content-Type: application/json" \
  -d '{"domains":["foo.com","bar.io","baz.dev"]}'
```

**Response:**
```json
{
  "results": [
    {
      "domain": "example.com",
      "available": false,
      "method": "dns",
      "reason": "NS records found"
    }
  ]
}
```

## How It Works

- **DNS first**: If a domain has NS records, it's taken (fast).
- **WHOIS fallback**: If no NS records exist, queries the TLD's WHOIS server for confirmation.
- **Recursive referrals**: Follows registrar referrals up to 3 hops, matching standard `whois` CLI behavior.
- **Self-healing TLD resolution**: Unknown TLDs are resolved via `whois.iana.org` and cached.

### `GET /api/tlds`

List all delegated TLDs from IANA.

```bash
curl {{BASE_URL}}/api/tlds
```

**Response:**
```json
{
  "count": 1437,
  "tlds": ["aaa", "aarp", "abb", "abbott", "abogado", ...]
}
```

Cached for 24 hours.
