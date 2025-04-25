package database

import (
	"context"
	"planning-poker/internal/server/models"
	"time"

	"github.com/lucsky/cuid"
)

func (s *service) UpdateStory(id string, title string, description string) error {
	story := &models.UserStory{
		Title:       title,
		Description: &description,
	}

	_, err := s.db.NewUpdate().Column("title", "description").Where("id = ?", id).Model(story).Exec(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (s *service) CreateStory(sesh *models.Session, title string, description *string, index string) (*models.UserStory, error) {
	story := &models.UserStory{
		ID:          cuid.New(),
		SessionID:   sesh.ID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Title:       title,
		Description: description,
		Index:       index,
		Status:      "pending",
		Session:     sesh,
	}

	_, err := s.db.NewInsert().Model(story).Exec(context.Background())
	if err != nil {
		return nil, err
	}

	return story, nil
}

func (s *service) GetUserStory(id string) (*models.UserStory, error) {
	story := new(models.UserStory)
	err := s.db.NewSelect().
		Model(story).
		Relation("Votes.User").
		Where("id = ?", id).
		Scan(context.Background())
	if err != nil {
		return nil, err
	}
	return story, nil
}
