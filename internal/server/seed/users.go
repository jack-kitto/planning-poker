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

func seedUsers(ctx context.Context, tx bun.Tx) ([]*models.User, []*models.Organisation, error) {
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
		return nil, nil, fmt.Errorf("failed to seed users: %w", err)
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
		return nil, nil, fmt.Errorf("failed to seed organisations: %w", err)
	}
	return users, organisations, nil
}
