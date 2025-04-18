package models

import (
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID        string    `bun:"id,pk,type:varchar(25)"`
	Name      string    `bun:"name,notnull"`
	Email     string    `bun:"email,notnull,unique"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp"`

	// Relations
	OrganisationMembers []*OrganisationMember `bun:"rel:has-many,join:id=user_id"`
	SessionParticipants []*SessionParticipant `bun:"rel:has-many,join:id=user_id"`
	Votes               []*Vote               `bun:"rel:has-many,join:id=user_id"`
}

type Organisation struct {
	bun.BaseModel `bun:"table:organisations,alias:o"`

	ID        string    `bun:"id,pk,type:varchar(25)"`
	Name      string    `bun:"name,notnull"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp"`

	// Relations
	Members  []*OrganisationMember `bun:"rel:has-many,join:id=organisation_id"`
	Sessions []*Session            `bun:"rel:has-many,join:id=organisation_id"`
}

type OrganisationMember struct {
	bun.BaseModel `bun:"table:organisation_members,alias:om"`

	OrganisationID string    `bun:"organisation_id,pk,type:varchar(25)"`
	UserID         string    `bun:"user_id,pk,type:varchar(25)"`
	Role           string    `bun:"role,notnull,default:'member'"`
	JoinedAt       time.Time `bun:"joined_at,notnull,default:current_timestamp"`

	// Relations
	Organisation *Organisation `bun:"rel:belongs-to,join:organisation_id=id"`
	User         *User         `bun:"rel:belongs-to,join:user_id=id"`
}

type Session struct {
	bun.BaseModel `bun:"table:sessions,alias:s"`

	ID             string    `bun:"id,pk,type:varchar(25)"`
	OrganisationID string    `bun:"organisation_id,notnull,type:varchar(25)"`
	Name           string    `bun:"name,notnull"`
	CurrentStoryID *string   `bun:"current_story_id,nullzero,type:varchar(25)"` // Pointer for NULLable FK
	IsActive       bool      `bun:"is_active,notnull,default:true"`
	CreatedAt      time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt      time.Time `bun:"updated_at,notnull,default:current_timestamp"`

	// Relations
	Organisation *Organisation         `bun:"rel:belongs-to,join:organisation_id=id"`
	Participants []*SessionParticipant `bun:"rel:has-many,join:id=session_id"`
	UserStories  []*UserStory          `bun:"rel:has-many,join:id=session_id"`
	CurrentStory *UserStory            `bun:"rel:has-one,join:current_story_id=id"`
}

type SessionParticipant struct {
	bun.BaseModel `bun:"table:session_participants,alias:sp"`

	SessionID string    `bun:"session_id,pk,type:varchar(25)"`
	UserID    string    `bun:"user_id,pk,type:varchar(25)"`
	IsOnline  bool      `bun:"is_online,notnull,default:false"`
	JoinedAt  time.Time `bun:"joined_at,notnull,default:current_timestamp"`

	// Relations
	Session *Session `bun:"rel:belongs-to,join:session_id=id"`
	User    *User    `bun:"rel:belongs-to,join:user_id=id"`
}

type UserStory struct {
	bun.BaseModel `bun:"table:user_stories,alias:us"`

	ID            string    `bun:"id,pk,type:varchar(25)"`
	SessionID     string    `bun:"session_id,notnull,type:varchar(25)"`
	Title         string    `bun:"title,notnull"`
	Description   *string   `bun:"description,nullzero"` // Pointer for NULLable text
	Index         string    `bun:"story_index,notnull,default:0"`
	Status        string    `bun:"status,notnull,default:'pending'"`
	FinalEstimate *string   `bun:"final_estimate,nullzero,type:varchar(20)"` // Pointer for NULLable varchar
	CreatedAt     time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     time.Time `bun:"updated_at,notnull,default:current_timestamp"`

	// Relations
	Session *Session `bun:"rel:belongs-to,join:session_id=id"`
	Votes   []*Vote  `bun:"rel:has-many,join:id=user_story_id"`
}

type Vote struct {
	bun.BaseModel `bun:"table:votes,alias:v"`

	ID          string    `bun:"id,pk,type:varchar(25)"`
	UserStoryID string    `bun:"user_story_id,notnull,type:varchar(25)"`
	UserID      string    `bun:"user_id,notnull,type:varchar(25)"`
	SessionID   string    `bun:"session_id,notnull,type:varchar(25)"` // Included for potential query convenience
	Round       int       `bun:"round,notnull"`
	Estimate    string    `bun:"estimate,notnull,type:varchar(20)"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp"`

	// Relations
	UserStory *UserStory `bun:"rel:belongs-to,join:user_story_id=id"`
	User      *User      `bun:"rel:belongs-to,join:user_id=id"`
	Session   *Session   `bun:"rel:belongs-to,join:session_id=id"`
}
