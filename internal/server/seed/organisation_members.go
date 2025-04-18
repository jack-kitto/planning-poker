package seed

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"planning-poker/internal/server/models"
	"slices"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/uptrace/bun"
)

func seedOrganisationMembers(ctx context.Context, organisations []*models.Organisation, users []*models.User, tx bun.Tx) (map[string][]string, error) {
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
			return nil, fmt.Errorf("failed to seed organisation members: %w", err)
		}
	}
	return userOrgMap, nil
}
