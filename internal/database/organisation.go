package database

import (
	"context"
	"planning-poker/internal/server/models"
	"time"

	"github.com/lucsky/cuid"
)

func (s *service) CreateOrgMember(org *models.Organisation, user *models.User) (*models.OrganisationMember, error) {
	orgMember := &models.OrganisationMember{
		UserID:         user.ID,
		User:           user,
		Role:           "member",
		JoinedAt:       time.Now(),
		Organisation:   org,
		OrganisationID: org.ID,
	}

	_, err := s.db.NewInsert().Model(orgMember).Exec(context.Background())
	if err != nil {
		return nil, err
	}

	return orgMember, nil
}

func (s *service) CreateOrg(name string, user *models.User) (*models.Organisation, error) {
	org := &models.Organisation{
		ID:        cuid.New(),
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := s.db.NewInsert().Model(org).Exec(context.Background())
	if err != nil {
		return nil, err
	}

	return org, nil
}
