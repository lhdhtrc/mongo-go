package core

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Paging(page uint, pageSize uint, option *options.FindOptions) {
	if page != 0 {
		if pageSize > 50 {
			pageSize = 50
		}
		option.SetLimit(int64(pageSize))
		option.SetSkip(int64((page - 1) * pageSize))
	}
}

func Timer(start string, end string, d bson.D) {
	if len(start) != 0 && len(end) != 0 {
		d = append(d, bson.E{
			Key: "created_at",
			Value: bson.D{
				{"$gte", start},
				{"$lte", end},
			},
		})
	}
}
