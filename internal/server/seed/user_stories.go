package seed

import (
	"context"
	"fmt"
	"log"
	"planning-poker/internal/server/models"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/lucsky/cuid"
	"github.com/uptrace/bun"
)

func seedUserStories(ctx context.Context, sessions []*models.Session, sessionParticipantMap map[string][]string, tx bun.Tx) ([]*models.UserStory, error) {
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
			return nil, fmt.Errorf("failed to seed user stories: %w", err)
		}
	}
	return userStories, nil
}
