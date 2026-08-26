# URL Shortener with GoFr Framework

## Features

- URL shortening with generated or custom short codes
- JWT based user authentication
- User registration, login, profile updates, and API key management
- User-owned links with public/private visibility
- URL CRUD operations: create, list, get, update, and delete
- Public redirects for public links; private redirects require owner authentication
- Expiring links, custom domains, and click counters on URL records
- MongoDB persistence through GoFr
- Health checks and Swagger docs at `/.well-known/swagger`
- Unit and integration-style tests for handlers, services, and stores

## Project Structure

```text
main.go                 # Application entry point
handler/                # HTTP request handlers
service/                # Business logic layer
auth/                   # JWT handling and authenticated context helpers
store/                  # MongoDB data access layer
model/                  # Data models
static/openapi.json     # OpenAPI specification
configs/                # Configuration files
```

## Architecture

1. Handler layer (`handler/`)
   HTTP request parsing, response formatting, and authenticated user extraction.

2. Service layer (`service/`)
   Business logic for users, URLs, password hashing, JWT creation, API key generation, ownership checks, privacy checks, and expiry checks.

3. Store layer (`store/`)
   MongoDB CRUD operations for users, URLs, and API key storage.

## Quick Start

```bash
git clone https://github.com/sksmagr23/url-shortener.git
cd url-shortener-gofr
make tidy
make run
```

Set the following environment variables in `configs/.env`:

```env
MONGO_URI=mongodb://localhost:27017/
MONGO_DB=url_shortener
GOFR_TELEMETRY=false
SHORT_URL_HOST=http://localhost:8000/
JWT_SECRET=replace-with-a-long-random-secret
```

## Authentication

Protected endpoints require:

```http
Authorization: Bearer <jwt_token>
```

Private redirect links can also be accessed by the owner when the same bearer token is supplied.

## API

Base URL: `http://localhost:8000`

### Health

`GET /health`

### Users

`POST /users/register`

```json
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}
```

`POST /users/login`

```json
{
  "identifier": "test@example.com",
  "password": "password123"
}
```

`GET /users/profile`

`PUT /users/profile`

```json
{
  "username": "new-name",
  "email": "new@example.com",
  "password": "new-password"
}
```

### API Keys

`POST /users/api-key`

Generates and stores a new API key.

`GET /users/api-keys`

Lists API keys for the authenticated user.

`DELETE /users/api-keys/{api_key}`

Revokes an API key for the authenticated user.

### URLs

`POST /urls`

```json
{
  "original_url": "https://example.com/very-long-url",
  "custom_code": "my-code",
  "public": true,
  "custom_domain": "https://short.example.com",
  "expires_at": "2026-12-31T23:59:59Z"
}
```

`GET /urls?page=1&limit=10&sort=created_at&order=desc`

Lists URLs owned by the authenticated user.

`GET /urls/{short_code}`

Returns details for an owned URL.

`PUT /urls/{short_code}`

```json
{
  "original_url": "https://example.com/new-target",
  "public": false,
  "custom_domain": "https://short.example.com",
  "expires_at": "2026-12-31T23:59:59Z",
  "clear_expiry": false
}
```

`DELETE /urls/{short_code}`

Deletes an owned URL.

`GET /{short_code}`

Redirects to the original URL. Public links redirect without authentication; private links require the owner JWT.

## MongoDB Collections

`users`

```json
{
  "_id": "ObjectIdHex",
  "username": "testuser",
  "email": "test@example.com",
  "password_hash": "bcrypt-hash",
  "api_keys": ["usk_..."],
  "created_at": "2026-06-01T00:00:00Z",
  "updated_at": "2026-06-01T00:00:00Z"
}
```

`urls`

```json
{
  "_id": "ObjectIdHex",
  "original_url": "https://example.com/long-url",
  "short_code": "abc123",
  "user_id": "ObjectIdHex",
  "public": true,
  "custom_domain": "https://short.example.com",
  "expires_at": "2026-12-31T23:59:59Z",
  "total_clicks": 0,
  "unique_clicks": 0,
  "created_at": "2026-06-01T00:00:00Z",
  "updated_at": "2026-06-01T00:00:00Z"
}
```

## Development

```bash
make help          # Show all available commands
make test          # Run all tests
make lint          # Run linting
make lint-fix      # Run linting with auto-fix
make lint-format   # Run linting formatting
make run           # Run the application
make tidy          # Install dependencies
make setup         # Setup the project
make clean         # Clean build artifacts
make test-coverage # Run tests with coverage
```

### Test Coverage

- **Unit Tests**: Individual component testing
- **Integration Tests**: Service layer integration
- **Handler Tests**: HTTP endpoint testing
- **Mock Testing**: Using GoFr's built-in mocking

#### Author:- [Saksham Agrawal](https://github.com/sksmagr23)