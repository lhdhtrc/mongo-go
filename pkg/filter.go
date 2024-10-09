package mongo

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func WithPagingFilter(page, size uint64, option *options.FindOptions) {
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = 5
	}
	if size > 100 {
		size = 100
	}
	option.SetLimit(int64(size))
	option.SetSkip(int64((page - 1) * size))
}

func WithTimerFilter(start, end string, option bson.D) {
	if len(start) != 0 && len(end) != 0 {
		option = append(option, bson.E{
			Key: "created_at",
			Value: bson.D{
				{"$gte", start},
				{"$lte", end},
			},
		})
	}
}
