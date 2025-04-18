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

func seedSessions(ctx context.Context, organisations []*models.Organisation, userOrgMap map[string][]string, tx bun.Tx) ([]*models.Session, map[string][]string, error) {
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
			return nil, nil, fmt.Errorf("failed to seed sessions: %w", err)
		}
	}
	if len(sessionParticipants) > 0 {
		if _, err := tx.NewInsert().Model(&sessionParticipants).Exec(ctx); err != nil {
			return nil, nil, fmt.Errorf("failed to seed session participants: %w", err)
		}
	}
	return sessions, sessionParticipantMap, nil
}
