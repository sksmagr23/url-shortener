# API Documentation

## Base URL & Interactive Docs

- **Base URL**: `http://localhost:8000`
- **Interactive Swagger UI**: `http://localhost:8000/.well-known/swagger`
- **OpenAPI Specification File**: `static/openapi.json`

---

## Authentication

Protected endpoints require a Bearer token in the `Authorization` header:

```http
Authorization: Bearer <jwt_token>
```

Private short links (`"public": false`) also require the owner's Bearer JWT token when calling the redirection endpoint `GET /{short_code}`.

---

## Endpoints Overview

| Category | Method | Endpoint | Description | Auth Required |
|---|---|---|---|---|
| **Health** | `GET` | `/health` | Application health and MongoDB status | ❌ No |
| **Users** | `POST` | `/users/register` | Register a new user account | ❌ No |
| **Users** | `POST` | `/users/login` | Login and receive JWT token | ❌ No |
| **Users** | `GET` | `/users/profile` | Get authenticated user profile | ✅ Yes |
| **Users** | `PUT` | `/users/profile` | Update user profile details | ✅ Yes |
| **API Keys** | `POST` | `/users/api-key` | Generate a new developer API key | ✅ Yes |
| **API Keys** | `GET` | `/users/api-keys` | List active developer API keys | ✅ Yes |
| **API Keys** | `DELETE` | `/users/api-keys/{api_key}` | Revoke an active API key | ✅ Yes |
| **URLs** | `POST` | `/urls` | Create a shortened URL | ✅ Yes |
| **URLs** | `GET` | `/urls` | List owned URLs (paginated & sorted) | ✅ Yes |
| **URLs** | `GET` | `/urls/{short_code}` | Get URL details by short code | ✅ Yes |
| **URLs** | `PUT` | `/urls/{short_code}` | Update URL target, domain, or expiry | ✅ Yes |
| **URLs** | `DELETE` | `/urls/{short_code}` | Delete an owned short URL | ✅ Yes |
| **Redirection** | `GET` | `/{short_code}` | Redirect to original target URL | ⚠️ Public / Owner |

---

## 1. Health Endpoint

### `GET /health`
Returns the health status of the application and its connected datasources (e.g. MongoDB).

- **Authentication**: None
- **Response `200 OK`**:
  ```json
  {
    "data": {
      "status": "UP",
      "details": {
        "mongoDB": "UP"
      }
    }
  }
  ```

---

## 2. User Management Endpoints

### `POST /users/register`
Registers a new user account.

- **Authentication**: None
- **Request Body**:
  ```json
  {
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }
  ```
  - `username` *(string, **Required**)*: Unique username.
  - `email` *(string, **Required**)*: Valid email address.
  - `password` *(string, **Required**)*: User account password.

- **Response `201 Created`**:
  ```json
  {
    "data": {
      "id": "607f1f77bcf86cd799439011",
      "username": "testuser",
      "email": "test@example.com",
      "api_keys": [],
      "created_at": "2026-08-26T12:00:00Z",
      "updated_at": "2026-08-26T12:00:00Z"
    }
  }
  ```
- **Errors**: `400 Bad Request`, `409 Conflict` (username or email already registered).

---

### `POST /users/login`
Authenticates a user using email or username and returns a signed JWT token.

- **Authentication**: None
- **Request Body**:
  ```json
  {
    "identifier": "test@example.com",
    "password": "password123"
  }
  ```
  - `identifier` *(string, **Required**)*: Registered email or username.
  - `password` *(string, **Required**)*: Account password.

- **Response `200 OK`**:
  ```json
  {
    "data": {
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "user": {
        "id": "607f1f77bcf86cd799439011",
        "username": "testuser",
        "email": "test@example.com",
        "api_keys": []
      }
    }
  }
  ```
- **Errors**: `400 Bad Request`, `401 Unauthorized` (invalid credentials).

---

### `GET /users/profile`
Retrieves profile details for the authenticated user.

- **Authentication**: Required (`Bearer <jwt_token>`)
- **Response `200 OK`**:
  ```json
  {
    "data": {
      "id": "607f1f77bcf86cd799439011",
      "username": "testuser",
      "email": "test@example.com",
      "api_keys": ["usk_abc123..."],
      "created_at": "2026-08-26T12:00:00Z",
      "updated_at": "2026-08-26T12:00:00Z"
    }
  }
  ```
- **Errors**: `401 Unauthorized`.

---

