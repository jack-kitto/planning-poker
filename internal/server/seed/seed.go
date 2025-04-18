package seed

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"planning-poker/internal/server/models"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

const (
	numUsers                  = 50
	numOrganisations          = 5
	numSessionsPerOrg         = 3
	numStoriesPerSession      = 10
	maxMembersPerOrg          = 20
	maxParticipantsPerSession = 15
	maxVotesPerStory          = 15 // Should not exceed maxParticipantsPerSession
	maxVotingRounds           = 3
)

var estimateValues = []string{"0", "1", "2", "3", "5", "8", "13", "20", "?", "coffee"}

var modelsToDrop = []any{
	(*models.Vote)(nil),
	(*models.UserStory)(nil),
	(*models.SessionParticipant)(nil),
	(*models.Session)(nil),
	(*models.OrganisationMember)(nil),
	(*models.Organisation)(nil),
	(*models.User)(nil),
}

var modelsToCreate = []any{
	(*models.User)(nil),
	(*models.Organisation)(nil),
	(*models.OrganisationMember)(nil),
	(*models.Session)(nil),
	(*models.UserStory)(nil),
	(*models.SessionParticipant)(nil),
	(*models.Vote)(nil),
}

func Init(db *bun.DB) error {
	ctx := context.Background()

	// Optional: Add query logging
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))

	// Use a transaction for atomicity
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	log.Println("Starting database initialisation...")

	// --- Drop existing data (optional, use with caution!) ---
	err = dropModels(ctx, tx)
	if err != nil {
		return err
	}

	// --- Create Tables ---
	err = createModels(ctx, tx)
	if err != nil {
		return err
	}

	// --- Commit Transaction ---
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Println("Database initialisation completed successfully!")
	return nil
}

// Seed runs the database seeding process.
func Seed(db *bun.DB) error {
	ctx := context.Background()

	// Optional: Add query logging
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))

	// Use a transaction for atomicity
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	log.Println("Starting database seeding...")

	// --- Drop existing data (optional, use with caution!) ---
	err = dropModels(ctx, tx)
	if err != nil {
		return err
	}

	// --- Create Tables ---
	err = createModels(ctx, tx)
	if err != nil {
		return err
	}

	// --- Seed Users ---
	users, organisations, err := seedUsers(ctx, tx)
	if err != nil {
		return err
	}

	// --- Seed Organisation Members ---
	userOrgMap, err := seedOrganisationMembers(ctx, organisations, users, tx)
	if err != nil {
		return err
	}

	// --- Seed Sessions ---
	sessions, sessionParticipantMap, err := seedSessions(ctx, organisations, userOrgMap, tx)
	if err != nil {
		return err
	}

	// --- Seed User Stories ---
	userStories, err := seedUserStories(ctx, sessions, sessionParticipantMap, tx)
	if err != nil {
		return err
	}

	// --- Seed Votes ---
	err = seedVotes(ctx, userStories, sessionParticipantMap, sessions, tx)
	if err != nil {
		return err
	}

	// --- Commit Transaction ---
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Println("Database seeding completed successfully!")
	return nil
}

func createModels(ctx context.Context, tx bun.Tx) error {
	log.Println("Creating tables...")
	for _, model := range modelsToCreate {
		if _, err := tx.NewCreateTable().Model(model).IfNotExists().Exec(ctx); err != nil {
			// If Session fails due to UserStory FK, it might need adjustment or deferrable constraints
			return fmt.Errorf("failed to create table for %T: %w", model, err)
		}
	}
	// Re-apply FK constraint for sessions.current_story_id if needed after user_stories table exists
	// Or ensure FK is created as DEFERRABLE INITIALLY DEFERRED if schema managed outside Bun
	log.Println("Finished creating tables.")
	return nil
}

func dropModels(ctx context.Context, tx bun.Tx) error {
	log.Println("Dropping existing data...")
	for _, model := range modelsToDrop {
		// CASCADE might be needed depending on FK constraints if not dropping in reverse order
		if _, err := tx.NewDropTable().Model(model).IfExists().Exec(ctx); err != nil {
			return fmt.Errorf("failed to drop table for %T: %w", model, err)
		}
	}
	log.Println("Finished dropping tables.")
	return nil
}

// Helper function to connect Bun DB (can be moved to database package)
func ConnectBunDB(dsn string) (*bun.DB, *sql.DB) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true))) // Optional debug logging
	return db, sqldb
}
