package database

import (
	"context"
	"planning-poker/internal/server/models"
	"time"

	"github.com/lucsky/cuid"
)

func (s *service) CreateUser(name string, email string) (*models.User, error) {
	user := &models.User{
		ID:        cuid.New(),
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := s.db.NewInsert().Model(user).Exec(context.Background())
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *service) GetUser(email string) (*models.User, error) {
	user := new(models.User)
	println(email)
	err := s.db.NewSelect().Model(user).Where("email = ?", email).Scan(context.Background())
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *service) GetUserWithOrg(email string) (*models.User, error) {
	user := new(models.User)
	err := s.db.NewSelect().
		Model(user).
		Relation("OrganisationMembers.Organisation"). // Joins orgs through org members
		Where("u.email = ?", email).
		Scan(context.Background())
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *service) UpdateUser(name string, email string) (*models.User, error) {
	user := &models.User{
		Name: name,
	}

	_, err := s.db.NewUpdate().Model(user).Column("name").Where("email = ?", email).Exec(context.Background())
	if err != nil {
		return nil, err
	}

	return user, nil
}
