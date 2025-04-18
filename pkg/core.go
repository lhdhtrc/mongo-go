package mongo

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/lhdhtrc/mongo-go/pkg/internal"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"os"
	"time"
)

func New(config *Config) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var clientOptions options.ClientOptions

	if config.Username != "" && config.Password != "" {
		clientOptions.SetAuth(options.Credential{
			Username: config.Username,
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
	clientOptions.ApplyURI(uri)

	clientOptions.SetBSONOptions(&options.BSONOptions{
		UseLocalTimeZone: true,
	})

	clientOptions.SetMaxConnecting(uint64(config.MaxOpenConnects))
	clientOptions.SetMaxPoolSize(uint64(config.MaxIdleConnects))
	clientOptions.SetMaxConnIdleTime(time.Second * time.Duration(config.ConnMaxLifeTime))

	if config.Logger {
		loger := internal.New(internal.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      internal.Info,
			Colorful:      true,
			Database:      config.Database,
			Console:       config.loggerConsole,
		}, config.loggerHandle)

		var statement string
		clientOptions.Monitor = &event.CommandMonitor{
			Started: func(ctx context.Context, event *event.CommandStartedEvent) {
				statement = event.Command.String()
			},
			Succeeded: func(ctx context.Context, event *event.CommandSucceededEvent) {
				loger.Trace(ctx, event.RequestID, event.Duration, statement, "")
			},
			Failed: func(ctx context.Context, event *event.CommandFailedEvent) {
				loger.Trace(ctx, event.RequestID, event.Duration, statement, event.Failure)
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

func (config *Config) WithLoggerConsole(state bool) {
	config.loggerConsole = state
}

func (config *Config) WithLoggerHandle(handle func(b []byte)) {
	config.loggerHandle = handle
}
