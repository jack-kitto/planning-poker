package seed

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"planning-poker/internal/server/models"
	"slices"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/lucsky/cuid"
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
		err := tx.Rollback()
		if err != nil {
			panic(err)
		}
	}()

	log.Println("Starting database seeding...")

	// --- Drop existing data (optional, use with caution!) ---
	log.Println("Dropping existing data...")
	modelsToDrop := []any{
		(*models.Vote)(nil),
		(*models.UserStory)(nil),
		(*models.SessionParticipant)(nil),
		(*models.Session)(nil),
		(*models.OrganisationMember)(nil),
		(*models.Organisation)(nil),
		(*models.User)(nil),
	}
	for _, model := range modelsToDrop {
		// CASCADE might be needed depending on FK constraints if not dropping in reverse order
		if _, err := tx.NewDropTable().Model(model).IfExists().Exec(ctx); err != nil {
			return fmt.Errorf("failed to drop table for %T: %w", model, err)
		}
	}
	log.Println("Finished dropping tables.")

	// --- Create Tables ---
	log.Println("Creating tables...")
	modelsToCreate := []any{
		(*models.User)(nil),
		(*models.Organisation)(nil),
		(*models.OrganisationMember)(nil),
		(*models.Session)(nil),
		(*models.UserStory)(nil), // Session FK depends on UserStory for current_story_id, but FK is deferrable or checked at end of tx
		(*models.SessionParticipant)(nil),
		(*models.Vote)(nil),
	}
	for _, model := range modelsToCreate {
		if _, err := tx.NewCreateTable().Model(model).IfNotExists().Exec(ctx); err != nil {
			// If Session fails due to UserStory FK, it might need adjustment or deferrable constraints
			return fmt.Errorf("failed to create table for %T: %w", model, err)
		}
	}
	// Re-apply FK constraint for sessions.current_story_id if needed after user_stories table exists
	// Or ensure FK is created as DEFERRABLE INITIALLY DEFERRED if schema managed outside Bun
	log.Println("Finished creating tables.")

	// --- Seed Users ---
	log.Printf("Seeding %d Users...\n", numUsers)
	users := make([]*models.User, 0, numUsers)
	for range numUsers {
		user := &models.User{
			ID:    cuid.New(),
			Name:  gofakeit.Name(),
			Email: gofakeit.Email(),
			// CreatedAt/UpdatedAt handled by Bun
		}
		users = append(users, user)
	}
	if _, err := tx.NewInsert().Model(&users).Exec(ctx); err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	// --- Seed Organisations ---
	log.Printf("Seeding %d Organisations...\n", numOrganisations)
	organisations := make([]*models.Organisation, 0, numOrganisations)
	for range numOrganisations {
		org := &models.Organisation{
			ID:   cuid.New(),
			Name: gofakeit.Company() + " Org",
		}
		organisations = append(organisations, org)
	}
	if _, err := tx.NewInsert().Model(&organisations).Exec(ctx); err != nil {
		return fmt.Errorf("failed to seed organisations: %w", err)
	}

	// --- Seed Organisation Members ---
	log.Println("Seeding Organisation Members...")
	orgMembers := make([]*models.OrganisationMember, 0)
	userOrgMap := make(map[string][]string) // orgID -> []userID
	orgUserMap := make(map[string][]string) // userID -> []orgID

	for _, org := range organisations {
		memberCount := min(gofakeit.Number(1, maxMembersPerOrg), len(users))
		potentialMembers := rand.Perm(len(users)) // Shuffle user indices

		count := 0
		for _, userIdx := range potentialMembers {
			if count >= memberCount {
				break
			}
			user := users[userIdx]

			// Avoid adding same user multiple times to the same org implicitly
			// (though PK constraint would prevent it anyway)
			isAlreadyMember := slices.Contains(orgUserMap[user.ID], org.ID)
			if isAlreadyMember {
				continue
			}

			role := "member"
			if gofakeit.Bool() && count == 0 { // Make the first member sometimes an admin
				role = "admin"
			}

			member := &models.OrganisationMember{
				OrganisationID: org.ID,
				UserID:         user.ID,
				Role:           role,
				JoinedAt:       gofakeit.DateRange(time.Now().AddDate(0, -6, 0), time.Now()),
			}
			orgMembers = append(orgMembers, member)
			userOrgMap[org.ID] = append(userOrgMap[org.ID], user.ID)
			orgUserMap[user.ID] = append(orgUserMap[user.ID], org.ID)
			count++
		}
	}
	if len(orgMembers) > 0 {
		if _, err := tx.NewInsert().Model(&orgMembers).Exec(ctx); err != nil {
			return fmt.Errorf("failed to seed organisation members: %w", err)
		}
	}

	// --- Seed Sessions ---
	log.Println("Seeding Sessions...")
	sessions := make([]*models.Session, 0)
	sessionParticipants := make([]*models.SessionParticipant, 0)
	sessionOrgMap := make(map[string]string)           // sessionID -> orgID
	sessionParticipantMap := make(map[string][]string) // sessionID -> []userID

	for _, org := range organisations {
		orgMemberIDs := userOrgMap[org.ID]
		if len(orgMemberIDs) == 0 {
			continue // Skip orgs with no members
		}

		for i := range numSessionsPerOrg {
			session := &models.Session{
				ID:             cuid.New(),
				OrganisationID: org.ID,
				Name:           fmt.Sprintf("%s Planning %d", gofakeit.Word(), i+1),
				IsActive:       gofakeit.Bool(),
				// CurrentStoryID will be updated later if needed
			}
			sessions = append(sessions, session)
			sessionOrgMap[session.ID] = org.ID

			// Add participants from the organisation members
			participantCount := min(gofakeit.Number(1, maxParticipantsPerSession), len(orgMemberIDs))
			potentialParticipants := rand.Perm(len(orgMemberIDs)) // Shuffle member indices

			count := 0
			for _, memberIdx := range potentialParticipants {
				if count >= participantCount {
					break
				}
				userID := orgMemberIDs[memberIdx]
				participant := &models.SessionParticipant{
					SessionID: session.ID,
					UserID:    userID,
					IsOnline:  gofakeit.Bool(), // Random online status
					JoinedAt:  gofakeit.DateRange(time.Now().AddDate(0, -1, 0), time.Now()),
				}
				sessionParticipants = append(sessionParticipants, participant)
				sessionParticipantMap[session.ID] = append(sessionParticipantMap[session.ID], userID)
				count++
			}
		}
	}
	if len(sessions) > 0 {
		if _, err := tx.NewInsert().Model(&sessions).Exec(ctx); err != nil {
			return fmt.Errorf("failed to seed sessions: %w", err)
		}
	}
	if len(sessionParticipants) > 0 {
		if _, err := tx.NewInsert().Model(&sessionParticipants).Exec(ctx); err != nil {
			return fmt.Errorf("failed to seed session participants: %w", err)
		}
	}

	// --- Seed User Stories ---
	log.Println("Seeding User Stories...")
	userStories := make([]*models.UserStory, 0)
	storySessionMap := make(map[string]string) // storyID -> sessionID

	for _, session := range sessions {
		participants := sessionParticipantMap[session.ID]
		if len(participants) == 0 {
			continue // Skip sessions with no participants
		}

		for i := range numStoriesPerSession {
			desc := gofakeit.Sentence(15)
			status := "pending"
			if gofakeit.Bool() {
				status = "estimated" // Some stories already estimated
			} else if gofakeit.Bool() {
				status = "voting" // Some are currently being voted on
			}

			story := &models.UserStory{
				ID:          cuid.New(),
				SessionID:   session.ID,
				Title:       fmt.Sprintf("Story %d: %s", i+1, gofakeit.HackerPhrase()),
				Description: &desc,
				Index:       fmt.Sprintf("%d", i),
				Status:      status,
				// FinalEstimate will be set later for 'estimated' stories
			}
			userStories = append(userStories, story)
			storySessionMap[story.ID] = session.ID

			// Potentially set the session's current story ID
			if status == "voting" && session.CurrentStoryID == nil {
				session.CurrentStoryID = &story.ID // Point to the first 'voting' story
				// Update the session in the DB (might be better to do this in a separate update pass)
				// For simplicity here, we'll update later if needed or rely on app logic
			}
		}
	}
	if len(userStories) > 0 {
		if _, err := tx.NewInsert().Model(&userStories).Exec(ctx); err != nil {
			return fmt.Errorf("failed to seed user stories: %w", err)
		}
	}

	// --- Seed Votes ---
	log.Println("Seeding Votes...")
	votes := make([]*models.Vote, 0)
	storiesToUpdate := make(map[string]*models.UserStory) // Track stories needing final estimate or status update

	for _, story := range userStories {
		sessionID := story.SessionID
		participants := sessionParticipantMap[sessionID]
		if len(participants) == 0 || story.Status == "pending" {
			continue // No participants or story not voted on yet
		}

		numRounds := 1
		if story.Status == "estimated" || gofakeit.Bool() { // Simulate multiple rounds sometimes
			numRounds = gofakeit.Number(1, maxVotingRounds)
		}

		finalEstimate := "" // Track consensus for the last round

		for round := 1; round <= numRounds; round++ {
			roundVotes := make(map[string]string) // userID -> estimate for this round
			numVotesThisRound := min(gofakeit.Number(1, len(participants)), maxVotesPerStory)

			potentialVoters := rand.Perm(len(participants)) // Shuffle participant indices
			voterCount := 0

			for _, voterIdx := range potentialVoters {
				if voterCount >= numVotesThisRound {
					break
				}
				userID := participants[voterIdx]
				estimate := estimateValues[gofakeit.Number(0, len(estimateValues)-1)]

				vote := &models.Vote{
					ID:          cuid.New(),
					UserStoryID: story.ID,
					UserID:      userID,
					SessionID:   sessionID, // Denormalized
					Round:       round,
					Estimate:    estimate,
					CreatedAt:   gofakeit.DateRange(story.CreatedAt, time.Now()),
				}
				votes = append(votes, vote)
				roundVotes[userID] = estimate
				voterCount++
			}

			// Determine consensus for the final round if story is 'estimated'
			if round == numRounds && story.Status == "estimated" {
				// Simple consensus: most frequent vote (excluding '?')
				counts := make(map[string]int)
				maxCount := 0
				consensusEstimate := ""
				for _, est := range roundVotes {
					if est == "?" || est == "coffee" {
						continue
					} // Ignore non-numeric for consensus example
					counts[est]++
					if counts[est] > maxCount {
						maxCount = counts[est]
						consensusEstimate = est
					} else if counts[est] == maxCount {
						// Tie-breaker (e.g., highest value, or just keep first one found)
						// For simplicity, keep the first one found with max count
					}
				}
				if consensusEstimate != "" {
					finalEstimate = consensusEstimate
				} else if len(roundVotes) > 0 {
					// Fallback if only '?' or 'coffee' were voted, or no votes
					// Pick a random valid estimate as placeholder
					finalEstimate = estimateValues[gofakeit.Number(0, 7)] // Pick from 0-20
				}
			}
		}

		// Mark story for update if it was estimated
		if story.Status == "estimated" && finalEstimate != "" {
			story.FinalEstimate = &finalEstimate
			storiesToUpdate[story.ID] = story
		} else if story.Status == "voting" && sessionParticipantMap[story.SessionID] != nil {
			// If story is 'voting', ensure the session points to it if no other story is current
			s := findSessionByID(sessions, story.SessionID)
			if s != nil && s.CurrentStoryID == nil {
				s.CurrentStoryID = &story.ID
				// Need to update the session model itself
				_, err := tx.NewUpdate().Model(s).Column("current_story_id").WherePK().Exec(ctx)
				if err != nil {
					log.Printf("Warning: failed to update current_story_id for session %s: %v", s.ID, err)
					// Continue seeding, but log the issue
				}
			}
		}
	}

	if len(votes) > 0 {
		// Insert votes in batches if very large
		if _, err := tx.NewInsert().Model(&votes).Exec(ctx); err != nil {
			return fmt.Errorf("failed to seed votes: %w", err)
		}
	}

	// Update stories with final estimates
	if len(storiesToUpdate) > 0 {
		log.Println("Updating stories with final estimates...")
		for _, story := range storiesToUpdate {
			// Update only the final_estimate column
			_, err := tx.NewUpdate().
				Model(story).
				Column("final_estimate").
				WherePK().
				Exec(ctx)
			if err != nil {
				// Log warning but continue, maybe FK constraint issue or story deleted?
				log.Printf("Warning: failed to update final_estimate for story %s: %v", story.ID, err)
			}
		}
	}

	// --- Commit Transaction ---
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Println("Database seeding completed successfully!")
	return nil
}

// Helper function to find a session by ID in the slice (needed for updating current_story_id)
func findSessionByID(sessions []*models.Session, id string) *models.Session {
	for _, s := range sessions {
		if s.ID == id {
			return s
		}
	}
	return nil
}

// Helper function to connect Bun DB (can be moved to database package)
func ConnectBunDB(dsn string) (*bun.DB, *sql.DB) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true))) // Optional debug logging
	return db, sqldb
}
