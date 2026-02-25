package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"lootor_posts/internal/core/models"
	"strconv"
)

type PostsRepository struct {
	db *gorm.DB
}

func NewPostsRepository(db *gorm.DB) *PostsRepository {
	return &PostsRepository{db: db}
}

func (r *PostsRepository) CreateRecord(post *models.Posts) (*models.Posts, error) {

	err := r.db.Create(post)
	if err.Error != nil {
		return nil, err.Error
	}

	return post, nil
}

func (r *PostsRepository) FindRecordsByUser(login, limit, offset string, isDraft bool) ([]models.Posts, int64, error) {
	var posts []models.Posts
	var totalCount int64
	limitInt, _ := strconv.Atoi(limit)
	offsetInt, _ := strconv.Atoi(offset)
	query := r.db.Unscoped().Where("author = ?", login).Where("is_draft = ?", isDraft).Where("deleted_at IS NULL")
	query = query.Preload("Reactions")
	query = query.Order("posts.date DESC").Limit(limitInt).Offset(offsetInt)
	err := query.Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}
	err = r.db.Model(&models.Posts{}).Unscoped().Where("author = ?", login).Where(
		"is_draft = ?",
		isDraft,
	).Where("deleted_at IS NULL").Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, totalCount, nil
}

func (r *PostsRepository) DeleteRecord(post *models.Posts) error {
	err := r.db.Delete(&models.Posts{}, "id = ?", post.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (r *PostsRepository) FindRecordById(id string) (*models.Posts, error) {
	var post models.Posts

	query := r.db.Where("id = ?", id).Where("deleted_at IS NULL")
	query = query.Preload("Reactions")
	err := query.First(&post).Error
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (r *PostsRepository) FindRecordByTranslit(translit string) (*models.Posts, error) {
	var post models.Posts

	query := r.db.Where("translit = ?", translit).Where("deleted_at IS NULL")
	query = query.Preload("Reactions")
	err := query.First(&post).Error
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func (r *PostsRepository) IncrementReactions(reactionType models.ReactionType, id string) error {

	err := r.db.Model(&models.Posts{}).Where("id = ?", id).
		Update(string(reactionType+"_count"), gorm.Expr("COALESCE("+string(reactionType)+"_count, 0) + ?", 1)).Error

	if err != nil {
		return err
	}

	return nil
}

func (r *PostsRepository) DecrementLikes(reactionType models.ReactionType, id string) error {

	err := r.db.Model(&models.Posts{}).Where("id = ?", id).
		Update(
			string(reactionType+"_count"),
			gorm.Expr("GREATEST(COALESCE("+string(reactionType)+"_count, 0) - ?, 0)", 1),
		).Error

	if err != nil {
		return err
	}

	return nil
}

func (r *PostsRepository) FindRecords(order, limit, offset string) ([]models.Posts, int64, error) {
	var posts []models.Posts

	var totalCount int64

	limitInt, _ := strconv.Atoi(limit)
	offsetInt, _ := strconv.Atoi(offset)

	var orderBy string
	switch order {
	case "reactions":
		orderBy = "total_reactions"
	default:
		orderBy = order
	}

	query := r.db.Unscoped().Where("is_draft = ?", false).Order(orderBy + " DESC").Where("deleted_at IS NULL")
	query = query.Preload("Reactions")
	err := query.Limit(limitInt).Offset(offsetInt).Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}
	err = r.db.Model(&models.Posts{}).Unscoped().Where("deleted_at IS NULL").Where(
		"is_draft = ?",
		false,
	).Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}
	return posts, totalCount, err

}

func (r *PostsRepository) UpdateFull(existsPost *models.Posts) (*models.Posts, error) {
	err := r.db.Transaction(
		func(tx *gorm.DB) error {
			if err := tx.Model(existsPost).Select("*").Updates(existsPost).Error; err != nil {
				return err
			}
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	var result models.Posts

	if err = r.db.Model(&models.Posts{}).
		First(&result, "id = ?", existsPost.Id).
		Error; err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *PostsRepository) IncrementViews(ids []string) error {
	err := r.db.Model(&models.Posts{}).Where("id IN (?)", ids).
		Update("views", gorm.Expr("COALESCE(views, 0) + ?", 1)).Error

	if err != nil {
		return err
	}

	return nil
}

func (r *PostsRepository) IncrementCommentsCount(id string) error {
	err := r.db.Model(&models.Posts{}).Where("id = ?", id).
		Update("comments_count", gorm.Expr("COALESCE(comments_count, 0) + ?", 1)).Error

	if err != nil {
		return err
	}

	return nil
}

func (r *PostsRepository) DecrementCommentsCount(id string) error {
	err := r.db.Model(&models.Posts{}).Where("id = ?", id).
		Update("comments_count", gorm.Expr("GREATEST(COALESCE(comments_count, 0) - ?, 0)", 1)).Error

	if err != nil {
		return err
	}

	return nil
}

func (r *PostsRepository) GetCount(userLogin string) (int64, error) {
	var count int64
	err := r.db.
		Model(&models.Posts{}).
		Where("is_draft = ?", false).
		Where("author = ?", userLogin).
		Where("deleted_at IS NULL").
		Count(&count).Error

	return count, err
}

func (r *PostsRepository) GetPostsByIdsMap(ids []string) (map[string]models.Posts, error) {
	var posts []models.Posts
	var validIDs []string

	for _, id := range ids {
		if _, err := uuid.Parse(id); err == nil {
			validIDs = append(validIDs, id)
		}
	}

	if len(validIDs) == 0 {
		return make(map[string]models.Posts), nil
	}

	query := r.db.Where("deleted_at IS NULL").
		Where("id IN (?)", validIDs).
		Preload("Reactions")

	err := query.Find(&posts).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]models.Posts)
	for _, post := range posts {
		result[post.Id.String()] = post
	}
	return result, nil
}

func (r *PostsRepository) GetPostsCountByUserLogin(userLogin string) (int64, error) {
	var count int64
	err := r.db.
		Model(&models.Posts{}).
		Where("is_draft = ?", false).
		Where("author = ?", userLogin).
		Where("deleted_at IS NULL").
		Count(&count).Error

	return count, err
}

func (r *PostsRepository) GetReactionsCountByUserLogin(userLogin string) (int64, error) {
	var count int64
	err := r.db.
		Model(&models.Posts{}).
		Where("is_draft = ?", false).
		Where("author = ?", userLogin).
		Where("deleted_at IS NULL").
		Count(&count).Error

	return count, err
}

func (r *PostsRepository) GetTotalPostsReactionsCountByUserLogin(login string) (int64, error) {
	var totalLikes int64

	query := `
        SELECT COALESCE(SUM(total_reactions), 0) as total_likes
        FROM loot_posts.posts
        WHERE LOWER(author) = LOWER($1) 
          AND deleted_at IS NULL
          AND is_draft = false
    `

	err := r.db.Raw(query, login).Scan(&totalLikes).Error
	return totalLikes, err
}
