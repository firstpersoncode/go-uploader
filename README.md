# Go Uploader

A Go-based REST API service for managing financial transaction uploads, user authentication, and transaction analysis. This service allows users to upload CSV bank statements, calculate balances, and track transaction issues.

## Table of Contents

- [Features](#features)
- [Setup Instructions](#setup-instructions)
- [Architecture Decisions](#architecture-decisions)
- [API Documentation](#api-documentation)

## Features

- **User Authentication**: Secure signup/signin with JWT-based session management
- **CSV Upload**: Parse and store bank statement transactions from CSV files
- **Balance Calculation**: Calculate total credits, debits, and current balance
- **Issue Tracking**: Query and filter failed/pending transactions with pagination and sorting
- **Rate Limiting**: Built-in request rate limiting (20 requests per 30 seconds)
- **Session Management**: Cookie-based authentication with HTTP-only secure cookies

## Setup Instructions

### Prerequisites

- Go 1.24.0 or higher
- Git

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd go-uploader
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Configure environment variables**
   
   Create a `.env` file in the project root:
   ```bash
   cp .env.example .env
   ```
   
   Edit `.env` with your configuration:
   ```env
    SESSION_COOKIE_NAME=__session__
    ALLOWED_ORIGINS=http://localhost:3000
    HOST=127.0.0.1
    PORT=8080
   ```

4. **Run the application**
   ```bash
   go run main.go
   ```
   
   The server will start at `http://127.0.0.1:8080`

5. **Run tests**
   ```bash
   go test ./internal/modules... -v
   ```

### Development Setup

For development with auto-reload, you can use `air`:
```bash
go install github.com/air-verse/air@latest
air
```

## Architecture Decisions

### 1. Project Structure (Clean Architecture)

The project follows **Clean Architecture** principles with clear separation of concerns:

```
go-uploader/
├── domain/              # Business entities and interfaces
│   ├── user.go
│   ├── session.go
│   └── transaction.go
├── dto/                 # Data Transfer Objects
│   ├── response.go
│   ├── session/
│   └── transaction/
├── internal/            # Internal application logic
│   ├── config/          # Configuration management
│   ├── middlewares/     # HTTP middlewares
│   ├── modules/         # Feature modules
│   │   ├── auth/        # Authentication module
│   │   └── transaction/ # Transaction module
│   ├── repositories/    # Data persistence layer
│   └── util/            # Utility functions
└── main.go              # Application entry point
```

**Benefits:**
- **Testability**: Each layer can be tested independently
- **Maintainability**: Clear boundaries between business logic and infrastructure
- **Flexibility**: Easy to swap implementations (e.g., switch from in-memory to database)
- **Scalability**: New features can be added as modules without affecting existing code

---

### 2. Domain-Driven Design (DDD)

**Domain Layer** (`domain/`) contains:
- Core business entities (User, Session, Transaction)
- Repository interfaces
- Service interfaces
- Handler interfaces

**Rationale:**
- Business logic is independent of frameworks and external dependencies
- Interfaces in the domain layer allow for dependency inversion
- Easy to mock dependencies for testing

---

### 3. Dependency Injection

The application uses **constructor-based dependency injection**:

```go
userRepo := repositories.NewUserRepository()
sessionRepo := repositories.NewSessionRepository()
authService := auth.NewAuthService(userRepo, sessionRepo)
authHandler := auth.NewAuthHandler(authService)
```

**Benefits:**
- Loose coupling between components
- Easy to test with mock implementations
- Clear dependency graph

---

### 4. In-Memory Storage

Current implementation uses **in-memory repositories** (maps with mutex locks):

```go
// Example from repositories
type userRepository struct {
    users map[string]*domain.User
    mu    sync.RWMutex
}
```

**Rationale:**
- Simple for development and testing
- No external database dependencies
- Fast performance for prototyping

**Trade-offs:**
- Data is lost on restart
- Not suitable for production
- No persistence across instances

**Future Enhancement:** Can easily swap to database implementation (PostgreSQL, MongoDB) by implementing the same repository interfaces.

---

### 5. HTTP Framework: Fiber

Using **GoFiber** as the web framework:

**Reasons:**
- Express.js-like API (familiar for web developers)
- High performance (built on fasthttp)
- Built-in middleware support (rate limiting, CORS, etc.)
- Simple request parsing and validation
- Efficient memory usage

---

### 6. Authentication Strategy

**JWT + HTTP-Only Cookies**:

```go
cookie := &fiber.Cookie{
    Name:     "__session__",
    Value:    token,
    HTTPOnly: true,
    Secure:   false,  // Set to true in production with HTTPS
    SameSite: "Lax",
}
```

**Design Decisions:**
- **JWT**: Stateless authentication, contains session ID
- **HTTP-Only Cookies**: Prevents XSS attacks (JavaScript cannot access)
- **SameSite=Lax**: CSRF protection
- **24-hour expiry**: Balances security and user experience

**Security Note:** The JWT secret is currently hardcoded. In production, this should be:
- Stored in environment variables
- Rotated regularly
- Generated with cryptographically secure random bytes

---

### 7. Password Security

Using **bcrypt** for password hashing:

```go
hashedPassword, err := bcrypt.GenerateFromPassword(
    []byte(credentials.Password), 
    bcrypt.DefaultCost
)
```

**Benefits:**
- Industry-standard hashing algorithm
- Built-in salt generation
- Adaptive cost factor (future-proof against hardware improvements)
- Slow by design (prevents brute-force attacks)

---

### 8. Transaction Processing

**CSV Parsing Strategy:**
- Stream-based processing using `encoding/csv`
- Batch insert for efficiency
- Trim whitespace for data consistency
- Unix timestamp to time.Time conversion

**Status-Based Balance Calculation:**
- Only `SUCCESS` transactions affect balance
- Separate tracking of credits and debits
- Clear audit trail

---

### 9. Error Handling

**Consistent error response format:**

```go
type Response struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

**Benefits:**
- Predictable API responses
- Easy to parse on client-side
- Clear success/failure indication

---

### 10. Pagination & Sorting

**Flexible query system** for the issues endpoint:

```go
type PaginationDTO struct {
    Page  int `query:"page"`
    Limit int `query:"limit"`
}

type SortingDTO struct {
    Sort   SortDirection `query:"sort"`
    SortBy string        `query:"sortBy"`
}
```

**Benefits:**
- Prevents large payload responses
- Customizable result ordering
- Better user experience for large datasets

---

### 11. Rate Limiting

**Sliding window rate limiter:**
- 20 requests per 30 seconds per IP
- Prevents abuse and DoS attacks
- Configurable limits

**Implementation:**
```go
app.Use(limiter.New(limiter.Config{
    Max:               20,
    Expiration:        30 * time.Second,
    LimiterMiddleware: limiter.SlidingWindow{},
}))
```

---

### 12. Middleware Pattern

**Session middleware** for authentication:

```go
app.Get("/session", sessionMiddleware.Handle, authHandler.Session)
```

**Advantages:**
- Reusable authentication logic
- Clean separation from business logic
- Easy to add additional middleware (logging, CORS, etc.)

---

### 13. Type Safety

**Strong typing** with custom types:

```go
type TransactionType string
const (
    TransactionTypeDebit  TransactionType = "DEBIT"
    TransactionTypeCredit TransactionType = "CREDIT"
)
```

**Benefits:**
- Compile-time type checking
- Self-documenting code
- Prevents invalid values

---

### 14. Testing Strategy

Separate test files for services:
- `service_test.go` files alongside implementation
- Tests focus on business logic
- Repository interfaces allow easy mocking

**Example test structure:**
```go
func TestAuthService_RegisterUser(t *testing.T) {
    // Arrange: Create mock repository
    // Act: Call service method
    // Assert: Verify results
}
```

---

### 15. Configuration Management

**Environment-based configuration:**
- `.env` files for local development
- `godotenv` for loading configuration
- Structured config with dedicated loader

**Benefits:**
- 12-factor app compliance
- Easy deployment across environments
- No hardcoded configuration

---

## API Documentation

### Base URL
```
http://127.0.0.1:8080
```

### Response Format

All successful responses follow this structure:
```json
{
  "status": "ok",
  "message": "Operation description",
  "data": {}
}
```

All error responses follow this structure:
```json
{
  "status": "error",
  "message": "Error description",
  "data": null
}
```

---

### Health Check

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "ok"
}
```

---

### Authentication

#### 1. Sign Up

**Endpoint:** `POST /signup`

**Request Body:**
```json
{
  "username": "john_doe",
  "password": "securePassword123"
}
```

**Success Response:**
```json
{
  "status": "ok",
  "message": "User registered successfully",
  "data": {}
}
```

**Error Response:**
```json
{
  "status": "error",
  "message": "username and password are required",
  "data": null
}
```

---

#### 2. Sign In

**Endpoint:** `POST /signin`

**Request Body:**
```json
{
  "username": "john_doe",
  "password": "securePassword123"
}
```

**Success Response:**
```json
{
  "status": "ok",
  "message": "Signed in successfully",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expiry": "2025-11-17T10:30:00Z"
  }
}
```

**Note:** The session token is automatically set as an HTTP-only cookie.

**Error Response:**
```json
{
  "status": "error",
  "message": "invalid credentials",
  "data": null
}
```

---

#### 3. Sign Out

**Endpoint:** `POST /signout`

**Response:**
```json
{
  "message": "Signed out successfully"
}
```

---

#### 4. Get Session

**Endpoint:** `GET /session`

**Headers:**
- Cookie: `<cookie-name>=<token>`

**Success Response:**
```json
{
  "status": "ok",
  "message": "Session retrieved successfully",
  "data": {
    "id": "session-uuid",
    "user_id": "user-uuid",
    "user": {
      "id": "user-uuid",
      "username": "john_doe"
    },
    "refresh_token": ""
  }
}
```

**Error Response:**
```json
{
  "status": "error",
  "message": "Unauthorized: No session token",
  "data": null
}
```

---

### Transaction Management

All transaction endpoints require authentication (valid session cookie).

#### 1. Upload Bank Statement

**Endpoint:** `POST /upload`

**Headers:**
- Cookie: `<cookie-name>=<token>`
- Content-Type: `multipart/form-data`

**Request Body:**
- `file`: CSV file containing transaction data

**CSV Format:**
```csv
timestamp,name,type,amount,status,description
1609459200,Grocery Store,DEBIT,5000,SUCCESS,Weekly groceries
1609545600,Salary Deposit,CREDIT,50000,SUCCESS,Monthly salary
1609632000,Failed Payment,DEBIT,2000,FAILED,Insufficient funds
```

**Success Response:**
```json
{
  "status": "ok",
  "message": "Statement uploaded successfully",
  "data": {
    "total_rows": 3,
    "upload_status": "success"
  }
}
```

**Error Response:**
```json
{
  "status": "error",
  "message": "Only CSV files are allowed",
  "data": null
}
```

---

#### 2. Get Balance

**Endpoint:** `GET /balance`

**Headers:**
- Cookie: `<cookie-name>=<token>`

**Success Response:**
```json
{
  "status": "ok",
  "message": "Balance calculated successfully",
  "data": {
    "credits": 50000,
    "debits": 5000,
    "balance": 45000
  }
}
```

**Note:** Balance calculation only includes transactions with `SUCCESS` status. Amounts are in the smallest currency unit (e.g., cents).

---

#### 3. Get Issues (Failed/Pending Transactions)

**Endpoint:** `GET /issues`

**Headers:**
- Cookie: `<cookie-name>=<token>`

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 10)
- `sort` (optional): Sort direction (`ASC` or `DESC`)
- `sortBy` (optional): Field to sort by (e.g., `timestamp`, `amount`)

**Example Request:**
```
GET /issues?page=1&limit=10&sort=DESC&sortBy=timestamp
```

**Success Response:**
```json
{
  "status": "ok",
  "message": "Issues retrieved successfully",
  "data": {
    "transactions": [
      {
        "timestamp": "2021-01-03T00:00:00Z",
        "name": "Failed Payment",
        "type": "DEBIT",
        "amount": 2000,
        "status": "FAILED",
        "description": "Insufficient funds"
      }
    ],
    "total": 1
  }
}
```

---

### HTTP Status Codes

- `200` - Success
- `400` - Bad Request (invalid input, wrong file type, etc.)
- `401` - Unauthorized (missing or invalid session token)
- `429` - Too Many Requests (rate limit exceeded)
- `500` - Internal Server Error

---

### Rate Limiting

The API implements rate limiting:
- **Limit**: 20 requests per 30 seconds per IP address
- **Algorithm**: Sliding window

When rate limit is exceeded, you'll receive a `429 Too Many Requests` response.

---