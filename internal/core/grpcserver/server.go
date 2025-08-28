package grpcserver

import (
	"context"
	"google.golang.org/protobuf/types/known/structpb"
	"log"
	"lootor_posts/gen/go/posts"
	"lootor_posts/internal/core/models"
	"lootor_posts/internal/core/services"
	"lootor_posts/pkg/utils"
	"strconv"

	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PostsService struct {
	posts.UnimplementedPostsServiceServer
	service *services.PostsService
}

func NewPostsService(svc *services.PostsService) *PostsService {
	return &PostsService{service: svc}
}

func (s *PostsService) CreatePost(ctx context.Context, req *posts.CreatePostRequest) (*posts.PostResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	if req.GetAuthor() == "" {
		return nil, status.Error(codes.InvalidArgument, "author is required")
	}

	contentJSON := utils.NormalizeContent(req.GetContent())
	if contentJSON == nil {
		return nil, status.Error(codes.InvalidArgument, "content cannot be empty")
	}

	dto := &models.PostsRequest{
		Date:    req.GetDate(),
		Author:  req.GetAuthor(),
		Content: contentJSON,
	}

	item, err := s.service.CreatePost(dto)
	if err != nil {
		log.Printf("Failed to create comment: %v", err)
		return nil, status.Error(codes.Internal, "failed to create comment")
	}

	content, err := utils.GormJSONToProtoStruct(item.Data.Content)
	if err != nil {
		log.Printf("Content conversion failed: %v", err)
		content = &structpb.Struct{}
	}

	formatTime := func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.Format(time.RFC3339)
	}

	return &posts.PostResponse{
		Data: &posts.PostItem{
			Id:        item.Data.Id.String(),
			Date:      item.Data.Date,
			Content:   content,
			CreatedAt: formatTime(item.Data.CreatedAt),
			UpdatedAt: formatTime(item.Data.UpdatedAt),
			Author:    item.Data.Author,
		},
	}, nil
}

func (s *PostsService) GetPostsByUser(ctx context.Context, req *posts.GetPostsByUserRequest) (*posts.GetAllPostsResponse, error) {
	allPosts, err := s.service.GetPostsByUser(req.GetLogin(), req.GetLimit(), req.GetOffset(), req.GetIsPremium(), req.GetAuthUserLogin())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	normalizedPosts := convertPostsToProto(allPosts.Data)

	return &posts.GetAllPostsResponse{
		Data:  normalizedPosts,
		Total: strconv.FormatInt(allPosts.Total, 10),
	}, nil
}

