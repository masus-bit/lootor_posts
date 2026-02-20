package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type ReactionType string

const (
	Fire      ReactionType = "fire"
	Heart     ReactionType = "heart"
	Glasses   ReactionType = "glasses"
	Laugh     ReactionType = "laugh"
	Tears     ReactionType = "tears"
	PokerFace ReactionType = "pokerFace"
	Eyes      ReactionType = "eyes"
	Angry     ReactionType = "angry"
	Shit      ReactionType = "shit"
	Clown     ReactionType = "clown"
)

type PostReactions struct {
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt" gorm:"index"`

	Id        uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PostId    uuid.UUID    `gorm:"not null;uniqueIndex:idx_user_post" json:"postId"`
	UserLogin string       `gorm:"not null;uniqueIndex:idx_user_post" json:"userLogin"`
	Reaction  ReactionType `json:"reaction"`

	Post Posts `gorm:"foreignKey:PostId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
