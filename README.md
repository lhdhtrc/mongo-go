## Mongo Go
Provides easy to use API to operate Mongo db.

### How to use it?
`go get github.com/lhdhtrc/mongo-go`

```go
package main

import (
	mongo "github.com/lhdhtrc/mongo-go/pkg"
	"go.uber.org/zap"
)

// How to define a table
type TestEntity struct {
    mongo.Table
}

func main() {
	logger, _ := zap.NewProduction()
	instance := mongo.New(logger, &mongo.Config{})
}
```

### Finally
- If you feel good, click on star.
- If you have a good suggestion, please ask the issue.