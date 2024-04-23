package core

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

func Delete(ctx context.Context, collection *mongo.Collection, id string) {
	_, _ = collection.DeleteOne(ctx, bson.E{
		Key:   "_id",
		Value: id,
	})
}

func DeleteMany(ctx context.Context, collection *mongo.Collection, ids []string) {
	_, _ = collection.DeleteMany(ctx, bson.E{Key: "_id", Value: bson.E{
		Key:   "$in",
		Value: ids,
	}})
}

func SoftDelete(ctx context.Context, collection *mongo.Collection, id string) {
	timer := time.Now().Local()
	_, _ = collection.UpdateOne(ctx, bson.E{
		Key:   "_id",
		Value: id,
	}, bson.D{
		{"updated_at", timer},
		{"deleted_at", timer},
	})
}

func SoftDeleteMany(ctx context.Context, collection *mongo.Collection, ids []string) {
	timer := time.Now().Local()
	_, _ = collection.UpdateMany(ctx, bson.E{
		Key:   "_id",
		Value: bson.E{Key: "$in", Value: ids},
	}, bson.D{
		{"updated_at", timer},
		{"deleted_at", timer},
	})
}
