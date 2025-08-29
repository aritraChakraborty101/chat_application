package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"connectsphere-backend/internal/auth"
	"connectsphere-backend/internal/database"
	"connectsphere-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Server represents the API server
type Server struct {
	db         *database.DB
	jwtManager *auth.JWTManager
}

// NewServer creates a new API server
func NewServer(db *database.DB, jwtSecret string) *Server {
	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour) // 24 hour token expiry
	return &Server{
		db:         db,
		jwtManager: jwtManager,
	}
}

// SetupRoutes sets up all the API routes
func (s *Server) SetupRoutes() *gin.Engine {
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API v1 routes
	v1 := r.Group("/api/v1")

	// Auth routes (public)
	auth := v1.Group("/auth")
	{
		auth.POST("/register", s.register)
		auth.POST("/login", s.login)
	}

	// Protected routes
	users := v1.Group("/users")
	users.Use(s.authMiddleware())
	{
		users.GET("/me", s.getCurrentUser)
		users.PUT("/me", s.updateProfile)
		users.GET("/:id", s.getUserByID)
		users.GET("/search", s.searchUsers)
	}

	connections := v1.Group("/connections")
	connections.Use(s.authMiddleware())
	{
		connections.POST("/send-request/:addressee_id", s.sendConnectionRequest)
		connections.POST("/accept-request/:requester_id", s.acceptConnectionRequest)
		connections.POST("/decline-request/:requester_id", s.declineConnectionRequest)
		connections.DELETE("/remove-friend/:friend_id", s.removeConnection)
		connections.GET("", s.getConnections)
		connections.GET("/pending", s.getPendingRequests)
	}

	return r
}

// Auth middleware to validate JWT tokens
func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "unauthorized",
				Message: "Authorization header required",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "unauthorized",
				Message: "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		claims, err := s.jwtManager.ValidateToken(tokenParts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "unauthorized",
				Message: "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Next()
	}
}

// Auth handlers

func (s *Server) register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Check if user already exists
	if _, err := s.db.GetUserByEmail(c.Request.Context(), req.Email); err == nil {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error: "user_exists",
			Message: "User with this email already exists",
		})
		return
	}

	// Check if username is taken
	if _, err := s.db.GetUserByUsername(c.Request.Context(), req.Username); err == nil {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error: "username_taken",
			Message: "Username is already taken",
		})
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "internal_error",
			Message: "Failed to hash password",
		})
		return
	}

	// Create user
	user := &models.User{
		ID:             uuid.New(),
		Username:       req.Username,
		DisplayName:    req.DisplayName,
		Email:          req.Email,
		HashedPassword: hashedPassword,
	}

	if err := s.db.CreateUser(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "internal_error",
			Message: "Failed to create user",
		})
		return
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "internal_error",
			Message: "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusCreated, models.LoginResponse{
		Token: token,
		User:  user.ToAuth(),
	})
}

func (s *Server) login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Get user by email
	user, err := s.db.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "invalid_credentials",
			Message: "Invalid email or password",
		})
		return
	}

	// Check password
	if !auth.CheckPassword(user.HashedPassword, req.Password) {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "invalid_credentials",
			Message: "Invalid email or password",
		})
		return
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "internal_error",
			Message: "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		Token: token,
		User:  user.ToAuth(),
	})
}

// User handlers

func (s *Server) getCurrentUser(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	user, err := s.db.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "user_not_found",
			Message: "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, user.ToAuth())
}

func (s *Server) getUserByID(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid_id",
			Message: "Invalid user ID format",
		})
		return
	}

	user, err := s.db.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "user_not_found",
			Message: "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, user.ToPublic())
}

func (s *Server) updateProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid_request",
			Message: err.Error(),
		})
		return
	}

	if err := s.db.UpdateUser(c.Request.Context(), userID, req.DisplayName); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "internal_error",
			Message: "Failed to update profile",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Profile updated successfully",
	})
}

func (s *Server) searchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid_request",
			Message: "Search query parameter 'q' is required",
		})
		return
	}

	limit := 20 // Default limit
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	users, err := s.db.SearchUsers(c.Request.Context(), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "internal_error",
			Message: "Failed to search users",
		})
		return
	}

	c.JSON(http.StatusOK, users)
}

// Connection handlers

func (s *Server) sendConnectionRequest(c *gin.Context) {
	requesterID := c.MustGet("user_id").(uuid.UUID)
	
	addresseeIDParam := c.Param("addressee_id")
	addresseeID, err := uuid.Parse(addresseeIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid_id",
			Message: "Invalid addressee ID format",
		})
		return
	}

	// Can't send request to yourself
	if requesterID == addresseeID {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid_request",
			Message: "Cannot send connection request to yourself",
		})
		return
	}

	// Check if addressee exists
	if _, err := s.db.GetUserByID(c.Request.Context(), addresseeID); err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "user_not_found",
			Message: "User not found",
		})
		return
	}

	// Check if connection already exists
	if _, err := s.db.GetConnection(c.Request.Context(), requesterID, addresseeID); err == nil {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error: "connection_exists",
			Message: "Connection request already exists",
		})
		return
	}

	if err := s.db.CreateConnection(c.Request.Context(), requesterID, addresseeID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "internal_error",
			Message: "Failed to send connection request",
		})
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse{
		Message: "Connection request sent successfully",
	})
}

func (s *Server) acceptConnectionRequest(c *gin.Context) {
	addresseeID := c.MustGet("user_id").(uuid.UUID)
	
	requesterIDParam := c.Param("requester_id")
	requesterID, err := uuid.Parse(requesterIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid_id",
			Message: "Invalid requester ID format",
		})
		return
	}

	if err := s.db.AcceptConnection(c.Request.Context(), requesterID, addresseeID); err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "request_not_found",
			Message: "Pending connection request not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Connection request accepted successfully",
	})
}

func (s *Server) declineConnectionRequest(c *gin.Context) {
	addresseeID := c.MustGet("user_id").(uuid.UUID)
	
	requesterIDParam := c.Param("requester_id")
	requesterID, err := uuid.Parse(requesterIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid_id",
			Message: "Invalid requester ID format",
		})
		return
	}

	if err := s.db.DeclineConnection(c.Request.Context(), requesterID, addresseeID); err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "request_not_found",
			Message: "Pending connection request not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Connection request declined successfully",
	})
}

func (s *Server) removeConnection(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	
	friendIDParam := c.Param("friend_id")
	friendID, err := uuid.Parse(friendIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid_id",
			Message: "Invalid friend ID format",
		})
		return
	}

	if err := s.db.RemoveConnection(c.Request.Context(), userID, friendID); err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "friendship_not_found",
			Message: "Friendship not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Friendship removed successfully",
	})
}

func (s *Server) getConnections(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	connections, err := s.db.GetUserConnections(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "internal_error",
			Message: "Failed to get connections",
		})
		return
	}

	c.JSON(http.StatusOK, connections)
}

func (s *Server) getPendingRequests(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	requests, err := s.db.GetPendingConnectionRequests(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "internal_error",
			Message: "Failed to get pending requests",
		})
		return
	}

	c.JSON(http.StatusOK, requests)
}
