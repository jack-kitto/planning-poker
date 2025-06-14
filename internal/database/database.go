package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"planning-poker/internal/server/models"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Database connection management
	Health() map[string]string
	Close() error

	// User operations
	CreateUser(name string, email string) (*models.User, error)
	UpdateUser(name string, email string) (*models.User, error)
	GetUser(email string) (*models.User, error)
	GetUserById(id string) (*models.User, error)
	GetUserWithOrg(email string) (*models.User, error)

	// Organisation operations
	CreateOrg(name string, user *models.User) (*models.Organisation, error)
	GetOrg(id string) (*models.Organisation, error)
	CreateOrgMember(org *models.Organisation, user *models.User) (*models.OrganisationMember, error)

	// Session operations
	CreateSession(org *models.Organisation, owner *models.User, name string) (*models.Session, error)
	GetSession(id string) (*models.Session, error)
	UpdateSessionName(id string, name string) error
	CreateSessionParticipant(sesh *models.Session, owner *models.User) (*models.SessionParticipant, error)
	GetSessionsForUser(userID string) ([]*models.Session, error)
	DeactivateSessionParticipant(sesh *models.Session, user *models.User) error
	ActivateSessionParticipant(sesh *models.Session, user *models.User) error

	// User story operations
	CreateStory(sesh *models.Session, title string, description *string, index string) (*models.UserStory, error)
	GetUserStory(id string) (*models.UserStory, error)
	UpdateStory(id string, title string, description string) error
}

type service struct {
	db *bun.DB
}

var (
	database   = os.Getenv("BLUEPRINT_DB_DATABASE")
	password   = os.Getenv("BLUEPRINT_DB_PASSWORD")
	username   = os.Getenv("BLUEPRINT_DB_USERNAME")
	port       = os.Getenv("BLUEPRINT_DB_PORT")
	host       = os.Getenv("BLUEPRINT_DB_HOST")
	schema     = os.Getenv("BLUEPRINT_DB_SCHEMA")
	dbInstance *service
)

func New() Service {
	if dbInstance != nil {
		return dbInstance
	}

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s",
		username, password, host, port, database, schema,
	)
	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}

	bunDB := bun.NewDB(sqlDB, pgdialect.New())

	dbInstance = &service{
		db: bunDB,
	}
	return dbInstance
}

// Get the Bun DB instance
func BunDB() *bun.DB {
	if dbInstance == nil {
		New()
	}
	return dbInstance.db
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)
	// Ping the database
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", database)
	return s.db.Close()
}
