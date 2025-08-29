# ConnectSphere - User Management MVP

## Overview
ConnectSphere is a real-time messaging and video call application. This repository contains the foundational User Management MVP implementation with a Go backend and Flutter frontend.

## 🚀 Features

- **User Authentication**: Secure registration and login with JWT tokens
- **Profile Management**: Update user profiles and display information  
- **Social Connections**: Send and manage friend/connection requests
- **User Search**: Search and discover other users
- **Real-time Updates**: Modern, responsive UI with state management

## Technology Stack

### Backend
- **Language**: Go (Golang)
- **Framework**: Gin
- **Database**: PostgreSQL with pgx driver
- **Authentication**: JWT (JSON Web Tokens)
- **Containerization**: Docker

### Frontend
- **Framework**: Flutter
- **State Management**: Riverpod
- **HTTP Client**: Dio
- **Secure Storage**: flutter_secure_storage
- **Navigation**: go_router

## Project Structure

### Backend (`/connectsphere-backend`)
```
connectsphere-backend/
├── cmd/server/main.go           # Application entry point
├── internal/
│   ├── api/handlers.go          # HTTP handlers and routes
│   ├── auth/auth.go             # JWT and password utilities
│   ├── config/config.go         # Configuration management
│   ├── database/database.go     # Database operations
│   └── models/models.go         # Data models and DTOs
├── docker-compose.yml           # Multi-service setup
├── Dockerfile                   # Production container
├── init.sql                     # Database schema
└── go.mod                       # Go dependencies
```

### Frontend (`/connectsphere_frontend`)
```
connectsphere_frontend/
├── lib/
│   ├── src/
│   │   ├── core/api_client.dart           # HTTP client with auth
│   │   ├── features/
│   │   │   ├── authentication/            # Login/Register screens
│   │   │   ├── connections/               # Friends and search features
│   │   │   └── profile/                   # User profile management
│   │   ├── models/user.dart               # Data models
│   │   └── router.dart                    # App navigation
│   └── main.dart                          # App entry point
└── pubspec.yaml                           # Flutter dependencies
```

## Features Implemented

### Authentication
- User registration with validation
- User login with JWT tokens
- Secure token storage
- Automatic logout on token expiry

### User Management
- View and edit user profile
- Username and email uniqueness
- Secure password hashing (bcrypt)

### Connection System
- Send friend requests
- Accept/decline requests
- Remove existing friendships
- View friends list and pending requests

### User Discovery
- Search users by username or display name
- Real-time search results
- Send connection requests to found users

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login

### User Management (Protected)
- `GET /api/v1/users/me` - Get current user profile
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/me` - Update profile
- `GET /api/v1/users/search?q=<query>` - Search users

### Connections (Protected)
- `POST /api/v1/connections/send-request/:addressee_id` - Send friend request
- `POST /api/v1/connections/accept-request/:requester_id` - Accept request
- `POST /api/v1/connections/decline-request/:requester_id` - Decline request
- `DELETE /api/v1/connections/remove-friend/:friend_id` - Remove friendship
- `GET /api/v1/connections` - Get friends list
- `GET /api/v1/connections/pending` - Get pending requests

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.21+ (for local development)
- Flutter 3.10+ (for local development)

### Running with Docker

1. **Start the backend services:**
   ```bash
   cd connectsphere-backend
   docker-compose up -d
   ```

2. **The backend will be available at:** `http://localhost:8080`

3. **For Flutter development:**
   ```bash
   cd connectsphere_frontend
   flutter pub get
   flutter run
   ```

### Environment Variables

Copy `.env.example` to `.env` and configure:
```env
DATABASE_URL=postgres://connectsphere:connectsphere_password@localhost:5432/connectsphere_db?sslmode=disable
JWT_SECRET=your-super-secret-jwt-key-change-in-production
PORT=8080
GIN_MODE=debug
```

## Database Schema

### Users Table
- `id` (UUID, Primary Key)
- `username` (TEXT, Unique, Not Null)
- `display_name` (TEXT, Not Null)
- `email` (TEXT, Unique, Not Null)
- `hashed_password` (TEXT, Not Null)
- `created_at`, `updated_at` (TIMESTAMPTZ)

### User Connections Table
- `id` (UUID, Primary Key)
- `requester_id` (UUID, Foreign Key)
- `addressee_id` (UUID, Foreign Key)
- `status` (TEXT: 'pending' or 'accepted')
- `created_at`, `updated_at` (TIMESTAMPTZ)

## Security Features

- JWT-based authentication
- Bcrypt password hashing
- Input validation and sanitization
- SQL injection prevention with parameterized queries
- CORS configuration
- Secure token storage in Flutter

## Next Steps

This MVP provides the foundation for:
- Real-time messaging implementation
- Video call functionality
- Push notifications
- File sharing capabilities
- Group conversations

## Development Notes

- All API responses include proper error handling
- Frontend includes loading states and error messaging
- Database includes proper indexing for performance
- Docker setup allows for easy deployment
- Code is organized following clean architecture principles

## API Testing

You can test the API using tools like Postman or curl:

```bash
# Register a new user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john.doe",
    "display_name": "John Doe",
    "email": "john@example.com",
    "password": "password123"
  }'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

## Contributing

1. Follow Go and Flutter coding standards
2. Add tests for new features
3. Update documentation
4. Ensure Docker builds work
5. Validate API contracts
