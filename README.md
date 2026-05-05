# vacant

Check if domain names are available.

DNS first, WHOIS fallback. No third-party APIs.

## Usage

```bash
# CLI
./vacant check example.com foo.io

# Server
./vacant serve
```

## Build

```bash
cd web && npm install && npm run build
cd .. && go build
```

## Docker

```bash
docker build -t vacant .
docker run -p 8080:8080 vacant
```

## API

```bash
curl -X POST {{BASE_URL}}/api/check \
  -H "Content-Type: application/json" \
  -d '{"domains":["example.com","foo.io"]}'
```

Returns: `domain`, `available`, `method` (`dns` or `whois`), `reason`, `error`.

## License

MIT
