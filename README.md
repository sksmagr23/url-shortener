# URL Management Service

A production-ready, high-performance URL Management & Link Analytics backend service built with the [GoFr](https://gofr.dev) framework, Redis, and MongoDB.

---

## Features

### 1. URL Shortening & Link Management
- **Custom Aliases & Auto Generation**: Generate short 6-character random links or specify custom memorable aliases (3–64 characters) with real-time collision checks.
- **Custom Branded Domains**: Support custom short domains (e.g. `me.tech/my-code` or `short.brand.com/my-code`) with domain normalization and HTTP/HTTPS protocol formatting.
- **Link Expiration**: Set UTC expiration timestamps (`expires_at`) for temporary links. Expired links reject redirects with HTTP 404. Expiry dates can be updated or cleared.
- **Public & Private Visibility**: Public links redirect any visitor; private links restrict redirection strictly to the link owner via JWT.
- **Dynamic QR Code Generation**: Allow generating high-resolution PNG image bytes of short urls for instant UI embedding or mobile scanning.
- **URL Management**: Endpoints to create, retrieve details, update target URLs or settings, delete links, and list owned URLs with pagination and sorting.

### 2. Link Analytics & Tracking
- **Granular Click Logging**: Logs every redirection event with timestamp, short code, and target URL ID.
- **Total vs Unique Clicks**: Tracks total redirects (`total_clicks`) alongside distinct unique visitors (`unique_clicks`) based on IP history per link.
- **User-Agent & Device Classification**: Automatically detects and categorizes Browser, OS, and Device Type (Desktop, Mobile, Tablet).
- **Client IP & Proxy Resolution**: Extracts real client IP addresses through proxy headers (`X-Forwarded-For`, `X-Real-IP`).
- **Referrer & Geo Tracking**: Captures HTTP `Referer` traffic sources and resolves IP addresses to country codes.
- **Analytics Summary API**: `GET /urls/{short_code}/analytics` provides link owners with top rankings for Browsers, OS, Devices, Countries, and Referrers.
- **Timeseries Aggregation API**: `GET /urls/{short_code}/analytics/timeseries` aggregates click volume over customizable intervals (`hour` or `day`).

### 3. User Accounts
- **Registration & Security**: Account creation with unique username and email validation, securing passwords with `bcrypt`.
- **JWT Authentication**: JSON Web Token authentication for protected user profiles, API key management, and link operations.
- **Profile Management**: View and update profile details (username, email, password).

### 4. Developer API Keys & Automation
- **Why API Keys alongside JWT?**:
  - **JWT Tokens**: Short-lived login sessions for human users browsing the website interface.
  - **API Keys (`usk_...`)**: Long-lived access keys for automated scripts, Python code, Slack bots, and CI/CD tools to create and manage links programmatically without hardcoding account passwords.
- **Leak Protection & Instant Revocation**: Generate separate keys for different applications. If a script or key is ever leaked, you can instantly revoke that single key without changing your account password.

### 5. Redis Caching & High-Performance Redirects
- **Sub-Millisecond Redirection Caching**: Short code target lookups are cached in Redis (`url:<short_code>`), achieving high-throughput sub-millisecond redirection responses.
- **Automatic Cache Invalidation**: Modifying or deleting links immediately purges stale cache entries from Redis.
- **Resilient Fallback Strategy**: If Redis is offline or unconfigured, the application seamlessly falls back to MongoDB without throwing errors or dropping requests.

### 6. Security, Rate Limiting & Dual Authentication
- **Dual Header Authentication**: Supports both **JWT Bearer tokens** and **Developer API Keys** (`X-API-Key: usk_...`) across all protected endpoints.
- **Sliding-Window Rate Limiting**: Enforces strict rate limits (100 requests per minute per IP/User) returning `HTTP 429 Too Many Requests` with `Retry-After: 60` headers to prevent link abuse and brute-force attacks.
- **RFC 3986 URL Safety & Malicious Scheme Prevention**: Strict target URL parsing and validation rejecting unsafe protocols (`javascript:`, `data:`, `vbscript:`, `file:`, `blob:`).

---

## Project Structure

```text
main.go                 # Application entry point & route registrations
handler/                # HTTP request handlers (User, URL, Health)
service/                # Business logic layer (User, URL generation & validation)
auth/                   # JWT creation, token validation & middleware
store/                  # Database & Caching access layer (UserStore, URLStore, AnalyticsStore, URLCache)
model/                  # Domain models & request/response DTOs
static/openapi.json     # OpenAPI 3.0 specification for Swagger UI
configs/                # Configuration directory (.env file)
```

---

## Architecture

The application follows GoFr's recommended **3-Layer Clean Architecture**:

1. **Handler Layer (`handler/`)**: Parses HTTP requests, validates incoming JSON payloads, extracts authenticated user context, and formats JSON responses.
2. **Service Layer (`service/`)**: Contains core domain rules (URL validation, custom code collision checks, domain normalization, password hashing, JWT creation, expiry checks, and privacy permissions).
3. **Store & Cache Layer (`store/`)**: Encapsulates data operations against MongoDB (`users`, `urls`, `click_events`) and Redis cache (`url:<code`).

---

## 🚀 Quick Start

### 1. Clone & Setup
```bash
git clone https://github.com/sksmagr23/url-shortener.git
cd url-shortener
make tidy
```

### 2. Environment Configuration
Create or edit `configs/.env`:
```env
MONGO_URI=mongodb://localhost:27017/
MONGO_DB=url_shortener
GOFR_TELEMETRY=false
SHORT_URL_HOST=http://localhost:8000/
JWT_SECRET=replace-with-a-long-random-secret

# Redis Configuration (Optional: set REDIS_HOST to enable caching)
REDIS_HOST=localhost
REDIS_PORT=6379
```

> **Note on Redis**: If you have Redis running (or started via `docker run -p 6379:6379 -d redis:7`), GoFr automatically connects to it. If you don't run Redis locally, you can comment out `REDIS_HOST` in `.env` to run purely on MongoDB.

### 3. Run Application
```bash
make run
```
The server will start listening at `http://localhost:8000`. Swagger documentation is available at `http://localhost:8000/.well-known/swagger`.

---

## API Documentation

Detailed RESTful API specifications, headers, request/response schemas, and parameter references are documented separately:

👉 **[View Complete API Documentation (API_DOCS.md)](API_DOCS.md)**

Alternatively, when the application is running, you can access the interactive Swagger UI at:
```text
http://localhost:8000/.well-known/swagger
```

---

## Database Schemas (MongoDB)

### `users` Collection
```json
{
  "_id": "ObjectIdHex",
  "username": "testuser",
  "email": "test@example.com",
  "password_hash": "$2a$10$...",
  "api_keys": ["usk_..."],
  "created_at": "2026-06-01T00:00:00Z",
  "updated_at": "2026-06-01T00:00:00Z"
}
```

### `urls` Collection
```json
{
  "_id": "ObjectIdHex",
  "original_url": "https://gofr.dev",
  "short_code": "my-code",
  "user_id": "ObjectIdHex",
  "public": true,
  "custom_domain": "sksm.tech",
  "expires_at": "2028-12-31T23:59:59Z",
  "total_clicks": 42,
  "unique_clicks": 35,
  "created_at": "2026-06-01T00:00:00Z",
  "updated_at": "2026-06-01T00:00:00Z"
}
```

### `click_events` Collection
```json
{
  "_id": "ObjectIdHex",
  "url_id": "ObjectIdHex",
  "short_code": "my-code",
  "timestamp": "2026-08-26T12:00:00Z",
  "ip_address": "203.0.113.195",
  "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) ...",
  "browser": "Chrome",
  "os": "Windows",
  "device_type": "Desktop",
  "country": "US",
  "referrer": "https://google.com"
}
```

---

## 🛠️ Development & Testing

```bash
make help          # List available Makefile targets
make test          # Run all unit tests
make test-coverage # Run unit tests with coverage report
make lint          # Run golangci-lint checks
make lint-fix      # Automatically fix linting/formatting issues
make run           # Start application server
make tidy          # Download and clean Go modules
```

---

#### Author: [Saksham Agrawal](https://github.com/sksmagr23)