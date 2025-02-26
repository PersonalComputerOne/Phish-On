# BACKEND

```bash
go run ./cmd/api

curl -X POST http://localhost:8080/api/v1/levenshtein -H "Content-Type: application/json" -d '{"urls": ["http://githun.com", "http://github.com", "https://giiiiithdub.com", "https://linkeddin.com", "https://twitter.com"]}'
```
