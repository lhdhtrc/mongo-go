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
	"io"
	"log"
	"os"
	"strings"
	"time"
)

func Install(config *ConfigEntity) (*mongo.Database, error) {
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
		var writer io.Writer
		if config.loggerConsole {
			writer = os.Stdout
		} else {
			writer = &internal.CustomWriter{}
		}

		loger := internal.New(log.New(writer, "\r\n", log.LstdFlags), internal.Config{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      internal.Info,
			Colorful:      true,
		}, config.loggerHandle)

		var statement string
		clientOptions.Monitor = &event.CommandMonitor{
			Started: func(ctx context.Context, event *event.CommandStartedEvent) {
				var smt strings.Builder
				smt.WriteString(fmt.Sprintf("[Database:%s]", event.DatabaseName))
				smt.WriteString(fmt.Sprintf("[RequestId:%d]\n", event.RequestID))
				smt.WriteString(event.Command.String())
				statement = smt.String()
			},
			Succeeded: func(ctx context.Context, event *event.CommandSucceededEvent) {
				loger.Trace(ctx, event.Duration, statement, "")
			},
			Failed: func(ctx context.Context, event *event.CommandFailedEvent) {
				loger.Trace(ctx, event.Duration, statement, event.Failure)
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

func (config *ConfigEntity) WithLoggerConsole(state bool) {
	config.loggerConsole = state
}

func (config *ConfigEntity) WithLoggerHandle(handle func(b []byte)) {
	config.loggerHandle = handle
}
