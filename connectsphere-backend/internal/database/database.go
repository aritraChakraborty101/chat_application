package database

import (
	"context"
	"fmt"
	"log"

	"connectsphere-backend/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps the database connection pool
type DB struct {
	pool *pgxpool.Pool
}

// New creates a new database connection
func New(databaseURL string) (*DB, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to database")

	return &DB{pool: pool}, nil
}

// Close closes the database connection
func (db *DB) Close() {
	db.pool.Close()
}

// User operations

// CreateUser creates a new user in the database
func (db *DB) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, username, display_name, email, hashed_password)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at`

	err := db.pool.QueryRow(ctx, query,
		user.ID, user.Username, user.DisplayName, user.Email, user.HashedPassword,
	).Scan(&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByEmail retrieves a user by email
func (db *DB) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, display_name, email, hashed_password, created_at, updated_at
		FROM users WHERE email = $1`

	err := db.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.Email,
		&user.HashedPassword, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (db *DB) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, display_name, email, hashed_password, created_at, updated_at
		FROM users WHERE id = $1`

	err := db.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.Email,
		&user.HashedPassword, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (db *DB) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, display_name, email, hashed_password, created_at, updated_at
		FROM users WHERE username = $1`

	err := db.pool.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.Email,
		&user.HashedPassword, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return user, nil
}

