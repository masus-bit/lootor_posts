package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
)

type Posts struct {
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt" gorm:"index"`

	Id             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Date           string    `json:"date"`
	Author         string    `json:"author"`
	HeartCount     int       `gorm:"default:0" json:"heartCount"`
	FireCount      int       `gorm:"default:0" json:"fireCount"`
	GlassesCount   int       `gorm:"default:0" json:"glassesCount"`
	LaughCount     int       `gorm:"default:0" json:"laughCount"`
	TearsCount     int       `gorm:"default:0" json:"tearsCount"`
	PokerFaceCount int       `gorm:"default:0" json:"pokerFaceCount"`
	EyesCount      int       `gorm:"default:0" json:"eyesCount"`
	AngryCount     int       `gorm:"default:0" json:"angryCount"`
	ShitCount      int       `gorm:"default:0" json:"shitCount"`
	ClownCount     int       `gorm:"default:0" json:"clownCount"`
	IsDraft        bool      `gorm:"default:false" json:"isDraft"`
	Views          int       `gorm:"default:0" json:"views"`
	CommentsCount  int       `gorm:"default:0" json:"commentsCount"`

	TotalReactions int            `gorm:"->;type:integer GENERATED ALWAYS AS (heart_count + fire_count + glasses_count + laugh_count + tears_count + poker_face_count + eyes_count + angry_count + shit_count + clown_count) STORED" json:"totalReactions"`
	Content        datatypes.JSON `gorm:"type:jsonb" json:"content"`

	Reactions []PostReactions `gorm:"foreignKey:PostId" json:"reactions"`
}

type PostResponse struct {
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt" gorm:"index"`

	Id             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Date           string    `json:"date"`
	Author         string    `json:"author"`
	HeartCount     int       `gorm:"default:0" json:"heartCount"`
	FireCount      int       `gorm:"default:0" json:"fireCount"`
	GlassesCount   int       `gorm:"default:0" json:"glassesCount"`
	LaughCount     int       `gorm:"default:0" json:"laughCount"`
	TearsCount     int       `gorm:"default:0" json:"tearsCount"`
	PokerFaceCount int       `gorm:"default:0" json:"pokerFaceCount"`
	EyesCount      int       `gorm:"default:0" json:"eyesCount"`
	AngryCount     int       `gorm:"default:0" json:"angryCount"`
	ShitCount      int       `gorm:"default:0" json:"shitCount"`
	ClownCount     int       `gorm:"default:0" json:"clownCount"`
	Reacted        string    `json:"reacted"`
	IsDraft        bool      `gorm:"default:false" json:"isDraft"`
	Views          int       `gorm:"default:0" json:"views"`
	CommentsCount  int       `gorm:"default:0" json:"commentsCount"`

	TotalReactions int `gorm:"->;type:GENERATED ALWAYS AS (heart_count + fire_count + glasses_count + laugh_count + tears_count + poker_face_count + eyes_count + angry_count + shit_count + clown_count)" json:"totalReactions"`

	Content datatypes.JSON `gorm:"type:jsonb" json:"content"`

	Reactions []PostReactions `gorm:"foreignKey:PostId" json:"reactions"`
}

type PostsDataResponse struct {
	Data  []PostResponse `json:"data"`
	Total int64          `json:"total"`
}

type PostDataResponse struct {
	Data PostResponse `json:"data"`
}

type PostsRequest struct {
	Date    string `json:"date"`
	Author  string `json:"author"`
	IsDraft bool   `json:"isDraft"`

	Content datatypes.JSON `json:"content"`
}
