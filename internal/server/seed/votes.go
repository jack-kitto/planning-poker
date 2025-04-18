package seed

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"planning-poker/internal/server/models"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/lucsky/cuid"
	"github.com/uptrace/bun"
)

func seedVotes(ctx context.Context, userStories []*models.UserStory, sessionParticipantMap map[string][]string, sessions []*models.Session, tx bun.Tx) error {
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
