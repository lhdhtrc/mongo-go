package core

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func StrIdToObjectId(ids []string) []primitive.ObjectID {
	var result []primitive.ObjectID
	for _, idStr := range ids {
		if id, err := primitive.ObjectIDFromHex(idStr); err == nil {
			result = append(result, id)
		}
	}
	return result
}
