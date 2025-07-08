# GoFr Summer of Code 2025 Proposal
## URL Shortener with Insights - Template Track

---

## Problem Statement

**Title:** Building a Full-Featured URL Shortener with Analytics using GoFr Framework

**Problem Relevance/Interest:**
.
- URL shortening is a popular feature in contemporary web applications, but most available solutions keep users stuck on proprietary platforms with limited customizations and exposure of analytics. I am drawn to this project because it solves an actual problem and demonstrates GoFr's ability to develop production-level backend services.
- This project aligns with my interest in scalable web service creation and communication with modern Go frameworks. I am particularly thrilled about the analytics section since it involves real-time handling and visualization concepts that are crucial in modern backend development.
---

## Personal Info and Experience

#### Name :- Saksham Agrawal
#### Current Education :- Pre-final Year, B.Tech at IIT (BHU) Varanasi
#### Email :- sakshamag34@gmail.com
#### Location :- Varanasi, U.P. India

### Skillset

- **Languages** : C++, Python, JavaScript, Tpescript, Golang 
- **Technologies/Frameworks** : HTML, React.js, Node.js, Express, EJS, Next.js, Django, Tailwind CSS - 
- **Tools/Platforms** : Git, GitHub, Docker, MongoDB, Firebase, Linux, Postman, REST APIs 
- **Areas of Interest** : Software Development, Open-Source, Data Structures & Algorithms

