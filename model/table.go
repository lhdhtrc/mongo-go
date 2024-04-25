package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type MongoTableEntity struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
	DeletedAt *time.Time         `json:"deleted_at" bson:"deleted_at,omitempty"`
}

func (mm *MongoTableEntity) BeforeInset() {
	mm.ID = primitive.NewObjectID()
	timer := time.Now().UTC()
	mm.CreatedAt = timer
	mm.UpdatedAt = timer
	mm.DeletedAt = nil
}

func (mm *MongoTableEntity) BeforeUpdate() {
	mm.UpdatedAt = time.Now().UTC()
}

func (mm *MongoTableEntity) BeforeDelete() {
	timer := time.Now().UTC()
	mm.UpdatedAt = timer
	mm.DeletedAt = &timer
}
