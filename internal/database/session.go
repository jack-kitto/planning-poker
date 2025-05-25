package database

import (
	"context"
	"log"
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

func (s *service) DeactivateSessionParticipant(sesh *models.Session, user *models.User) error {
	log.Println("[DeactivateSessionParticipant] start")
	seshParticipant := &models.SessionParticipant{
		IsOnline: false,
	}

	_, err := s.db.NewUpdate().Column("is_online").Where("user_id = ?", user.ID).Model(seshParticipant).Exec(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (s *service) ActivateSessionParticipant(sesh *models.Session, user *models.User) error {
	seshParticipant := &models.SessionParticipant{
		IsOnline: true,
	}

	_, err := s.db.NewUpdate().Column("is_online").Where("user_id = ?", user.ID).Model(seshParticipant).Exec(context.Background())
	if err != nil {
		return err
	}

	return nil
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

func (s *service) GetSessionsForUser(userID string) ([]*models.Session, error) {
	var sessions []*models.Session

	err := s.db.NewSelect().
		Model(&sessions).
		Join("JOIN session_participants sp ON sp.session_id = s.id").
		Where("sp.user_id = ?", userID).
		Relation("Organisation").
		Relation("Participants.User").
		OrderExpr("s.created_at DESC").
		Scan(context.Background())
	if err != nil {
		return nil, err
	}
	return sessions, nil
}
