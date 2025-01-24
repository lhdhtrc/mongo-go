package mongo

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

func (table *Table) BeforeInset() {
	table.ID = primitive.NewObjectID()
	timer := time.Now().UTC()
	table.CreatedAt = timer
	table.UpdatedAt = timer
	table.DeletedAt = nil
}

func (table *Table) BeforeUpdate() {
	table.UpdatedAt = time.Now().UTC()
}
