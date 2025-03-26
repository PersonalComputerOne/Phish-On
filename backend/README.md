# BACKEND

#### 1. Install Goose for Database Migrations

Goose is used for managing database migrations. To install it, run the following:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Once installed, run the migrations:

```bash
goose up
```

#### 2. Seed the Database

To seed the database with initial data, use the following command:

```bash
go run ./seed
```

#### 3. Start the API

To run the API server, execute:

```bash
go run main.go
```

#### 4. Test the API

To test the API, you can make a `POST` request with `curl`:

```bash
curl -X POST http://localhost:8080/api/v1/levenshtein/sequential \
-H "Content-Type: application/json" \
-d '{"urls": ["http://githun.com", "http://github.com", "https://giiiiithdub.com", "https://linkeddin.com", "https://twitter.com"]}'

curl -X POST http://localhost:8080/api/v1/levenshtein/parallel \
-H "Content-Type: application/json" \
-d '{"urls": ["http://githun.com", "http://github.com", "https://giiiiithdub.com", "https://linkeddin.com", "https://twitter.com"]}'
```
