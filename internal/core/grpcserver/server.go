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
		IsDraft: req.GetIsDraft(),
		Title:   req.GetTitle(),
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
			IsDraft:   item.Data.IsDraft,
			Title:     item.Data.Title,
			Translit:  item.Data.Translit,
		},
	}, nil
}

func (s *PostsService) GetPostsByUser(
	ctx context.Context,
	req *posts.GetPostsByUserRequest,
) (*posts.GetAllPostsResponse, error) {
	allPosts, err := s.service.GetPostsByUser(
		req.GetLogin(),
		req.GetLimit(),
		req.GetOffset(),
		req.GetIsPremium(),
		req.GetAuthUserLogin(),
		req.GetIsDraft(),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	normalizedPosts := convertPostsToProto(allPosts.Data)

	return &posts.GetAllPostsResponse{
		Data:  normalizedPosts,
		Total: strconv.FormatInt(allPosts.Total, 10),
	}, nil
}

func (s *PostsService) DeletePost(ctx context.Context, req *posts.DeletePostRequest) (
	*posts.DeletePostResponse,
	error,
) {
	result, err := s.service.DeletePost(req.GetId(), ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &posts.DeletePostResponse{
		Success:    true,
		ReactCount: result.ReactCount,
		Title:      result.Title,
	}, nil
}

func (s *PostsService) GetAllPosts(ctx context.Context, req *posts.GetAllPostsRequest) (
	*posts.GetAllPostsResponse,
	error,
) {
	allPosts, err := s.service.GetAllPosts(
		req.GetOrder(),
		req.GetLimit(),
		req.GetOffset(),
		req.GetIsPremium(),
		req.GetAuthUserLogin(),
	)
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
		Data: utils.FillPostItem(&post.Data, content, protoReacts),
	}, nil
}

func (s *PostsService) GetPostByTranslit(ctx context.Context, req *posts.PostRequestByTranslit) (
	*posts.PostResponse,
	error,
) {
	post, err := s.service.GetPostByTranslit(ctx, req.GetTranslit(), req.GetIsPremium(), req.GetAuthUserLogin())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	content, _ := utils.GormJSONToProtoStruct(post.Data.Content)
	protoReacts := convertReactionsToProto(post.Data.Reactions)
	return &posts.PostResponse{
		Data: utils.FillPostItem(&post.Data, content, protoReacts),
	}, nil
}

func (s *PostsService) IncrementReaction(ctx context.Context, req *posts.ReactRequest) (*posts.ReactResponse, error) {
	author, err := s.service.IncrementReactions(
		req.GetPostId(),
		models.ReactionType(req.GetReaction()),
		ctx,
		req.GetUserLogin(),
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &posts.ReactResponse{
		Success:   true,
		UserLogin: author,
	}, nil
}

func (s *PostsService) DecrementReaction(ctx context.Context, req *posts.ReactDecrementRequest) (
	*posts.ReactResponse,
	error,
) {
	author, err := s.service.DecrementReactions(
		req.GetPostId(),
		models.ReactionType(req.GetReaction()),
		ctx,
		req.GetUserLogin(),
	)
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
	post, err := s.service.UpdatePost(
		req.GetId(), &models.PostsRequest{
			Content:  utils.NormalizeContent(req.GetContent()),
			IsDraft:  req.GetIsDraft(),
			Title:    req.GetTitle(),
			Translit: req.GetTranslit(),
		}, ctx,
	)
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
		Data: utils.FillPostItem(&post.Data, protoContent, protoReacts),
	}, nil
}

func (s *PostsService) IncrementViews(ctx context.Context, req *posts.ViewsRequest) (*posts.ViewsResponse, error) {
	err := s.service.IncrementViews(req.GetPostIds(), ctx)
	if err != nil {
		return nil, err
	}

	return &posts.ViewsResponse{
		Success: true,
	}, nil
}

func (s *PostsService) IncrementCommentsCount(
	ctx context.Context,
	req *posts.CommentsCountRequest,
) (*posts.CommentsCountResponse, error) {
	err := s.service.IncrementCommentsCount(req.GetPostId(), ctx)
	if err != nil {
		return nil, err
	}

	return &posts.CommentsCountResponse{
		Success: true,
	}, nil
}

func (s *PostsService) DecrementCommentsCount(
	ctx context.Context,
	req *posts.CommentsCountRequest,
) (*posts.CommentsCountResponse, error) {
	err := s.service.DecrementCommentsCount(req.GetPostId(), ctx)
	if err != nil {
		return nil, err
	}

	return &posts.CommentsCountResponse{
		Success: true,
	}, nil
}

func (s *PostsService) GetCountByUser(ctx context.Context, req *posts.CountRequest) (*posts.CountResponse, error) {
	count, err := s.service.GetCount(req.UserLogin, ctx)
	if err != nil {
		return nil, err
	}
	return &posts.CountResponse{Count: count}, nil
}

func (s *PostsService) GetPostsByIds(
	ctx context.Context,
	req *posts.GetPostsByIdsMapRequest,
) (*posts.GetPostsByIdsMapResponse, error) {
	postList, err := s.service.GetByIds(req.PostIds, req.AuthUserLogin, req.IsPremium, ctx)
	if err != nil {
		return nil, err
	}

	var resultMap map[string]*posts.PostItem

	resultMap = make(map[string]*posts.PostItem)
	for key, post := range postList.Posts {
		content, _ := utils.GormJSONToProtoStruct(post.Content)
		resultMap[key] = utils.FillPostItem(&post, content, convertReactionsToProto(post.Reactions))
	}
	return &posts.GetPostsByIdsMapResponse{Data: resultMap}, nil
}

func (s *PostsService) GetPostsCountByUserLogin(
	ctx context.Context,
	req *posts.GetPostsCountByUserLoginRequest,
) (*posts.GetPostsCountByUserLoginResponse, error) {
	count, err := s.service.GetPostsCountByUserLogin(req.UserLogin, ctx)
	if err != nil {
		return nil, err
	}
	return &posts.GetPostsCountByUserLoginResponse{Count: count}, nil
}

func (s *PostsService) GetReactionsCountByUserLogin(
	ctx context.Context,
	req *posts.GetReactionsCountByUserLoginRequest,
) (*posts.GetReactionsCountByUserLoginResponse, error) {
	count, err := s.service.GetReactionsCountByUserLogin(req.UserLogin, ctx)
	if err != nil {
		return nil, err
	}
	return &posts.GetReactionsCountByUserLoginResponse{Count: count}, nil
}

func convertPostsToProto(postList []models.PostResponse) []*posts.PostItem {
	var protoPostList []*posts.PostItem

	for _, post := range postList {
		content, _ := utils.GormJSONToProtoStruct(post.Content)
		protoReacts := convertReactionsToProto(post.Reactions)
		protoPostList = append(protoPostList, utils.FillPostItem(&post, content, protoReacts))
	}
	return protoPostList
}

func convertReactionsToProto(reactions []models.PostReactions) []*posts.React {
	var protoReactions []*posts.React
	for _, reaction := range reactions {
		protoReactions = append(
			protoReactions, &posts.React{
				UserLogin: reaction.UserLogin,
				PostId:    reaction.PostId.String(),
				Id:        reaction.Id.String(),
				CreatedAt: reaction.CreatedAt.Format(time.RFC3339),
				Reaction:  string(reaction.Reaction),
			},
		)
	}
	return protoReactions
}