### Relevant Projects
- Codexhange - inspired by Stack Overflow, built with Node.js, MongoDB, EJS, and Tailwind CSS. It features Google and email/password login, question-answer CRUD, user profiles, and Stack Overflow API integration. Includes markdown support and a persistent dark/light theme toggle. [GitHub](https://github.com/sksmagr23/CodeExchange)  [Deployment](https://codeexchange-3s2g.onrender.com)
  
- AI-powered renewable energy prediction system built with Next.js, using GRU and XGBoost models via Groq SDK and Mapbox API for location-based forecasts. Features secure access with Next-auth and a responsive UI with MapboxGL, Recharts, and Framer Motion. [GitHub](https://github.com/sksmagr23/shannon_frontend) [Deployment](https://shannonntpc.vercel.app/)  

- A Golang-based weather application that fetches real-time weather data for any city using the WeatherAPI. It displays current weather details to the user and optionally stores the data in a MongoDB database for future reference. [GitHub](https://github.com/sksmagr23/go-weather-api) [Deployment](https://weather-goapi.netlify.app/)  


### Notable Achievements Open-source Contributions
- Secured 2nd position in the Hack It Out hackathon at Technex’25 IIT BHU, Varanasi
- Paricipated in College summer of code last year and a main contributor in the mern project.
- Have done many open source contributions for college clubs and fests , which you can find on my github and resume given below

---

## Resume / Portfolio

- **Resume:** [https://portfolio-sksm.vercel.app/resume.pdf](https://portfolio-sksm.vercel.app/resume.pdf)
- **Portfolio:** [https://portfolio-sksm.vercel.app/](https://portfolio-sksm.vercel.app/)
- **GitHub Profile:** [sksmagr23](https://github.com/sksmagr23)

---

## Time Commitment

- Weekly Availability: 22-25 hours per week
- Estimated Total Hours: 90-100 hours over the program duration
- Availability Period: Full-time during summer break, flexible schedule during academic year

---

## Implementation Plan

### Planned Features/Deliverables

1. **URL Shortening Engine**
   - **Custom Short Links**: Generate short, memorable URLs like `short.domain/user123` instead of long URLs
   - **Configurable Length**: Users can choose how short they want their links (5-10 characters)
   - **Custom Branded Domains**: Use your own domain like `sksm.tech/user123` instead of generic shorteners
   - **Link Validation**: Automatically check if URLs are valid and safe before shortening
   - **Expiration Dates**: Set links to expire after a certain time (useful for temporary promotions)

2. **URL Analytics**
   - **Click Tracking**: Count how many times each link is clicked
   - **Real-time Statistics**: See live updates of link performance
   - **Geographic Location**: Know which countries your visitors are from
   - **Browser & Device Detection**: See what browsers (Chrome, Firefox) and devices (mobile, desktop) visitors use
   - **Referrer Tracking**: Know which websites are sending traffic to your links
   - **Time-based Analytics**: View click patterns by hour, day, or month to understand peak usage times

3. **User Management**
   - **Public/Private Links**: Make links visible to everyone or just yourself
   - **User Authentication**: Secure login system to protect user accounts
   - **Link Ownership**: Each user can only see and manage their own links

4. **REST API**
   - **Complete CRUD Operations**: Create, read, update, and delete links through API calls
   - **Analytics Data Endpoints**: Get click statistics and user behavior data via API
   - **User Management Endpoints**: Register, login, and manage user accounts through API
   - **Rate Limiting**: Prevent abuse by limiting how many API calls users can make
   - **API Key Management**: Secure access to the API using authentication keys

5. **Testing and linting**
   - **Unit Tests**: Test individual components (handlers, services, database) in isolation
   - **Performance Testing**: Ensure the system can handle high traffic without crashing
   - **Linting**: Use tools like golangci-lint to enforce code quality and consistency

### Development Phases

**Week 1: Foundation & Core Functionality**
- Project setup with GoFr, establish three-layer architecture.
- Design and connect MongoDB database.
- Implement core URL shortening engine (generation, redirection).
- Develop basic REST API for URL CRUD operations.

**Week 2: User Management & Analytics**
- Implement user authentication (registration, login, JWT).
- Develop user profile management and API key generation endpoints.
- Build the analytics tracking system for clicks, geographic location, and device info.
- Implement link ownership and public/private link settings.

**Week 3: Advanced Features & Security**
- Add support for custom short links, configurable length, and expiration dates.
- Implement security features like rate limiting and URL validation.
- Enhance analytics with referrer tracking and time-based data.
- Begin work on custom branded domain support.

**Week 4: Testing, Documentation & Polish**
- Write comprehensive unit and integration tests.
- Use `golangci-lint` for code quality checks and refactoring.
- Finalize API documentation and project README.
- Bug fixing, performance optimization, and final project submission preparation.

### Timeline with Key Milestones

**1:** Project setup, basic GoFr application structure, MongoDB connection
**2:** Core URL shortening engine, basic REST API endpoints
**3:** User authentication system, JWT implementation
**4:** Analytics tracking, user profile management
**5:** Advanced features (custom links, expiration, rate limiting)
**6:** Custom domain support, enhanced analytics
**7:** Comprehensive testing suite, performance optimization
**8:** Documentation, linting, final submission

---

## Complete API Endpoints Specification

**URL Management Endpoints:**

**1. Create Short URL**
```
POST /api/urls
Authorization: Bearer <jwt_token>

Request Body:
{
    "original_url": "https://test.com/long-url",
    "custom_code": "my-code",
    "public": true,
    "custom_domain": "custom-domain.com",
    "expires_at": "2024-12-31T23:59:59Z"
}

Response (200):
{
    "data": {
        "id": "507f1f77bcf86cd799439011",
        "short_code": "user123",
        "short_url": "https://short.domain/user123",
        "original_url": "https://test.com/long-url",
        "public": true,
        "custom_domain": "custom-domain.com",
        "expires_at": "2024-12-31T23:59:59Z",
        "created_at": "2024-01-01T00:00:00Z",
        "total_clicks": 0,
        "unique_clicks": 0
    }
}
```

**2. Get URL Details**
```
GET /api/urls/{short_code}
Authorization: Bearer <jwt_token>

Response (200):
{
    "data": {
        "id": "507f1f77bcf86cd799439011",
        "short_code": "user123",
        "short_url": "https://short.domain/user123",
        "original_url": "https://test.com/long-url",
        "public": true,
        "custom_domain": "custom-domain.com",
        "expires_at": "2024-12-31T23:59:59Z",
        "created_at": "2024-01-01T00:00:00Z",
        "total_clicks": 150,
        "unique_clicks": 120
    }
}
```

**3. Update URL**
```
PUT /api/urls/{short_code}
Authorization: Bearer <jwt_token>

Request Body:
{
    "public": false,
    "expires_at": "2024-12-31T23:59:59Z"
}

Response (200):
{
    "data": {
        "id": "507f1f77bcf86cd799439011",
        "message": "URL updated successfully"
    }
}
```

**4. Delete URL**
```
DELETE /api/urls/{short_code}
Authorization: Bearer <jwt_token>

Response (200):
{
    "data": {
        "message": "URL deleted successfully"
    }
}
```

**5. List User URLs**
```
GET /api/urls?page=1&limit=10&sort=created_at&order=desc
Authorization: Bearer <jwt_token>

Response (200):
{
    "data": [
        {
            "id": "507f1f77bcf86cd799439011",
            "short_code": "abc123",
            "short_url": "https://short.domain/abc123",
            "original_url": "https://example.com/very-long-url",
            "total_clicks": 150,
            "unique_clicks": 120,
            "created_at": "2024-01-01T00:00:00Z"
        },
        ...more
    ],
    "pagination": {
        "page": 1,
        "limit": 10,
        "total": 20,
        "total_pages": 2
    }
}
```

**6. Redirect to Original URL**
```
GET /{short_code}

Response (301):
Location: https://example.com/very-long-url
```

**Analytics Endpoints:**

**1. Get URL Analytics**
```
GET /api/urls/{short_code}/analytics?period=7d&group_by=day
Authorization: Bearer <jwt_token>

Response (200):
{
    "data": {
        "total_clicks": 150,
        "unique_clicks": 120,
        "top_countries": [
            {"country": "US", "clicks": 50, "percentage": 33.33},
            {"country": "IN", "clicks": 30, "percentage": 20.00}
        ],
        "top_browsers": [
            {"browser": "Chrome", "clicks": 80, "percentage": 53.33},
            {"browser": "Safari", "clicks": 40, "percentage": 26.67}
        ],
        "top_devices": [
            {"device": "desktop", "clicks": 90, "percentage": 60.00},
            {"device": "mobile", "clicks": 60, "percentage": 40.00}
        ],
        "daily_stats": [
            {"date": "2024-01-01", "clicks": 20, "unique_clicks": 18},
            {"date": "2024-01-02", "clicks": 25, "unique_clicks": 22}
        ],
        "hourly_stats": [
            {"hour": 10, "clicks": 5},
            {"hour": 11, "clicks": 8}
        ]
    }
}
```

**2. Get Real-time Analytics**
```
GET /api/urls/{short_code}/analytics/realtime
Authorization: Bearer <jwt_token>

Response (200):
{
    "data": {
        "clicks_today": 15,
        "clicks_this_hour": 3,
        "clicks_this_minute": 1,
        "recent_clicks": [
            {
                "ip_address": "192.168.1.1",
                "country": "US",
                "browser": "Chrome",
                "device": "desktop",
                "referrer": "https://google.com",
                "clicked_at": "2024-01-01T12:30:00Z"
            }
        ]
    }
}
```

**3. Get Analytics Summary**
```
GET /api/urls/{short_code}/analytics/summary
Authorization: Bearer <jwt_token>

Response (200):
{
    "data": {
        "total_clicks": 150,
        "unique_clicks": 120,
        "click_through_rate": 85.5,
        "average_clicks_per_day": 21.4,
        "peak_hour": 14,
        "peak_day": "Monday",
        "top_referrer": "https://google.com",
        "top_country": "US",
        "top_browser": "Chrome"
    }
}
```

**User Management Endpoints:**

**1. User Registration**
```
POST /api/users/register

Request Body:
{
    "username": "user123",
    "email": "user123@gmail.com",
    "password": "1X1X1X"
}

Response (201):
{
    "data": {
        "message": "User registered successfully",
        "user": {
            "id": "507f1f77bcf86cd799439012",
            "username": "user123",
            "email": "user123@gmail.com",
            "created_at": "2024-01-01T00:00:00Z"
        }
    }
}
```

**2. User Login**
```
POST /api/users/login

Request Body:
{
    "email": "user123@gmail.com",
    "password": "1X1X1X"
}

Response (200):
{
    "data": {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
        "user": {
            "id": "507f1f77bcf86cd799439012",
            "username": "user123",
            "email": "user123@gmail.com",
            "api_key": "api_key_here",
            "created_at": "2024-01-01T00:00:00Z"
        }
    }
}
```

**3. Get User Profile**
```
GET /api/users/profile
Authorization: Bearer <jwt_token>

Response (200):
{
    "data": {
        "id": "507f1f77bcf86cd799439012",
        "username": "user123",
        "email": "user123@gmail.com",
        "api_key": "api_key_here",
        "total_urls": 25,
        "total_clicks": 1500,
        "created_at": "2024-01-01T00:00:00Z"
    }
}
```

**4. Update User Profile**
```
PUT /api/users/profile
Authorization: Bearer <jwt_token>

Request Body:
{
    "username": "user123",
    "email": "user123@gmail.com"
}

Response (200):
{
    "data": {
        "message": "Profile updated successfully"
    }
}
```

**5. Generate New API Key**
```
POST /api/users/api-key
Authorization: Bearer <jwt_token>

Response (200):
{
    "data": {
        "api_key": "new_api_key"
    }
}
```


**Health and Monitoring Endpoints:**

**1. Health Check**
```
GET /health

Response (200):
{
    "status": "healthy",
    "timestamp": "2024-01-01T00:00:00Z",
    "services": {
        "mongoDB": "connected",
        "redis": "connected"
    }
}
```

**2. API Status**
```
GET /api/status

Response (200):
{
    "data": {
        "version": "1.0.0",
        "uptime": "24h 30m 15s",
        "total_urls": 15000,
        "total_clicks": 500000,
        "active_users": 250
    }
}
```

---

## Architecture Diagram

[Download as PDF](architecture-diagram.pdf)


**Technical Components Breakdown:**

1.  **URL Shortening Engine (Service Layer):** Handles core logic for generating, validating, and redirecting short URLs, supporting custom aliases, configurable length, and expiration dates.

2.  **URL Analytics Engine (Service Layer):** Asynchronously captures and processes click metadata (geo-location, device, referrer) for real-time and aggregated analytics using Redis and MongoDB.

3.  **User Management (Service & Handler Layers):** Manages user registration, login, JWT-based authentication, link ownership, and API key generation.

4.  **REST API (Handler Layer):** Exposes RESTful CRUD endpoints for URLs and analytics, protected by middleware for authentication, rate limiting, and logging.

5.  **Data Persistence (Store Layer):** Utilizes MongoDB for primary storage (users, URLs, analytics) and Redis for high-speed caching and temporary event processing.

---

## Fallback Plan

**Strategy for Missed Milestones:**
- Prioritize core functionality over advanced features
- Timely progress tracking and early communication with mentors
- Break down complex features into smaller, manageable tasks

**Approach to Resolving Blockers:**
- Regular mentor check-ins for technical guidance
- Research and documentation review for unfamiliar concepts
- Utilize GoFr's built-in features and best practices

**Risk Mitigation:**
- Focus on core URL shortening functionality first
- Implement comprehensive testing early in the development cycle
- Maintain regular backups and version control

---

## Mentorship Expectations

**Frequency of Guidance:**
- Weekly mentor meetings for progress updates
- Timely code review sessions
- On-demand support for technical blockers

**Areas Where Mentor Help is Expected:**

1. **Design and Architecture:**
   - GoFr framework best practices
   - Three-layer architecture optimization
   - API pattern design

2. **Code Review:**
   - Code quality and optimization suggestions
   - Security best practices

3. **Technical Guidance:**
   - GoFr-specific features and capabilities
   - Deployment and production considerations

**Communication Preferences:**
- Slack/Discord for quick questions and updates
- GitHub issues/discussions for bug tracking and feature requests

---

## My statement

This URL Shortener with Insights project represents an excellent opportunity to contribute to the GoFr ecosystem while building a practical, real-world application. The three-layer architecture will demonstrate GoFr's capabilities for building scalable backend services, while the analytics features will showcase the framework's data processing abilities.

I am excited about the opportunity to work with the GoFr community and contribute to building a production-ready template that other developers can use as a reference for their own projects. The combination of practical utility, technical complexity, and learning opportunities makes this project an ideal fit for my skills and interests.

I am committed to delivering a high-quality, well-tested, and thoroughly documented solution that will serve as a valuable addition to the GoFr template ecosystem. 