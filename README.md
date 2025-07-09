# URL Shortener with GoFr Framework

## Features - PR1

- **URL Shortening**: Create short URLs from long URLs
- **Custom Host Support**: Configurable short URL host via environment variables
- **Health Checks**: Built-in health monitoring endpoints
- **Comprehensive Testing**: Unit tests for all components
- **Code Quality**: Linting with golangci-lint
- **Three-Layer Architecture**

## Project Structure

```
├── main.go                 # Application entry point
├── go.mod
├── internal/
│   ├── handler/           # HTTP request handlers
│   ├── service/           # Business logic layer
│   ├── store/            # Data access layer
│   ├── model/            # Data models
│   └── tests/            # Test files
├── configs/              # Configuration files
├── .golangci.yml         # Linting configuration
├── Makefile              # Development tasks
└── README.md
```

## Architecture

### Three-Layer Architecture

1. **Handler Layer** (`internal/handler/`)
   - HTTP request/response handling
   - Input validation
   - Error handling

2. **Service Layer** (`internal/service/`)
   - Business logic
   - URL generation and validation
   - Environment configuration

3. **Store Layer** (`internal/store/`)
   - Data persistence
   - MongoDB operations
   - Database abstraction

## Quick Start

### 1. Setup

```bash
git clone https://github.com/sksmagr23/url-shortener.git
cd url-shortener-gofr

make tidy
```

### 2. Configuration

Set the following environment variables:

```env
MONGO_URI=mongodb://localhost:27017/
MONGO_DB=url_shortener
GOFR_TELEMETRY=false
SHORT_URL_HOST=http://localhost:8000/
```

### 3. Run the Application

```bash
make run
# Or
go run main.go
```

## API Documentation

#### Base URL : `http://localhost:8000`


### 1. Health Check

**Endpoint:** `GET /health`
**Description:** Check the health status of the application and its dependencies.

**Response:**
```json
{
  "data": {
    "status": "healthy",
    "timestamp": "2024-01-01T12:00:00Z",
    "services": {
      "mongoDB": "connected" // or disconnected
    }
  }
}
```

### 2. Create Short URL

**Endpoint:** `POST /api/urls`
**Description:** Create a new short URL from a long URL.
**Request Body:**
```json
{
  "original_url": "https://example.com/very-long-url-that-needs-to_shorten"
}
```

**Success Response (200):**
```json
{
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "original_url": "https://example.com/very-long-url-that-needs-shortening",
    "short_code": "abc123",
    "short_url": "http://localhost:8000/abc123",
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

**Error Response (400) - Invalid URL:**
```json
{
  "error": {
    "message": "invalid URL"
  }
}
```

### 3. Get URL Details

**Endpoint:** `GET /api/urls/{short_code}`
**Description:** Retrieve details of a short URL by its short code.

**Success Response (200):**
```json
{
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "original_url": "https://example.com/very-long-url-that-needs-shortening",
    "short_code": "abc123",
    "short_url": "http://localhost:8000/abc123",
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

**Error Response (404) - URL Not Found:**
```json
{
  "error": {
    "message": "mongo: no documents in result"
  }
}
```

### 4. Redirect to Original URL

**Endpoint:** `GET /{short_code}`
**Description:** Redirect to the original URL using the short code.

**Success Response (302):**
```
HTTP/1.1 302 Found
Location: https://example.com/very-long-url-that-needs-shortening
```

**Error Response (404) - URL Not Found:**
```json
{
  "error": {
    "message": "mongo: no documents in result"
  }
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


### MongoDB Urls collection
```json
{
  "_id": "ObjectId",
  "original_url": "https://example.com/long-url",
  "short_code": "abc123",
  "created_at": "2024-01-01T00:00:00Z"
}
```