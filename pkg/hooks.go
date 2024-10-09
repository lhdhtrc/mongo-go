package mongo

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

func (table *TableEntity) BeforeInset() {
	table.ID = primitive.NewObjectID()
	timer := time.Now().UTC().UnixMilli()
	table.CreatedAt = timer
	table.UpdatedAt = timer
	table.DeletedAt = nil
}

func (table *TableEntity) BeforeUpdate() {
	table.UpdatedAt = time.Now().UTC().UnixMilli()
}
