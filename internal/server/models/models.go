package models

import (
	"time"

	"github.com/uptrace/bun"
)

// --- Base Model ---
// Provides common fields like ID, CreatedAt, UpdatedAt.
// We'll embed this in other models. Bun handles CreatedAt/UpdatedAt automatically.

// --- User ---
// Represents a user in the system.
//
// SQL Schema:
// CREATE TABLE users (
//
//	id VARCHAR(25) PRIMARY KEY, -- CUID length
//	name VARCHAR(255) NOT NULL,
//	email VARCHAR(255) NOT NULL UNIQUE,
//	created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
//	updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
//
// );
type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID        string    `bun:"id,pk,type:varchar(25)"` // CUIDs are typically 25 chars
	Name      string    `bun:"name,notnull"`
	Email     string    `bun:"email,notnull,unique"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp"`

	// Relations
	OrganisationMembers []*OrganisationMember `bun:"rel:has-many,join:id=user_id"`
	SessionParticipants []*SessionParticipant `bun:"rel:has-many,join:id=user_id"`
	Votes               []*Vote               `bun:"rel:has-many,join:id=user_id"`
}

// --- Organisation ---
// Represents a team or company.
//
// SQL Schema:
// CREATE TABLE organisations (
//
//	id VARCHAR(25) PRIMARY KEY,
//	name VARCHAR(255) NOT NULL,
//	created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
//	updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
//
// );
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

// --- OrganisationMember (Join Table) ---
// Links Users to Organisations.
//
// SQL Schema:
// CREATE TABLE organisation_members (
//
//	organisation_id VARCHAR(25) NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
//	user_id VARCHAR(25) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
//	role VARCHAR(50) NOT NULL DEFAULT 'member', -- e.g., 'admin', 'member'
//	joined_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
//	PRIMARY KEY (organisation_id, user_id)
//
// );
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

// --- Session ---
// Represents a Planning Poker session.
//
// SQL Schema:
// CREATE TABLE sessions (
//
//	id VARCHAR(25) PRIMARY KEY,
//	organisation_id VARCHAR(25) NOT NULL REFERENCES organisations(id) ON DELETE CASCADE,
//	name VARCHAR(255) NOT NULL,
//	current_story_id VARCHAR(25) NULL REFERENCES user_stories(id) ON DELETE SET NULL, -- Story currently being voted on
//	is_active BOOLEAN NOT NULL DEFAULT TRUE,
//	created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
//	updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
//
// );
// CREATE INDEX idx_sessions_organisation_id ON sessions(organisation_id);
// CREATE INDEX idx_sessions_current_story_id ON sessions(current_story_id);
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

// --- SessionParticipant (Join Table) ---
// Links Users (participants) to Sessions.
//
// SQL Schema:
// CREATE TABLE session_participants (
//
//	session_id VARCHAR(25) NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
//	user_id VARCHAR(25) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
//	is_online BOOLEAN NOT NULL DEFAULT FALSE, -- For live presence tracking
//	joined_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
//	PRIMARY KEY (session_id, user_id)
//
// );
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

// --- UserStory ---
// Represents a story/task to be estimated.
//
// SQL Schema:
// CREATE TABLE user_stories (
//
//	id VARCHAR(25) PRIMARY KEY,
//	session_id VARCHAR(25) NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
//	title VARCHAR(255) NOT NULL,
//	description TEXT,
//	story_order INTEGER NOT NULL DEFAULT 0, -- For ordering within a session
//	status VARCHAR(50) NOT NULL DEFAULT 'pending', -- e.g., 'pending', 'voting', 'estimated'
//	final_estimate VARCHAR(20), -- Store the consensus value (string to allow "?", "coffee")
//	created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
//	updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
//
// );
// CREATE INDEX idx_user_stories_session_id ON user_stories(session_id);
type UserStory struct {
	bun.BaseModel `bun:"table:user_stories,alias:us"`

	ID            string    `bun:"id,pk,type:varchar(25)"`
	SessionID     string    `bun:"session_id,notnull,type:varchar(25)"`
	Title         string    `bun:"title,notnull"`
	Description   *string   `bun:"description,nullzero"` // Pointer for NULLable text
	Order         int       `bun:"story_order,notnull,default:0"`
	Status        string    `bun:"status,notnull,default:'pending'"`
	FinalEstimate *string   `bun:"final_estimate,nullzero,type:varchar(20)"` // Pointer for NULLable varchar
	CreatedAt     time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt     time.Time `bun:"updated_at,notnull,default:current_timestamp"`

	// Relations
	Session *Session `bun:"rel:belongs-to,join:session_id=id"`
	Votes   []*Vote  `bun:"rel:has-many,join:id=user_story_id"`
}

// --- Vote ---
// Represents a single estimate vote by a user for a specific story in a session round.
//
// SQL Schema:
// CREATE TABLE votes (
//
//	id VARCHAR(25) PRIMARY KEY,
//	user_story_id VARCHAR(25) NOT NULL REFERENCES user_stories(id) ON DELETE CASCADE,
//	user_id VARCHAR(25) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
//	session_id VARCHAR(25) NOT NULL REFERENCES sessions(id) ON DELETE CASCADE, -- Denormalized for easier querying? Or derive via user_story_id -> session_id
//	round INTEGER NOT NULL, -- To distinguish votes in re-voting rounds
//	estimate VARCHAR(20) NOT NULL, -- e.g., "1", "3", "5", "?", "coffee"
//	created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
//	-- No updated_at needed for votes typically
//
// );
// CREATE INDEX idx_votes_user_story_id ON votes(user_story_id);
// CREATE INDEX idx_votes_user_id ON votes(user_id);
// CREATE INDEX idx_votes_session_id ON votes(session_id);
// CREATE UNIQUE INDEX uidx_votes_story_user_round ON votes(user_story_id, user_id, round); -- Ensure one vote per user per story per round
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
