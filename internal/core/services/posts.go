package services

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"lootor_posts/internal/core/models"
	"lootor_posts/internal/core/repositories"
	"lootor_posts/pkg/types"
)

type PostsService struct {
	postsRepo     *repositories.PostsRepository
	reactionsRepo *repositories.PostReactionsRepository
}

func NewPostsService(
	postsRepo *repositories.PostsRepository,
	reactionsRepo *repositories.PostReactionsRepository,
) *PostsService {
	return &PostsService{
		postsRepo:     postsRepo,
		reactionsRepo: reactionsRepo,
	}
}

func (s *PostsService) CreatePost(dto *models.PostsRequest) (*models.PostDataResponse, error) {
	post := &models.Posts{
		Date:    dto.Date,
		Author:  dto.Author,
		Content: dto.Content,
		IsDraft: dto.IsDraft,
		Title:   dto.Title,
	}

	post, err := s.postsRepo.CreateRecord(post)
	if err != nil {
		return nil, err
	}
	var postResponse models.PostResponse
	err = mapstructure.Decode(post, &postResponse)
	if err != nil {
		return nil, err
	}
	postResponse.Reacted = ""
	return &models.PostDataResponse{Data: postResponse}, nil
}

func (s *PostsService) GetPostsByUser(
	userLogin, limit, offset string,
	authUserIsPremium bool,
	authUserLogin string,
	isDraft bool,
) (*models.PostsDataResponse, error) {
	records, total, err := s.postsRepo.FindRecordsByUser(userLogin, limit, offset, isDraft)
	if err != nil {
		return nil, err
	}
	var result []models.PostResponse
	for _, record := range records {
		var postResponse models.PostResponse
		err = mapstructure.Decode(record, &postResponse)
		if err != nil {
			return nil, err
		}
		postResponse.Reacted = ""
		for _, reaction := range record.Reactions {
			if reaction.UserLogin == authUserLogin {
				postResponse.Reacted = string(reaction.Reaction)
				break
			}
		}
		if !authUserIsPremium {
			postResponse.Reactions = nil
		}
		result = append(result, postResponse)
	}

	return &models.PostsDataResponse{Data: result, Total: total}, nil
}

func (s *PostsService) GetAllPosts(
	order, limit, offset string,
	authUserIsPremium bool,
	authUserLogin string,
) (*models.PostsDataResponse, error) {
	records, total, err := s.postsRepo.FindRecords(order, limit, offset)
	if err != nil {
		return nil, err
	}
	var result []models.PostResponse
	for _, record := range records {
		var postResponse models.PostResponse
		err = mapstructure.Decode(record, &postResponse)
		if err != nil {
			return nil, err
		}
		postResponse.Reacted = ""
		for _, reaction := range record.Reactions {
			if reaction.UserLogin == authUserLogin {
				postResponse.Reacted = string(reaction.Reaction)
				break
			}
		}
		if !authUserIsPremium {
			postResponse.Reactions = nil
		}

		result = append(result, postResponse)
	}

	return &models.PostsDataResponse{
		Data:  result,
		Total: total,
	}, nil
}

func (s *PostsService) GetPost(
	ctx context.Context,
	id string,
	authUserIsPremium bool,
	authUserLogin string,
) (*models.PostDataResponse, error) {
	post, err := s.postsRepo.FindRecordById(id)
	if err != nil {
		return nil, err
	}
	var postResponse models.PostResponse
	err = mapstructure.Decode(post, &postResponse)
	if err != nil {
		return nil, err
	}
	postResponse.Reacted = ""
	for _, reaction := range post.Reactions {
		if reaction.UserLogin == authUserLogin {
			postResponse.Reacted = string(reaction.Reaction)
			break
		}
	}
	if !authUserIsPremium {
		postResponse.Reactions = nil
	}

	return &models.PostDataResponse{Data: postResponse}, nil
}

func (s *PostsService) GetPostByTranslit(
	ctx context.Context,
	translit string,
	authUserIsPremium bool,
	authUserLogin string,
) (*models.PostDataResponse, error) {
	post, err := s.postsRepo.FindRecordByTranslit(translit)
	if err != nil {
		return nil, err
	}

	var postResponse models.PostResponse
	err = mapstructure.Decode(post, &postResponse)
	if err != nil {
		return nil, err
	}
	postResponse.Reacted = ""
	for _, reaction := range post.Reactions {
		if reaction.UserLogin == authUserLogin {
			postResponse.Reacted = string(reaction.Reaction)
			break
		}
	}
	if !authUserIsPremium {
		postResponse.Reactions = nil
	}

	return &models.PostDataResponse{Data: postResponse}, nil
}

