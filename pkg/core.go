package mongo

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
	"os"
	"time"
)

func Install(remote bool, logger *zap.Logger, config *ConfigEntity) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var clientOptions options.ClientOptions

	if config.Account != "" && config.Password != "" {
		clientOptions.SetAuth(options.Credential{
			Username: config.Account,
			Password: config.Password,
		})
	}
	if config.Tls.CaCert != "" && config.Tls.ClientCert != "" && config.Tls.ClientCertKey != "" {
		certPool := x509.NewCertPool()
		CAFile, CAErr := os.ReadFile(config.Tls.CaCert)
		if CAErr != nil {
			return nil, CAErr
		}
		certPool.AppendCertsFromPEM(CAFile)

		clientCert, clientCertErr := tls.LoadX509KeyPair(config.Tls.ClientCert, config.Tls.ClientCertKey)
		if clientCertErr != nil {
			return nil, clientCertErr
		}

		tlsConfig := tls.Config{
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      certPool,
		}
		clientOptions.SetTLSConfig(&tlsConfig)
	}

	uri := fmt.Sprintf("mongodb://%s", config.Address)
	if config.Mode { // cluster
	} else {
	} // stand alone
	clientOptions.ApplyURI(uri)

	clientOptions.SetBSONOptions(&options.BSONOptions{
		UseLocalTimeZone: true,
	})

	clientOptions.SetMaxConnecting(uint64(config.MaxOpenConnects))
	clientOptions.SetMaxPoolSize(uint64(config.MaxIdleConnects))
	clientOptions.SetMaxConnIdleTime(time.Second * time.Duration(config.ConnMaxLifeTime))

	if config.LoggerEnable {
		var fields []zap.Field
		var statement string
		clientOptions.Monitor = &event.CommandMonitor{
			Started: func(ctx context.Context, event *event.CommandStartedEvent) {
				statement = event.Command.String()
				if remote {
					fields = append(fields, zap.String("Type", "Mongo"), zap.String("Database", event.DatabaseName), zap.String("Statement", statement))
				}
			},
			Succeeded: func(ctx context.Context, event *event.CommandSucceededEvent) {
				if remote {
					fields = append(fields, zap.String("Result", "success"), zap.String("Timer", event.Duration.String()))
				}
				logger.Info(fmt.Sprintf("[Mongo:%s][RequestID:%d][Timer:%s]\n%s", event.DatabaseName, event.RequestID, event.Duration.String(), statement), fields...)
				fields = []zap.Field{}
			},
			Failed: func(ctx context.Context, event *event.CommandFailedEvent) {
				if remote {
					fields = append(fields, zap.String("Result", event.Failure), zap.String("Timer", event.Duration.String()))
				}
				logger.Error(fmt.Sprintf("[Mongo:%s][RequestID:%d][Timer:%s]\n%s", event.DatabaseName, event.RequestID, event.Duration.String(), event.Failure), fields...)
				fields = []zap.Field{}
			},
		}
	}

	client, cErr := mongo.Connect(ctx, &clientOptions)
	if cErr != nil {
		return nil, cErr
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}

	db := client.Database(config.Database)
	return db, nil
}
