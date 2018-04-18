package baseMySQL

import (
	"sync"
	"github.com/sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"fmt"
	"time"
	"github.com/pkg/errors"
	"context"
)

// Config is used to pass connection settings to NewDatabase
type Config struct {
	Secret, Username, Host, Database string
	TLS                              bool
}

// New will return new mysql connection and will start ping goroutine to monitor connection state
// On ctx.Done() db object will be automatically closed
func New(ctx context.Context, wg *sync.WaitGroup, logger logrus.FieldLogger, dbCfg *Config) (*sqlx.DB, error) {
	db, err := sqlx.ConnectContext(ctx, "mysql", fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?parseTime=true&tls=%t&allowNativePasswords=true",
		dbCfg.Username, dbCfg.Secret, dbCfg.Host, dbCfg.Database, dbCfg.TLS,
	))

	if err != nil {
		return nil, errors.Wrap(err, "error connecting to database")
	}

	db.SetConnMaxLifetime(3 * time.Minute)

	pingAndClose(ctx, wg, logger, db)

	return db, nil
}

func pingAndClose(ctx context.Context, wg *sync.WaitGroup, logger logrus.FieldLogger, db *sqlx.DB) {
	wg.Add(1)
	go func() {
		ticker := time.NewTicker(time.Second * 3)
		defer func() {
			ticker.Stop()
			wg.Done()
		}()

		for {
			select {
			case <-ctx.Done():
				err := db.Close()
				if err != nil {
					logger.Warn(errors.Wrap(err, "failed to close DB proreply"))
				}
				return
			case <-ticker.C:
				if err := db.PingContext(ctx); err != nil {
					logger.Warn(errors.Wrap(err, "failed to ping mysql"))
				}
			}
		}
	}()
}