func (s *PostsService) DeletePost(ctx context.Context, req *posts.DeletePostRequest) (*posts.DeletePostResponse, error) {
	_, err := s.service.DeletePost(req.GetId(), ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &posts.DeletePostResponse{
		Success: true,
	}, nil
}

func (s *PostsService) GetAllPosts(ctx context.Context, req *posts.GetAllPostsRequest) (*posts.GetAllPostsResponse, error) {
	allPosts, err := s.service.GetAllPosts(req.GetOrder(), req.GetLimit(), req.GetOffset(), req.GetIsPremium(), req.GetAuthUserLogin())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	normalizedPosts := convertPostsToProto(allPosts.Data)

	return &posts.GetAllPostsResponse{
		Data:  normalizedPosts,
		Total: strconv.FormatInt(allPosts.Total, 10),
	}, nil
}

func (s *PostsService) GetPost(ctx context.Context, req *posts.PostRequest) (*posts.PostResponse, error) {
	post, err := s.service.GetPost(ctx, req.GetId(), req.GetIsPremium(), req.GetAuthUserLogin())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	content, _ := utils.GormJSONToProtoStruct(post.Data.Content)
	protoReacts := convertReactionsToProto(post.Data.Reactions)
	return &posts.PostResponse{
		Data: &posts.PostItem{
			Id:             post.Data.Id.String(),
			Date:           post.Data.Date,
			Content:        content,
			CreatedAt:      post.Data.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      post.Data.UpdatedAt.Format(time.RFC3339),
			Author:         post.Data.Author,
			Reactions:      protoReacts,
			Reacted:        post.Data.Reacted,
			FireCount:      int64(post.Data.FireCount),
			LaughCount:     int64(post.Data.LaughCount),
			HeartCount:     int64(post.Data.HeartCount),
			AngryCount:     int64(post.Data.AngryCount),
			ShitCount:      int64(post.Data.ShitCount),
			ClownCount:     int64(post.Data.ClownCount),
			GlassesCount:   int64(post.Data.GlassesCount),
			EyesCount:      int64(post.Data.EyesCount),
			PokerFaceCount: int64(post.Data.PokerFaceCount),
			TotalReactions: int64(post.Data.TotalReactions),
			TearsCount:     int64(post.Data.TearsCount),
		},
	}, nil
}

func (s *PostsService) IncrementReaction(ctx context.Context, req *posts.ReactRequest) (*posts.ReactResponse, error) {
	author, err := s.service.IncrementReactions(req.GetPostId(), models.ReactionType(req.GetReaction()), ctx, req.GetUserLogin())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &posts.ReactResponse{
		Success:   true,
		UserLogin: author,
	}, nil
}

func (s *PostsService) DecrementReaction(ctx context.Context, req *posts.ReactDecrementRequest) (*posts.ReactResponse, error) {
	author, err := s.service.DecrementReactions(req.GetPostId(), models.ReactionType(req.GetReaction()), ctx, req.GetUserLogin())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &posts.ReactResponse{
		Success:   true,
		UserLogin: author,
	}, nil
}

func (s *PostsService) UpdatePost(ctx context.Context, req *posts.UpdatePostRequest) (*posts.PostResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if req.GetContent() == nil {
		return nil, status.Error(codes.InvalidArgument, "content is required")
	}
	post, err := s.service.UpdatePost(req.GetId(), &models.PostsRequest{
		Date:    "",
		Author:  "",
		Content: utils.NormalizeContent(req.GetContent()),
	}, ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoContent, err := utils.GormJSONToProtoStruct(post.Data.Content)
	if err != nil {
		log.Printf("Content conversion failed: %v", err)
		protoContent = &structpb.Struct{}
	}
	protoReacts := convertReactionsToProto(post.Data.Reactions)
	return &posts.PostResponse{
		Data: &posts.PostItem{
			Id:        post.Data.Id.String(),
			Date:      post.Data.Date,
			Content:   protoContent,
			CreatedAt: post.Data.CreatedAt.Format(time.RFC3339),
			UpdatedAt: post.Data.UpdatedAt.Format(time.RFC3339),
			Author:    post.Data.Author,
			Reactions: protoReacts,
		},
	}, nil
}

func convertPostsToProto(postList []models.PostResponse) []*posts.PostItem {
	var protoPostList []*posts.PostItem

	for _, post := range postList {
		content, _ := utils.GormJSONToProtoStruct(post.Content)
		protoReacts := convertReactionsToProto(post.Reactions)
		protoPostList = append(protoPostList, &posts.PostItem{
			Id:             post.Id.String(),
			Date:           post.Date,
			Content:        content,
			CreatedAt:      post.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      post.UpdatedAt.Format(time.RFC3339),
			Author:         post.Author,
			Reactions:      protoReacts,
			TearsCount:     int64(post.TearsCount),
			Reacted:        post.Reacted,
			FireCount:      int64(post.FireCount),
			LaughCount:     int64(post.LaughCount),
			HeartCount:     int64(post.HeartCount),
			AngryCount:     int64(post.AngryCount),
			ShitCount:      int64(post.ShitCount),
			ClownCount:     int64(post.ClownCount),
			GlassesCount:   int64(post.GlassesCount),
			EyesCount:      int64(post.EyesCount),
			PokerFaceCount: int64(post.PokerFaceCount),
			TotalReactions: int64(post.TotalReactions),
		})
	}
	return protoPostList
}

func convertReactionsToProto(reactions []models.PostReactions) []*posts.React {
	var protoReactions []*posts.React
	for _, reaction := range reactions {
		protoReactions = append(protoReactions, &posts.React{
			UserLogin: reaction.UserLogin,
			PostId:    reaction.PostId,
			Id:        reaction.Id.String(),
			CreatedAt: reaction.CreatedAt.Format(time.RFC3339),
			Reaction:  string(reaction.Reaction),
		})
	}
	return protoReactions
}
