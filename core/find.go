package core

import "go.mongodb.org/mongo-driver/mongo/options"

func Paging(page uint, pageSize uint, option *options.FindOptions) {
	if page != 0 {
		if pageSize > 50 {
			pageSize = 50
		}
		option.SetLimit(int64(pageSize))
		option.SetSkip(int64((page - 1) * pageSize))
	}
}
