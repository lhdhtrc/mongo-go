package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

func Delete(ctx context.Context, collection *mongo.Collection, id string) {
	if ObjectID, err := primitive.ObjectIDFromHex(id); err == nil {
		_, _ = collection.DeleteOne(ctx, bson.D{
			{"_id", ObjectID},
		})
	}
}

func DeleteMany(ctx context.Context, collection *mongo.Collection, ids []string) {
	_, _ = collection.DeleteMany(ctx, bson.D{
		{"_id", bson.E{Key: "$in", Value: StrIdToObjectId(ids)}},
	})
}

func SoftDelete(ctx context.Context, collection *mongo.Collection, id string) {
	if ObjectID, err := primitive.ObjectIDFromHex(id); err == nil {
		timer := time.Now().UTC()
		_, _ = collection.UpdateOne(ctx, bson.D{
			{"_id", ObjectID},
		}, bson.D{
			{"$set", bson.D{
				{"updated_at", timer},
				{"deleted_at", timer},
			}},
		})
	}
}

func SoftDeleteMany(ctx context.Context, collection *mongo.Collection, ids []string) {
	timer := time.Now().UTC()
	_, _ = collection.UpdateMany(ctx, bson.D{
		{"_id", bson.E{Key: "$in", Value: StrIdToObjectId(ids)}},
	}, bson.D{
		{"$set", bson.D{
			{"updated_at", timer},
			{"deleted_at", timer},
		}},
	})
}
