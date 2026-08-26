# URL Shortener with GoFr Framework

A production-ready, feature-rich URL shortener backend service built with the [GoFr](https://gofr.dev) framework and MongoDB.

---

## Features

### 1. URL Shortening & Link Management
- **Custom Aliases & Auto Generation**: Generate short 6-character random links or specify custom memorable aliases (3–64 characters) with real-time collision checks.
- **Custom Branded Domains**: Support custom short domains (e.g. `me.tech/my-code` or `short.brand.com/my-code`) with domain normalization and HTTP/HTTPS protocol formatting.
- **Link Expiration**: Set UTC expiration timestamps (`expires_at`) for temporary links. Expired links reject redirects with HTTP 404. Expiry dates can be updated or cleared.
- **Public & Private Visibility**: Public links redirect any visitor; private links restrict redirection strictly to the link owner via JWT.
- **URL Management**: Endpoints to create, retrieve details, update target URLs or settings, delete links, and list owned URLs with pagination and sorting.

### 2. Link Analytics & Click Tracking
- **Granular Click Logging**: Logs every redirection event with timestamp, short code, and target URL ID.
- **Total vs Unique Clicks**: Tracks total redirects (`total_clicks`) alongside distinct unique visitors (`unique_clicks`) based on IP history per link.
- **User-Agent & Device Classification**: Automatically detects and categorizes:
  - **Browser**: Chrome, Safari, Firefox, Edge, Opera, Other.
  - **OS**: Windows, macOS, Linux, iOS, Android, Unknown.
  - **Device Type**: Desktop, Mobile, Tablet.
- **Client IP & Proxy Resolution**: Extracts real client IP addresses through proxy headers (`X-Forwarded-For`, `X-Real-IP`).
- **Referrer & Geo Tracking**: Captures HTTP `Referer` traffic sources and resolves IP addresses to country codes.

### 3. User Accounts
- **Registration & Security**: Account creation with unique username and email validation, securing passwords with `bcrypt`.
- **JWT Authentication**: JSON Web Token authentication for protected user profiles, API key management, and link operations.
- **Profile Management**: View and update profile details (username, email, password).

### 4. Developer API Keys
- **API Key Generation**: Cryptographically generated API keys (`usk_...`) for external API integration.
- **Key Management**: List active keys and instantly revoke compromised or obsolete API keys.

---

## 📁 Project Structure

```text
main.go                 # Application entry point & route registrations
handler/                # HTTP request handlers (User, URL, Health)
service/                # Business logic layer (User, URL generation & validation)
auth/                   # JWT creation, token validation & middleware
store/                  # MongoDB data access layer (UserStore & URLStore)
model/                  # Domain models & request/response DTOs
static/openapi.json     # OpenAPI 3.0 specification for Swagger UI
configs/                # Configuration directory (.env file)
```

---

## 🏛️ Architecture

The application follows GoFr's recommended **3-Layer Clean Architecture**:

1. **Handler Layer (`handler/`)**: Parses HTTP requests, validates incoming JSON payloads, extracts authenticated user context, and formats JSON responses.
2. **Service Layer (`service/`)**: Contains core domain rules (URL validation, custom code collision checks, domain normalization, password hashing, JWT creation, expiry checks, and privacy permissions).
3. **Store Layer (`store/`)**: Encapsulates database CRUD operations against MongoDB collections (`users` and `urls`).

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
```

### 3. Run Application
```bash
make run
```
The server will start listening at `http://localhost:8000`. Swagger documentation is available at `http://localhost:8000/.well-known/swagger`.

---

## 🔐 Authentication

Protected endpoints require a Bearer token in the `Authorization` header:

```http
Authorization: Bearer <jwt_token>
```

Private short links (`"public": false`) require the owner's JWT token when accessing the redirection endpoint `GET /{short_code}`.

---

## 📖 API Documentation

### 1. Health Endpoint

#### `GET /health`
Returns system status and datasource health (e.g., MongoDB connectivity).
- **Authentication**: None

---

### 2. User Management Endpoints

#### `POST /users/register`
Registers a new user account.
- **Authentication**: None
- **Request Body**:
  - `username` *(string, **Required**)*: Unique username.
  - `email` *(string, **Required**)*: Valid email address.
  - `password` *(string, **Required**)*: User account password.

#### `POST /users/login`
Authenticates a user and returns a JWT token.
- **Authentication**: None
- **Request Body**:
  - `identifier` *(string, **Required**)*: Registered email or username.
  - `password` *(string, **Required**)*: Account password.

#### `GET /users/profile`
Retrieves profile details for the authenticated user.
- **Authentication**: Required (`Bearer <jwt_token>`)

#### `PUT /users/profile`
Updates profile information for the authenticated user.
- **Authentication**: Required (`Bearer <jwt_token>`)
- **Request Body** *(All fields optional)*:
  - `username` *(string, Optional)*: New username.
  - `email` *(string, Optional)*: New email address.
  - `password` *(string, Optional)*: New password.

---

### 3. API Key Management Endpoints

#### `POST /users/api-key`
Generates a new developer API key.
- **Authentication**: Required (`Bearer <jwt_token>`)

#### `GET /users/api-keys`
Lists all active API keys owned by the user.
- **Authentication**: Required (`Bearer <jwt_token>`)

#### `DELETE /users/api-keys/{api_key}`
Revokes an active API key.
- **Authentication**: Required (`Bearer <jwt_token>`)
- **Path Parameters**:
  - `api_key` *(string, **Required**)*: The API key string to revoke.

---

### 4. URL Management Endpoints

#### `POST /urls`
Creates a shortened URL.
- **Authentication**: Required (`Bearer <jwt_token>`)
- **Request Body**:
  - `original_url` *(string, **Required**)*: Destination HTTP/HTTPS target URL.
  - `custom_code` *(string, Optional)*: Desired custom alias (3–64 chars). Auto-generated if omitted.
  - `public` *(boolean, Optional, default: `false`)*: Link visibility (`true` for public, `false` for owner-only).
  - `custom_domain` *(string, Optional)*: Custom domain alias (e.g. `sksm.tech` or `short.example.com`).
  - `expires_at` *(string RFC3339 date-time, Optional)*: UTC expiration timestamp.

#### `GET /urls`
Lists all URLs owned by the authenticated user with pagination and sorting.
- **Authentication**: Required (`Bearer <jwt_token>`)
- **Query Parameters**:
  - `page` *(integer, Optional, default: `1`)*: Page number.
  - `limit` *(integer, Optional, default: `10`, max: `100`)*: Items per page.
  - `sort` *(string, Optional, default: `"created_at"`)*: Sort field (`"created_at"`, `"short_code"`, `"total_clicks"`).
  - `order` *(string, Optional, default: `"desc"`)*: Sort direction (`"asc"` or `"desc"`).

#### `GET /urls/{short_code}`
Retrieves URL details for an owned link.
- **Authentication**: Required (`Bearer <jwt_token>`)
- **Path Parameters**:
  - `short_code` *(string, **Required**)*: Short code identifier.

#### `PUT /urls/{short_code}`
Updates configuration for an owned URL.
- **Authentication**: Required (`Bearer <jwt_token>`)
- **Path Parameters**:
  - `short_code` *(string, **Required**)*: Short code identifier.
- **Request Body** *(All fields optional)*:
  - `original_url` *(string, Optional)*: Updated target URL.
  - `public` *(boolean, Optional)*: Update link visibility.
  - `custom_domain` *(string, Optional)*: Update custom domain alias.
  - `clear_custom_domain` *(boolean, Optional, default: `false`)*: Set `true` to remove custom domain.
  - `expires_at` *(string RFC3339 date-time, Optional)*: Update expiration timestamp.
  - `clear_expiry` *(boolean, Optional, default: `false`)*: Set `true` to remove link expiration date.

#### `DELETE /urls/{short_code}`
Deletes an owned URL record.
- **Authentication**: Required (`Bearer <jwt_token>`)
- **Path Parameters**:
  - `short_code` *(string, **Required**)*: Short code identifier.

---

### 5. Redirection Endpoint

#### `GET /{short_code}`
Redirects callers to the original target URL and increments click count.
- **Authentication**: None for public links; Required (`Bearer <jwt_token>`) for private links.
- **Response**: `HTTP 302 Found` with `Location` header pointing to `original_url`.

---

## 🗄️ Database Schemas (MongoDB)

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