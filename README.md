# URL Shortener with GoFr Framework

## Features

- URL shortening with public redirects
- JWT based user authentication
- User registration, login, profile updates, and API key generation
- User-scoped URL management so users can only access their own shortened URLs
- MongoDB persistence through GoFr
- Health checks
- Swagger docs at `/.well-known/swagger`
- Unit and integration-style tests for handlers & services

## Project Structure

```
├── main.go                 # Application entry point
├── go.mod
├── handler/                # HTTP request handlers
├── service/                # Business logic layer
├── auth/                   # JWT handling and authenticated context helpers
├── store/                  # Data access layer
├── model/                  # Data models
├── static/openapi.json     # OpenAPI specification
├── configs/                # Configuration files
├── .golangci.yaml          # Linting configuration
├── Makefile                # Development tasks
├── README.md
```

## Architecture

### Three-Layer Architecture

1. **Handler layer** (`handler/`)
   - HTTP request handling
   - Request parsing and response formatting
   - Authentication and authorization wiring

2. **Service layer** (`service/`)
   - Business logic for users and URLs
   - Password hashing, JWT creation, and API key generation
   - Authorization checks and access controls

3. **Store layer** (`store/`)
   - MongoDB CRUD operations
   - User and URL persistence
   - Data access abstractions

## Quick Start

### 1. Clone the repository

```bash
git clone https://github.com/sksmagr23/url-shortener.git
cd url-shortener-gofr
make tidy
```

### 2. Configure environment

Set the following environment variables in `configs/.env`:

```env
MONGO_URI=
MONGO_DB=
GOFR_TELEMETRY=false
SHORT_URL_HOST=
JWT_SECRET=
```

### 3. Run the application

```bash
make run
# or
go run main.go
```

---

## API Documentation

Base URL: `http://localhost:8000`

Protected endpoints require:

```http
Authorization: Bearer <jwt_token>
```

### 1. Health Check

**Endpoint:** `GET /health`

**Description:** Returns service health and MongoDB connection status.

### 2. Register

**Endpoint:** `POST /users/register`

**Request Body:**

```json
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}
```

### 3. Login

**Endpoint:** `POST /users/login`

**Description:** Authenticate with `identifier` plus `password`.

**Request Body:**

```json
{
  "identifier": "test@example.com",
  "password": "password123"
}
```

**Success Response:**

```json
{
  "data": {
    "token": "<jwt_token>",
    "user": {
      "id": "507f1f77bcf86cd799439011",
      "username": "testuser",
      "email": "test@example.com"
    }
  }
}
```

### 4. Profile

**Endpoints:**
- `GET /users/profile`
- `PUT /users/profile`

**Request Body for update:**

```json
{
  "username": "new-name",
  "email": "new@example.com",
  "password": "new-password"
}
```

### 5. API Key

**Endpoint:** `POST /users/api-key`

**Description:** Generate a new API key for the authenticated user.

### 6. Create Short URL

**Endpoint:** `POST /urls`

**Description:** Create a short URL from a long URL. Requires JWT.

**Request Body:**

```json
{
  "original_url": "https://example.com/very-long-url"
}
```

### 7. Get URL Details

**Endpoint:** `GET /urls/{short_code}`

**Description:** Retrieve URL details for the authenticated user.

### 8. Redirect

**Endpoint:** `GET /{short_code}`

**Description:** Public redirect to the original URL.

--- 

## MongoDB Collections

### `users`

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

### `urls`

```json
{
  "_id": "ObjectId",
  "original_url": "https://example.com/long-url",
  "short_code": "abc123",
  "user_id": "ObjectIdHex",
  "created_at": "2026-06-01T00:00:00Z"
}
```

### Available make Commands

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

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test file
go test ./internal/tests/health_test.go -v
```

### Linting

```bash
# Run linting
make lint

# Run linting with auto-fix
make lint-fix

# Run lint formatting
make lint-format
```

### Test Coverage

- **Unit Tests**: Individual component testing
- **Integration Tests**: Service layer integration
- **Handler Tests**: HTTP endpoint testing
- **Mock Testing**: Using GoFr's built-in mocking

#### Author:- [Saksham Agrawal](https://github.com/sksmagr23)