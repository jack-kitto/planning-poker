package database

import (
	"context"
	"planning-poker/internal/server/models"
	"time"

	"github.com/lucsky/cuid"
)

func (s *service) CreateSession(org *models.Organisation, owner *models.User, name string) (*models.Session, error) {
	sesh := &models.Session{
		ID:             cuid.New(),
		Name:           name,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Organisation:   org,
		OrganisationID: org.ID,
	}

	_, err := s.db.NewInsert().Model(sesh).Exec(context.Background())
	if err != nil {
		return nil, err
	}

	return sesh, nil
}

func (s *service) UpdateSessionName(id string, name string) error {
	sesh := &models.Session{
		Name: name,
	}

	_, err := s.db.NewUpdate().Column("name").Where("id = ?", id).Model(sesh).Exec(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (s *service) CreateSessionParticipant(sesh *models.Session, owner *models.User) (*models.SessionParticipant, error) {
	seshParticipant := &models.SessionParticipant{
		User:      owner,
		Session:   sesh,
		SessionID: sesh.ID,
		UserID:    owner.ID,
		IsOnline:  true,
		JoinedAt:  time.Now(),
	}

	_, err := s.db.NewInsert().Model(seshParticipant).Exec(context.Background())
	if err != nil {
		return nil, err
	}

	return seshParticipant, nil
}

func (s *service) GetSession(id string) (*models.Session, error) {
	session := new(models.Session)
	err := s.db.NewSelect().
		Model(session).
		Relation("Participants.User").
		Relation("Organisation").
		Relation("UserStories.Votes").
		Relation("CurrentStory").
		Where("s.id = ?", id).
		Scan(context.Background())
	if err != nil {
		return nil, err
	}
	return session, nil
}
