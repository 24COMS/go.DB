package baseMySQL

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Config is used to pass connection settings to NewDatabase
type Config struct {
	Secret, Username, Host, Database string
	NoTLS                            bool
}

// New will return new mysql connection and will start ping goroutine to monitor connection state
// On ctx.Done() db object will be automatically closed
func New(ctx context.Context, wg *sync.WaitGroup, logger logrus.FieldLogger, dbCfg *Config) (*sqlx.DB, error) {
	tlsCfgName := tlsConfigNameLocal
	if !dbCfg.NoTLS {
		tlsCfgName = tlsConfigName

		rootCertPool := x509.NewCertPool()
		if ok := rootCertPool.AppendCertsFromPEM([]byte(msRootCert)); !ok {
			return nil, errors.New("failed to create new certificate pool")
		}

		err := mysql.RegisterTLSConfig(tlsCfgName, &tls.Config{RootCAs: rootCertPool})
		if err != nil {
			return nil, errors.Wrap(err, "failed to register TLS config")
		}
	}

	db, err := sqlx.ConnectContext(ctx, "mysql", fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?parseTime=true&tls=%s&allowNativePasswords=true",
		dbCfg.Username, dbCfg.Secret, dbCfg.Host, dbCfg.Database, tlsCfgName,
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

const (
	tlsConfigName      = "custom"
	tlsConfigNameLocal = "false"
	msRootCert         = `-----BEGIN CERTIFICATE-----
MIIDdzCCAl+gAwIBAgIEAgAAuTANBgkqhkiG9w0BAQUFADBaMQswCQYDVQQGEwJJ
RTESMBAGA1UEChMJQmFsdGltb3JlMRMwEQYDVQQLEwpDeWJlclRydXN0MSIwIAYD
VQQDExlCYWx0aW1vcmUgQ3liZXJUcnVzdCBSb290MB4XDTAwMDUxMjE4NDYwMFoX
DTI1MDUxMjIzNTkwMFowWjELMAkGA1UEBhMCSUUxEjAQBgNVBAoTCUJhbHRpbW9y
ZTETMBEGA1UECxMKQ3liZXJUcnVzdDEiMCAGA1UEAxMZQmFsdGltb3JlIEN5YmVy
VHJ1c3QgUm9vdDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKMEuyKr
mD1X6CZymrV51Cni4eiVgLGw41uOKymaZN+hXe2wCQVt2yguzmKiYv60iNoS6zjr
IZ3AQSsBUnuId9Mcj8e6uYi1agnnc+gRQKfRzMpijS3ljwumUNKoUMMo6vWrJYeK
mpYcqWe4PwzV9/lSEy/CG9VwcPCPwBLKBsua4dnKM3p31vjsufFoREJIE9LAwqSu
XmD+tqYF/LTdB1kC1FkYmGP1pWPgkAx9XbIGevOF6uvUA65ehD5f/xXtabz5OTZy
dc93Uk3zyZAsuT3lySNTPx8kmCFcB5kpvcY67Oduhjprl3RjM71oGDHweI12v/ye
jl0qhqdNkNwnGjkCAwEAAaNFMEMwHQYDVR0OBBYEFOWdWTCCR1jMrPoIVDaGezq1
BE3wMBIGA1UdEwEB/wQIMAYBAf8CAQMwDgYDVR0PAQH/BAQDAgEGMA0GCSqGSIb3
DQEBBQUAA4IBAQCFDF2O5G9RaEIFoN27TyclhAO992T9Ldcw46QQF+vaKSm2eT92
9hkTI7gQCvlYpNRhcL0EYWoSihfVCr3FvDB81ukMJY2GQE/szKN+OMY3EU/t3Wgx
jkzSswF07r51XgdIGn9w/xZchMB5hbgF/X++ZRGjD8ACtPhSNzkE1akxehi/oCr0
Epn3o0WC4zxe9Z2etciefC7IpJ5OCBRLbf1wbWsaY71k5h+3zvDyny67G7fyUIhz
ksLi4xaNmjICq44Y3ekQEe5+NauQrz4wlHrQMz2nZQ/1/I6eYs9HRCwBXbsdtTLS
R9I4LtD+gdwyah617jzV/OeBHRnDJELqYzmp
-----END CERTIFICATE-----`
)
