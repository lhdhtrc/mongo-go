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
	"sync"
	"time"
)

func New(conf *Conf) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var clientOptions options.ClientOptions

	if conf.Username != "" && conf.Password != "" {
		clientOptions.SetAuth(options.Credential{
			Username: conf.Username,
			Password: conf.Password,
		})
	}
	if conf.Tls.CaCert != "" && conf.Tls.ClientCert != "" && conf.Tls.ClientCertKey != "" {
		certPool := x509.NewCertPool()
		CAFile, CAErr := os.ReadFile(conf.Tls.CaCert)
		if CAErr != nil {
			return nil, CAErr
		}
		certPool.AppendCertsFromPEM(CAFile)

		clientCert, clientCertErr := tls.LoadX509KeyPair(conf.Tls.ClientCert, conf.Tls.ClientCertKey)
		if clientCertErr != nil {
			return nil, clientCertErr
		}

		tlsConfig := tls.Config{
			Certificates: []tls.Certificate{clientCert},
			RootCAs:      certPool,
		}
		clientOptions.SetTLSConfig(&tlsConfig)
	}

	uri := fmt.Sprintf("mongodb://%s", conf.Address)
	clientOptions.ApplyURI(uri)

	clientOptions.SetBSONOptions(&options.BSONOptions{
		UseLocalTimeZone: false,
	})

	clientOptions.SetMaxConnecting(uint64(conf.MaxOpenConnects))
	clientOptions.SetMaxPoolSize(uint64(conf.MaxIdleConnects))
	clientOptions.SetMaxConnIdleTime(time.Second * time.Duration(conf.ConnMaxLifeTime))

	if conf.Logger {
		loger := internal.New(internal.Conf{
			SlowThreshold: 200 * time.Millisecond,
			LogLevel:      internal.Info,
			Colorful:      true,
			Database:      conf.Database,
			Console:       conf.loggerConsole,
		}, conf.loggerHandle)

		var stmts sync.Map
		clientOptions.Monitor = &event.CommandMonitor{
			Started: func(ctx context.Context, e *event.CommandStartedEvent) {
				stmts.Store(e.RequestID, e.Command.String())
			},
			Succeeded: func(ctx context.Context, e *event.CommandSucceededEvent) {
				var smt string
				if v, ok := stmts.Load(e.RequestID); ok {
					smt, _ = v.(string)
					stmts.Delete(e.RequestID)
				}
				loger.Trace(ctx, e.RequestID, e.Duration, smt, "")
			},
			Failed: func(ctx context.Context, e *event.CommandFailedEvent) {
				var smt string
				if v, ok := stmts.Load(e.RequestID); ok {
					smt, _ = v.(string)
					stmts.Delete(e.RequestID)
				}
				loger.Trace(ctx, e.RequestID, e.Duration, smt, e.Failure)
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

	db := client.Database(conf.Database)

	return db, nil
}

func (config *Conf) WithLoggerConsole(state bool) {
	config.loggerConsole = state
}

func (config *Conf) WithLoggerHandle(handle func(b []byte)) {
	config.loggerHandle = handle
}