### `PUT /users/profile`
Updates profile information for the authenticated user.

- **Authentication**: Required (`Bearer <jwt_token>`)
- **Request Body** *(All fields optional; supply only fields to change)*:
  ```json
  {
    "username": "newusername",
    "email": "newemail@example.com",
    "password": "newpassword123"
  }
  ```
  - `username` *(string, Optional)*: Updated username.
  - `email` *(string, Optional)*: Updated email address.
  - `password` *(string, Optional)*: Updated password.

- **Response `200 OK`**:
  ```json
  {
    "data": {
      "id": "607f1f77bcf86cd799439011",
      "username": "newusername",
      "email": "newemail@example.com"
    }
  }
  ```
- **Errors**: `400 Bad Request`, `401 Unauthorized`.

---

## 3. API Key Management Endpoints

### `POST /users/api-key`
Generates a new developer API key for programmatic access.

- **Authentication**: Required (`Bearer <jwt_token>`)
- **Response `201 Created`**:
  ```json
  {
    "data": {
      "api_key": "usk_9f83a71b4c02e56d8123..."
    }
  }
  ```
- **Errors**: `401 Unauthorized`.

---

### `GET /users/api-keys`
Lists all active API keys owned by the authenticated user.

- **Authentication**: Required (`Bearer <jwt_token>`)
- **Response `200 OK`**:
  ```json
  {
    "data": {
      "api_keys": [
        "usk_9f83a71b4c02e56d8123...",
        "usk_7a81b20c3d4e5f6g7h8i..."
      ]
    }
  }
  ```
- **Errors**: `401 Unauthorized`.

---

### `DELETE /users/api-keys/{api_key}`
Revokes an active API key.

- **Authentication**: Required (`Bearer <jwt_token>`)
- **Path Parameters**:
  - `api_key` *(string, **Required**)*: The exact API key string to revoke.
- **Response `200 OK`**:
  ```json
  {
    "data": {
      "message": "API key revoked successfully"
    }
  }
  ```
- **Errors**: `400 Bad Request`, `401 Unauthorized`.

---

## 4. URL Management Endpoints

### `POST /urls`
Creates a shortened URL.

- **Authentication**: Required (`Bearer <jwt_token>`)
- **Request Body**:
  ```json
  {
    "original_url": "https://gofr.dev",
    "custom_code": "my-code",
    "public": true,
    "custom_domain": "sksm.tech",
    "expires_at": "2028-12-31T23:59:59Z"
  }
  ```
  - `original_url` *(string, **Required**)*: Valid HTTP or HTTPS target URL.
  - `custom_code` *(string, Optional)*: Custom short code alias (3–64 alphanumeric characters). Auto-generated 6-char random code if omitted.
  - `public` *(boolean, Optional, default: `false`)*: Link visibility permissions (`true` = public redirect, `false` = owner auth required).
  - `custom_domain` *(string, Optional)*: Custom branded domain (e.g. `sksm.tech` or `short.domain.com`).
  - `expires_at` *(string RFC3339 date-time, Optional)*: UTC expiration timestamp.

- **Response `201 Created`**:
  ```json
  {
    "data": {
      "id": "607f1f77bcf86cd799439011",
      "original_url": "https://gofr.dev",
      "short_code": "my-code",
      "user_id": "607f1f77bcf86cd799439012",
      "public": true,
      "custom_domain": "sksm.tech",
      "expires_at": "2028-12-31T23:59:59Z",
      "total_clicks": 0,
      "unique_clicks": 0,
      "short_url": "https://sksm.tech/my-code",
      "created_at": "2026-08-26T12:00:00Z",
      "updated_at": "2026-08-26T12:00:00Z"
    }
  }
  ```
- **Errors**: `400 Bad Request`, `401 Unauthorized`, `409 Conflict` (custom code already in use).

---

### `GET /urls`
Lists all URLs owned by the authenticated user with pagination and sorting.

- **Authentication**: Required (`Bearer <jwt_token>`)
- **Query Parameters**:
  - `page` *(integer, Optional, default: `1`)*: Page number.
  - `limit` *(integer, Optional, default: `10`, max: `100`)*: Number of items per page.
  - `sort` *(string, Optional, default: `"created_at"`)*: Sorting field (`"created_at"`, `"short_code"`, `"total_clicks"`).
  - `order` *(string, Optional, default: `"desc"`)*: Sorting direction (`"asc"` or `"desc"`).

