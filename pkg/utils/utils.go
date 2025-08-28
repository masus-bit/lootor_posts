package utils

import (
	"encoding/json"
	"fmt"
	"google.golang.org/protobuf/types/known/structpb"
	"gorm.io/datatypes"
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
