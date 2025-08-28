package repositories

import (
	"gorm.io/gorm"
	"lootor_posts/internal/core/models"
	"time"
)

type PostReactionsRepository struct {
	db *gorm.DB
}

func NewPostReactionsRepository(db *gorm.DB) *PostReactionsRepository {
	return &PostReactionsRepository{db: db}
}

func (r *PostReactionsRepository) CreateRecord(reaction *models.PostReactions) (*models.PostReactions, error) {

	err := r.db.Create(reaction)
	if err.Error != nil {
		return nil, err.Error
	}

	return reaction, nil
}

func (r *PostReactionsRepository) DeleteRecord(reaction *models.PostReactions) error {
	deletedDate := time.Now().Format("2006-01-02 15:04:05")
	err := r.db.Model(&models.PostReactions{}).Where("id = ?", reaction.Id).Update("deleted_at", deletedDate).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *PostReactionsRepository) FindReaction(postId, login string) (*models.PostReactions, error) {

	var reaction models.PostReactions

	err := r.db.Model(&models.PostReactions{}).Where("post_id = ? AND user_login = ?", postId, login).Preload("Post").First(&reaction).Error
	if err != nil {
		return nil, err
	}
	return &reaction, nil
}
