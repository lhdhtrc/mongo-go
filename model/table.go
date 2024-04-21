package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type MongoTableModeEntity struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
	DeletedAt time.Time          `json:"deleted_at" bson:"deleted_at"`
}

func (mm *MongoTableModeEntity) BeforeInset() {
	mm.ID = primitive.NewObjectID()
	mm.CreatedAt = time.Now().Local()
	mm.UpdatedAt = time.Now().Local()
}

func (mm *MongoTableModeEntity) BeforeUpdate() {
	mm.UpdatedAt = time.Now().Local()
}

func (mm *MongoTableModeEntity) BeforeDelete() {
	mm.UpdatedAt = time.Now().Local()
	mm.DeletedAt = time.Now().Local()
}
