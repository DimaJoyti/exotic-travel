# API Documentation

This document provides comprehensive documentation for the Exotic Travel Booking Platform REST API.

## Base URL

- **Development**: `http://localhost:8080`
- **Production**: `https://api.yourdomain.com`

## Authentication

The API uses JWT (JSON Web Tokens) for authentication. Include the token in the Authorization header:

```http
Authorization: Bearer <your-jwt-token>
```

### Token Lifecycle
- **Access Token**: 15 minutes expiration
- **Refresh Token**: 7 days expiration
- **Automatic Key Rotation**: 24 hours

## Response Format

All API responses follow a consistent format:

```json
{
  "success": true,
  "data": {},
  "message": "Success message",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

### Error Response Format
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": {}
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## Authentication Endpoints

### Register User
```http
POST /api/auth/register
```

**Request Body:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "SecurePassword123!",
  "phone": "+1234567890"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "role": "user",
      "created_at": "2024-01-01T00:00:00Z"
    },
    "tokens": {
      "access_token": "eyJhbGciOiJSUzI1NiIs...",
      "refresh_token": "eyJhbGciOiJSUzI1NiIs...",
      "expires_at": "2024-01-01T00:15:00Z",
      "token_type": "Bearer"
    }
  }
}
```

### Login User
```http
POST /api/auth/login
```

**Request Body:**
```json
{
  "email": "john@example.com",
  "password": "SecurePassword123!"
}
```

**Response:** Same as register response

### Refresh Token
```http
POST /api/auth/refresh
```

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJSUzI1NiIs..."
}
```

**Response:** New token pair

### Logout
```http
POST /api/auth/logout
```

**Headers:** `Authorization: Bearer <token>`

**Response:**
```json
{
  "success": true,
  "message": "Successfully logged out"
}
```

## User Endpoints

### Get Current User
```http
GET /api/users/me
```

**Headers:** `Authorization: Bearer <token>`

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "+1234567890",
    "role": "user",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### Update User Profile
```http
PUT /api/users/me
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**
```json
{
  "name": "John Smith",
  "phone": "+1234567891"
}
```

### Change Password
```http
PUT /api/users/me/password
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**
```json
{
  "current_password": "OldPassword123!",
  "new_password": "NewPassword123!"
}
```

## Destination Endpoints

### Get All Destinations
```http
GET /api/destinations
```

**Query Parameters:**
- `page` (int): Page number (default: 1)
- `limit` (int): Items per page (default: 10, max: 100)
- `search` (string): Search term for name/description
- `country` (string): Filter by country
- `min_price` (float): Minimum price filter
- `max_price` (float): Maximum price filter
- `sort` (string): Sort field (name, price, created_at)
- `order` (string): Sort order (asc, desc)

**Response:**
```json
{
  "success": true,
  "data": {
    "destinations": [
      {
        "id": 1,
        "name": "Bali Paradise Resort",
        "description": "Luxury resort in tropical Bali",
        "country": "Indonesia",
        "city": "Ubud",
        "price": 299.99,
        "duration": 7,
        "max_guests": 4,
        "images": [
          "https://example.com/image1.jpg",
          "https://example.com/image2.jpg"
        ],
        "features": ["Pool", "Spa", "Beach Access"],
        "rating": 4.8,
        "review_count": 156,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 50,
      "total_pages": 5,
      "has_next": true,
      "has_prev": false
    }
  }
}
```

### Get Destination by ID
```http
GET /api/destinations/{id}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "Bali Paradise Resort",
    "description": "Luxury resort in tropical Bali with stunning views...",
    "country": "Indonesia",
    "city": "Ubud",
    "price": 299.99,
    "duration": 7,
    "max_guests": 4,
    "images": ["https://example.com/image1.jpg"],
    "features": ["Pool", "Spa", "Beach Access"],
    "rating": 4.8,
    "review_count": 156,
    "availability": {
      "available_dates": ["2024-02-01", "2024-02-08"],
      "unavailable_dates": ["2024-01-15", "2024-01-22"]
    },
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

## Booking Endpoints

### Create Booking
```http
POST /api/bookings
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**
```json
{
  "destination_id": 1,
  "start_date": "2024-02-01",
  "end_date": "2024-02-08",
  "guests": 2,
  "special_requests": "Vegetarian meals please"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 1,
    "destination_id": 1,
    "start_date": "2024-02-01",
    "end_date": "2024-02-08",
    "guests": 2,
    "total_price": 2099.93,
    "status": "pending",
    "special_requests": "Vegetarian meals please",
    "payment_intent_id": "pi_1234567890",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

### Get User Bookings
```http
GET /api/bookings
```

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**
- `page` (int): Page number
- `limit` (int): Items per page
- `status` (string): Filter by status (pending, confirmed, cancelled, completed)

### Get Booking by ID
```http
GET /api/bookings/{id}
```

**Headers:** `Authorization: Bearer <token>`

### Cancel Booking
```http
DELETE /api/bookings/{id}
```

**Headers:** `Authorization: Bearer <token>`

## Review Endpoints

### Create Review
```http
POST /api/destinations/{id}/reviews
```

**Headers:** `Authorization: Bearer <token>`

**Request Body:**
```json
{
  "rating": 5,
  "comment": "Amazing experience! Highly recommended.",
  "booking_id": 1
}
```

### Get Destination Reviews
```http
GET /api/destinations/{id}/reviews
```

**Query Parameters:**
- `page` (int): Page number
- `limit` (int): Items per page
- `sort` (string): Sort by (rating, created_at)
- `order` (string): Sort order (asc, desc)

## Admin Endpoints

### Get All Users (Admin Only)
```http
GET /api/admin/users
```

**Headers:** `Authorization: Bearer <admin-token>`

### Create Destination (Admin Only)
```http
POST /api/admin/destinations
```

**Headers:** `Authorization: Bearer <admin-token>`

**Request Body:**
```json
{
  "name": "New Destination",
  "description": "Amazing place to visit",
  "country": "Thailand",
  "city": "Bangkok",
  "price": 199.99,
  "duration": 5,
  "max_guests": 6,
  "images": ["https://example.com/image.jpg"],
  "features": ["Pool", "Restaurant"]
}
```

### Update Destination (Admin Only)
```http
PUT /api/admin/destinations/{id}
```

### Delete Destination (Admin Only)
```http
DELETE /api/admin/destinations/{id}
```

## Health & Monitoring Endpoints

### Health Check
```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "version": "1.0.0",
  "services": {
    "database": "healthy",
    "redis": "healthy"
  }
}
```

### Metrics (Prometheus Format)
```http
GET /metrics
```

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Request validation failed |
| `UNAUTHORIZED` | 401 | Authentication required |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Resource conflict |
| `RATE_LIMITED` | 429 | Too many requests |
| `INTERNAL_ERROR` | 500 | Internal server error |

## Rate Limiting

- **Default**: 10 requests per second per IP
- **Burst**: 20 requests
- **Headers**: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`

## Security Headers

All responses include security headers:
- `X-Frame-Options: DENY`
- `X-Content-Type-Options: nosniff`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security: max-age=31536000`
- `Content-Security-Policy: default-src 'self'`