func (s *PostsService) IncrementReactions(
	id string,
	reactionType models.ReactionType,
	ctx context.Context,
	userLogin string,
) (string, error) {
	existsReaction, _ := s.reactionsRepo.FindReaction(id, userLogin)
	author := ""
	if existsReaction != nil {
		_, err := s.DecrementReactions(id, existsReaction.Reaction, ctx, userLogin)
		if err != nil {
			return "", err
		}
		author = existsReaction.Post.Author
	}
	err := s.postsRepo.IncrementReactions(reactionType, id)
	if err != nil {
		return "", err
	}
	uuidVal, _ := uuid.Parse(id)
	_, err = s.reactionsRepo.CreateRecord(
		&models.PostReactions{
			PostId:    uuidVal,
			UserLogin: userLogin,
			Reaction:  reactionType,
		},
	)
	if err != nil {
		return "", err
	}

	return author, nil
}

func (s *PostsService) DecrementReactions(
	postId string,
	reactionType models.ReactionType,
	ctx context.Context,
	userLogin string,
) (string, error) {
	err := s.postsRepo.DecrementLikes(reactionType, postId)
	author := ""
	if err != nil {
		return "", err
	}
	react, _ := s.reactionsRepo.FindReaction(postId, userLogin)
	if react == nil {
		return "", errors.New("reaction not found")
	}
	author = react.Post.Author
	uuidVal, _ := uuid.Parse(postId)
	err = s.reactionsRepo.DeleteRecord(
		&models.PostReactions{
			Id:        react.Id,
			PostId:    uuidVal,
			UserLogin: userLogin,
			Reaction:  reactionType,
		},
	)
	if err != nil {
		return "", err
	}
	return author, nil
}

func (s *PostsService) DeletePost(id string, ctx context.Context) (*types.CommonResponse, error) {
	post, err := s.postsRepo.FindRecordById(id)
	if err != nil {
		return nil, err
	}
	err = s.postsRepo.DeleteRecord(post)
	if err != nil {
		return nil, err
	}
	return &types.CommonResponse{
		Data:       types.Resp{Success: true},
		ReactCount: int64(post.TotalReactions),
		Title:      post.Title,
	}, nil
}

func (s *PostsService) UpdatePost(id string, dto *models.PostsRequest, ctx context.Context) (
	*models.PostDataResponse,
	error,
) {
	post, err := s.postsRepo.FindRecordById(id)
	if err != nil {
		return nil, err
	}
	post.Content = dto.Content
	post.IsDraft = dto.IsDraft
	post.Title = dto.Title
	if dto.Translit != "" && dto.Translit != post.Translit {
		post.Translit = dto.Translit
	}
	_, err = s.postsRepo.UpdateFull(post)
	if err != nil {
		return nil, err
	}
	var postResponse models.PostResponse
	err = mapstructure.Decode(post, &postResponse)
	if err != nil {
		return nil, err
	}
	return &models.PostDataResponse{Data: postResponse}, nil
}

func (s *PostsService) IncrementViews(ids []string, ctx context.Context) error {
	return s.postsRepo.IncrementViews(ids)
}

func (s *PostsService) IncrementCommentsCount(id string, ctx context.Context) error {
	return s.postsRepo.IncrementCommentsCount(id)
}

func (s *PostsService) DecrementCommentsCount(id string, ctx context.Context) error {
	return s.postsRepo.DecrementCommentsCount(id)
}

func (s *PostsService) GetCount(userLogin string, ctx context.Context) (int64, error) {
	count, err := s.postsRepo.GetCount(userLogin)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *PostsService) GetByIds(
	ids []string,
	authUserLogin string,
	authUserIsPremium bool,
	ctx context.Context,
) (*models.PostsMap, error) {
	posts, err := s.postsRepo.GetPostsByIdsMap(ids)
	postsMap := make(map[string]models.PostResponse)
	for key, record := range posts {
		var postResponse models.PostResponse
		err = mapstructure.Decode(record, &postResponse)
		if err != nil {
			return nil, err
		}
		postResponse.Reacted = ""
		for _, reaction := range record.Reactions {
			if reaction.UserLogin == authUserLogin {
				postResponse.Reacted = string(reaction.Reaction)
				break
			}
		}
		if !authUserIsPremium {
			postResponse.Reactions = nil
		}
		postsMap[key] = postResponse
	}
	if err != nil {
		return nil, err
	}
	return &models.PostsMap{Posts: postsMap}, nil
}

func (s *PostsService) GetPostsCountByUserLogin(userLogin string, ctx context.Context) (int64, error) {
	count, err := s.postsRepo.GetPostsCountByUserLogin(userLogin)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *PostsService) GetReactionsCountByUserLogin(userLogin string, ctx context.Context) (int64, error) {
	count, err := s.postsRepo.GetTotalPostsReactionsCountByUserLogin(userLogin)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *PostsService) GetPostsCountsByUserLogins(logins []string, ctx context.Context) (map[string]int64, error) {
	counts, err := s.postsRepo.GetCountsByLogins(logins)
	if err != nil {
		return nil, err
	}
	return counts, nil
}
