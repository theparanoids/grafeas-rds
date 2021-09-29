// Copyright Yahoo 2021
// Licensed under the terms of the Apache License 2.0.
// See LICENSE file in project root for terms.
package rds

import (
	"database/sql/driver"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/rds/rdsutils"
	"golang.org/x/net/context"

	"github.com/theparanoids/grafeas-rds/rds/config"
)

const (
	// A temporary DB password requested via IAM auth is only valid for 15 minutes.
	// Ref: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/UsingWithRDS.IAMDBAuth.Connecting.html
	refreshAuthTokenInterval = 10 * time.Minute

	errMsgCreateCredentials = "failed to create AWS credentials"
	errMsgRefreshAuthToken  = "failed to refresh auth token"
	errMsgSetupIAMAuth      = "failed to set up IAM auth"

	logsOptInIAMAuth = "Opt in IAM Authentication..."
)

// connector implements driver.Connector
// Reference implementation: sql.dsnConnector.
type connector struct {
	host        string
	port        int
	dbName      string
	user        string
	password    string
	sslMode     string
	sslRootCert string

	driver driver.Driver
	// dsn refers to data source name.
	// Only this variable should be accessed concurrently.
	dsn string
	// reader: when a new connection to DB is needed.
	// writer: when the AWS auth token is refreshed.
	dsnLock sync.RWMutex
}

func newConnector(ctx context.Context, conf *config.Config, driver driver.Driver, cc CredentialsCreator, logger *log.Logger) (*connector, error) {
	c := &connector{
		host:        conf.Host,
		port:        conf.Port,
		dbName:      conf.DBName,
		user:        conf.User,
		password:    conf.Password,
		sslMode:     conf.SSLMode,
		sslRootCert: conf.SSLRootCert,
		driver:      driver,
	}
	if cc == nil {
		c.updateDSN()
	} else {
		logger.Printf("%s", logsOptInIAMAuth)
		if err := c.setupIAMAuth(ctx, conf.IAMAuth, cc, logger); err != nil {
			return nil, fmt.Errorf("%s, err: %v", errMsgSetupIAMAuth, err)
		}
	}
	return c, nil
}

func (c *connector) Connect(context.Context) (driver.Conn, error) {
	dsn := c.readDSN()
	return c.driver.Open(dsn)
}

func (c *connector) Driver() driver.Driver {
	return c.driver
}

func (c *connector) setupIAMAuth(ctx context.Context, conf config.IAMAuthConfig, cc CredentialsCreator, logger *log.Logger) error {
	var err error
	creds, err := cc.Create(conf)
	if err != nil {
		return fmt.Errorf("%s, err: %v", errMsgCreateCredentials, err)
	}
	if err := c.refreshAuthToken(creds, conf.Region); err != nil {
		return fmt.Errorf("%s, err: %v", errMsgRefreshAuthToken, err)
	}
	go c.refreshAuthTokenPeriodically(ctx, creds, conf.Region, refreshAuthTokenInterval, logger)
	return nil
}

func (c *connector) refreshAuthToken(creds *credentials.Credentials, region string) error {
	endpoint := fmt.Sprintf("%s:%d", c.host, c.port)
	authToken, err := rdsutils.BuildAuthToken(endpoint, region, c.user, creds)
	if err != nil {
		return err
	}
	c.updatePassword(authToken)
	return nil
}

func (c *connector) refreshAuthTokenPeriodically(ctx context.Context, creds *credentials.Credentials, region string, interval time.Duration, logger *log.Logger) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			logger.Println("try to refresh auth token")
			if err := c.refreshAuthToken(creds, region); err != nil {
				logger.Printf("%s, err: %v", errMsgRefreshAuthToken, err)
			}
		}
	}
}

// updatePassword should only be invoked by updateAuthToken.
func (c *connector) updatePassword(password string) {
	c.password = password
	c.updateDSN()
}

func (c *connector) readDSN() string {
	c.dsnLock.RLock()
	defer c.dsnLock.RUnlock()
	return c.dsn
}

func (c *connector) updateDSN() {
	dsn := c.assembleDSN()
	c.dsnLock.Lock()
	defer c.dsnLock.Unlock()
	c.dsn = dsn
}

func (c *connector) assembleDSN() string {
	dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		c.host, c.port, c.dbName, c.user, c.password, c.sslMode,
	)
	if c.sslRootCert != "" {
		dsn = fmt.Sprintf("%s sslrootcert=%s", dsn, c.sslRootCert)
	}
	return dsn
}
