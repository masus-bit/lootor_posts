package utils

import (
	"encoding/json"
	"fmt"
	"google.golang.org/protobuf/types/known/structpb"
	"gorm.io/datatypes"
	"lootor_posts/gen/go/posts"
	"lootor_posts/internal/core/models"
	"time"
)

func NormalizeContent(content *structpb.Struct) []byte {
	contentMap := content.AsMap()
	contentJSON, err := json.Marshal(contentMap)

	if err != nil {
		return nil
	}
	return contentJSON
}

func GormJSONToProtoStruct(jsonData datatypes.JSON) (*structpb.Struct, error) {
	if len(jsonData) == 0 {
		return &structpb.Struct{Fields: make(map[string]*structpb.Value)}, nil
	}

	var contentMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &contentMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %w", err)
	}

	return structpb.NewStruct(contentMap)
}

func FillPostItem(post *models.PostResponse, content *structpb.Struct, protoReacts []*posts.React) *posts.PostItem {
	return &posts.PostItem{
		Id:             post.Id,
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
		IsDraft:        post.IsDraft,
		Views:          int64(post.Views),
		CommentsCount:  int64(post.CommentsCount),
		Title:          post.Title,
		Translit:       post.Translit,
	}
}