- **Response `200 OK`**:
  ```json
  {
    "data": [
      {
        "id": "607f1f77bcf86cd799439011",
        "original_url": "https://gofr.dev",
        "short_code": "my-code",
        "public": true,
        "custom_domain": "sksm.tech",
        "total_clicks": 42,
        "unique_clicks": 35,
        "short_url": "https://sksm.tech/my-code",
        "created_at": "2026-08-26T12:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 1,
      "total_pages": 1
    }
  }
  ```
- **Errors**: `401 Unauthorized`.

---

### `GET /urls/{short_code}`
Retrieves URL details for a specific URL owned by the authenticated user.

- **Authentication**: Required (`Bearer <jwt_token>`)
- **Path Parameters**:
  - `short_code` *(string, **Required**)*: Short code identifier.

- **Response `200 OK`**:
  ```json
  {
    "data": {
      "id": "607f1f77bcf86cd799439011",
      "original_url": "https://gofr.dev",
      "short_code": "my-code",
      "user_id": "607f1f77bcf86cd799439012",
      "public": true,
      "custom_domain": "sksm.tech",
      "total_clicks": 42,
      "unique_clicks": 35,
      "short_url": "https://sksm.tech/my-code",
      "created_at": "2026-08-26T12:00:00Z",
      "updated_at": "2026-08-26T12:00:00Z"
    }
  }
  ```
- **Errors**: `401 Unauthorized`, `404 Not Found`.

---

### `PUT /urls/{short_code}`
Updates configuration or target URL for an owned link.

- **Authentication**: Required (`Bearer <jwt_token>`)
- **Path Parameters**:
  - `short_code` *(string, **Required**)*: Short code identifier.
- **Request Body** *(All fields optional)*:
  ```json
  {
    "original_url": "https://gofr.dev/docs",
    "public": false,
    "custom_domain": "links.mysite.com",
    "clear_custom_domain": false,
    "expires_at": "2028-12-31T23:59:59Z",
    "clear_expiry": false
  }
  ```
  - `original_url` *(string, Optional)*: Updated target URL.
  - `public` *(boolean, Optional)*: Update link visibility (`true`/`false`).
  - `custom_domain` *(string, Optional)*: Update custom branded domain.
  - `clear_custom_domain` *(boolean, Optional, default: `false`)*: Set `true` to remove custom domain and revert to default host.
  - `expires_at` *(string RFC3339 date-time, Optional)*: Update UTC expiration timestamp.
  - `clear_expiry` *(boolean, Optional, default: `false`)*: Set `true` to remove link expiration date.

- **Response `200 OK`**:
  ```json
  {
    "data": {
      "id": "607f1f77bcf86cd799439011",
      "original_url": "https://gofr.dev/docs",
      "short_code": "my-code",
      "public": false,
      "custom_domain": "links.mysite.com",
      "short_url": "https://links.mysite.com/my-code",
      "updated_at": "2026-08-26T13:00:00Z"
    }
  }
  ```
- **Errors**: `400 Bad Request`, `401 Unauthorized`, `404 Not Found`.

---

### `DELETE /urls/{short_code}`
Deletes a URL owned by the authenticated user.

- **Authentication**: Required (`Bearer <jwt_token>`)
- **Path Parameters**:
  - `short_code` *(string, **Required**)*: Short code identifier.

- **Response `200 OK`**:
  ```json
  {
    "data": {
      "message": "URL deleted successfully"
    }
  }
  ```
- **Errors**: `401 Unauthorized`, `404 Not Found`.

---

## 5. Redirection Endpoint

### `GET /{short_code}`
Redirects callers to the original target URL.
- **Performance Optimization**: Served with **sub-millisecond latency** via Redis cache (`url:<short_code>`).
- **Asynchronous Processing**: Click metadata logging (IP, User-Agent, Referrer, Browser, OS, Device Type, Country) and unique click updates run in background workers without blocking HTTP redirection.

- **Authentication**: None for public links (`"public": true`); Required (`Bearer <jwt_token>`) for private links (`"public": false`).
- **Path Parameters**:
  - `short_code` *(string, **Required**)*: Short code identifier.

- **Response `302 Found`**:
  - **Headers**:
    ```http
    Location: https://gofr.dev/docs
    ```

- **Errors**:
  - `401 Unauthorized`: Private URL accessed without valid owner Bearer JWT token.
  - `404 Not Found`: URL does not exist or link has expired (`expires_at < current_time`).