// UpdateUser updates a user's profile
func (db *DB) UpdateUser(ctx context.Context, id uuid.UUID, displayName string) error {
	query := `
		UPDATE users 
		SET display_name = $1, updated_at = NOW()
		WHERE id = $2`

	result, err := db.pool.Exec(ctx, query, displayName, id)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// SearchUsers searches for users by username or display name with improved matching
func (db *DB) SearchUsers(ctx context.Context, query string, limit int) ([]models.UserPublic, error) {
	// Enhanced search query with better ranking and matching
	searchQuery := `
		SELECT id, username, display_name, created_at,
		       -- Ranking system: exact matches first, then prefix matches, then partial matches
		       CASE 
		           WHEN LOWER(username) = LOWER($1) OR LOWER(display_name) = LOWER($1) THEN 1
		           WHEN LOWER(username) LIKE LOWER($1) || '%' OR LOWER(display_name) LIKE LOWER($1) || '%' THEN 2
		           WHEN LOWER(username) LIKE '%' || LOWER($1) || '%' OR LOWER(display_name) LIKE '%' || LOWER($1) || '%' THEN 3
		           ELSE 4
		       END as rank
		FROM users 
		WHERE LOWER(username) LIKE '%' || LOWER($1) || '%' 
		   OR LOWER(display_name) LIKE '%' || LOWER($1) || '%'
		ORDER BY rank ASC, 
		         -- Secondary ordering: exact matches first, then by length (shorter names first), then alphabetically
		         CASE WHEN LOWER(username) = LOWER($1) THEN 0 ELSE 1 END,
		         CASE WHEN LOWER(display_name) = LOWER($1) THEN 0 ELSE 1 END,
		         LENGTH(username), 
		         LENGTH(display_name),
		         username
		LIMIT $2`

	rows, err := db.pool.Query(ctx, searchQuery, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	defer rows.Close()

	var users []models.UserPublic
	for rows.Next() {
		var user models.UserPublic
		var rank int // We don't need to return this, just for the query
		err := rows.Scan(&user.ID, &user.Username, &user.DisplayName, &user.CreatedAt, &rank)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// Connection operations

// CreateConnection creates a new connection request
func (db *DB) CreateConnection(ctx context.Context, requesterID, addresseeID uuid.UUID) error {
	query := `
		INSERT INTO user_connections (requester_id, addressee_id, status)
		VALUES ($1, $2, $3)`

	_, err := db.pool.Exec(ctx, query, requesterID, addresseeID, models.StatusPending)
	if err != nil {
		return fmt.Errorf("failed to create connection: %w", err)
	}

	return nil
}

// GetConnection retrieves a connection between two users
func (db *DB) GetConnection(ctx context.Context, requesterID, addresseeID uuid.UUID) (*models.UserConnection, error) {
	connection := &models.UserConnection{}
	query := `
		SELECT id, requester_id, addressee_id, status, created_at, updated_at
		FROM user_connections 
		WHERE (requester_id = $1 AND addressee_id = $2) OR (requester_id = $2 AND addressee_id = $1)`

	err := db.pool.QueryRow(ctx, query, requesterID, addresseeID).Scan(
		&connection.ID, &connection.RequesterID, &connection.AddresseeID,
		&connection.Status, &connection.CreatedAt, &connection.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("connection not found")
		}
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	return connection, nil
}

// AcceptConnection accepts a pending connection request
func (db *DB) AcceptConnection(ctx context.Context, requesterID, addresseeID uuid.UUID) error {
	query := `
		UPDATE user_connections 
		SET status = $1, updated_at = NOW()
		WHERE requester_id = $2 AND addressee_id = $3 AND status = $4`

	result, err := db.pool.Exec(ctx, query, models.StatusAccepted, requesterID, addresseeID, models.StatusPending)
	if err != nil {
		return fmt.Errorf("failed to accept connection: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("pending connection request not found")
	}

	return nil
}

// DeclineConnection declines/cancels a connection request
func (db *DB) DeclineConnection(ctx context.Context, requesterID, addresseeID uuid.UUID) error {
	query := `
		DELETE FROM user_connections 
		WHERE requester_id = $1 AND addressee_id = $2 AND status = $3`

	result, err := db.pool.Exec(ctx, query, requesterID, addresseeID, models.StatusPending)
	if err != nil {
		return fmt.Errorf("failed to decline connection: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("pending connection request not found")
	}

	return nil
}

// RemoveConnection removes an existing friendship
func (db *DB) RemoveConnection(ctx context.Context, userID, friendID uuid.UUID) error {
	query := `
		DELETE FROM user_connections 
		WHERE ((requester_id = $1 AND addressee_id = $2) OR (requester_id = $2 AND addressee_id = $1))
		AND status = $3`

	result, err := db.pool.Exec(ctx, query, userID, friendID, models.StatusAccepted)
	if err != nil {
		return fmt.Errorf("failed to remove connection: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("friendship not found")
	}

	return nil
}

// GetUserConnections retrieves all accepted connections for a user
func (db *DB) GetUserConnections(ctx context.Context, userID uuid.UUID) ([]models.ConnectionWithUser, error) {
	query := `
		SELECT uc.id, uc.requester_id, uc.addressee_id, uc.status, uc.created_at, uc.updated_at,
		       u.id, u.username, u.display_name, u.created_at
		FROM user_connections uc
		JOIN users u ON (
			CASE 
				WHEN uc.requester_id = $1 THEN u.id = uc.addressee_id 
				ELSE u.id = uc.requester_id 
			END
		)
		WHERE (uc.requester_id = $1 OR uc.addressee_id = $1) AND uc.status = $2
		ORDER BY u.display_name`

	rows, err := db.pool.Query(ctx, query, userID, models.StatusAccepted)
	if err != nil {
		return nil, fmt.Errorf("failed to get user connections: %w", err)
	}
	defer rows.Close()

	var connections []models.ConnectionWithUser
	for rows.Next() {
		var conn models.ConnectionWithUser
		err := rows.Scan(
			&conn.Connection.ID, &conn.Connection.RequesterID, &conn.Connection.AddresseeID,
			&conn.Connection.Status, &conn.Connection.CreatedAt, &conn.Connection.UpdatedAt,
			&conn.User.ID, &conn.User.Username, &conn.User.DisplayName, &conn.User.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan connection: %w", err)
		}
		connections = append(connections, conn)
	}

	return connections, nil
}

// GetPendingConnectionRequests retrieves all pending incoming connection requests for a user
func (db *DB) GetPendingConnectionRequests(ctx context.Context, userID uuid.UUID) ([]models.ConnectionWithUser, error) {
	query := `
		SELECT uc.id, uc.requester_id, uc.addressee_id, uc.status, uc.created_at, uc.updated_at,
		       u.id, u.username, u.display_name, u.created_at
		FROM user_connections uc
		JOIN users u ON u.id = uc.requester_id
		WHERE uc.addressee_id = $1 AND uc.status = $2
		ORDER BY uc.created_at DESC`

	rows, err := db.pool.Query(ctx, query, userID, models.StatusPending)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending requests: %w", err)
	}
	defer rows.Close()

	var requests []models.ConnectionWithUser
	for rows.Next() {
		var req models.ConnectionWithUser
		err := rows.Scan(
			&req.Connection.ID, &req.Connection.RequesterID, &req.Connection.AddresseeID,
			&req.Connection.Status, &req.Connection.CreatedAt, &req.Connection.UpdatedAt,
			&req.User.ID, &req.User.Username, &req.User.DisplayName, &req.User.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pending request: %w", err)
		}
		requests = append(requests, req)
	}

	return requests, nil
}
